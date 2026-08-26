// Package security implements permission checks and audit logging.
package security

import (
	"fmt"
	"strings"

	"github.com/j4ckzh0u/opslang/internal/ast"
	"github.com/j4ckzh0u/opslang/internal/opsspec"
)

// OperationType represents the type of operation being performed.
type OperationType string

const (
	OpRead   OperationType = "read"
	OpWrite  OperationType = "write"
	OpExec   OperationType = "exec"
	OpAdmin  OperationType = "admin"
	OpSystem OperationType = "system"
)

// operationPermissions maps operations to their required privilege levels.
var operationPermissions = map[OperationType]ast.PrivilegeLevel{
	OpRead:   ast.PrivilegeReadOnly,
	OpWrite:  ast.PrivilegeAdmin,
	OpExec:   ast.PrivilegeAdmin,
	OpAdmin:  ast.PrivilegeRoot,
	OpSystem: ast.PrivilegeRoot,
}

// CheckPrivilege verifies that the script's privilege level allows the operation.
func CheckPrivilege(scriptPriv ast.PrivilegeLevel, op OperationType) error {
	required, ok := operationPermissions[op]
	if !ok {
		return fmt.Errorf("unknown operation type: %s", op)
	}

	if !privilegeAllows(scriptPriv, required) {
		return fmt.Errorf("privilege denied: operation %s requires %s, but script has %s",
			op, required, scriptPriv)
	}

	return nil
}

// privilegeAllows checks if 'has' privilege level allows 'required' level.
// Hierarchy: root > admin > read_only
func privilegeAllows(has, required ast.PrivilegeLevel) bool {
	switch required {
	case ast.PrivilegeReadOnly:
		return true // Any level allows read_only
	case ast.PrivilegeAdmin:
		return has == ast.PrivilegeAdmin || has == ast.PrivilegeRoot
	case ast.PrivilegeRoot:
		return has == ast.PrivilegeRoot
	default:
		return false
	}
}

// CheckFuncPrivilege verifies that the script's privilege level allows
// calling the named OpsLang function. It is the enforcement hook shared by
// the interpreter (runtime), the AOT compiler (static) and the runner
// (remote second check); the mutating classification itself comes from the
// canonical opsspec table, so all engines agree by construction.
//
// Functions unknown to opsspec (user-defined functions, host-injected
// custom builtins like print/len) are never restricted here: the canonical
// table is the single source of truth for which operations mutate state.
func CheckFuncPrivilege(scriptPriv ast.PrivilegeLevel, funcName string) error {
	mutating, known := opsspec.Mutating(funcName)
	if !known || !mutating {
		return nil
	}

	// Every mutating operation requires at least admin privilege.
	required := ast.PrivilegeAdmin
	if !privilegeAllows(scriptPriv, required) {
		return fmt.Errorf("privilege denied: %s requires %s privilege, but script privilege is %s (add \"privilege: %s\" at the top of the script)",
			funcName, required, scriptPriv, required)
	}
	return nil
}

// GetScriptPrivilege extracts the privilege level from a program.
// Returns read_only if no privilege statement is found (default).
func GetScriptPrivilege(prog *ast.Program) ast.PrivilegeLevel {
	for _, stmt := range prog.Statements {
		if priv, ok := stmt.(*ast.PrivilegeStatement); ok {
			return priv.Level
		}
		// Privilege must be the first statement (or after imports)
		if _, isImport := stmt.(*ast.ImportStatement); !isImport {
			break
		}
	}
	return ast.PrivilegeReadOnly // Default
}

// ClassifyOperation determines the operation type for a given function call.
// The read-vs-mutating decision comes from the canonical opsspec table (the
// same metadata the three execution engines enforce privileges with); the
// write/exec split below is advisory classification only — both categories
// require the same admin privilege. Names unknown to the table fall back to
// the conservative root-only OpSystem.
func ClassifyOperation(funcName string) OperationType {
	mutating, known := opsspec.Mutating(funcName)
	if !known {
		return OpSystem
	}
	if !mutating {
		return OpRead
	}
	if isExecKind(funcName) {
		return OpExec
	}
	return OpWrite
}

// isExecKind reports whether a mutating function is an execution-style op
// (spawns/kills processes, manages services or packages) rather than a
// data-write op. Both kinds require admin; the split only feeds reporting.
func isExecKind(funcName string) bool {
	canonical := funcName
	if c, ok := opsspec.Aliases[funcName]; ok {
		canonical = c
	}
	for _, prefix := range []string{"process.", "service.", "pkg."} {
		if strings.HasPrefix(canonical, prefix) {
			return true
		}
	}
	return canonical == "binary.exec"
}
