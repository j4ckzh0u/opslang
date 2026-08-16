package runner

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/opslang/opslang/internal/ast"
	"github.com/opslang/opslang/internal/opsspec"
	"github.com/opslang/opslang/pkg/ops-core-sdk/file"
	opsjson "github.com/opslang/opslang/pkg/ops-core-sdk/json"
	opsnet "github.com/opslang/opslang/pkg/ops-core-sdk/net"
	opspkg "github.com/opslang/opslang/pkg/ops-core-sdk/pkg"
	"github.com/opslang/opslang/pkg/ops-core-sdk/process"
	"github.com/opslang/opslang/pkg/ops-core-sdk/service"
	"github.com/opslang/opslang/pkg/ops-core-sdk/sys"
	optime "github.com/opslang/opslang/pkg/ops-core-sdk/time"
	opsyaml "github.com/opslang/opslang/pkg/ops-core-sdk/yaml"
)

// Registry holds all registered operations and provides lookup and execution.
// Operation names follow the canonical table in internal/opsspec; historical
// aliases are resolved transparently at lookup time.
type Registry struct {
	ops map[string]OperationFunc
}

// NewRegistry creates a new registry with all standard operations registered.
func NewRegistry() *Registry {
	r := &Registry{
		ops: make(map[string]OperationFunc),
	}
	r.registerAll()
	return r
}

// Register adds an operation to the registry.
func (r *Registry) Register(name string, fn OperationFunc) {
	r.ops[name] = fn
}

// Get retrieves an operation from the registry, resolving canonical aliases.
func (r *Registry) Get(name string) (OperationFunc, bool) {
	if fn, ok := r.ops[name]; ok {
		return fn, true
	}
	if canonical, ok := opsspec.Aliases[name]; ok {
		fn, ok := r.ops[canonical]
		return fn, ok
	}
	return nil, false
}

// Has reports whether an operation (or its alias) is registered.
func (r *Registry) Has(name string) bool {
	_, ok := r.Get(name)
	return ok
}

// ListOperations returns the names of all registered operations.
func (r *Registry) ListOperations() []string {
	names := make([]string, 0, len(r.ops))
	for name := range r.ops {
		names = append(names, name)
	}
	return names
}

// registerAll registers all standard library operations grouped by SDK package.
func (r *Registry) registerAll() {
	r.registerSysOps()
	r.registerFileOps()
	r.registerNetOps()
	r.registerProcessOps()
	r.registerServiceOps()
	r.registerPkgOps()
	r.registerTimeOps()
	r.registerJSONOps()
	r.registerYAMLOps()
	r.registerBuiltinOps()
}

// ============================================================
// sys operations
// ============================================================

func (r *Registry) registerSysOps() {
	r.Register("sys.cpu.usage", func(_ map[string]interface{}) (interface{}, error) {
		return sys.GetCPUUsage()
	})
	r.Register("sys.cpu.count", func(_ map[string]interface{}) (interface{}, error) {
		return sys.GetCPUCount()
	})
	r.Register("sys.cpu.info", func(_ map[string]interface{}) (interface{}, error) {
		return sys.GetCPUInfo()
	})
	r.Register("sys.memory.info", func(_ map[string]interface{}) (interface{}, error) {
		return sys.GetMemoryInfo()
	})
	r.Register("sys.disk.usage", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("sys.disk.usage: %w", err)
		}
		return sys.GetDiskUsage(path)
	})
	r.Register("sys.disk.partitions", func(_ map[string]interface{}) (interface{}, error) {
		return sys.GetDiskPartitions()
	})
	r.Register("sys.os", func(_ map[string]interface{}) (interface{}, error) {
		return sys.GetHostInfo()
	})
	r.Register("sys.load", func(_ map[string]interface{}) (interface{}, error) {
		return sys.GetLoadAvg()
	})
	r.Register("sys.net.interfaces", func(_ map[string]interface{}) (interface{}, error) {
		return sys.GetNetInterfaces()
	})
	r.Register("sys.users", func(_ map[string]interface{}) (interface{}, error) {
		return sys.Users()
	})
	r.Register("sys.uptime", func(_ map[string]interface{}) (interface{}, error) {
		return sys.Uptime()
	})
	r.Register("sys.hostname", func(_ map[string]interface{}) (interface{}, error) {
		return sys.Hostname()
	})
}

// ============================================================
// file operations
// ============================================================

func (r *Registry) registerFileOps() {
	r.Register("file.read", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.read: %w", err)
		}
		return file.Read(path)
	})
	r.Register("file.write", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.write: %w", err)
		}
		content, err := argString(args, "content")
		if err != nil {
			return nil, fmt.Errorf("file.write: %w", err)
		}
		return file.Write(path, content)
	})
	r.Register("file.append", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.append: %w", err)
		}
		content, err := argString(args, "content")
		if err != nil {
			return nil, fmt.Errorf("file.append: %w", err)
		}
		return file.Append(path, content)
	})
	r.Register("file.exists", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.exists: %w", err)
		}
		return file.Exists(path)
	})
	r.Register("file.copy", func(args map[string]interface{}) (interface{}, error) {
		src, err := argString(args, "src")
		if err != nil {
			return nil, fmt.Errorf("file.copy: %w", err)
		}
		dst, err := argString(args, "dst")
		if err != nil {
			return nil, fmt.Errorf("file.copy: %w", err)
		}
		return file.Copy(src, dst)
	})
	r.Register("file.move", func(args map[string]interface{}) (interface{}, error) {
		src, err := argString(args, "src")
		if err != nil {
			return nil, fmt.Errorf("file.move: %w", err)
		}
		dst, err := argString(args, "dst")
		if err != nil {
			return nil, fmt.Errorf("file.move: %w", err)
		}
		return file.Move(src, dst)
	})
	r.Register("file.delete", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.delete: %w", err)
		}
		return file.Delete(path)
	})
	r.Register("file.stat", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.stat: %w", err)
		}
		return file.Stat(path)
	})
	r.Register("file.list", func(args map[string]interface{}) (interface{}, error) {
		dir, err := argString(args, "dir")
		if err != nil {
			return nil, fmt.Errorf("file.list: %w", err)
		}
		return file.List(dir)
	})
	r.Register("file.mkdir", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.mkdir: %w", err)
		}
		return file.Mkdir(path)
	})
	r.Register("file.chmod", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.chmod: %w", err)
		}
		modeStr, err := argString(args, "mode")
		if err != nil {
			return nil, fmt.Errorf("file.chmod: %w", err)
		}
		var mode uint64
		if _, err := fmt.Sscanf(modeStr, "%o", &mode); err != nil {
			return nil, fmt.Errorf("file.chmod: mode must be an octal string like \"0755\", got %q", modeStr)
		}
		return file.Chmod(path, uint32(mode))
	})
	r.Register("file.template", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.template: %w", err)
		}
		vars, _ := args["vars"].(map[string]interface{})
		return file.Template(path, vars)
	})
	r.Register("file.checksum", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.checksum: %w", err)
		}
		algo := getStringArg(args, "algo", "sha256")
		return file.Checksum(path, algo)
	})
}

// ============================================================
// net operations
// ============================================================

func (r *Registry) registerNetOps() {
	r.Register("net.http_get", func(args map[string]interface{}) (interface{}, error) {
		url, err := argString(args, "url")
		if err != nil {
			return nil, fmt.Errorf("net.http_get: %w", err)
		}
		return opsnet.HTTPGet(url)
	})
	r.Register("net.http_post", func(args map[string]interface{}) (interface{}, error) {
		url, err := argString(args, "url")
		if err != nil {
			return nil, fmt.Errorf("net.http_post: %w", err)
		}
		body, err := argString(args, "body")
		if err != nil {
			return nil, fmt.Errorf("net.http_post: %w", err)
		}
		return opsnet.HTTPPost(url, body)
	})
	r.Register("net.tcp_check", func(args map[string]interface{}) (interface{}, error) {
		host, err := argString(args, "host")
		if err != nil {
			return nil, fmt.Errorf("net.tcp_check: %w", err)
		}
		port, err := argInt(args, "port")
		if err != nil {
			return nil, fmt.Errorf("net.tcp_check: %w", err)
		}
		return opsnet.TCPConnect(host, port)
	})
	r.Register("net.dns_lookup", func(args map[string]interface{}) (interface{}, error) {
		host, err := argString(args, "host")
		if err != nil {
			return nil, fmt.Errorf("net.dns_lookup: %w", err)
		}
		return opsnet.DNSLookup(host)
	})
	r.Register("net.interfaces", func(_ map[string]interface{}) (interface{}, error) {
		return opsnet.Interfaces()
	})
}

// ============================================================
// process operations
// ============================================================

func (r *Registry) registerProcessOps() {
	r.Register("process.list", func(_ map[string]interface{}) (interface{}, error) {
		return process.List()
	})
	r.Register("process.find_by_name", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("process.find_by_name: %w", err)
		}
		return process.FindByName(name)
	})
	r.Register("process.find_by_port", func(args map[string]interface{}) (interface{}, error) {
		port, err := argInt(args, "port")
		if err != nil {
			return nil, fmt.Errorf("process.find_by_port: %w", err)
		}
		return process.FindByPort(port)
	})
	r.Register("process.kill", func(args map[string]interface{}) (interface{}, error) {
		pid, err := argInt(args, "pid")
		if err != nil {
			return nil, fmt.Errorf("process.kill: %w", err)
		}
		signal := getStringArg(args, "signal", "TERM")
		return process.Kill(pid, signal)
	})
	r.Register("process.exec", func(args map[string]interface{}) (interface{}, error) {
		command, err := argString(args, "command")
		if err != nil {
			return nil, fmt.Errorf("process.exec: %w", err)
		}
		var procArgs []string
		if a, ok := args["args"].([]interface{}); ok {
			for _, v := range a {
				if s, ok := v.(string); ok {
					procArgs = append(procArgs, s)
				}
			}
		}
		return process.Exec(command, procArgs)
	})
}

// ============================================================
// service operations
// ============================================================

func (r *Registry) registerServiceOps() {
	r.Register("service.status", serviceOp("service.status", service.Status))
	r.Register("service.start", serviceOp("service.start", service.Start))
	r.Register("service.stop", serviceOp("service.stop", service.Stop))
	r.Register("service.restart", serviceOp("service.restart", service.Restart))
	r.Register("service.enable", serviceOp("service.enable", service.Enable))
	r.Register("service.disable", serviceOp("service.disable", service.Disable))
}

// serviceOp adapts a service SDK function taking just a name.
func serviceOp[T any](opName string, fn func(string) (T, error)) OperationFunc {
	return func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("%s: %w", opName, err)
		}
		return fn(name)
	}
}

// ============================================================
// pkg operations
// ============================================================

func (r *Registry) registerPkgOps() {
	r.Register("pkg.install", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("pkg.install: %w", err)
		}
		return opspkg.Install(name)
	})
	r.Register("pkg.remove", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("pkg.remove: %w", err)
		}
		return opspkg.Remove(name)
	})
	r.Register("pkg.info", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("pkg.info: %w", err)
		}
		return opspkg.Info(name)
	})
	r.Register("pkg.list", func(_ map[string]interface{}) (interface{}, error) {
		return opspkg.List()
	})
}

// ============================================================
// time operations
// ============================================================

func (r *Registry) registerTimeOps() {
	r.Register("time.now", func(_ map[string]interface{}) (interface{}, error) {
		return optime.Now(), nil
	})
	r.Register("time.format", func(args map[string]interface{}) (interface{}, error) {
		ts, err := argInt64(args, "ts")
		if err != nil {
			return nil, fmt.Errorf("time.format: %w", err)
		}
		layout := getStringArg(args, "layout", "2006-01-02 15:04:05")
		return optime.Format(ts, layout)
	})
	r.Register("time.parse", func(args map[string]interface{}) (interface{}, error) {
		layout, err := argString(args, "layout")
		if err != nil {
			return nil, fmt.Errorf("time.parse: %w", err)
		}
		value, err := argString(args, "value")
		if err != nil {
			return nil, fmt.Errorf("time.parse: %w", err)
		}
		return optime.Parse(layout, value)
	})
	r.Register("time.diff", func(args map[string]interface{}) (interface{}, error) {
		t1, err := argInt64(args, "t1")
		if err != nil {
			return nil, fmt.Errorf("time.diff: %w", err)
		}
		t2, err := argInt64(args, "t2")
		if err != nil {
			return nil, fmt.Errorf("time.diff: %w", err)
		}
		return optime.Diff(t1, t2), nil
	})
	r.Register("time.since", func(args map[string]interface{}) (interface{}, error) {
		ts, err := argInt64(args, "ts")
		if err != nil {
			return nil, fmt.Errorf("time.since: %w", err)
		}
		return optime.Since(ts), nil
	})
	r.Register("time.sleep", func(args map[string]interface{}) (interface{}, error) {
		ms, err := argInt(args, "ms")
		if err != nil {
			return nil, fmt.Errorf("time.sleep: %w", err)
		}
		return optime.Sleep(ms)
	})
}

// ============================================================
// json operations
// ============================================================

func (r *Registry) registerJSONOps() {
	r.Register("json.encode", func(args map[string]interface{}) (interface{}, error) {
		data, ok := args["data"]
		if !ok {
			data, ok = args["value"]
			if !ok {
				return nil, fmt.Errorf("json.encode: argument \"value\" is required")
			}
		}
		return opsjson.Encode(data)
	})
	r.Register("json.decode", func(args map[string]interface{}) (interface{}, error) {
		input, err := argString(args, "input")
		if err != nil {
			return nil, fmt.Errorf("json.decode: %w", err)
		}
		return opsjson.Decode(input)
	})
}

// ============================================================
// yaml operations
// ============================================================

func (r *Registry) registerYAMLOps() {
	r.Register("yaml.encode", func(args map[string]interface{}) (interface{}, error) {
		data, ok := args["data"]
		if !ok {
			data, ok = args["value"]
			if !ok {
				return nil, fmt.Errorf("yaml.encode: argument \"value\" is required")
			}
		}
		return opsyaml.Encode(data)
	})
	r.Register("yaml.decode", func(args map[string]interface{}) (interface{}, error) {
		input, err := argString(args, "input")
		if err != nil {
			return nil, fmt.Errorf("yaml.decode: %w", err)
		}
		return opsyaml.Decode(input)
	})
}

// ============================================================
// built-in operations (log, alert, set, report, binary.exec)
// ============================================================

func (r *Registry) registerBuiltinOps() {
	// report: collects named values into the output data map. The executor
	// resolves "$name" variable references before it runs.
	r.Register("report", func(args map[string]interface{}) (interface{}, error) {
		result := make(map[string]interface{}, len(args))
		for key, value := range args {
			result[key] = value
		}
		return result, nil
	})

	// set: stores a value; the executor handles variable assignment.
	r.Register("set", func(args map[string]interface{}) (interface{}, error) {
		v, ok := args["value"]
		if !ok {
			return nil, fmt.Errorf("set: argument \"value\" is required")
		}
		return v, nil
	})

	r.Register("log", func(args map[string]interface{}) (interface{}, error) {
		return getStringArg(args, "message", ""), nil
	})

	r.Register("alert", func(args map[string]interface{}) (interface{}, error) {
		msg, err := argString(args, "message")
		if err != nil {
			return nil, fmt.Errorf("alert: %w", err)
		}
		return msg, nil
	})

	// binary.exec: executes a compiled OpsLang binary (AOT deploy mode) and
	// returns its parsed JSON output. A non-zero exit or startup failure is
	// an ERROR, not a result — deployment must not report success when the
	// program failed.
	r.Register("binary.exec", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("binary.exec: %w", err)
		}

		var execArgs []string
		if a, ok := args["args"].([]interface{}); ok {
			for _, v := range a {
				if s, ok := v.(string); ok {
					execArgs = append(execArgs, s)
				}
			}
		}

		cmd := exec.Command(path, execArgs...)
		output, err := cmd.Output()
		if err != nil {
			stderr := ""
			if exitErr, ok := err.(*exec.ExitError); ok {
				stderr = string(exitErr.Stderr)
			}
			return nil, fmt.Errorf("binary.exec: %s failed: %v%s", path, err, stderr)
		}

		// The AOT binary prints its report as JSON on stdout.
		var result interface{}
		if jsonErr := json.Unmarshal(output, &result); jsonErr == nil {
			return result, nil
		}
		return map[string]interface{}{
			"output": string(output),
		}, nil
	})
}

// ============================================================
// Argument helper functions
// ============================================================

// getStringArg returns the string value of an optional arg, or defaultVal.
func getStringArg(args map[string]interface{}, key string, defaultVal string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultVal
}

// argString returns a required string arg or a descriptive error. Silent
// empty-string fallbacks masked broken instruction packages in the past;
// missing required arguments must fail loudly.
func argString(args map[string]interface{}, key string) (string, error) {
	v, ok := args[key]
	if !ok {
		return "", fmt.Errorf("missing required argument %q", key)
	}
	switch s := v.(type) {
	case string:
		return s, nil
	default:
		return "", fmt.Errorf("argument %q must be a string, got %T", key, v)
	}
}

// argInt returns a required integer arg, accepting the numeric types that
// JSON unmarshalling can produce (float64) as well as native ints.
func argInt(args map[string]interface{}, key string) (int, error) {
	v, ok := args[key]
	if !ok {
		return 0, fmt.Errorf("missing required argument %q", key)
	}
	switch n := v.(type) {
	case float64:
		return int(n), nil
	case int:
		return n, nil
	case int64:
		return int(n), nil
	default:
		return 0, fmt.Errorf("argument %q must be a number, got %T", key, v)
	}
}

// argInt64 is argInt with an int64 result.
func argInt64(args map[string]interface{}, key string) (int64, error) {
	v, ok := args[key]
	if !ok {
		return 0, fmt.Errorf("missing required argument %q", key)
	}
	switch n := v.(type) {
	case float64:
		return int64(n), nil
	case int:
		return int64(n), nil
	case int64:
		return n, nil
	default:
		return 0, fmt.Errorf("argument %q must be a number, got %T", key, v)
	}
}

// ValidatePackage checks if an instruction package is valid.
func ValidatePackage(pkg *InstructionPackage) error {
	if pkg.Version == "" {
		return fmt.Errorf("version is required")
	}
	if pkg.Version != "1.0" {
		return fmt.Errorf("unsupported version: %s", pkg.Version)
	}
	if len(pkg.Instructions) == 0 {
		return fmt.Errorf("at least one instruction is required")
	}
	// The privilege field is optional (legacy packages omit it), but when
	// present it must be a declared level so the runner's second check
	// cannot be silently downgraded by a typo.
	switch pkg.Privilege {
	case "", string(ast.PrivilegeReadOnly), string(ast.PrivilegeAdmin), string(ast.PrivilegeRoot):
	default:
		return fmt.Errorf("invalid privilege %q (expected read_only, admin, or root)", pkg.Privilege)
	}
	registry := NewRegistry()
	for i, inst := range pkg.Instructions {
		if inst.Op == "" {
			return fmt.Errorf("instruction %d: op is required", i)
		}
		if !registry.Has(inst.Op) {
			return fmt.Errorf("instruction %d: unknown operation %q", i, inst.Op)
		}
	}
	return nil
}
