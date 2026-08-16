// Package security implements permission checks and audit logging.
package security

import (
	"fmt"

	"github.com/opslang/opslang/internal/ast"
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
func ClassifyOperation(funcName string) OperationType {
	// Read operations
	if isReadOperation(funcName) {
		return OpRead
	}
	// Write operations
	if isWriteOperation(funcName) {
		return OpWrite
	}
	// Exec operations
	if isExecOperation(funcName) {
		return OpExec
	}
	// Admin operations
	if isAdminOperation(funcName) {
		return OpAdmin
	}
	// System operations
	return OpSystem
}

func isReadOperation(name string) bool {
	readOps := []string{
		"sys.cpu", "sys.memory", "sys.disk", "sys.host",
		"sys.load", "sys.net", "sys.users", "sys.uptime",
		"sys.hostname", "sys.os",
		"file.read", "file.exists", "file.stat", "file.list",
		"file.checksum",
		"process.list", "process.find",
		"net.http", "net.tcp", "net.dns", "net.interfaces",
		"service.status", "pkg.list", "pkg.info",
		"time.", "json.", "yaml.",
	}
	for _, op := range readOps {
		if len(name) >= len(op) && name[:len(op)] == op {
			return true
		}
	}
	return false
}

func isWriteOperation(name string) bool {
	writeOps := []string{
		"file.write", "file.append", "file.copy", "file.move",
		"file.delete", "file.mkdir", "file.chmod",
		// file.template only READS the template and returns rendered text;
		// it does not modify any file. Keep it out of write ops.
	}
	for _, op := range writeOps {
		if len(name) >= len(op) && name[:len(op)] == op {
			return true
		}
	}
	return false
}

func isExecOperation(name string) bool {
	execOps := []string{
		"process.exec", "process.kill",
		"service.start", "service.stop", "service.restart",
		"service.enable", "service.disable",
		"pkg.install", "pkg.remove",
	}
	for _, op := range execOps {
		if len(name) >= len(op) && name[:len(op)] == op {
			return true
		}
	}
	return false
}

func isAdminOperation(name string) bool {
	// Admin operations require explicit admin privilege
	return false // For now, no special admin operations
}
