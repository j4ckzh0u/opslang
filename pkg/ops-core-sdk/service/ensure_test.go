package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// stubSystemctlScript is a stateful systemctl emulator. Unit state lives in a
// flat append log (last write wins): "<unit>.active=1", "<unit>.enabled=0".
// Every verb invocation is also appended to a call log so tests can assert
// exactly which verbs ran — the core of idempotency verification. Units whose
// name starts with "noreload" fail `reload` the way real units without a
// reload handler do.
const stubSystemctlScript = `#!/bin/sh
S=$OPSLANG_STUB_STATE
L=$OPSLANG_STUB_LOG
echo "$1 $2" >> "$L"
last() { grep "^$2.$1=" "$S" 2>/dev/null | tail -1; }
case "$1" in
show)
  if [ "$(last active "$2")" = "$2.active=1" ]; then
    printf 'ActiveState=active\nSubState=running\n'
  else
    printf 'ActiveState=inactive\nSubState=dead\n'
  fi
  printf 'LoadState=loaded\nMainPID=0\nDescription=stub unit\n'
  ;;
is-enabled)
  if [ "$(last enabled "$2")" = "$2.enabled=1" ]; then echo enabled; exit 0; fi
  echo disabled
  exit 1
  ;;
start|restart)
  echo "$2.active=1" >> "$S"
  ;;
reload)
  case "$2" in
  noreload*) echo "unit does not support reload" >&2; exit 1 ;;
  *) echo "$2.active=1" >> "$S" ;;
  esac
  ;;
stop)
  echo "$2.active=0" >> "$S"
  ;;
enable)
  echo "$2.enabled=1" >> "$S"
  ;;
disable)
  echo "$2.enabled=0" >> "$S"
  ;;
esac
exit 0
`

// stubSystemctl installs the emulator and returns a reader of mutation verbs
// (everything except show/is-enabled) that actually executed.
func stubSystemctl(t *testing.T) func() []string {
	t.Helper()
	dir := t.TempDir()
	stubPath := filepath.Join(dir, "systemctl")
	if err := os.WriteFile(stubPath, []byte(stubSystemctlScript), 0755); err != nil {
		t.Fatalf("write systemctl stub: %v", err)
	}
	statePath := filepath.Join(dir, "state")
	logPath := filepath.Join(dir, "calls")
	for _, p := range []string{statePath, logPath} {
		if err := os.WriteFile(p, nil, 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("OPSLANG_STUB_STATE", statePath)
	t.Setenv("OPSLANG_STUB_LOG", logPath)

	old := systemctlBin
	systemctlBin = stubPath
	t.Cleanup(func() { systemctlBin = old })

	return func() []string {
		data, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("read call log: %v", err)
		}
		var mutations []string
		for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
			if line == "" {
				continue
			}
			verb := strings.Fields(line)[0]
			switch verb {
			case "show", "is-enabled":
			default:
				mutations = append(mutations, line)
			}
		}
		return mutations
	}
}

func TestEnsureRejectsInvalidState(t *testing.T) {
	mutations := stubSystemctl(t)

	if _, err := Ensure("nginx", "paused"); err == nil {
		t.Fatal("invalid state must be rejected")
	}
	if got := mutations(); len(got) != 0 {
		t.Errorf("invalid state must not touch systemctl, got %v", got)
	}
}

func TestEnsureRejectsEmptyName(t *testing.T) {
	if _, err := Ensure("", "started"); err == nil {
		t.Fatal("empty name must be rejected")
	}
}

func TestEnsureStartedConvergesThenNoOps(t *testing.T) {
	mutations := stubSystemctl(t)

	first, err := Ensure("nginx", "started")
	if err != nil {
		t.Fatalf("first started: %v", err)
	}
	if !first.Changed || !first.Active {
		t.Errorf("inactive->started must change, got %+v", first)
	}
	if len(first.Actions) != 1 || first.Actions[0] != "start" {
		t.Errorf("actions must be [start], got %v", first.Actions)
	}

	second, err := Ensure("nginx", "started")
	if err != nil {
		t.Fatalf("second started: %v", err)
	}
	if second.Changed {
		t.Error("second converged run must be changed=false")
	}
	if second.Message != `service "nginx" already active` {
		t.Errorf("unexpected message %q", second.Message)
	}
	all := mutations()
	if len(all) != 1 || all[0] != "start nginx" {
		t.Errorf("exactly one start total expected, got %v", all)
	}
}

func TestEnsureStoppedConvergesThenNoOps(t *testing.T) {
	mutations := stubSystemctl(t)

	// Seed the unit as active and enabled, the usual production drift.
	state := os.Getenv("OPSLANG_STUB_STATE")
	os.WriteFile(state, []byte("chrony.active=1\nchrony.enabled=1\n"), 0644)

	first, err := Ensure("chrony", "stopped")
	if err != nil {
		t.Fatalf("first stopped: %v", err)
	}
	if !first.Changed || first.Active {
		t.Errorf("active->stopped must change and end inactive, got %+v", first)
	}
	if first.Enabled != true {
		t.Error("stop must not alter enablement")
	}

	second, err := Ensure("chrony", "stopped")
	if err != nil {
		t.Fatalf("second stopped: %v", err)
	}
	if second.Changed {
		t.Error("already-stopped unit must be changed=false")
	}
	if all := mutations(); len(all) != 1 || all[0] != "stop chrony" {
		t.Errorf("exactly one stop total expected, got %v", all)
	}
}

func TestEnsureRestartedAlwaysActs(t *testing.T) {
	mutations := stubSystemctl(t)
	os.WriteFile(os.Getenv("OPSLANG_STUB_STATE"), []byte("nginx.active=1\n"), 0644)

	for i := 0; i < 2; i++ {
		res, err := Ensure("nginx", "restarted")
		if err != nil {
			t.Fatalf("restart #%d: %v", i+1, err)
		}
		if !res.Changed {
			t.Errorf("restart is never a no-op (run %d)", i+1)
		}
	}
	if all := mutations(); len(all) != 2 {
		t.Errorf("two restarts expected, got %v", all)
	}
}

func TestEnsureReloadedFallsBackToRestart(t *testing.T) {
	stubSystemctl(t)

	res, err := Ensure("noreload.service", "reloaded")
	if err != nil {
		t.Fatalf("reload fallback: %v", err)
	}
	if !res.Changed {
		t.Error("reload with fallback must report changed=true")
	}
	if len(res.Actions) != 2 || res.Actions[0] != "reload" || res.Actions[1] != "restart" {
		t.Errorf("actions must be [reload restart], got %v", res.Actions)
	}
	if !strings.Contains(res.Message, "restarted instead") {
		t.Errorf("fallback must be explained, got %q", res.Message)
	}
}

func TestEnsureEnabledConvergesBothWays(t *testing.T) {
	mutations := stubSystemctl(t)
	on, err := EnsureEnabled("nginx", true)
	if err != nil {
		t.Fatalf("enable: %v", err)
	}
	if !on.Changed || !on.Enabled {
		t.Errorf("enable must converge, got %+v", on)
	}
	onAgain, err := EnsureEnabled("nginx", true)
	if err != nil {
		t.Fatalf("enable again: %v", err)
	}
	if onAgain.Changed {
		t.Error("already-enabled must be changed=false")
	}

	off, err := EnsureEnabled("nginx", false)
	if err != nil {
		t.Fatalf("disable: %v", err)
	}
	if !off.Changed || off.Enabled {
		t.Errorf("disable must converge, got %+v", off)
	}
	offAgain, err := EnsureEnabled("nginx", false)
	if err != nil {
		t.Fatalf("disable again: %v", err)
	}
	if offAgain.Changed {
		t.Error("already-disabled must be changed=false")
	}
	if all := mutations(); len(all) != 2 {
		t.Errorf("expected exactly [enable disable], got %v", all)
	}
}

func TestEnsureEnabledRejectsEmptyName(t *testing.T) {
	if _, err := EnsureEnabled("", true); err == nil {
		t.Fatal("empty name must be rejected")
	}
}
