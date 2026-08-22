package group

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// stubGroupEnv installs a fixture /etc/group plus stub groupadd/groupdel
// binaries that log every invocation. Returns the invocation log path.
func stubGroupEnv(t *testing.T, groupContent string) string {
	t.Helper()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "invocations.log")

	groupPath := filepath.Join(dir, "group")
	if err := os.WriteFile(groupPath, []byte(groupContent), 0644); err != nil {
		t.Fatalf("write group fixture: %v", err)
	}
	for _, bin := range []string{"groupadd", "groupdel"} {
		script := "#!/bin/sh\necho \"$0 $*\" >> " + logPath + "\n"
		if err := os.WriteFile(filepath.Join(dir, bin), []byte(script), 0755); err != nil {
			t.Fatalf("write stub %s: %v", bin, err)
		}
	}

	oldFile, oldAdd, oldDel := groupFile, groupaddBin, groupdelBin
	groupFile = groupPath
	groupaddBin = filepath.Join(dir, "groupadd")
	groupdelBin = filepath.Join(dir, "groupdel")
	t.Cleanup(func() {
		groupFile = oldFile
		groupaddBin, groupdelBin = oldAdd, oldDel
	})
	return logPath
}

func groupInvocations(t *testing.T, logPath string) []string {
	t.Helper()
	data, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read invocation log: %v", err)
	}
	var out []string
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

var ensureGroupFixture = `root:x:0:
adm:x:4:syslog
developers:x:1500:alice
`

func TestEnsureRejectsEmptyName(t *testing.T) {
	if _, err := Ensure("", nil); err == nil {
		t.Fatal("Ensure(\"\") must fail")
	}
}

func TestEnsureCreatesMissingGroup(t *testing.T) {
	logPath := stubGroupEnv(t, ensureGroupFixture)

	res, err := Ensure("opslang-demo", map[string]string{})
	if err != nil {
		t.Fatalf("Ensure create: %v", err)
	}
	if !res.Changed || !res.Present {
		t.Errorf("create must report changed=true present=true, got %+v", res)
	}
	if res.Message != "group created" {
		t.Errorf("unexpected message %q", res.Message)
	}
	invocations := groupInvocations(t, logPath)
	if len(invocations) != 1 || !strings.Contains(invocations[0], "groupadd opslang-demo") {
		t.Errorf("expected one groupadd call, got %v", invocations)
	}
}

func TestEnsureExistingGroupIsNoOp(t *testing.T) {
	logPath := stubGroupEnv(t, ensureGroupFixture)

	res, err := Ensure("developers", nil)
	if err != nil {
		t.Fatalf("Ensure no-op: %v", err)
	}
	if res.Changed {
		t.Error("existing group must report changed=false")
	}
	if res.GID != 1500 {
		t.Errorf("should echo existing GID 1500, got %d", res.GID)
	}
	if got := groupInvocations(t, logPath); len(got) != 0 {
		t.Errorf("idempotent run must execute zero commands, got %v", got)
	}
}

func TestAddIsNowIdempotent(t *testing.T) {
	logPath := stubGroupEnv(t, ensureGroupFixture)

	res, err := Add("developers", nil)
	if err != nil {
		t.Fatalf("Add existing: %v", err)
	}
	if res.Changed {
		t.Error("Add on existing group must report changed=false (Ansible semantics)")
	}
	if got := groupInvocations(t, logPath); len(got) != 0 {
		t.Errorf("no groupadd may run for existing group, got %v", got)
	}
}

func TestAbsentRemovesThenNoOps(t *testing.T) {
	logPath := stubGroupEnv(t, ensureGroupFixture)

	res, err := Absent("developers")
	if err != nil {
		t.Fatalf("Absent remove: %v", err)
	}
	if !res.Changed || res.Present {
		t.Errorf("removal must report changed=true present=false, got %+v", res)
	}

	// Mirror what the real groupdel did to /etc/group before re-running.
	if err := os.WriteFile(groupFile, []byte("root:x:0:\n"), 0644); err != nil {
		t.Fatal(err)
	}
	res2, err := Absent("developers")
	if err != nil {
		t.Fatalf("Absent no-op: %v", err)
	}
	if res2.Changed {
		t.Error("second Absent must be changed=false")
	}
	if res2.Message != "group already absent" {
		t.Errorf("unexpected message %q", res2.Message)
	}
	invocations := groupInvocations(t, logPath)
	if len(invocations) != 1 || !strings.Contains(invocations[0], "groupdel developers") {
		t.Errorf("expected exactly one groupdel total, got %v", invocations)
	}
}

func TestAbsentAlreadyMissingRunsNothing(t *testing.T) {
	logPath := stubGroupEnv(t, "root:x:0:\n")

	res, err := Absent("ghost-group")
	if err != nil {
		t.Fatalf("Absent missing: %v", err)
	}
	if res.Changed || res.Present {
		t.Errorf("missing group must be changed=false present=false, got %+v", res)
	}
	if got := groupInvocations(t, logPath); len(got) != 0 {
		t.Errorf("no command may run, got %v", got)
	}
}

func TestAbsentRejectsEmptyName(t *testing.T) {
	if _, err := Absent(""); err == nil {
		t.Fatal("Absent(\"\") must fail")
	}
}

func TestEnsureResultJSONFields(t *testing.T) {
	// Field names are a stable contract consumed by ops scripts; lock them.
	res := EnsureResult{Name: "x", Present: true, Changed: false, Action: "ensure", GID: 42, Message: "m"}
	b, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"name"`, `"present"`, `"changed"`, `"action"`, `"gid"`, `"message"`} {
		if !strings.Contains(string(b), key) {
			t.Errorf("JSON must contain %s, got %s", key, b)
		}
	}
}
