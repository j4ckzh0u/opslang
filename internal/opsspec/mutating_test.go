package opsspec_test

import (
	"testing"

	"github.com/opslang/opslang/internal/opsspec"
)

// The mutating flag is the metadata every engine's privilege enforcement
// derives from, so it must be complete (every canonical name and alias is
// classifiable) and explicit about the boundary calls.
func TestMutatingClassificationComplete(t *testing.T) {
	for _, f := range opsspec.Funcs {
		if _, known := opsspec.Mutating(f.Name); !known {
			t.Errorf("Mutating(%q) reports unknown for a canonical function", f.Name)
		}
	}
	for alias := range opsspec.Aliases {
		if _, known := opsspec.Mutating(alias); !known {
			t.Errorf("Mutating(%q) reports unknown for a registered alias", alias)
		}
	}
	for _, op := range opsspec.BuiltinOps {
		if _, known := opsspec.Mutating(op); !known {
			t.Errorf("Mutating(%q) reports unknown for a builtin op", op)
		}
	}
	// Unknown names must report unknown so engines skip enforcement for
	// custom builtins instead of blocking them.
	if _, known := opsspec.Mutating("definitely.not.an.op"); known {
		t.Error("Mutating must report unknown for names outside the table")
	}
}

func TestMutatingSet(t *testing.T) {
	wantMutating := map[string]bool{
		// file writes
		"file.write": true, "file.append": true, "file.copy": true,
		"file.move": true, "file.delete": true, "file.mkdir": true,
		"file.chmod": true, "file.distribute": true, "file.collect": true,
		// net
		"net.http_post": true,
		// process
		"process.exec": true, "process.kill": true,
		// service
		"service.start": true, "service.stop": true, "service.restart": true,
		"service.enable": true, "service.disable": true,
		// pkg
		"pkg.install": true, "pkg.remove": true,
		// builtin VM op running an arbitrary binary
		"binary.exec": true,
	}
	wantReadOnly := map[string]bool{
		"sys.cpu.usage": true, "sys.users": true, "sys.disk.usage": true,
		"file.read": true, "file.exists": true, "file.stat": true,
		"file.list": true, "file.checksum": true,
		// file.template only renders text; it never writes a file.
		"file.template": true,
		"net.http_get":  true, "net.tcp_check": true, "net.dns_lookup": true,
		"net.interfaces": true,
		"process.list":   true, "process.find_by_name": true, "process.find_by_port": true,
		"service.status": true, "pkg.info": true, "pkg.list": true,
		"time.now": true, "time.sleep": true,
		"json.encode": true, "yaml.decode": true,
		"log": true, "alert": true, "set": true, "report": true,
	}

	for name := range wantMutating {
		mutating, known := opsspec.Mutating(name)
		if !known || !mutating {
			t.Errorf("Mutating(%q) = (%v, %v), want (true, true)", name, mutating, known)
		}
	}
	for name := range wantReadOnly {
		mutating, known := opsspec.Mutating(name)
		if !known || mutating {
			t.Errorf("Mutating(%q) = (%v, %v), want (false, true)", name, mutating, known)
		}
	}

	// Aliases must resolve to the canonical classification.
	if mutating, _ := opsspec.Mutating("net.http.post"); !mutating {
		t.Error(`alias "net.http.post" must be mutating like canonical net.http_post`)
	}
	if mutating, _ := opsspec.Mutating("net.http.get"); mutating {
		t.Error(`alias "net.http.get" must be read-only like canonical net.http_get`)
	}
}
