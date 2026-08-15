package runner

import (
	"fmt"

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

// Get retrieves an operation from the registry.
func (r *Registry) Get(name string) (OperationFunc, bool) {
	fn, ok := r.ops[name]
	return fn, ok
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
		path := getStringArg(args, "path", "/")
		return sys.GetDiskUsage(path)
	})
	r.Register("sys.disk.partitions", func(_ map[string]interface{}) (interface{}, error) {
		return sys.GetDiskPartitions()
	})
	r.Register("sys.host.info", func(_ map[string]interface{}) (interface{}, error) {
		return sys.GetHostInfo()
	})
	r.Register("sys.load.avg", func(_ map[string]interface{}) (interface{}, error) {
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
		path := requireStringArg(args, "path")
		return file.Read(path)
	})
	r.Register("file.write", func(args map[string]interface{}) (interface{}, error) {
		path := requireStringArg(args, "path")
		content := requireStringArg(args, "content")
		return file.Write(path, content)
	})
	r.Register("file.exists", func(args map[string]interface{}) (interface{}, error) {
		path := requireStringArg(args, "path")
		return file.Exists(path)
	})
	r.Register("file.copy", func(args map[string]interface{}) (interface{}, error) {
		src := requireStringArg(args, "src")
		dst := requireStringArg(args, "dst")
		return file.Copy(src, dst)
	})
	r.Register("file.move", func(args map[string]interface{}) (interface{}, error) {
		src := requireStringArg(args, "src")
		dst := requireStringArg(args, "dst")
		return file.Move(src, dst)
	})
	r.Register("file.delete", func(args map[string]interface{}) (interface{}, error) {
		path := requireStringArg(args, "path")
		return file.Delete(path)
	})
	r.Register("file.info", func(args map[string]interface{}) (interface{}, error) {
		path := requireStringArg(args, "path")
		return file.Stat(path)
	})
	r.Register("file.list", func(args map[string]interface{}) (interface{}, error) {
		dir := requireStringArg(args, "dir")
		return file.List(dir)
	})
	r.Register("file.mkdir", func(args map[string]interface{}) (interface{}, error) {
		path := requireStringArg(args, "path")
		return file.Mkdir(path)
	})
	r.Register("file.checksum", func(args map[string]interface{}) (interface{}, error) {
		path := requireStringArg(args, "path")
		algo := getStringArg(args, "algo", "sha256")
		return file.Checksum(path, algo)
	})
}

// ============================================================
// net operations
// ============================================================

func (r *Registry) registerNetOps() {
	r.Register("net.http.get", func(args map[string]interface{}) (interface{}, error) {
		url := requireStringArg(args, "url")
		return opsnet.HTTPGet(url)
	})
	r.Register("net.http.post", func(args map[string]interface{}) (interface{}, error) {
		url := requireStringArg(args, "url")
		body := requireStringArg(args, "body")
		return opsnet.HTTPPost(url, body)
	})
	r.Register("net.tcp.ping", func(args map[string]interface{}) (interface{}, error) {
		host := requireStringArg(args, "host")
		port := requireIntArg(args, "port")
		return opsnet.TCPConnect(host, port)
	})
	r.Register("net.dns.resolve", func(args map[string]interface{}) (interface{}, error) {
		host := requireStringArg(args, "host")
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
	r.Register("process.find.by_name", func(args map[string]interface{}) (interface{}, error) {
		name := requireStringArg(args, "name")
		return process.FindByName(name)
	})
	r.Register("process.find.by_port", func(args map[string]interface{}) (interface{}, error) {
		port := requireIntArg(args, "port")
		return process.FindByPort(port)
	})
	r.Register("process.exec", func(args map[string]interface{}) (interface{}, error) {
		command := requireStringArg(args, "command")
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
	r.Register("service.status", func(args map[string]interface{}) (interface{}, error) {
		name := requireStringArg(args, "name")
		return service.Status(name)
	})
	r.Register("service.start", func(args map[string]interface{}) (interface{}, error) {
		name := requireStringArg(args, "name")
		return service.Start(name)
	})
	r.Register("service.stop", func(args map[string]interface{}) (interface{}, error) {
		name := requireStringArg(args, "name")
		return service.Stop(name)
	})
	r.Register("service.restart", func(args map[string]interface{}) (interface{}, error) {
		name := requireStringArg(args, "name")
		return service.Restart(name)
	})
	r.Register("service.enable", func(args map[string]interface{}) (interface{}, error) {
		name := requireStringArg(args, "name")
		return service.Enable(name)
	})
	r.Register("service.disable", func(args map[string]interface{}) (interface{}, error) {
		name := requireStringArg(args, "name")
		return service.Disable(name)
	})
}

// ============================================================
// pkg operations
// ============================================================

func (r *Registry) registerPkgOps() {
	r.Register("pkg.install", func(args map[string]interface{}) (interface{}, error) {
		name := requireStringArg(args, "name")
		return opspkg.Install(name)
	})
	r.Register("pkg.remove", func(args map[string]interface{}) (interface{}, error) {
		name := requireStringArg(args, "name")
		return opspkg.Remove(name)
	})
	r.Register("pkg.search", func(args map[string]interface{}) (interface{}, error) {
		// pkg.Info is the closest to "search" available in the SDK
		name := requireStringArg(args, "name")
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
		layout := requireStringArg(args, "layout")
		unix := requireInt64Arg(args, "ts")
		return optime.Format(unix, layout)
	})
	r.Register("time.parse", func(args map[string]interface{}) (interface{}, error) {
		layout := requireStringArg(args, "layout")
		value := requireStringArg(args, "value")
		return optime.Parse(layout, value)
	})
	r.Register("time.diff", func(args map[string]interface{}) (interface{}, error) {
		t1 := requireInt64Arg(args, "t1")
		t2 := requireInt64Arg(args, "t2")
		return optime.Diff(t1, t2), nil
	})
	r.Register("time.sleep", func(args map[string]interface{}) (interface{}, error) {
		ms := requireIntArg(args, "ms")
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
			return nil, fmt.Errorf("data is required")
		}
		return opsjson.Encode(data)
	})
	r.Register("json.decode", func(args map[string]interface{}) (interface{}, error) {
		input := requireStringArg(args, "input")
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
			return nil, fmt.Errorf("data is required")
		}
		return opsyaml.Encode(data)
	})
	r.Register("yaml.decode", func(args map[string]interface{}) (interface{}, error) {
		input := requireStringArg(args, "input")
		return opsyaml.Decode(input)
	})
}

// ============================================================
// built-in operations (report, log)
// ============================================================

func (r *Registry) registerBuiltinOps() {
	// report: collects named variables into the output data.
	// Each arg value is treated as a variable name; the variable's value
	// is looked up from the executor context during execution.
	r.Register("report", func(args map[string]interface{}) (interface{}, error) {
		result := make(map[string]interface{})
		for key, value := range args {
			result[key] = value
		}
		return result, nil
	})

	// log: outputs a message to warnings.
	r.Register("log", func(args map[string]interface{}) (interface{}, error) {
		msg := getStringArg(args, "message", "")
		if msg == "" {
			msg = getStringArg(args, "msg", "")
		}
		return msg, nil
	})
}

// ============================================================
// Argument helper functions
// ============================================================

// getStringArg returns the string value of an arg, or the default if missing/not a string.
func getStringArg(args map[string]interface{}, key string, defaultVal string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultVal
}

// requireStringArg returns the string value of an arg, or an error if missing/not a string.
func requireStringArg(args map[string]interface{}, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
		return fmt.Sprintf("%v", v)
	}
	return ""
}

// requireIntArg returns an arg as int, supporting both float64 (JSON numbers) and int.
func requireIntArg(args map[string]interface{}, key string) int {
	v, ok := args[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	default:
		return 0
	}
}

// requireInt64Arg returns an arg as int64, supporting both float64 (JSON numbers) and int.
func requireInt64Arg(args map[string]interface{}, key string) int64 {
	v, ok := args[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int:
		return int64(n)
	case int64:
		return n
	default:
		return 0
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
	for i, inst := range pkg.Instructions {
		if inst.Op == "" {
			return fmt.Errorf("instruction %d: op is required", i)
		}
	}
	return nil
}
