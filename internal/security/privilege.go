package security

import (
	"fmt"
	"strings"
)

// Privilege represents the privilege level of a script
type Privilege int

const (
	ReadOnly Privilege = iota
	Admin
	Root
)

// String returns the string representation of the privilege level
func (p Privilege) String() string {
	switch p {
	case ReadOnly:
		return "read_only"
	case Admin:
		return "admin"
	case Root:
		return "root"
	default:
		return "unknown"
	}
}

// ParsePrivilege parses a string into a Privilege level
func ParsePrivilege(s string) (Privilege, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "read_only", "readonly":
		return ReadOnly, nil
	case "admin":
		return Admin, nil
	case "root":
		return Root, nil
	default:
		return ReadOnly, fmt.Errorf("unknown privilege level: %s", s)
	}
}

// mutationOps contains operations that modify system state
var mutationOps = map[string]bool{
	"file.write":      true,
	"file.delete":     true,
	"file.move":       true,
	"file.copy":       true,
	"file.append":     true,
	"file.touch":      true,
	"file.template":   true,
	"service.start":   true,
	"service.stop":    true,
	"service.restart": true,
	"service.enable":  true,
	"pkg.install":     true,
	"pkg.remove":      true,
	"process.kill":    true,
	"process.exec":    true,
}

// IsMutationOp checks if an operation is a mutation operation
func IsMutationOp(op string) bool {
	// Check exact match first
	if mutationOps[op] {
		return true
	}

	// Check prefix match for service.*
	if strings.HasPrefix(op, "service.") {
		return true
	}

	return false
}

// CanCall checks if the privilege level allows calling the operation
func (p Privilege) CanCall(op string) bool {
	if !IsMutationOp(op) {
		return true // Read operations are always allowed
	}

	switch p {
	case Root:
		return true // Root can call everything
	case Admin:
		return true // Admin can call most mutations
	case ReadOnly:
		return false // ReadOnly cannot call mutations
	default:
		return false
	}
}

// Checker validates operations against privilege requirements
type Checker struct {
	required Privilege
}

// NewChecker creates a new privilege checker with the required privilege level
func NewChecker(required Privilege) *Checker {
	return &Checker{required: required}
}

// ValidateCall checks if an operation is allowed for the required privilege level
func (c *Checker) ValidateCall(op string) error {
	if c.required.CanCall(op) {
		return nil
	}

	return fmt.Errorf("operation %q not allowed with privilege level %s", op, c.required.String())
}

// Required returns the required privilege level
func (c *Checker) Required() Privilege {
	return c.required
}
