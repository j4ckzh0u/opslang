package runner

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/opslang/opslang/internal/ast"
	"github.com/opslang/opslang/internal/opsspec"
	sdkarchive "github.com/opslang/opslang/pkg/ops-core-sdk/archive"
	opscron "github.com/opslang/opslang/pkg/ops-core-sdk/cron"
	sdkdisk "github.com/opslang/opslang/pkg/ops-core-sdk/disk"
	sdkdocker "github.com/opslang/opslang/pkg/ops-core-sdk/docker"
	"github.com/opslang/opslang/pkg/ops-core-sdk/file"
	opsgit "github.com/opslang/opslang/pkg/ops-core-sdk/git"
	opsgrp "github.com/opslang/opslang/pkg/ops-core-sdk/group"
	opshosts "github.com/opslang/opslang/pkg/ops-core-sdk/hosts"
	opsjson "github.com/opslang/opslang/pkg/ops-core-sdk/json"
	sdkkernel "github.com/opslang/opslang/pkg/ops-core-sdk/kernel"
	sdklocale "github.com/opslang/opslang/pkg/ops-core-sdk/locale"
	opsnet "github.com/opslang/opslang/pkg/ops-core-sdk/net"
	sdkpip "github.com/opslang/opslang/pkg/ops-core-sdk/pip"
	opspkg "github.com/opslang/opslang/pkg/ops-core-sdk/pkg"
	"github.com/opslang/opslang/pkg/ops-core-sdk/process"
	"github.com/opslang/opslang/pkg/ops-core-sdk/service"
	sdkssh "github.com/opslang/opslang/pkg/ops-core-sdk/ssh"
	"github.com/opslang/opslang/pkg/ops-core-sdk/sys"
	sdksysctl "github.com/opslang/opslang/pkg/ops-core-sdk/sysctl"
	optime "github.com/opslang/opslang/pkg/ops-core-sdk/time"
	opsuser "github.com/opslang/opslang/pkg/ops-core-sdk/user"
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
	r.registerGitOps()
	r.registerBuiltinOps()
	r.registerPlatformOps()
	r.registerExtensions()
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
	r.Register("file.lineinfile", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.lineinfile: %w", err)
		}
		line, err := argString(args, "line")
		if err != nil {
			return nil, fmt.Errorf("file.lineinfile: %w", err)
		}
		present, err := argBool(args, "present")
		if err != nil {
			return nil, fmt.Errorf("file.lineinfile: %w", err)
		}
		rx := getStringArg(args, "regexp", "")
		return file.LineInFile(path, line, present, rx)
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
	r.Register("net.wait_for", func(args map[string]interface{}) (interface{}, error) {
		host, err := argString(args, "host")
		if err != nil {
			return nil, fmt.Errorf("net.wait_for: %w", err)
		}
		port, err := argInt(args, "port")
		if err != nil {
			return nil, fmt.Errorf("net.wait_for: %w", err)
		}
		timeout, err := argInt(args, "timeout")
		if err != nil {
			return nil, fmt.Errorf("net.wait_for: %w", err)
		}
		return opsnet.WaitFor(host, port, timeout)
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
		r, _ := opspkg.Install(name)
		return r, nil
	})
	r.Register("pkg.remove", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("pkg.remove: %w", err)
		}
		r, _ := opspkg.Remove(name)
		return r, nil
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
// git operations
// ============================================================

func (r *Registry) registerGitOps() {
	r.Register("git.clone", func(args map[string]interface{}) (interface{}, error) {
		url, err := argString(args, "url")
		if err != nil {
			return nil, fmt.Errorf("git.clone: %w", err)
		}
		dest, err := argString(args, "dest")
		if err != nil {
			return nil, fmt.Errorf("git.clone: %w", err)
		}
		var opts map[string]string
		if raw, ok := args["opts"]; ok && raw != nil {
			if m, ok := raw.(map[string]interface{}); ok {
				opts = make(map[string]string, len(m))
				for k, v := range m {
					opts[k] = fmt.Sprintf("%v", v)
				}
			}
		}
		return opsgit.Clone(url, dest, opts)
	})
	r.Register("git.pull", func(args map[string]interface{}) (interface{}, error) {
		repoPath, err := argString(args, "repo_path")
		if err != nil {
			return nil, fmt.Errorf("git.pull: %w", err)
		}
		remote := getStringArg(args, "remote", "")
		branch := getStringArg(args, "branch", "")
		return opsgit.Pull(repoPath, remote, branch)
	})
}

// ============================================================
// platform operations (user, group, cron, sysctl, mount, firewall, etc.)
// ============================================================

func (r *Registry) registerPlatformOps() {
	// ── user.* ─────────────────────────────────────────────────────────
	r.Register("user.info", func(args map[string]interface{}) (interface{}, error) {
		username, err := argString(args, "username")
		if err != nil {
			return nil, fmt.Errorf("user.info: %w", err)
		}
		return opsuser.Info(username)
	})
	r.Register("user.list", func(_ map[string]interface{}) (interface{}, error) {
		return opsuser.List()
	})
	r.Register("user.add", func(args map[string]interface{}) (interface{}, error) {
		username, err := argString(args, "username")
		if err != nil {
			return nil, fmt.Errorf("user.add: %w", err)
		}
		opts := toStringMapArg(args, "opts")
		return opsuser.Add(username, opts)
	})
	r.Register("user.remove", func(args map[string]interface{}) (interface{}, error) {
		username, err := argString(args, "username")
		if err != nil {
			return nil, fmt.Errorf("user.remove: %w", err)
		}
		removeHome, _ := args["remove_home"].(bool)
		return opsuser.Remove(username, removeHome)
	})
	r.Register("user.modify", func(args map[string]interface{}) (interface{}, error) {
		username, err := argString(args, "username")
		if err != nil {
			return nil, fmt.Errorf("user.modify: %w", err)
		}
		opts := toStringMapArg(args, "opts")
		return opsuser.Modify(username, opts)
	})
	r.Register("user.exists", func(args map[string]interface{}) (interface{}, error) {
		username, err := argString(args, "username")
		if err != nil {
			return nil, fmt.Errorf("user.exists: %w", err)
		}
		return opsuser.Exists(username)
	})

	// ── group.* ────────────────────────────────────────────────────────
	r.Register("group.info", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("group.info: %w", err)
		}
		return opsgrp.Info(name)
	})
	r.Register("group.list", func(_ map[string]interface{}) (interface{}, error) {
		return opsgrp.List()
	})
	r.Register("group.add", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("group.add: %w", err)
		}
		opts := toStringMapArg(args, "opts")
		return opsgrp.Add(name, opts)
	})
	r.Register("group.remove", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("group.remove: %w", err)
		}
		return opsgrp.Remove(name)
	})
	r.Register("group.exists", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("group.exists: %w", err)
		}
		return opsgrp.Exists(name)
	})

	// ── cron.* ─────────────────────────────────────────────────────────
	r.Register("cron.list", func(args map[string]interface{}) (interface{}, error) {
		user, err := argString(args, "user")
		if err != nil {
			return nil, fmt.Errorf("cron.list: %w", err)
		}
		return opscron.List(user)
	})
	r.Register("cron.add", func(args map[string]interface{}) (interface{}, error) {
		user, err := argString(args, "user")
		if err != nil {
			return nil, fmt.Errorf("cron.add: %w", err)
		}
		entryMap, _ := args["entry"].(map[string]interface{})
		entry := opscron.CronEntry{
			Minute:     mapStrArg(entryMap, "minute", "*"),
			Hour:       mapStrArg(entryMap, "hour", "*"),
			DayOfMonth: mapStrArg(entryMap, "day_of_month", "*"),
			Month:      mapStrArg(entryMap, "month", "*"),
			DayOfWeek:  mapStrArg(entryMap, "day_of_week", "*"),
			Command:    mapStrArg(entryMap, "command", ""),
		}
		return opscron.Add(user, entry)
	})
	r.Register("cron.remove", func(args map[string]interface{}) (interface{}, error) {
		user, err := argString(args, "user")
		if err != nil {
			return nil, fmt.Errorf("cron.remove: %w", err)
		}
		lineMatch, err := argString(args, "line_match")
		if err != nil {
			return nil, fmt.Errorf("cron.remove: %w", err)
		}
		return opscron.Remove(user, lineMatch)
	})

	// ── sysctl.* ───────────────────────────────────────────────────────
	r.Register("sysctl.get", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("sysctl.get: %w", err)
		}
		return sdksysctl.Get(name)
	})
	r.Register("sysctl.set", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("sysctl.set: %w", err)
		}
		value, err := argString(args, "value")
		if err != nil {
			return nil, fmt.Errorf("sysctl.set: %w", err)
		}
		return sdksysctl.Set(name, value)
	})
	r.Register("sysctl.list", func(_ map[string]interface{}) (interface{}, error) {
		return sdksysctl.List()
	})

	// ── sys.mount / sys.unmount / sys.list_mounts ──────────────────────
	r.Register("sys.mount", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("sys.mount: %w", err)
		}
		mountpoint, err := argString(args, "mountpoint")
		if err != nil {
			return nil, fmt.Errorf("sys.mount: %w", err)
		}
		fsType := getStringArg(args, "fs_type", "")
		opts := toStringMapArg(args, "opts")
		return sys.Mount(device, mountpoint, fsType, opts)
	})
	r.Register("sys.unmount", func(args map[string]interface{}) (interface{}, error) {
		mountpoint, err := argString(args, "mountpoint")
		if err != nil {
			return nil, fmt.Errorf("sys.unmount: %w", err)
		}
		return sys.Unmount(mountpoint)
	})
	r.Register("sys.list_mounts", func(_ map[string]interface{}) (interface{}, error) {
		return sys.ListMounts()
	})

	// ── sys.hostname_set ───────────────────────────────────────────────
	r.Register("sys.hostname_set", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("sys.hostname_set: %w", err)
		}
		return sys.HostnameSet(name)
	})

	// ── firewall.rule ──────────────────────────────────────────────────
	r.Register("firewall.rule", func(args map[string]interface{}) (interface{}, error) {
		action, err := argString(args, "action")
		if err != nil {
			return nil, fmt.Errorf("firewall.rule: %w", err)
		}
		protocol, err := argString(args, "protocol")
		if err != nil {
			return nil, fmt.Errorf("firewall.rule: %w", err)
		}
		port, _ := argInt(args, "port")
		source := getStringArg(args, "source", "")
		return sys.FirewallRule(action, protocol, port, source)
	})
}

// ============================================================
// extension operations (Ansible-aligned capabilities)
// ============================================================

func (r *Registry) registerExtensions() {
	// ── file.find ────────────────────────────────────────────────────────
	r.Register("file.find", func(args map[string]interface{}) (interface{}, error) {
		opts := file.FindOptions{}
		if paths, ok := args["paths"].([]interface{}); ok {
			for _, p := range paths {
				if s, ok := p.(string); ok {
					opts.Paths = append(opts.Paths, s)
				}
			}
		} else if p, ok := args["paths"].(string); ok {
			opts.Paths = []string{p}
		}
		if pats, ok := args["patterns"].([]interface{}); ok {
			for _, p := range pats {
				if s, ok := p.(string); ok {
					opts.Patterns = append(opts.Patterns, s)
				}
			}
		} else if p, ok := args["patterns"].(string); ok {
			if p != "" {
				opts.Patterns = []string{p}
			}
		}
		if rx, ok := args["regex"].(string); ok {
			opts.Regex = rx
		}
		if ft, ok := args["file_type"].(string); ok {
			opts.FileType = ft
		}
		if md, ok := args["max_depth"].(float64); ok {
			opts.MaxDepth = int(md)
		}
		if age, ok := args["age"].(float64); ok {
			opts.Age = int64(age)
		}
		if sz, ok := args["size"].(float64); ok {
			opts.Size = int64(sz)
		}
		return file.Find(opts)
	})

	// ── file.replace ─────────────────────────────────────────────────────
	r.Register("file.replace", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.replace: %w", err)
		}
		pattern, err := argString(args, "pattern")
		if err != nil {
			return nil, fmt.Errorf("file.replace: %w", err)
		}
		replacement, err := argString(args, "replacement")
		if err != nil {
			return nil, fmt.Errorf("file.replace: %w", err)
		}
		after := getStringArg(args, "after", "")
		before := getStringArg(args, "before", "")
		return file.Replace(path, pattern, replacement, after, before)
	})

	// ── file.blockinfile ─────────────────────────────────────────────────
	r.Register("file.blockinfile", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.blockinfile: %w", err)
		}
		marker, err := argString(args, "marker")
		if err != nil {
			return nil, fmt.Errorf("file.blockinfile: %w", err)
		}
		content, err := argString(args, "content")
		if err != nil {
			return nil, fmt.Errorf("file.blockinfile: %w", err)
		}
		present := true
		if v, ok := args["present"].(bool); ok {
			present = v
		}
		insertAfter := getStringArg(args, "insert_after", "")
		insertBefore := getStringArg(args, "insert_before", "")
		return file.BlockInFile(path, marker, content, present, insertAfter, insertBefore)
	})

	// ── file.ini_get ─────────────────────────────────────────────────────
	r.Register("file.ini_get", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.ini_get: %w", err)
		}
		section, err := argString(args, "section")
		if err != nil {
			return nil, fmt.Errorf("file.ini_get: %w", err)
		}
		key, err := argString(args, "key")
		if err != nil {
			return nil, fmt.Errorf("file.ini_get: %w", err)
		}
		return file.IniGet(path, section, key)
	})

	// ── file.ini_set ─────────────────────────────────────────────────────
	r.Register("file.ini_set", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.ini_set: %w", err)
		}
		section, err := argString(args, "section")
		if err != nil {
			return nil, fmt.Errorf("file.ini_set: %w", err)
		}
		key, err := argString(args, "key")
		if err != nil {
			return nil, fmt.Errorf("file.ini_set: %w", err)
		}
		value, err := argString(args, "value")
		if err != nil {
			return nil, fmt.Errorf("file.ini_set: %w", err)
		}
		return file.IniSet(path, section, key, value)
	})

	// ── archive.create ───────────────────────────────────────────────────
	r.Register("archive.create", func(args map[string]interface{}) (interface{}, error) {
		dest, err := argString(args, "dest")
		if err != nil {
			return nil, fmt.Errorf("archive.create: %w", err)
		}
		var sources []string
		switch v := args["sources"].(type) {
		case string:
			sources = []string{v}
		case []interface{}:
			for _, s := range v {
				if str, ok := s.(string); ok {
					sources = append(sources, str)
				}
			}
		}
		return sdkarchive.Create(dest, sources)
	})

	// ── archive.extract ──────────────────────────────────────────────────
	r.Register("archive.extract", func(args map[string]interface{}) (interface{}, error) {
		src, err := argString(args, "src")
		if err != nil {
			return nil, fmt.Errorf("archive.extract: %w", err)
		}
		dest, err := argString(args, "dest")
		if err != nil {
			return nil, fmt.Errorf("archive.extract: %w", err)
		}
		return sdkarchive.Extract(src, dest)
	})

	// ── net.download ─────────────────────────────────────────────────────
	r.Register("net.download", func(args map[string]interface{}) (interface{}, error) {
		url, err := argString(args, "url")
		if err != nil {
			return nil, fmt.Errorf("net.download: %w", err)
		}
		dest, err := argString(args, "dest")
		if err != nil {
			return nil, fmt.Errorf("net.download: %w", err)
		}
		algo := getStringArg(args, "checksum_algo", "")
		expected := getStringArg(args, "checksum_expected", "")
		return opsnet.Download(url, dest, algo, expected)
	})

	// ── net.wait_for_connection ──────────────────────────────────────────
	r.Register("net.wait_for_connection", func(args map[string]interface{}) (interface{}, error) {
		host, err := argString(args, "host")
		if err != nil {
			return nil, fmt.Errorf("net.wait_for_connection: %w", err)
		}
		port, err := argInt(args, "port")
		if err != nil {
			return nil, fmt.Errorf("net.wait_for_connection: %w", err)
		}
		timeout := 30
		if t, ok := args["timeout"].(float64); ok {
			timeout = int(t)
		}
		return opsnet.WaitForConnection(host, port, timeout)
	})

	// ── sys.timezone_get ─────────────────────────────────────────────────
	r.Register("sys.timezone_get", func(_ map[string]interface{}) (interface{}, error) {
		return sys.TimezoneGet()
	})

	// ── sys.timezone_set ─────────────────────────────────────────────────
	r.Register("sys.timezone_set", func(args map[string]interface{}) (interface{}, error) {
		tz, err := argString(args, "timezone")
		if err != nil {
			return nil, fmt.Errorf("sys.timezone_set: %w", err)
		}
		return sys.TimezoneSet(tz)
	})

	// ── sys.reboot ───────────────────────────────────────────────────────
	r.Register("sys.reboot", func(_ map[string]interface{}) (interface{}, error) {
		return sys.Reboot()
	})

	// ── ssh.authorized_key_add ───────────────────────────────────────────
	r.Register("ssh.authorized_key_add", func(args map[string]interface{}) (interface{}, error) {
		user, err := argString(args, "user")
		if err != nil {
			return nil, fmt.Errorf("ssh.authorized_key_add: %w", err)
		}
		key, err := argString(args, "key")
		if err != nil {
			return nil, fmt.Errorf("ssh.authorized_key_add: %w", err)
		}
		exclusive, _ := args["exclusive"].(bool)
		return sdkssh.AuthorizedKeyAdd(user, key, exclusive)
	})

	// ── ssh.authorized_key_remove ────────────────────────────────────────
	r.Register("ssh.authorized_key_remove", func(args map[string]interface{}) (interface{}, error) {
		user, err := argString(args, "user")
		if err != nil {
			return nil, fmt.Errorf("ssh.authorized_key_remove: %w", err)
		}
		key, err := argString(args, "key")
		if err != nil {
			return nil, fmt.Errorf("ssh.authorized_key_remove: %w", err)
		}
		return sdkssh.AuthorizedKeyRemove(user, key)
	})

	// ── ssh.authorized_key_list ──────────────────────────────────────────
	r.Register("ssh.authorized_key_list", func(args map[string]interface{}) (interface{}, error) {
		user, err := argString(args, "user")
		if err != nil {
			return nil, fmt.Errorf("ssh.authorized_key_list: %w", err)
		}
		return sdkssh.AuthorizedKeyList(user)
	})

	// ── kernel.module_list ───────────────────────────────────────────────
	r.Register("kernel.module_list", func(_ map[string]interface{}) (interface{}, error) {
		return sdkkernel.ModuleList()
	})

	// ── kernel.module_load ───────────────────────────────────────────────
	r.Register("kernel.module_load", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("kernel.module_load: %w", err)
		}
		return sdkkernel.ModuleLoad(name)
	})

	// ── kernel.module_unload ─────────────────────────────────────────────
	r.Register("kernel.module_unload", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("kernel.module_unload: %w", err)
		}
		return sdkkernel.ModuleUnload(name)
	})

	// ── disk.filesystem ──────────────────────────────────────────────────
	r.Register("disk.filesystem", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("disk.filesystem: %w", err)
		}
		fsType := getStringArg(args, "fs_type", "ext4")
		return sdkdisk.FilesystemCreate(device, fsType)
	})

	// ── disk.part_list ───────────────────────────────────────────────────
	r.Register("disk.part_list", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("disk.part_list: %w", err)
		}
		return sdkdisk.PartList(device)
	})

	// ── docker.container_list ─────────────────────────────────────────────
	r.Register("docker.container_list", func(args map[string]interface{}) (interface{}, error) {
		all := getBoolArg(args, "all", false)
		return sdkdocker.ContainerList(all)
	})

	// ── docker.container_exists ───────────────────────────────────────────
	r.Register("docker.container_exists", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("docker.container_exists: %w", err)
		}
		return sdkdocker.ContainerExists(name)
	})

	// ── docker.container_run ──────────────────────────────────────────────
	r.Register("docker.container_run", func(args map[string]interface{}) (interface{}, error) {
		name := getStringArg(args, "name", "")
		image, err := argString(args, "image")
		if err != nil {
			return nil, fmt.Errorf("docker.container_run: %w", err)
		}
		opts := toStringMapArg(args, "opts")
		return sdkdocker.ContainerRun(name, image, opts)
	})

	// ── docker.container_stop ─────────────────────────────────────────────
	r.Register("docker.container_stop", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("docker.container_stop: %w", err)
		}
		return sdkdocker.ContainerStop(name)
	})

	// ── docker.container_remove ───────────────────────────────────────────
	r.Register("docker.container_remove", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("docker.container_remove: %w", err)
		}
		force := getBoolArg(args, "force", false)
		return sdkdocker.ContainerRemove(name, force)
	})

	// ── docker.image_list ─────────────────────────────────────────────────
	r.Register("docker.image_list", func(args map[string]interface{}) (interface{}, error) {
		return sdkdocker.ImageList()
	})

	// ── docker.image_pull ─────────────────────────────────────────────────
	r.Register("docker.image_pull", func(args map[string]interface{}) (interface{}, error) {
		image, err := argString(args, "image")
		if err != nil {
			return nil, fmt.Errorf("docker.image_pull: %w", err)
		}
		return sdkdocker.ImagePull(image)
	})

	// ── docker.image_remove ───────────────────────────────────────────────
	r.Register("docker.image_remove", func(args map[string]interface{}) (interface{}, error) {
		image, err := argString(args, "image")
		if err != nil {
			return nil, fmt.Errorf("docker.image_remove: %w", err)
		}
		force := getBoolArg(args, "force", false)
		return sdkdocker.ImageRemove(image, force)
	})

	// ── hosts.list ────────────────────────────────────────────────────────
	r.Register("hosts.list", func(args map[string]interface{}) (interface{}, error) {
		return opshosts.List()
	})

	// ── hosts.exists ──────────────────────────────────────────────────────
	r.Register("hosts.exists", func(args map[string]interface{}) (interface{}, error) {
		hostname, err := argString(args, "hostname")
		if err != nil {
			return nil, fmt.Errorf("hosts.exists: %w", err)
		}
		return opshosts.Exists(hostname)
	})

	// ── hosts.add ─────────────────────────────────────────────────────────
	r.Register("hosts.add", func(args map[string]interface{}) (interface{}, error) {
		ip, err := argString(args, "ip")
		if err != nil {
			return nil, fmt.Errorf("hosts.add: %w", err)
		}
		hostnamesRaw, ok := args["hostnames"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("hosts.add: hostnames must be an array")
		}
		hostnames := make([]string, len(hostnamesRaw))
		for i, v := range hostnamesRaw {
			if s, ok := v.(string); ok {
				hostnames[i] = s
			}
		}
		return opshosts.Add(ip, hostnames)
	})

	// ── hosts.remove ──────────────────────────────────────────────────────
	r.Register("hosts.remove", func(args map[string]interface{}) (interface{}, error) {
		hostnamesRaw, ok := args["hostnames"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("hosts.remove: hostnames must be an array")
		}
		hostnames := make([]string, len(hostnamesRaw))
		for i, v := range hostnamesRaw {
			if s, ok := v.(string); ok {
				hostnames[i] = s
			}
		}
		return opshosts.Remove(hostnames)
	})

	// ── locale.get ────────────────────────────────────────────────────────
	r.Register("locale.get", func(args map[string]interface{}) (interface{}, error) {
		return sdklocale.Get()
	})

	// ── locale.available ──────────────────────────────────────────────────
	r.Register("locale.available", func(args map[string]interface{}) (interface{}, error) {
		return sdklocale.Available()
	})

	// ── locale.set ────────────────────────────────────────────────────────
	r.Register("locale.set", func(args map[string]interface{}) (interface{}, error) {
		locale, err := argString(args, "locale")
		if err != nil {
			return nil, fmt.Errorf("locale.set: %w", err)
		}
		return sdklocale.Set(locale)
	})

	// ── pip.list ──────────────────────────────────────────────────────────
	r.Register("pip.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkpip.List()
	})

	// ── pip.exists ────────────────────────────────────────────────────────
	r.Register("pip.exists", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("pip.exists: %w", err)
		}
		return sdkpip.Exists(name)
	})

	// ── pip.install ───────────────────────────────────────────────────────
	r.Register("pip.install", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("pip.install: %w", err)
		}
		version := getStringArg(args, "version", "")
		return sdkpip.Install(name, version)
	})

	// ── pip.uninstall ─────────────────────────────────────────────────────
	r.Register("pip.uninstall", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("pip.uninstall: %w", err)
		}
		return sdkpip.Uninstall(name)
	})
}

// toStringMapArg extracts a map[string]string from args[key].
func toStringMapArg(args map[string]interface{}, key string) map[string]string {
	result := make(map[string]string)
	raw, ok := args[key]
	if !ok || raw == nil {
		return result
	}
	m, ok := raw.(map[string]interface{})
	if !ok {
		return result
	}
	for k, v := range m {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result
}

// mapStrArg extracts a string from a map with a default.
func mapStrArg(m map[string]interface{}, key string, def string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
		return fmt.Sprintf("%v", v)
	}
	return def
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

// getBoolArg returns the bool value of an optional arg, or defaultVal.
func getBoolArg(args map[string]interface{}, key string, defaultVal bool) bool {
	if v, ok := args[key]; ok {
		if b, ok := v.(bool); ok {
			return b
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

// argBool returns a boolean arg, defaulting to false if missing.
func argBool(args map[string]interface{}, key string) (bool, error) {
	v, ok := args[key]
	if !ok {
		return false, nil
	}
	switch b := v.(type) {
	case bool:
		return b, nil
	case float64:
		return b != 0, nil
	default:
		return false, fmt.Errorf("argument %q must be a boolean, got %T", key, v)
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
