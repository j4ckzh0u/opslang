package security

import (
	"testing"
)

func TestParsePrivilege(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Privilege
		wantErr bool
	}{
		{"read_only", "read_only", ReadOnly, false},
		{"readonly", "readonly", ReadOnly, false},
		{"READ_ONLY", "READ_ONLY", ReadOnly, false},
		{"admin", "admin", Admin, false},
		{"Admin", "Admin", Admin, false},
		{"root", "root", Root, false},
		{"ROOT", "ROOT", Root, false},
		{"invalid", "invalid", ReadOnly, true},
		{"empty", "", ReadOnly, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePrivilege(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePrivilege(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParsePrivilege(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestPrivilegeString(t *testing.T) {
	tests := []struct {
		priv Privilege
		want string
	}{
		{ReadOnly, "read_only"},
		{Admin, "admin"},
		{Root, "root"},
		{Privilege(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.priv.String(); got != tt.want {
				t.Errorf("Privilege.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrivilegeCanCall(t *testing.T) {
	tests := []struct {
		name string
		priv Privilege
		op   string
		want bool
	}{
		// Read operations - always allowed
		{"ReadOnly can read", ReadOnly, "sys.cpu.usage", true},
		{"Admin can read", Admin, "sys.cpu.usage", true},
		{"Root can read", Root, "sys.cpu.usage", true},

		// Mutation operations
		{"ReadOnly cannot write", ReadOnly, "file.write", false},
		{"Admin can write", Admin, "file.write", true},
		{"Root can write", Root, "file.write", true},

		{"ReadOnly cannot delete", ReadOnly, "file.delete", false},
		{"Admin can delete", Admin, "file.delete", true},
		{"Root can delete", Root, "file.delete", true},

		{"ReadOnly cannot service.start", ReadOnly, "service.start", false},
		{"Admin can service.start", Admin, "service.start", true},
		{"Root can service.start", Root, "service.start", true},

		{"ReadOnly cannot pkg.install", ReadOnly, "pkg.install", false},
		{"Admin can pkg.install", Admin, "pkg.install", true},
		{"Root can pkg.install", Root, "pkg.install", true},

		{"ReadOnly cannot process.kill", ReadOnly, "process.kill", false},
		{"Admin can process.kill", Admin, "process.kill", true},
		{"Root can process.kill", Root, "process.kill", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.priv.CanCall(tt.op); got != tt.want {
				t.Errorf("Privilege(%v).CanCall(%q) = %v, want %v", tt.priv, tt.op, got, tt.want)
			}
		})
	}
}

func TestIsMutationOp(t *testing.T) {
	tests := []struct {
		op   string
		want bool
	}{
		{"file.write", true},
		{"file.delete", true},
		{"file.move", true},
		{"service.start", true},
		{"service.stop", true},
		{"pkg.install", true},
		{"process.kill", true},
		{"process.exec", true},

		// Read operations
		{"sys.cpu.usage", false},
		{"sys.memory.info", false},
		{"file.read", false},
		{"file.exists", false},
	}

	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			if got := IsMutationOp(tt.op); got != tt.want {
				t.Errorf("IsMutationOp(%q) = %v, want %v", tt.op, got, tt.want)
			}
		})
	}
}

func TestChecker(t *testing.T) {
	t.Run("ReadOnly checker blocks mutations", func(t *testing.T) {
		checker := NewChecker(ReadOnly)
		if err := checker.ValidateCall("file.write"); err == nil {
			t.Error("Expected error for ReadOnly calling file.write")
		}
		if err := checker.ValidateCall("sys.cpu.usage"); err != nil {
			t.Errorf("Expected no error for ReadOnly reading, got: %v", err)
		}
	})

	t.Run("Admin checker allows most operations", func(t *testing.T) {
		checker := NewChecker(Admin)
		if err := checker.ValidateCall("file.write"); err != nil {
			t.Errorf("Expected no error for Admin calling file.write, got: %v", err)
		}
		if err := checker.ValidateCall("sys.cpu.usage"); err != nil {
			t.Errorf("Expected no error for Admin reading, got: %v", err)
		}
	})

	t.Run("Root checker allows all operations", func(t *testing.T) {
		checker := NewChecker(Root)
		if err := checker.ValidateCall("file.write"); err != nil {
			t.Errorf("Expected no error for Root calling file.write, got: %v", err)
		}
		if err := checker.ValidateCall("pkg.install"); err != nil {
			t.Errorf("Expected no error for Root calling pkg.install, got: %v", err)
		}
	})

	t.Run("Checker Required returns correct privilege", func(t *testing.T) {
		checker := NewChecker(Admin)
		if got := checker.Required(); got != Admin {
			t.Errorf("Checker.Required() = %v, want %v", got, Admin)
		}
	})
}
