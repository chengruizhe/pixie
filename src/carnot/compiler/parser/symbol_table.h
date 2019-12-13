#pragma once

#include <algorithm>
#include <iostream>
#include <memory>
#include <set>
#include <string>
#include <unordered_map>
#include <unordered_set>
#include <utility>
#include <vector>

#include <pypa/ast/ast.hh>
#include <pypa/ast/tree_walker.hh>

#include "src/carnot/compiler/compiler_state/compiler_state.h"
#include "src/carnot/compiler/ir/ast_utils.h"
#include "src/carnot/compiler/ir/ir_nodes.h"

namespace pl {
namespace carnot {
namespace compiler {

/**
 * @brief Symbol Table is an abstraction used to access any table structure.
 * The two main virtualn functions to implement are HasSymbol and GetSymbol.
 *
 * AddSymbol is deliberately left out from the interface because there is a case
 * where we would like to store QLObjectPtrs but can't create a circular dependency in the system to
 * access them
 *
 */
class SymbolTable {
 public:
  bool HasSymbol(const std::string& name);
  IRNode* GetSymbol(const std::string& name);

 protected:
  bool HasSymbolImpl(const)
};
using SymbolTablePtr = std::shared_ptr<SymbolTable>;

}  // namespace compiler
}  // namespace carnot
}  // namespace pl
