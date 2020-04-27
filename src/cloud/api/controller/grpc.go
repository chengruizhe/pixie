package controller

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"

	uuid "github.com/satori/go.uuid"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	artifacttrackerpb "pixielabs.ai/pixielabs/src/cloud/artifact_tracker/artifacttrackerpb"
	"pixielabs.ai/pixielabs/src/cloud/autocomplete"
	"pixielabs.ai/pixielabs/src/cloud/cloudapipb"
	"pixielabs.ai/pixielabs/src/cloud/scriptmgr/scriptmgrpb"
	"pixielabs.ai/pixielabs/src/cloud/vzmgr/vzmgrpb"
	uuidpb "pixielabs.ai/pixielabs/src/common/uuid/proto"
	versionspb "pixielabs.ai/pixielabs/src/shared/artifacts/versionspb"
	"pixielabs.ai/pixielabs/src/shared/cvmsgspb"
	"pixielabs.ai/pixielabs/src/shared/services/authcontext"
	"pixielabs.ai/pixielabs/src/shared/services/utils"
	pbutils "pixielabs.ai/pixielabs/src/utils"
)

func init() {
	pflag.String("vizier_image_secret_path", "/vizier-image-secret", "[WORKAROUND] The path the the image secrets")
	pflag.String("vizier_image_secret_file", "vizier_image_secret.json", "[WORKAROUND] The image secret file")
}

// VizierImageAuthServer is the GRPC server responsible for providing access to Vizier images.
type VizierImageAuthServer struct{}

// GetImageCredentials fetches image credentials for vizier.
func (v VizierImageAuthServer) GetImageCredentials(context.Context, *cloudapipb.GetImageCredentialsRequest) (*cloudapipb.GetImageCredentialsResponse, error) {
	// TODO(zasgar/michelle): Fix this to create creds for user.
	// This is a workaround implementation to just give them access based on static keys.
	p := viper.GetString("vizier_image_secret_path")
	f := viper.GetString("vizier_image_secret_file")

	absP, err := filepath.Abs(p)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to parse creds paths")
	}
	credsFile := filepath.Join(absP, f)
	b, err := ioutil.ReadFile(credsFile)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to read creds file")
	}

	return &cloudapipb.GetImageCredentialsResponse{Creds: string(b)}, nil
}

// ArtifactTrackerServer is the GRPC server responsible for providing access to artifacts.
type ArtifactTrackerServer struct {
	ArtifactTrackerClient artifacttrackerpb.ArtifactTrackerClient
}

func getArtifactTypeFromCloudProto(a cloudapipb.ArtifactType) versionspb.ArtifactType {
	switch a {
	case cloudapipb.AT_LINUX_AMD64:
		return versionspb.AT_LINUX_AMD64
	case cloudapipb.AT_DARWIN_AMD64:
		return versionspb.AT_DARWIN_AMD64
	case cloudapipb.AT_CONTAINER_SET_YAMLS:
		return versionspb.AT_CONTAINER_SET_YAMLS
	case cloudapipb.AT_CONTAINER_SET_LINUX_AMD64:
		return versionspb.AT_CONTAINER_SET_LINUX_AMD64
	default:
		return versionspb.AT_UNKNOWN
	}
}

func getArtifactTypeFromVersionsProto(a versionspb.ArtifactType) cloudapipb.ArtifactType {
	switch a {
	case versionspb.AT_LINUX_AMD64:
		return cloudapipb.AT_LINUX_AMD64
	case versionspb.AT_DARWIN_AMD64:
		return cloudapipb.AT_DARWIN_AMD64
	case versionspb.AT_CONTAINER_SET_YAMLS:
		return cloudapipb.AT_CONTAINER_SET_YAMLS
	case versionspb.AT_CONTAINER_SET_LINUX_AMD64:
		return cloudapipb.AT_CONTAINER_SET_LINUX_AMD64
	default:
		return cloudapipb.AT_UNKNOWN
	}
}

func getServiceCredentials(signingKey string) (string, error) {
	claims := utils.GenerateJWTForService("API Service")
	return utils.SignJWTClaims(claims, signingKey)
}

// GetArtifactList gets the set of artifact versions for the given artifact.
func (a ArtifactTrackerServer) GetArtifactList(ctx context.Context, req *cloudapipb.GetArtifactListRequest) (*cloudapipb.ArtifactSet, error) {
	atReq := &artifacttrackerpb.GetArtifactListRequest{
		ArtifactType: getArtifactTypeFromCloudProto(req.ArtifactType),
		ArtifactName: req.ArtifactName,
		Limit:        req.Limit,
	}

	serviceAuthToken, err := getServiceCredentials(viper.GetString("jwt_signing_key"))
	if err != nil {
		return nil, err
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization",
		fmt.Sprintf("bearer %s", serviceAuthToken))

	resp, err := a.ArtifactTrackerClient.GetArtifactList(ctx, atReq)
	if err != nil {
		return nil, err
	}

	cloudpbArtifacts := make([]*cloudapipb.Artifact, len(resp.Artifact))
	for i, artifact := range resp.Artifact {
		availableArtifacts := make([]cloudapipb.ArtifactType, len(artifact.AvailableArtifacts))
		for j, a := range artifact.AvailableArtifacts {
			availableArtifacts[j] = getArtifactTypeFromVersionsProto(a)
		}
		cloudpbArtifacts[i] = &cloudapipb.Artifact{
			Timestamp:          artifact.Timestamp,
			CommitHash:         artifact.CommitHash,
			VersionStr:         artifact.VersionStr,
			Changelog:          artifact.Changelog,
			AvailableArtifacts: availableArtifacts,
		}
	}

	return &cloudapipb.ArtifactSet{
		Name:     resp.Name,
		Artifact: cloudpbArtifacts,
	}, nil
}

// GetDownloadLink gets the download link for the given artifact.
func (a ArtifactTrackerServer) GetDownloadLink(ctx context.Context, req *cloudapipb.GetDownloadLinkRequest) (*cloudapipb.GetDownloadLinkResponse, error) {
	atReq := &artifacttrackerpb.GetDownloadLinkRequest{
		ArtifactName: req.ArtifactName,
		VersionStr:   req.VersionStr,
		ArtifactType: getArtifactTypeFromCloudProto(req.ArtifactType),
	}

	serviceAuthToken, err := getServiceCredentials(viper.GetString("jwt_signing_key"))
	if err != nil {
		return nil, err
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization",
		fmt.Sprintf("bearer %s", serviceAuthToken))

	resp, err := a.ArtifactTrackerClient.GetDownloadLink(ctx, atReq)
	if err != nil {
		return nil, err
	}

	return &cloudapipb.GetDownloadLinkResponse{
		Url:        resp.Url,
		SHA256:     resp.SHA256,
		ValidUntil: resp.ValidUntil,
	}, nil
}

// VizierClusterInfo is the server that implements the VizierClusterInfo gRPC service.
type VizierClusterInfo struct {
	VzMgr vzmgrpb.VZMgrServiceClient
}

// CreateCluster creates a cluster for the current org.
func (v *VizierClusterInfo) CreateCluster(ctx context.Context, request *cloudapipb.CreateClusterRequest) (*cloudapipb.CreateClusterResponse, error) {
	sCtx, err := authcontext.FromContext(ctx)
	orgIDstr := sCtx.Claims.GetUserClaims().OrgID

	ctx = metadata.AppendToOutgoingContext(ctx, "authorization",
		fmt.Sprintf("bearer %s", sCtx.AuthToken))

	clusterID, err := v.VzMgr.CreateVizierCluster(ctx, &vzmgrpb.CreateVizierClusterRequest{
		OrgID: pbutils.ProtoFromUUIDStrOrNil(orgIDstr),
	})
	if err != nil {
		return nil, err
	}
	return &cloudapipb.CreateClusterResponse{ClusterID: clusterID}, nil
}

// GetClusterInfo returns information about Vizier clusters.
func (v *VizierClusterInfo) GetClusterInfo(ctx context.Context, request *cloudapipb.GetClusterInfoRequest) (*cloudapipb.GetClusterInfoResponse, error) {
	sCtx, err := authcontext.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	orgIDstr := sCtx.Claims.GetUserClaims().OrgID
	orgID, err := uuid.FromString(orgIDstr)
	if err != nil {
		return nil, err
	}

	ctx = metadata.AppendToOutgoingContext(ctx, "authorization",
		fmt.Sprintf("bearer %s", sCtx.AuthToken))

	vzIDs := make([]*uuidpb.UUID, 0)
	if request.ID != nil {
		vzIDs = append(vzIDs, request.ID)
	} else {
		viziers, err := v.VzMgr.GetViziersByOrg(ctx, pbutils.ProtoFromUUID(&orgID))
		if err != nil {
			return nil, err
		}
		vzIDs = viziers.VizierIDs
	}

	return v.getClusterInfoForViziers(ctx, vzIDs)
}

func (v *VizierClusterInfo) getClusterInfoForViziers(ctx context.Context, ids []*uuidpb.UUID) (*cloudapipb.GetClusterInfoResponse, error) {
	resp := &cloudapipb.GetClusterInfoResponse{}
	for _, id := range ids {
		// TODO(zasgar/michelle): Make these requests parallel
		vzInfo, err := v.VzMgr.GetVizierInfo(ctx, id)
		if err != nil {
			return nil, err
		}

		s := vzStatusToClusterStatus(vzInfo.Status)
		resp.Clusters = append(resp.Clusters, &cloudapipb.ClusterInfo{
			ID:              id,
			Status:          s,
			LastHeartbeatNs: vzInfo.LastHeartbeatNs,
			Config: &cloudapipb.VizierConfig{
				PassthroughEnabled: vzInfo.Config.PassthroughEnabled,
			},
		})
	}
	return resp, nil
}

// GetClusterConnectionInfo returns information about connections to Vizier cluster.
func (v *VizierClusterInfo) GetClusterConnectionInfo(ctx context.Context, request *cloudapipb.GetClusterConnectionInfoRequest) (*cloudapipb.GetClusterConnectionInfoResponse, error) {
	id := request.ID

	sCtx, err := authcontext.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	ctx = metadata.AppendToOutgoingContext(ctx, "authorization",
		fmt.Sprintf("bearer %s", sCtx.AuthToken))

	ci, err := v.VzMgr.GetVizierConnectionInfo(ctx, id)
	if err != nil {
		return nil, err
	}

	return &cloudapipb.GetClusterConnectionInfoResponse{
		IPAddress: ci.IPAddress,
		Token:     ci.Token,
	}, nil
}

// UpdateClusterVizierConfig supports updates of VizierConfig for a cluster
func (v *VizierClusterInfo) UpdateClusterVizierConfig(ctx context.Context, req *cloudapipb.UpdateClusterVizierConfigRequest) (*cloudapipb.UpdateClusterVizierConfigResponse, error) {
	sCtx, err := authcontext.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", fmt.Sprintf("bearer %s", sCtx.AuthToken))

	_, err = v.VzMgr.UpdateVizierConfig(ctx, &cvmsgspb.UpdateVizierConfigRequest{
		VizierID: req.ID,
		ConfigUpdate: &cvmsgspb.VizierConfigUpdate{
			PassthroughEnabled: req.ConfigUpdate.PassthroughEnabled,
		},
	})
	if err != nil {
		return nil, err
	}

	return &cloudapipb.UpdateClusterVizierConfigResponse{}, nil
}

// UpdateOrInstallCluster updates or installs the given vizier cluster to the specified version.
func (v *VizierClusterInfo) UpdateOrInstallCluster(ctx context.Context, req *cloudapipb.UpdateOrInstallClusterRequest) (*cloudapipb.UpdateOrInstallClusterResponse, error) {
	sCtx, err := authcontext.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Version == "" {
		return nil, status.Errorf(codes.InvalidArgument, "version cannot be empty")
	}

	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", fmt.Sprintf("bearer %s", sCtx.AuthToken))

	resp, err := v.VzMgr.UpdateOrInstallVizier(ctx, &cvmsgspb.UpdateOrInstallVizierRequest{
		VizierID: req.ClusterID,
		Version:  req.Version,
	})
	if err != nil {
		return nil, err
	}

	return &cloudapipb.UpdateOrInstallClusterResponse{
		UpdateStarted: resp.UpdateStarted,
	}, nil
}

func vzStatusToClusterStatus(s cvmsgspb.VizierInfo_Status) cloudapipb.ClusterStatus {
	switch s {
	case cvmsgspb.VZ_ST_HEALTHY:
		return cloudapipb.CS_HEALTHY
	case cvmsgspb.VZ_ST_UNHEALTHY:
		return cloudapipb.CS_UNHEALTHY
	case cvmsgspb.VZ_ST_DISCONNECTED:
		return cloudapipb.CS_DISCONNECTED
	default:
		return cloudapipb.CS_UNKNOWN
	}
}

// AutocompleteServer is the server that implements the Autocomplete gRPC service.
type AutocompleteServer struct {
	Suggester autocomplete.Suggester
}

// Autocomplete returns a formatted string and autocomplete suggestions.
func (a *AutocompleteServer) Autocomplete(ctx context.Context, req *cloudapipb.AutocompleteRequest) (*cloudapipb.AutocompleteResponse, error) {
	sCtx, err := authcontext.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	orgIDstr := sCtx.Claims.GetUserClaims().OrgID
	orgID, err := uuid.FromString(orgIDstr)
	if err != nil {
		return nil, err
	}

	fmtString, executable, suggestions, err := autocomplete.Autocomplete(req.Input, int(req.CursorPos), req.Action, a.Suggester, orgID)
	if err != nil {
		return nil, err
	}

	return &cloudapipb.AutocompleteResponse{
		FormattedInput: fmtString,
		IsExecutable:   executable,
		TabSuggestions: suggestions,
	}, nil
}

// ScriptMgrServer is the server that implements the ScriptMgr gRPC service.
type ScriptMgrServer struct {
	ScriptMgr scriptmgrpb.ScriptMgrServiceClient
}

// GetLiveViews returns a list of all available live views.
func (s *ScriptMgrServer) GetLiveViews(ctx context.Context, req *cloudapipb.GetLiveViewsReq) (*cloudapipb.GetLiveViewsResp, error) {
	sCtx, err := authcontext.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", fmt.Sprintf("bearer %s", sCtx.AuthToken))

	smReq := &scriptmgrpb.GetLiveViewsReq{}
	smResp, err := s.ScriptMgr.GetLiveViews(ctx, smReq)
	if err != nil {
		return nil, err
	}
	resp := &cloudapipb.GetLiveViewsResp{
		LiveViews: make([]*cloudapipb.LiveViewMetadata, len(smResp.LiveViews)),
	}
	for i, liveView := range smResp.LiveViews {
		resp.LiveViews[i] = &cloudapipb.LiveViewMetadata{
			ID:   pbutils.UUIDFromProtoOrNil(liveView.ID).String(),
			Name: liveView.Name,
			Desc: liveView.Desc,
		}
	}
	return resp, nil
}

// GetLiveViewContents returns the pxl script, vis info, and metdata for a live view.
func (s *ScriptMgrServer) GetLiveViewContents(ctx context.Context, req *cloudapipb.GetLiveViewContentsReq) (*cloudapipb.GetLiveViewContentsResp, error) {
	sCtx, err := authcontext.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", fmt.Sprintf("bearer %s", sCtx.AuthToken))

	smReq := &scriptmgrpb.GetLiveViewContentsReq{
		LiveViewID: pbutils.ProtoFromUUIDStrOrNil(req.LiveViewID),
	}
	smResp, err := s.ScriptMgr.GetLiveViewContents(ctx, smReq)
	if err != nil {
		return nil, err
	}

	return &cloudapipb.GetLiveViewContentsResp{
		Metadata: &cloudapipb.LiveViewMetadata{
			ID:   req.LiveViewID,
			Name: smResp.Metadata.Name,
			Desc: smResp.Metadata.Desc,
		},
		PxlContents: smResp.PxlContents,
		Vis:         smResp.Vis,
	}, nil
}

// GetScripts returns a list of all available scripts.
func (s *ScriptMgrServer) GetScripts(ctx context.Context, req *cloudapipb.GetScriptsReq) (*cloudapipb.GetScriptsResp, error) {
	sCtx, err := authcontext.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", fmt.Sprintf("bearer %s", sCtx.AuthToken))

	smReq := &scriptmgrpb.GetScriptsReq{}
	smResp, err := s.ScriptMgr.GetScripts(ctx, smReq)
	if err != nil {
		return nil, err
	}
	resp := &cloudapipb.GetScriptsResp{
		Scripts: make([]*cloudapipb.ScriptMetadata, len(smResp.Scripts)),
	}
	for i, script := range smResp.Scripts {
		resp.Scripts[i] = &cloudapipb.ScriptMetadata{
			ID:          pbutils.UUIDFromProtoOrNil(script.ID).String(),
			Name:        script.Name,
			Desc:        script.Desc,
			HasLiveView: script.HasLiveView,
		}
	}
	return resp, nil
}

// GetScriptContents returns the pxl string of the script.
func (s *ScriptMgrServer) GetScriptContents(ctx context.Context, req *cloudapipb.GetScriptContentsReq) (*cloudapipb.GetScriptContentsResp, error) {
	sCtx, err := authcontext.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", fmt.Sprintf("bearer %s", sCtx.AuthToken))

	smReq := &scriptmgrpb.GetScriptContentsReq{
		ScriptID: pbutils.ProtoFromUUIDStrOrNil(req.ScriptID),
	}
	smResp, err := s.ScriptMgr.GetScriptContents(ctx, smReq)
	if err != nil {
		return nil, err
	}
	return &cloudapipb.GetScriptContentsResp{
		Metadata: &cloudapipb.ScriptMetadata{
			ID:          req.ScriptID,
			Name:        smResp.Metadata.Name,
			Desc:        smResp.Metadata.Desc,
			HasLiveView: smResp.Metadata.HasLiveView,
		},
		Contents: smResp.Contents,
	}, nil
}
