package user

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// stubBinDir writes executable stub scripts named after each binary into a
// temp dir and returns (dir, logPath). Every invocation appends
// "<binary> <args...>" to the log so tests can assert exactly which
// commands ran — the core of idempotency verification.
func stubBinDir(t *testing.T, binaries ...string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "invocations.log")
	for _, bin := range binaries {
		script := "#!/bin/sh\necho \"$0 $*\" >> " + logPath + "\n"
		p := filepath.Join(dir, bin)
		if err := os.WriteFile(p, []byte(script), 0755); err != nil {
			t.Fatalf("write stub %s: %v", bin, err)
		}
	}
	return dir, logPath
}

func readInvocations(t *testing.T, logPath string) []string {
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

// fixtureEnv swaps passwdFile and the command binaries for stubs, restores on cleanup.
func fixtureEnv(t *testing.T, passwdContent string, binaries ...string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "fixture")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	passwd := filepath.Join(dir, "passwd")
	if err := os.WriteFile(passwd, []byte(passwdContent), 0644); err != nil {
		t.Fatal(err)
	}
	stubDir, logPath := stubBinDir(t, binaries...)

	oldPasswd := passwdFile
	passwdFile = passwd
	oldAdd, oldMod, oldDel := useraddBin, usermodBin, userdelBin
	useraddBin = filepath.Join(stubDir, "useradd")
	usermodBin = filepath.Join(stubDir, "usermod")
	userdelBin = filepath.Join(stubDir, "userdel")
	t.Cleanup(func() {
		passwdFile = oldPasswd
		useraddBin, usermodBin, userdelBin = oldAdd, oldMod, oldDel
	})
	return logPath
}

var ensurePasswdFixture = `root:x:0:0:root:/root:/bin/bash
daemon:x:1:1:daemon:/usr/sbin:/usr/sbin/nologin
deploy:x:1001:1001::/home/deploy:/bin/bash
`

func TestEnsureRejectsEmptyUsername(t *testing.T) {
	res, err := Ensure("", nil)
	if err == nil {
		t.Fatal("Ensure(\"\") must fail")
	}
	if res.Error == "" {
		t.Error("result should carry the error description")
	}
}

func TestEnsureCreatesMissingUser(t *testing.T) {
	logPath := fixtureEnv(t, ensurePasswdFixture, "useradd", "usermod", "userdel")

	res, err := Ensure("opslang-demo", map[string]string{"shell": "/bin/bash", "create_home": "true"})
	if err != nil {
		t.Fatalf("Ensure create: %v", err)
	}
	if !res.Changed || !res.Present {
		t.Errorf("create must report changed=true present=true, got %+v", res)
	}
	if res.Message != "user created" {
		t.Errorf("unexpected message %q", res.Message)
	}
	invocations := readInvocations(t, logPath)
	if len(invocations) != 1 || !strings.Contains(invocations[0], "useradd") {
		t.Errorf("expected exactly one useradd call, got %v", invocations)
	}
	if !strings.Contains(invocations[0], "-s /bin/bash") {
		t.Errorf("useradd must receive shell opt, got %q", invocations[0])
	}
}

func TestEnsureIsIdempotentForConvergedUser(t *testing.T) {
	logPath := fixtureEnv(t, ensurePasswdFixture, "useradd", "usermod", "userdel")

	res, err := Ensure("deploy", map[string]string{"shell": "/bin/bash"})
	if err != nil {
		t.Fatalf("Ensure no-op: %v", err)
	}
	if res.Changed {
		t.Error("already-converged user must report changed=false")
	}
	if res.Message != "user already up to date" {
		t.Errorf("unexpected message %q", res.Message)
	}
	if got := readInvocations(t, logPath); len(got) != 0 {
		t.Errorf("idempotent run must execute zero commands, got %v", got)
	}
	if res.Shell != "/bin/bash" || res.Home != "/home/deploy" {
		t.Errorf("should echo current attributes, got shell=%q home=%q", res.Shell, res.Home)
	}
}

func TestEnsureConvergesShellDrift(t *testing.T) {
	logPath := fixtureEnv(t, ensurePasswdFixture, "useradd", "usermod", "userdel")

	res, err := Ensure("deploy", map[string]string{"shell": "/usr/sbin/nologin"})
	if err != nil {
		t.Fatalf("Ensure drift: %v", err)
	}
	if !res.Changed {
		t.Error("shell drift must be converged with changed=true")
	}
	if res.Message != "user attributes converged" {
		t.Errorf("unexpected message %q", res.Message)
	}
	invocations := readInvocations(t, logPath)
	if len(invocations) != 1 || !strings.Contains(invocations[0], "usermod -s /usr/sbin/nologin deploy") {
		t.Errorf("expected one usermod shell call, got %v", invocations)
	}
}

func TestAbsentRemovesThenNoOps(t *testing.T) {
	logPath := fixtureEnv(t, ensurePasswdFixture, "useradd", "usermod", "userdel")

	res, err := Absent("deploy", false)
	if err != nil {
		t.Fatalf("Absent remove: %v", err)
	}
	if !res.Changed || res.Present {
		t.Errorf("removal must report changed=true present=false, got %+v", res)
	}

	// The stub userdel cannot edit the fixture, so mirror what the real
	// userdel did to /etc/passwd before asserting idempotency.
	if err := os.WriteFile(passwdFile, []byte("root:x:0:0:root:/root:/bin/bash\n"), 0644); err != nil {
		t.Fatal(err)
	}

	res2, err := Absent("deploy", false)
	if err != nil {
		t.Fatalf("Absent no-op: %v", err)
	}
	if res2.Changed {
		t.Error("second Absent must be changed=false")
	}
	if res2.Message != "user already absent" {
		t.Errorf("unexpected message %q", res2.Message)
	}
	invocations := readInvocations(t, logPath)
	if len(invocations) != 1 || !strings.Contains(invocations[0], "userdel") {
		t.Errorf("expected exactly one userdel total, got %v", invocations)
	}
}

func TestAbsentAlreadyMissingIsNoOp(t *testing.T) {
	logPath := fixtureEnv(t, "root:x:0:0:root:/root:/bin/bash\n", "useradd", "usermod", "userdel")

	res, err := Absent("ghost-user", false)
	if err != nil {
		t.Fatalf("Absent missing: %v", err)
	}
	if res.Changed || res.Present {
		t.Errorf("missing user must be changed=false present=false, got %+v", res)
	}
	if got := readInvocations(t, logPath); len(got) != 0 {
		t.Errorf("no command may run for missing user, got %v", got)
	}
}

func TestAbsentRefusesRoot(t *testing.T) {
	fixtureEnv(t, ensurePasswdFixture, "useradd", "usermod", "userdel")

	if _, err := Absent("root", true); err == nil {
		t.Fatal("Absent(root) must be refused")
	}
}

func TestAbsentRejectsEmptyUsername(t *testing.T) {
	if _, err := Absent("", true); err == nil {
		t.Fatal("Absent(\"\") must fail")
	}
}

func TestAddBindsExistingSameNameGroup(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "fixture")
	os.MkdirAll(dir, 0755)
	// deploy user absent from passwd, but a group named "deploy" exists.
	os.WriteFile(filepath.Join(dir, "passwd"), []byte("root:x:0:0:root:/root:/bin/bash\n"), 0644)
	os.WriteFile(filepath.Join(dir, "group"), []byte("deploy:x:1500:\n"), 0644)
	stubDir, logPath := stubBinDir(t, "useradd", "usermod", "userdel")

	oldPasswd, oldGroup := passwdFile, groupFile
	passwdFile = filepath.Join(dir, "passwd")
	groupFile = filepath.Join(dir, "group")
	oldAdd, oldMod, oldDel := useraddBin, usermodBin, userdelBin
	useraddBin, usermodBin, userdelBin = filepath.Join(stubDir, "useradd"), filepath.Join(stubDir, "usermod"), filepath.Join(stubDir, "userdel")
	t.Cleanup(func() {
		passwdFile, groupFile = oldPasswd, oldGroup
		useraddBin, usermodBin, userdelBin = oldAdd, oldMod, oldDel
	})

	res, err := Add("deploy", map[string]string{"shell": "/bin/bash"})
	if err != nil {
		t.Fatalf("Add with same-name group: %v", err)
	}
	if !res.Changed {
		t.Error("user must be created")
	}
	invocations := readInvocations(t, logPath)
	if len(invocations) != 1 || !strings.Contains(invocations[0], "-g 1500") {
		t.Errorf("useradd must bind existing same-name group via -g 1500, got %v", invocations)
	}
}
