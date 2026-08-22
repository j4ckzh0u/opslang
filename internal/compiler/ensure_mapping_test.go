package compiler

import "testing"

// sdkMapping is a curated subset of the opsspec table (AOT mode), so full
// coverage cannot be required — but the Ansible-parity ensure family is the
// documented flagship of AOT mode and must never silently drop out of it.
func TestEnsureFamilyInCodegen(t *testing.T) {
	for _, name := range []string{
		"pkg.ensure",
		"user.ensure", "user.absent",
		"group.ensure", "group.absent",
		"service.ensure", "service.ensure_enabled",
		"file.ensure",
	} {
		if _, ok := sdkMapping[name]; !ok {
			t.Errorf("AOT codegen must map %s to an SDK call", name)
		}
	}
}

// Omitted trailing optional arguments are padded with zero values so the
// interpreter's optional-argument semantics hold in AOT binaries too.
func TestGenerateSDKCallPadsMissingTrailingArgs(t *testing.T) {
	m := sdkMapping["user.ensure"] // params: s, ms
	call := m.generateSDKCall("opsuser", []string{`"svc"`})
	want := `opsuser.Ensure(opsStr("svc"), nil)`
	if call != want {
		t.Errorf("padded call = %q, want %q", call, want)
	}

	f := sdkMapping["file.ensure"] // params: s, s, s
	call = f.generateSDKCall("file", []string{`"/d"`, `"directory"`})
	want = `file.Ensure(opsStr("/d"), opsStr("directory"), "")`
	if call != want {
		t.Errorf("padded call = %q, want %q", call, want)
	}
}
