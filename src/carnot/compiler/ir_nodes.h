#pragma once
#include <pypa/ast/ast.hh>

#include <map>
#include <memory>
#include <string>
#include <unordered_map>
#include <unordered_set>
#include <utility>
#include <vector>

#include "src/carnot/plan/dag.h"
#include "src/carnot/plan/operators.h"
#include "src/carnot/plan/relation.h"
#include "src/common/statusor.h"

namespace pl {
namespace carnot {
namespace compiler {

class IRNode;
using IRNodePtr = std::unique_ptr<IRNode>;
using ColExprMap = std::unordered_map<std::string, IRNode*>;

/**
 * IR contains the intermediate representation of the query
 * before compiling into the logical plan.
 */
class IR {
 public:
  /**
   * @brief Node factory that adds a node to the list,
   * updates an id, then returns a pointer to manipulate.
   *
   * The object will be owned by the IR object that created it.
   *
   * @tparam TOperator the type of the operator.
   * @return StatusOr<TOperator *> - the node will be owned
   * by this IR object.
   */
  template <typename TOperator>
  StatusOr<TOperator*> MakeNode() {
    auto id = id_node_map_.size();
    auto node = std::make_unique<TOperator>(id);
    dag_.AddNode(node->id());
    node->SetGraphPtr(this);
    TOperator* raw = node.get();
    id_node_map_.emplace(node->id(), std::move(node));
    return raw;
  }

  Status AddEdge(int64_t parent, int64_t child);
  Status AddEdge(IRNode* parent, IRNode* child);
  plan::DAG& dag() { return dag_; }
  std::string DebugString();
  IRNode* Get(int64_t id) const { return id_node_map_.at(id).get(); }
  size_t size() const { return id_node_map_.size(); }

 private:
  plan::DAG dag_;
  std::unordered_map<int64_t, IRNodePtr> id_node_map_;
};

enum IRNodeType {
  MemorySourceType,
  MemorySinkType,
  RangeType,
  MapType,
  AggType,
  StringType,
  FloatType,
  IntType,
  BoolType,
  BinFuncType,
  FuncType,
  ListType,
  LambdaType,
  ColumnType,
  FuncNameType
};
static constexpr const char* IRNodeString[] = {
    "MemorySourceType", "MemorySinkType", "RangeType",  "MapType",    "AggType",
    "StringType",       "FloatType",      "IntType",    "BoolType",   "BinFuncType",
    "FuncType",         "ListType",       "LambdaType", "ColumnType", "FuncNameType"};

/**
 * @brief Node class for the IR.
 *
 * Each Operator that overlaps IR and LogicalPlan can notify the compiler by returning true in the
 * overloaded HasLogicalRepr method.
 */
class IRNode {
 public:
  IRNode() = delete;
  explicit IRNode(int64_t id, IRNodeType type) : id_(id), type_(type) {}
  virtual ~IRNode() = default;
  /**
   * @return whether or not the node has a logical representation.
   */
  virtual bool HasLogicalRepr() const = 0;
  void SetLineCol(int64_t line, int64_t col);
  int64_t line() const { return line_; }
  int64_t col() const { return col_; }
  bool line_col_set() const { return line_col_set_; }
  virtual std::string DebugString(int64_t depth) const = 0;
  virtual bool IsOp() const = 0;
  IRNodeType type() const { return type_; }
  std::string type_string() const { return IRNodeString[type()]; }
  /**
   * @brief Set the pointer to the graph.
   * The pointer is passed in by the Node factory of the graph
   * (see IR::MakeNode) so that we can add edges between this
   * object and any other objects created later on.
   *
   * @param graph_ptr : pointer to the graph object.
   */
  void SetGraphPtr(IR* graph_ptr) { graph_ptr_ = graph_ptr; }
  // Returns the ID of the operator.
  int64_t id() const { return id_; }
  IR* graph_ptr() { return graph_ptr_; }

 private:
  int64_t id_;
  // line and column where the parser read the data for this node.
  // used for highlighting errors in queries.
  int64_t line_;
  int64_t col_;
  IR* graph_ptr_;
  IRNodeType type_;
  bool line_col_set_ = false;
};

/**
 * @brief Node class for the operator
 *
 */
class OperatorIR : public IRNode {
 public:
  OperatorIR() = delete;
  bool IsOp() const { return true; }
  plan::Relation relation() const { return relation_; }
  Status SetRelation(plan::Relation relation) {
    relation_ = relation;
    return Status::OK();
  }
  bool IsRelationInit() const { return relation_init_; }
  bool HasParent() const { return has_parent_; }
  OperatorIR* parent() const { return parent_; }
  Status SetParent(IRNode* parent);

 protected:
  explicit OperatorIR(int64_t id, IRNodeType type, bool has_parent)
      : IRNode(id, type), has_parent_(has_parent) {}

 private:
  plan::Relation relation_;
  bool relation_init_ = false;
  bool has_parent_;
  OperatorIR* parent_;
};

/**
 * @brief ColumnIR wraps around columns found in the lambda functions.
 *
 */
class ColumnIR : public IRNode {
 public:
  ColumnIR() = delete;
  explicit ColumnIR(int64_t id) : IRNode(id, ColumnType) {}
  Status Init(const std::string col_name);
  bool HasLogicalRepr() const override;
  std::string col_name() const { return col_name_; }
  std::string DebugString(int64_t depth) const override;
  bool IsOp() const override { return false; }
  void SetColumnIdx(int64_t col_idx) {
    col_idx_ = col_idx;
    col_idx_set_ = true;
  }
  bool col_idx_set() const { return col_idx_set_; }
  int64_t col_idx() const { return col_idx_; }

  void SetColumnType(types::DataType type) {
    type_ = type;
    col_type_set_ = true;
  }
  bool col_type_set() const { return col_type_set_; }
  types::DataType type() const { return type_; }

 private:
  std::string col_name_;
  // The column index in the relation.
  int64_t col_idx_;
  // The data type in the relation.
  types::DataType type_;
  bool col_idx_set_ = false;
  bool col_type_set_ = false;
};

/**
 * @brief The MemorySourceIR is a dual logical plan
 * and IR node operator. It inherits from both classes
 */
class MemorySourceIR : public OperatorIR {
 public:
  MemorySourceIR() = delete;
  explicit MemorySourceIR(int64_t id) : OperatorIR(id, MemorySourceType, false) {}
  Status Init(IRNode* table_node, IRNode* select);
  bool HasLogicalRepr() const override;
  std::string DebugString(int64_t depth) const override;
  IRNode* table_node() { return table_node_; }
  IRNode* select() { return select_; }
  void SetTime(int64_t time_start_ms, int64_t time_stop_ms) {
    time_start_ms_ = time_start_ms;
    time_stop_ms_ = time_stop_ms;
    time_set_ = true;
  }
  bool IsTimeSet() const { return time_set_; }
  bool columns_set() const { return columns_set_; }
  void SetColumns(std::vector<ColumnIR*> columns) {
    columns_set_ = true;
    columns_ = columns;
  }

 private:
  IRNode* table_node_;
  IRNode* select_;
  bool time_set_;
  int64_t time_start_ms_;
  int64_t time_stop_ms_;
  // in conjunction with the relation, we can get the idx, names, and types of this column.
  std::vector<ColumnIR*> columns_;
  bool columns_set_ = false;
};

/**
 * The MemorySinkIR describes the MemorySink operator.
 */
class MemorySinkIR : public OperatorIR {
 public:
  MemorySinkIR() = delete;
  explicit MemorySinkIR(int64_t id) : OperatorIR(id, MemorySinkType, false) {}
  Status Init(IRNode* parent);
  bool HasLogicalRepr() const override;
  std::string DebugString(int64_t depth) const override;
  IRNode* parent() { return parent_; }
  void SetName(std::string name) {
    name_ = name;
    name_set_ = true;
  }
  bool name_set() const { return name_set_; }
  std::string name() const { return name_; }

 private:
  IRNode* parent_;
  std::string name_;
  bool name_set_ = false;
};

/**
 * @brief The RangeIR describe the range()
 * operator, which is combined with a Source
 * when converted to the Logical Plan.
 *
 */
class RangeIR : public OperatorIR {
 public:
  RangeIR() = delete;
  explicit RangeIR(int64_t id) : OperatorIR(id, RangeType, false) {}
  Status Init(IRNode* parent, IRNode* time_repr);
  bool HasLogicalRepr() const override;
  std::string DebugString(int64_t depth) const override;
  IRNode* time_repr() { return time_repr_; }

 private:
  IRNode* time_repr_;
};

/**
 * @brief The RangeIR describe the range()
 * operator, which is combined with a Source
 * when converted to the Logical Plan.
 *
 */
class MapIR : public OperatorIR {
 public:
  MapIR() = delete;
  explicit MapIR(int64_t id) : OperatorIR(id, MapType, false) {}
  Status Init(IRNode* parent, IRNode* lambda_func);
  bool HasLogicalRepr() const override;
  std::string DebugString(int64_t depth) const override;
  IRNode* lambda_func() const { return lambda_func_; }
  void SetColExprMap(ColExprMap col_expr_map) {
    col_expr_map_ = col_expr_map;
    col_expr_map_set_ = true;
  }
  bool col_expr_map_set() const { return col_expr_map_set_; }

 private:
  IRNode* lambda_func_;
  // The map from new column_names to expressions.
  ColExprMap col_expr_map_;
  bool col_expr_map_set_ = false;
};

/**
 * @brief The RangeIR describe the range()
 * operator, which is combined with a Source
 * when converted to the Logical Plan.
 *
 */
class AggIR : public OperatorIR {
 public:
  AggIR() = delete;
  explicit AggIR(int64_t id) : OperatorIR(id, AggType, false) {}
  Status Init(IRNode* parent, IRNode* by_func, IRNode* agg_func);
  bool HasLogicalRepr() const override;
  std::string DebugString(int64_t depth) const override;
  IRNode* by_func() const { return by_func_; }
  IRNode* agg_func() const { return agg_func_; }
  void SetGroups(std::vector<ColumnIR*> groups) {
    groups_ = groups;
    groups_set_ = true;
  }
  bool groups_set() const { return groups_set_; }
  void SetAggValMap(ColExprMap agg_val_map) {
    agg_val_map_ = agg_val_map;
    agg_val_map_set_ = true;
  }
  bool agg_val_map_set() const { return agg_val_map_set_; }

 private:
  IRNode* by_func_;
  IRNode* agg_func_;
  // contains group_names and groups columns.
  std::vector<ColumnIR*> groups_;
  bool groups_set_ = false;
  // The map from value_names to values
  ColExprMap agg_val_map_;
  bool agg_val_map_set_ = false;
};

/**
 * @brief StringIR wraps around the String AST node
 * and only contains the value of that string.
 *
 */
class StringIR : public IRNode {
 public:
  StringIR() = delete;
  explicit StringIR(int64_t id) : IRNode(id, StringType) {}
  Status Init(const std::string str);
  bool HasLogicalRepr() const override;
  std::string str() const { return str_; }
  std::string DebugString(int64_t depth) const override;
  bool IsOp() const override { return false; }

 private:
  std::string str_;
};

/**
 * @brief ColumnIR wraps around columns found in the lambda functions.
 * @brief StringIR wraps around the String AST node
 * and only contains the value of that string.
 *
 */
class FuncNameIR : public IRNode {
 public:
  FuncNameIR() = delete;
  explicit FuncNameIR(int64_t id) : IRNode(id, FuncNameType) {}
  Status Init(const std::string func_name);
  bool HasLogicalRepr() const override;
  std::string func_name() const { return func_name_; }
  std::string DebugString(int64_t depth) const override;
  bool IsOp() const override { return false; }

 private:
  std::string func_name_;
};

/**
 * @brief ListIR wraps around lists. Will maintain a
 * vector of pointers to the contained nodes in the
 * list.
 *
 */
class ListIR : public IRNode {
 public:
  ListIR() = delete;
  explicit ListIR(int64_t id) : IRNode(id, ListType) {}
  bool HasLogicalRepr() const override;
  Status AddListItem(IRNode* node);
  std::string DebugString(int64_t depth) const override;
  std::vector<IRNode*> children() { return children_; }
  bool IsOp() const override { return false; }

 private:
  std::vector<IRNode*> children_;
};

/**
 * @brief IR representation for a Lambda
 * function. Should contain an expected
 * Relation based on which columns are called
 * within the contained relation.
 *
 */
class LambdaIR : public IRNode {
 public:
  LambdaIR() = delete;
  explicit LambdaIR(int64_t id) : IRNode(id, LambdaType) {}
  Status Init(std::unordered_set<std::string> column_names, ColExprMap expr_map);
  /**
   * @brief Init for the Lambda called elsewhere. Uses a default value for the key to the expression
   * map.
   */
  Status Init(std::unordered_set<std::string> expected_column_names, IRNode* node);
  /**
   * @brief Returns the one_expr_ if it has only one expr in the col_expr_map, otherwise returns an
   * error.
   *
   * @return StatusOr<IRNode*>
   */
  StatusOr<IRNode*> GetDefaultExpr();
  bool HasLogicalRepr() const override;
  bool HasDictBody() const;
  std::string DebugString(int64_t depth) const override;
  bool IsOp() const override { return false; }

 private:
  static constexpr const char* default_key = "_default";
  std::unordered_set<std::string> expected_column_names_;
  ColExprMap col_expr_map_;
  bool has_dict_body_;
};

/**
 * @brief Represents functions with arbitrary number of values
 */
class FuncIR : public IRNode {
 public:
  FuncIR() = delete;
  explicit FuncIR(int64_t id) : IRNode(id, FuncType) {}
  Status Init(std::string func_name, std::vector<IRNode*> args);
  bool HasLogicalRepr() const override;
  std::string DebugString(int64_t depth) const override;
  std::string func_name() const { return func_name_; }
  const std::vector<IRNode*>& args() { return args_; }

  bool IsOp() const override { return false; }

 private:
  std::string func_name_;
  std::vector<IRNode*> args_;
};

/**
 * @brief Primitive values.
 */
class FloatIR : public IRNode {
 public:
  FloatIR() = delete;
  explicit FloatIR(int64_t id) : IRNode(id, FloatType) {}
  Status Init(double val);
  bool HasLogicalRepr() const override;
  std::string DebugString(int64_t depth) const override;
  double val() const { return val_; }
  bool IsOp() const override { return false; }

 private:
  double val_;
};

class IntIR : public IRNode {
 public:
  IntIR() = delete;
  explicit IntIR(int64_t id) : IRNode(id, IntType) {}
  Status Init(int64_t val);
  bool HasLogicalRepr() const override;
  std::string DebugString(int64_t depth) const override;
  int64_t val() const { return val_; }
  bool IsOp() const override { return false; }

 private:
  int64_t val_;
};

class BoolIR : public IRNode {
 public:
  BoolIR() = delete;
  explicit BoolIR(int64_t id) : IRNode(id, BoolType) {}
  Status Init(bool val);
  bool HasLogicalRepr() const override;
  std::string DebugString(int64_t depth) const override;
  bool val() const { return val_; }
  bool IsOp() const override { return false; }

 private:
  bool val_;
};
}  // namespace compiler
}  // namespace carnot
}  // namespace pl
