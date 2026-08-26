package security

import (
	"testing"

	"github.com/j4ckzh0u/opslang/internal/ast"
)

func TestCheckPrivilege(t *testing.T) {
	tests := []struct {
		name      string
		priv      ast.PrivilegeLevel
		op        OperationType
		wantErr   bool
		errSubstr string
	}{
		{"read_only allows read", ast.PrivilegeReadOnly, OpRead, false, ""},
		{"read_only denies write", ast.PrivilegeReadOnly, OpWrite, true, "privilege denied"},
		{"read_only denies exec", ast.PrivilegeReadOnly, OpExec, true, "privilege denied"},
		{"read_only denies admin", ast.PrivilegeReadOnly, OpAdmin, true, "privilege denied"},
		{"read_only denies system", ast.PrivilegeReadOnly, OpSystem, true, "privilege denied"},

		{"admin allows read", ast.PrivilegeAdmin, OpRead, false, ""},
		{"admin allows write", ast.PrivilegeAdmin, OpWrite, false, ""},
		{"admin allows exec", ast.PrivilegeAdmin, OpExec, false, ""},
		{"admin denies admin op", ast.PrivilegeAdmin, OpAdmin, true, "privilege denied"},
		{"admin denies system", ast.PrivilegeAdmin, OpSystem, true, "privilege denied"},

		{"root allows read", ast.PrivilegeRoot, OpRead, false, ""},
		{"root allows write", ast.PrivilegeRoot, OpWrite, false, ""},
		{"root allows exec", ast.PrivilegeRoot, OpExec, false, ""},
		{"root allows admin", ast.PrivilegeRoot, OpAdmin, false, ""},
		{"root allows system", ast.PrivilegeRoot, OpSystem, false, ""},

		{"unknown operation", ast.PrivilegeRoot, OperationType("bogus"), true, "unknown operation type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPrivilege(tt.priv, tt.op)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.errSubstr != "" && !containsStr(err.Error(), tt.errSubstr) {
					t.Errorf("error = %q, want substring %q", err.Error(), tt.errSubstr)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestPrivilegeAllows(t *testing.T) {
	tests := []struct {
		name     string
		has      ast.PrivilegeLevel
		required ast.PrivilegeLevel
		want     bool
	}{
		{"read_only satisfies read_only", ast.PrivilegeReadOnly, ast.PrivilegeReadOnly, true},
		{"admin satisfies read_only", ast.PrivilegeAdmin, ast.PrivilegeReadOnly, true},
		{"root satisfies read_only", ast.PrivilegeRoot, ast.PrivilegeReadOnly, true},
		{"read_only does not satisfy admin", ast.PrivilegeReadOnly, ast.PrivilegeAdmin, false},
		{"admin satisfies admin", ast.PrivilegeAdmin, ast.PrivilegeAdmin, true},
		{"root satisfies admin", ast.PrivilegeRoot, ast.PrivilegeAdmin, true},
		{"read_only does not satisfy root", ast.PrivilegeReadOnly, ast.PrivilegeRoot, false},
		{"admin does not satisfy root", ast.PrivilegeAdmin, ast.PrivilegeRoot, false},
		{"root satisfies root", ast.PrivilegeRoot, ast.PrivilegeRoot, true},
		{"unknown required level", ast.PrivilegeRoot, ast.PrivilegeLevel("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := privilegeAllows(tt.has, tt.required)
			if got != tt.want {
				t.Errorf("privilegeAllows(%q, %q) = %v, want %v", tt.has, tt.required, got, tt.want)
			}
		})
	}
}

func TestGetScriptPrivilege(t *testing.T) {
	tests := []struct {
		name string
		prog *ast.Program
		want ast.PrivilegeLevel
	}{
		{
			name: "explicit root privilege",
			prog: &ast.Program{
				Statements: []ast.Statement{
					&ast.PrivilegeStatement{Level: ast.PrivilegeRoot},
				},
			},
			want: ast.PrivilegeRoot,
		},
		{
			name: "explicit admin privilege",
			prog: &ast.Program{
				Statements: []ast.Statement{
					&ast.PrivilegeStatement{Level: ast.PrivilegeAdmin},
				},
			},
			want: ast.PrivilegeAdmin,
		},
		{
			name: "explicit read_only privilege",
			prog: &ast.Program{
				Statements: []ast.Statement{
					&ast.PrivilegeStatement{Level: ast.PrivilegeReadOnly},
				},
			},
			want: ast.PrivilegeReadOnly,
		},
		{
			name: "no privilege statement defaults to read_only",
			prog: &ast.Program{
				Statements: []ast.Statement{},
			},
			want: ast.PrivilegeReadOnly,
		},
		{
			name: "privilege after import",
			prog: &ast.Program{
				Statements: []ast.Statement{
					&ast.ImportStatement{},
					&ast.PrivilegeStatement{Level: ast.PrivilegeAdmin},
				},
			},
			want: ast.PrivilegeAdmin,
		},
		{
			name: "privilege must be before non-import statements",
			prog: &ast.Program{
				Statements: []ast.Statement{
					&ast.ExpressionStatement{},
					&ast.PrivilegeStatement{Level: ast.PrivilegeRoot},
				},
			},
			want: ast.PrivilegeReadOnly,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetScriptPrivilege(tt.prog)
			if got != tt.want {
				t.Errorf("GetScriptPrivilege() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClassifyOperation(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		want     OperationType
	}{
		{"sys.cpu is read", "sys.cpu.usage", OpRead},
		{"sys.memory is read", "sys.memory.info", OpRead},
		{"sys.hostname is read", "sys.hostname", OpRead},
		{"file.read is read", "file.read", OpRead},
		{"file.exists is read", "file.exists", OpRead},
		{"file.checksum is read", "file.checksum", OpRead},
		{"process.list is read", "process.list", OpRead},
		{"net.http.get is read", "net.http.get", OpRead},
		{"net.http_get is read", "net.http_get", OpRead},
		{"service.status is read", "service.status", OpRead},
		{"pkg.list is read", "pkg.list", OpRead},
		{"time.sleep is read (delays, changes nothing)", "time.sleep", OpRead},
		{"file.template is read (renders, never writes)", "file.template", OpRead},

		{"file.write is write", "file.write", OpWrite},
		{"file.delete is write", "file.delete", OpWrite},
		{"file.copy is write", "file.copy", OpWrite},
		{"file.mkdir is write", "file.mkdir", OpWrite},
		{"file.append is write", "file.append", OpWrite},
		{"file.chmod is write", "file.chmod", OpWrite},
		{"file.distribute is write (writes on remote hosts)", "file.distribute", OpWrite},
		{"file.collect is write (writes collected archives)", "file.collect", OpWrite},
		{"net.http_post is write (changes remote state)", "net.http_post", OpWrite},

		{"process.exec is exec", "process.exec", OpExec},
		{"process.kill is exec", "process.kill", OpExec},
		{"service.start is exec", "service.start", OpExec},
		{"service.stop is exec", "service.stop", OpExec},
		{"service.enable is exec", "service.enable", OpExec},
		{"pkg.install is exec", "pkg.install", OpExec},
		{"pkg.remove is exec", "pkg.remove", OpExec},
		{"binary.exec is exec (arbitrary binary)", "binary.exec", OpExec},

		{"unknown falls to system", "custom.unknown.func", OpSystem},
		{"empty falls to system", "", OpSystem},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyOperation(tt.funcName)
			if got != tt.want {
				t.Errorf("ClassifyOperation(%q) = %q, want %q", tt.funcName, got, tt.want)
			}
		})
	}
}

func TestCheckFuncPrivilege(t *testing.T) {
	tests := []struct {
		name      string
		priv      ast.PrivilegeLevel
		funcName  string
		wantErr   bool
		errSubstr string
	}{
		// read_only: every read-only function is allowed...
		{"read_only allows sys.cpu.usage", ast.PrivilegeReadOnly, "sys.cpu.usage", false, ""},
		{"read_only allows file.read", ast.PrivilegeReadOnly, "file.read", false, ""},
		{"read_only allows file.template", ast.PrivilegeReadOnly, "file.template", false, ""},
		{"read_only allows net.http_get", ast.PrivilegeReadOnly, "net.http_get", false, ""},
		{"read_only allows process.list", ast.PrivilegeReadOnly, "process.list", false, ""},
		{"read_only allows time.sleep", ast.PrivilegeReadOnly, "time.sleep", false, ""},
		// ...and every mutating function is denied, with its name in the error.
		{"read_only denies file.write", ast.PrivilegeReadOnly, "file.write", true, "file.write"},
		{"read_only denies file.mkdir", ast.PrivilegeReadOnly, "file.mkdir", true, "privilege denied"},
		{"read_only denies process.exec", ast.PrivilegeReadOnly, "process.exec", true, "privilege denied"},
		{"read_only denies service.start", ast.PrivilegeReadOnly, "service.start", true, "privilege denied"},
		{"read_only denies pkg.install", ast.PrivilegeReadOnly, "pkg.install", true, "privilege denied"},
		{"read_only denies net.http_post", ast.PrivilegeReadOnly, "net.http_post", true, "privilege denied"},
		{"read_only denies file.distribute", ast.PrivilegeReadOnly, "file.distribute", true, "privilege denied"},
		{"read_only denies binary.exec", ast.PrivilegeReadOnly, "binary.exec", true, "privilege denied"},
		// Aliases resolve before the check.
		{"read_only denies alias net.http.post", ast.PrivilegeReadOnly, "net.http.post", true, "privilege denied"},
		{"read_only allows alias net.http.get", ast.PrivilegeReadOnly, "net.http.get", false, ""},

		// admin and root may mutate.
		{"admin allows file.write", ast.PrivilegeAdmin, "file.write", false, ""},
		{"admin allows process.exec", ast.PrivilegeAdmin, "process.exec", false, ""},
		{"root allows file.delete", ast.PrivilegeRoot, "file.delete", false, ""},

		// Unknown names (custom builtins, core print/len) are not OpsLang
		// operations and must not be restricted by the op table.
		{"read_only allows unknown custom builtin", ast.PrivilegeReadOnly, "double", false, ""},
		{"read_only allows print", ast.PrivilegeReadOnly, "print", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckFuncPrivilege(tt.priv, tt.funcName)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.errSubstr != "" && !containsStr(err.Error(), tt.errSubstr) {
					t.Errorf("error = %q, want substring %q", err.Error(), tt.errSubstr)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstr(s, substr)
}

func searchSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
