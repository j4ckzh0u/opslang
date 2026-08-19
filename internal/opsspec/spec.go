// Package opsspec defines the canonical registry of every OpsLang atomic
// operation: its dotted DSL name, its positional argument names, and where
// it can run. This table is the single source of truth shared by the three
// execution engines (interpreter, runner registry, AOT code generator).
// A cross-engine consistency test enforces that all engines agree with it.
package opsspec

// Availability describes which engines expose a function.
type Availability int

const (
	// All means the function is available in the interpreter (controller),
	// the remote runner registry, and the AOT code generator.
	All Availability = iota
	// ControllerOnly means the function only makes sense on the controller
	// machine (e.g. it fans out over SSH itself); the remote runner and the
	// AOT code generator must not expose it.
	ControllerOnly
)

// Func describes one atomic operation.
type Func struct {
	// Name is the canonical dotted DSL name, e.g. "sys.cpu.usage".
	// All engines must resolve exactly this name; historical aliases are
	// tolerated at lookup time but never generated.
	Name string
	// Args lists positional argument names in call order, e.g. ["path"].
	// Engines use these names when building instruction argument maps.
	Args []string
	// Avail restricts which engines expose the function.
	Avail Availability
	// Mutating marks functions that change system or remote state (files,
	// processes, services, packages, or the state of a remote HTTP
	// endpoint). Mutating functions require at least admin privilege;
	// non-mutating ones are available to every privilege level. This flag
	// is the single source of truth shared by the interpreter, the
	// instruction generator, the AOT compiler and the runner — engines
	// must not keep their own mutation lists.
	Mutating bool
}

// Funcs is the canonical table. Keep sorted by category then name.
var Funcs = []Func{
	// ── cron ──────────────────────────────────────────────────────────
	{Name: "cron.add", Args: []string{"user", "entry"}, Mutating: true},
	{Name: "cron.list", Args: []string{"user"}},
	{Name: "cron.remove", Args: []string{"user", "line_match"}, Mutating: true},

	// ── file ──────────────────────────────────────────────────────────
	{Name: "file.append", Args: []string{"path", "content"}, Mutating: true},
	{Name: "file.checksum", Args: []string{"path", "algo"}},
	{Name: "file.chmod", Args: []string{"path", "mode"}, Mutating: true},
	{Name: "file.collect", Args: []string{"source", "targets", "options"}, Avail: ControllerOnly, Mutating: true},
	{Name: "file.copy", Args: []string{"src", "dst"}, Mutating: true},
	{Name: "file.delete", Args: []string{"path"}, Mutating: true},
	// distribute/collect fan out from the controller over SSH; a remote
	// runner executing them would need controller credentials.
	{Name: "file.distribute", Args: []string{"source", "targets", "options"}, Avail: ControllerOnly, Mutating: true},
	{Name: "file.exists", Args: []string{"path"}},
	{Name: "file.lineinfile", Args: []string{"path", "line", "present", "regexp"}, Mutating: true},
	{Name: "file.list", Args: []string{"dir"}},
	{Name: "file.mkdir", Args: []string{"path"}, Mutating: true},
	{Name: "file.move", Args: []string{"src", "dst"}, Mutating: true},
	{Name: "file.read", Args: []string{"path"}},
	{Name: "file.stat", Args: []string{"path"}},
	// file.template only READS the template and returns the rendered text;
	// it never writes a file, so it is not mutating.
	{Name: "file.template", Args: []string{"path", "vars"}},
	{Name: "file.write", Args: []string{"path", "content"}, Mutating: true},

	// ── firewall ──────────────────────────────────────────────────────
	{Name: "firewall.rule", Args: []string{"action", "protocol", "port", "source"}, Mutating: true},

	// ── git ───────────────────────────────────────────────────────────
	{Name: "git.clone", Args: []string{"url", "dest", "opts"}, Mutating: true},
	{Name: "git.pull", Args: []string{"repo_path", "remote", "branch"}, Mutating: true},

	// ── group ─────────────────────────────────────────────────────────
	{Name: "group.add", Args: []string{"name", "opts"}, Mutating: true},
	{Name: "group.exists", Args: []string{"name"}},
	{Name: "group.info", Args: []string{"name"}},
	{Name: "group.list"},
	{Name: "group.remove", Args: []string{"name"}, Mutating: true},

	// ── json ──────────────────────────────────────────────────────────
	{Name: "json.decode", Args: []string{"input"}},
	{Name: "json.encode", Args: []string{"value"}},

	// ── net ───────────────────────────────────────────────────────────
	{Name: "net.dns_lookup", Args: []string{"host"}},
	{Name: "net.http_get", Args: []string{"url"}},
	// net.http_post is classified as mutating even though it only changes
	// REMOTE state: a POST submits data (deploys, webhooks, form
	// submissions) and is the writing counterpart to the read-only GET.
	{Name: "net.http_post", Args: []string{"url", "body"}, Mutating: true},
	{Name: "net.interfaces"},
	{Name: "net.tcp_check", Args: []string{"host", "port"}},
	{Name: "net.wait_for", Args: []string{"host", "port", "timeout"}},

	// ── pkg ───────────────────────────────────────────────────────────
	{Name: "pkg.info", Args: []string{"name"}},
	{Name: "pkg.install", Args: []string{"name"}, Mutating: true},
	{Name: "pkg.list"},
	{Name: "pkg.remove", Args: []string{"name"}, Mutating: true},

	// ── process ───────────────────────────────────────────────────────
	{Name: "process.exec", Args: []string{"command", "args"}, Mutating: true},
	{Name: "process.find_by_name", Args: []string{"name"}},
	{Name: "process.find_by_port", Args: []string{"port"}},
	{Name: "process.kill", Args: []string{"pid", "signal"}, Mutating: true},
	{Name: "process.list"},

	// ── service ───────────────────────────────────────────────────────
	{Name: "service.disable", Args: []string{"name"}, Mutating: true},
	{Name: "service.enable", Args: []string{"name"}, Mutating: true},
	{Name: "service.restart", Args: []string{"name"}, Mutating: true},
	{Name: "service.start", Args: []string{"name"}, Mutating: true},
	{Name: "service.status", Args: []string{"name"}},
	{Name: "service.stop", Args: []string{"name"}, Mutating: true},

	// ── sys ───────────────────────────────────────────────────────────
	{Name: "sys.cpu.count"},
	{Name: "sys.cpu.info"},
	{Name: "sys.cpu.usage"},
	{Name: "sys.disk.partitions"},
	{Name: "sys.disk.usage", Args: []string{"path"}},
	{Name: "sys.hostname"},
	{Name: "sys.hostname_set", Args: []string{"name"}, Mutating: true},
	{Name: "sys.list_mounts"},
	{Name: "sys.load"},
	{Name: "sys.memory.info"},
	{Name: "sys.mount", Args: []string{"device", "mountpoint", "fs_type", "opts"}, Mutating: true},
	{Name: "sys.net.interfaces"},
	{Name: "sys.os"},
	{Name: "sys.unmount", Args: []string{"mountpoint"}, Mutating: true},
	{Name: "sys.uptime"},
	{Name: "sys.users"},

	// ── sysctl ────────────────────────────────────────────────────────
	{Name: "sysctl.get", Args: []string{"name"}},
	{Name: "sysctl.list"},
	{Name: "sysctl.set", Args: []string{"name", "value"}, Mutating: true},

	// ── time ──────────────────────────────────────────────────────────
	{Name: "time.diff", Args: []string{"t1", "t2"}},
	{Name: "time.format", Args: []string{"ts", "layout"}},
	{Name: "time.now"},
	{Name: "time.parse", Args: []string{"layout", "value"}},
	{Name: "time.since", Args: []string{"ts"}},
	{Name: "time.sleep", Args: []string{"ms"}},

	// ── user ──────────────────────────────────────────────────────────
	{Name: "user.add", Args: []string{"username", "opts"}, Mutating: true},
	{Name: "user.exists", Args: []string{"username"}},
	{Name: "user.info", Args: []string{"username"}},
	{Name: "user.list"},
	{Name: "user.modify", Args: []string{"username", "opts"}, Mutating: true},
	{Name: "user.remove", Args: []string{"username", "remove_home"}, Mutating: true},

	// ── yaml ──────────────────────────────────────────────────────────
	{Name: "yaml.decode", Args: []string{"input"}},
	{Name: "yaml.encode", Args: []string{"value"}},
}

// BuiltinOps are runner instruction ops that are not SDK calls.
var BuiltinOps = []string{"log", "alert", "set", "report", "binary.exec"}

// byName indexes Funcs by canonical name.
var byName = func() map[string]Func {
	m := make(map[string]Func, len(Funcs))
	for _, f := range Funcs {
		m[f.Name] = f
	}
	return m
}()

// Lookup returns the spec for a canonical function name.
func Lookup(name string) (Func, bool) {
	f, ok := byName[name]
	return f, ok
}

// Mutating reports whether an operation (SDK call or builtin VM op, with
// historical aliases resolved) changes system or remote state. known is
// false for names outside the canonical table and BuiltinOps — callers
// must skip privilege enforcement for them (they are custom builtins, not
// OpsLang operations). binary.exec counts as mutating: it runs an
// arbitrary binary, which can change anything.
func Mutating(op string) (isMutating, known bool) {
	if canonical, isAlias := Aliases[op]; isAlias {
		op = canonical
	}
	if f, ok := byName[op]; ok {
		return f.Mutating, true
	}
	switch op {
	case "binary.exec":
		return true, true
	case "log", "alert", "set", "report":
		return false, true // output/bookkeeping ops only
	}
	return false, false
}

// ArgNames returns the positional argument names for an op (SDK call or
// builtin). The second return is false when the op is unknown, in which
// case callers must reject the call rather than guess.
func ArgNames(op string) ([]string, bool) {
	if f, ok := byName[op]; ok {
		return f.Args, true
	}
	for _, b := range BuiltinOps {
		if op == b {
			switch op {
			case "log", "alert":
				return []string{"message"}, true
			case "set":
				return []string{"value"}, true
			case "binary.exec":
				return []string{"path", "args"}, true
			}
		}
	}
	return nil, false
}

// Names returns all canonical function names, optionally filtered by
// availability.
func Names(avail *Availability) []string {
	out := make([]string, 0, len(Funcs))
	for _, f := range Funcs {
		if avail != nil && f.Avail != *avail {
			continue
		}
		out = append(out, f.Name)
	}
	return out
}

// Aliases maps historical runner-registry names to canonical names. Engines
// accept these at lookup time for backward compatibility with existing
// instruction packages, but generators only emit canonical names.
var Aliases = map[string]string{
	"sys.load.avg":         "sys.load",
	"sys.host.info":        "sys.os",
	"net.http.get":         "net.http_get",
	"net.http.post":        "net.http_post",
	"net.tcp.ping":         "net.tcp_check",
	"net.dns.resolve":      "net.dns_lookup",
	"process.find.by_name": "process.find_by_name",
	"process.find.by_port": "process.find_by_port",
	"file.info":            "file.stat",
	"pkg.search":           "pkg.info",
}
