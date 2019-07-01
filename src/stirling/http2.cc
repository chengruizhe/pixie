#include "src/stirling/http2.h"

#include <arpa/inet.h>
extern "C" {
#include <nghttp2/nghttp2_frame.h>
#include <nghttp2/nghttp2_hd.h>
#include <nghttp2/nghttp2_helper.h>
}

#include <algorithm>
#include <map>
#include <utility>
#include <vector>

#include "absl/strings/str_cat.h"
#include "src/common/base/error.h"
#include "src/common/base/status.h"

namespace pl {
namespace stirling {
namespace http2 {

namespace {

bool IsPadded(const nghttp2_frame_hd& hd) { return hd.flags & NGHTTP2_FLAG_PADDED; }
bool IsEndHeaders(const nghttp2_frame_hd& hd) { return hd.flags & NGHTTP2_FLAG_END_HEADERS; }
bool IsEndStream(const nghttp2_frame_hd& hd) { return hd.flags & NGHTTP2_FLAG_END_STREAM; }

// HTTP2 frame header size, see https://http2.github.io/http2-spec/#FrameHeader.
constexpr size_t kFrameHeaderSizeInBytes = 9;

void UnpackData(const uint8_t* buf, Frame* f) {
  nghttp2_data& frame = f->frame.data;
  size_t frame_body_length = frame.hd.length;

  DCHECK(frame.hd.type == NGHTTP2_DATA) << "Must be a DATA frame, got: " << frame.hd.type;

  if (IsPadded(frame.hd)) {
    frame.padlen = *buf;
    ++buf;
    --frame_body_length;
  }

  if (frame_body_length >= frame.padlen) {
    f->u8payload.assign(buf, frame_body_length - frame.padlen);
  } else {
    LOG(DFATAL) << "Pad length cannot be larger than frame body";
    f->u8payload.clear();
  }
}

class InflaterRAIIWrapper {
 public:
  explicit InflaterRAIIWrapper(nghttp2_hd_inflater* inflater) : inflater_(inflater) {}
  ~InflaterRAIIWrapper() {
    if (inflater_ != nullptr) {
      nghttp2_hd_inflate_free(inflater_);
    }
  }
  Status Init() {
    int rv = nghttp2_hd_inflate_init(inflater_, nghttp2_mem_default());
    if (rv != 0) {
      return error::Internal("Failed to initialize nghttp2_hd_inflater");
    }
    return Status::OK();
  }

 private:
  nghttp2_hd_inflater* inflater_;
};

void UnpackHeaders(const uint8_t* buf, Frame* f) {
  nghttp2_headers& frame = f->frame.headers;
  size_t frame_body_length = frame.hd.length;

  DCHECK(frame.hd.type == NGHTTP2_HEADERS) << "Must be a HEADERS frame, got: " << frame.hd.type;

  if (IsPadded(frame.hd)) {
    frame.padlen = *buf;
    ++buf;
    --frame_body_length;
  }

  nghttp2_frame_unpack_headers_payload(&frame, buf);

  const size_t offset = nghttp2_frame_headers_payload_nv_offset(&frame);
  buf += offset;
  frame_body_length -= offset;

  if (frame_body_length >= frame.padlen) {
    f->u8payload.assign(buf, frame_body_length - frame.padlen);
  } else {
    LOG(DFATAL) << "Pad length cannot be larger than frame body";
    f->u8payload.clear();
  }
}

void UnpackContinuation(const uint8_t* buf, Frame* frame) {
  frame->u8payload.assign(buf, frame->frame.hd.length);
}

/**
 * @brief Inflates a complete header block. If the input header block fragment, then the results are
 * undefined.
 *
 * This code follows: https://github.com/nghttp2/nghttp2/blob/master/examples/deflate.c.
 */
Status InflateHeaderBlock(u8string_view buf, nghttp2_hd_inflater* inflater, NVMap* nv_map) {
  // TODO(yzhao): Experiment continuous parsing of multiple header block fragments from different
  // HTTP2 streams.
  constexpr bool in_final = true;
  for (;;) {
    int inflate_flags = 0;
    nghttp2_nv nv = {};

    const int rv =
        nghttp2_hd_inflate_hd2(inflater, &nv, &inflate_flags, buf.data(), buf.size(), in_final);
    if (rv < 0) {
      return error::Internal("Failed to inflate header");
    }
    buf.remove_prefix(rv);

    if (inflate_flags & NGHTTP2_HD_INFLATE_EMIT) {
      nv_map->insert(
          std::make_pair(std::string(reinterpret_cast<const char*>(nv.name), nv.namelen),
                         std::string(reinterpret_cast<const char*>(nv.value), nv.valuelen)));
    }
    if (inflate_flags & NGHTTP2_HD_INFLATE_FINAL) {
      nghttp2_hd_inflate_end_headers(inflater);
      break;
    }
    if ((inflate_flags & NGHTTP2_HD_INFLATE_EMIT) == 0 && buf.empty()) {
      break;
    }
  }
  return Status::OK();
}

}  // namespace

std::string_view FrameTypeName(uint8_t type) {
  switch (type) {
    case NGHTTP2_DATA:
      return "DATA";
    case NGHTTP2_HEADERS:
      return "HEADERS";
    case NGHTTP2_PRIORITY:
      return "PRIORITY";
    case NGHTTP2_RST_STREAM:
      return "RST_STREAM";
    case NGHTTP2_SETTINGS:
      return "SETTINGS";
    case NGHTTP2_PUSH_PROMISE:
      return "PUSH_PROMISE";
    case NGHTTP2_PING:
      return "PING";
    case NGHTTP2_GOAWAY:
      return "GOAWAY";
    case NGHTTP2_WINDOW_UPDATE:
      return "WINDOW_UPDATE";
    case NGHTTP2_CONTINUATION:
      return "CONTINUATION";
    case NGHTTP2_ALTSVC:
      return "ALTSVC";
    case NGHTTP2_ORIGIN:
      return "ORIGIN";
    default:
      return "unknown";
  }
}

// frame{} zero initialize the member, which is needed to make sure default value is sensible.
Frame::Frame() : frame{} {}
Frame::~Frame() {
  if (frame.hd.type == NGHTTP2_HEADERS) {
    nghttp2_frame_headers_free(&frame.headers, nghttp2_mem_default());
  }
}

Status UnpackFrame(std::string_view* buf, Frame* frame) {
  if (buf->size() < kFrameHeaderSizeInBytes) {
    return error::ResourceUnavailable(absl::Substitute(
        "Not enough data to parse, got: $0 must be >= $1", buf->size(), kFrameHeaderSizeInBytes));
  }

  const uint8_t* u8_buf = reinterpret_cast<const uint8_t*>(buf->data());
  nghttp2_frame_unpack_frame_hd(&frame->frame.hd, u8_buf);

  if (buf->size() < kFrameHeaderSizeInBytes + frame->frame.hd.length) {
    return error::ResourceUnavailable(
        absl::Substitute("Not enough u8_buf to parse, got: $0 must be >= $1", buf->size(),
                         kFrameHeaderSizeInBytes + frame->frame.hd.length));
  }

  buf->remove_prefix(kFrameHeaderSizeInBytes + frame->frame.hd.length);
  u8_buf += kFrameHeaderSizeInBytes;

  const uint8_t type = frame->frame.hd.type;
  switch (type) {
    case NGHTTP2_DATA:
      UnpackData(u8_buf, frame);
      break;
    case NGHTTP2_HEADERS:
      UnpackHeaders(u8_buf, frame);
      break;
    case NGHTTP2_CONTINUATION:
      UnpackContinuation(u8_buf, frame);
      break;
    case NGHTTP2_PRIORITY:
    case NGHTTP2_RST_STREAM:
    case NGHTTP2_SETTINGS:
    case NGHTTP2_PUSH_PROMISE:
    case NGHTTP2_PING:
    case NGHTTP2_GOAWAY:
    case NGHTTP2_WINDOW_UPDATE:
    case NGHTTP2_ALTSVC:
    case NGHTTP2_ORIGIN:
      return error::Cancelled(absl::StrCat("Ignored frame type: ", FrameTypeName(type)));
    default:
      return error::Cancelled(absl::StrCat("Unknown frame type: ", type));
  }
  return Status::OK();
}

Status UnpackFrames(std::string_view* buf, std::vector<std::unique_ptr<Frame>>* frames) {
  while (!buf->empty()) {
    auto frame = std::make_unique<Frame>();
    Status s = UnpackFrame(buf, frame.get());
    if (error::IsCancelled(s)) {
      continue;
    }
    PL_RETURN_IF_ERROR(s);
    frames->push_back(std::move(frame));
  }
  return Status::OK();
}

namespace {

Status CheckGRPCMessage(u8string_view u8buf) {
  if (u8buf.size() < kGRPCMessageHeaderSizeInBytes) {
    return error::InvalidArgument(absl::StrCat("Needs at least 5 bytes, got: ", u8buf.size()));
  }
  if (u8buf[0] != 0) {
    return error::Unimplemented("GRPC message decompression is unimplemented");
  }
  const uint32_t len = nghttp2_get_uint32(u8buf.data() + 1);
  if (u8buf.size() != kGRPCMessageHeaderSizeInBytes + len) {
    return error::InvalidArgument(absl::Substitute(
        "Expect bytes: $0, got: $1", kGRPCMessageHeaderSizeInBytes + len, u8buf.size()));
  }
  return Status::OK();
}

/**
 * @brief Given a list of frames for one stream, stitches them together into gRPC request and
 * response messages. Also mark consumed frames, so the caller can destroy them afterwards.
 */
Status StitchFrames(const std::vector<Frame*>& frames, nghttp2_hd_inflater* inflater,
                    std::vector<GRPCMessage>* msgs) {
  size_t header_block_size = 0;
  std::vector<Frame*> header_block_frames;
  GRPCMessage msg;
  bool header_block_started = false;

  auto handle_headers_or_continuation = [inflater, &header_block_size, &header_block_frames, &msg,
                                         &header_block_started](Frame* f) {
    header_block_size += f->u8payload.size();
    header_block_frames.push_back(f);
    if (IsEndHeaders(f->frame.hd)) {
      header_block_started = false;
      u8string u8buf;
      u8buf.reserve(header_block_size);
      for (auto* f : header_block_frames) {
        f->consumed = true;
        u8buf.append(f->u8payload);
      }
      header_block_frames.clear();
      header_block_size = 0;
      msg.parse_succeeded = InflateHeaderBlock(u8buf, inflater, &msg.headers).ok();
    }
  };

  for (Frame* f : frames) {
    const uint8_t type = f->frame.hd.type;
    switch (type) {
      case NGHTTP2_DATA:
        DCHECK(!header_block_started) << "DATA frame must not follow a unended HEADERS frame.";
        f->consumed = true;
        msg.message.append(f->u8payload.data(), f->u8payload.size());
        // gRPC request EOS (end-of-stream) is indicated by END_STREAM flag on the last DATA frame.
        // This also indicates this message is a request. Keep appending data, and then export when
        // END_STREAM is seen.
        if (IsEndStream(f->frame.hd)) {
          msg.type = MessageType::kRequests;
          msg.parse_succeeded = CheckGRPCMessage(msg.message).ok();
          msgs->emplace_back(std::move(msg));
        }
        break;
      case NGHTTP2_HEADERS:
        DCHECK(!header_block_started) << "HEADERS frame must not follow another unended HEADERS "
                                         "frame.";
        header_block_started = true;
        handle_headers_or_continuation(f);
        // gRPC response EOS (end-of-stream) is indicated by END_STREAM flag on the last HEADERS
        // frame. This also indicates this message is a response.
        // No CONTINUATION frame will be used.
        if (IsEndStream(f->frame.hd)) {
          msg.type = MessageType::kResponses;
          msg.parse_succeeded = CheckGRPCMessage(msg.message).ok();
          msgs->emplace_back(std::move(msg));
        }
        break;
      case NGHTTP2_CONTINUATION:
        DCHECK(header_block_started)
            << "CONTINUATION frame must follow a HEADERS or CONTINUATION frame.";
        handle_headers_or_continuation(f);
        // No need to handle END_STREAM as CONTINUATION frame does not define END_STREAM flag.
        break;
      default:
        constexpr char kErr[] =
            "This function does not accept any frame types other than "
            "DATA, HEADERS, and CONTINUATION";
        DCHECK(false) << kErr;
        return error::InvalidArgument(absl::StrCat(kErr, ". Got: ", FrameTypeName(type)));
    }
  }
  return Status::OK();
}

}  // namespace

Status StitchGRPCStreamFrames(std::vector<std::unique_ptr<Frame>>* frames,
                              std::map<uint32_t, std::vector<GRPCMessage>>* stream_msgs) {
  // TODO(yzhao): In next diff, move inflater into part of ConnectionTracker or DataStream.
  nghttp2_hd_inflater inflater;
  InflaterRAIIWrapper inflater_wrapper{&inflater};
  PL_RETURN_IF_ERROR(inflater_wrapper.Init());

  std::map<uint32_t, std::vector<Frame*>> stream_frames;

  // Collect frames for each stream.
  for (const std::unique_ptr<Frame>& f : *frames) {
    stream_frames[f->frame.hd.stream_id].push_back(f.get());
  }
  for (auto& [stream_id, frame_ptrs] : stream_frames) {
    std::vector<GRPCMessage>* msgs = &(*stream_msgs)[stream_id];
    PL_RETURN_IF_ERROR(StitchFrames(frame_ptrs, &inflater, msgs));
  }
  frames->erase(std::remove_if(frames->begin(), frames->end(),
                               [](const std::unique_ptr<Frame>& f) { return f->consumed; }),
                frames->end());
  return Status::OK();
}

}  // namespace http2
}  // namespace stirling
}  // namespace pl
