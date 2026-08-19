package runner

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/opslang/opslang/internal/ast"
	"github.com/opslang/opslang/internal/opsspec"
	sdkaptrepo "github.com/opslang/opslang/pkg/ops-core-sdk/apt_repo"
	sdkarchive "github.com/opslang/opslang/pkg/ops-core-sdk/archive"
	opscron "github.com/opslang/opslang/pkg/ops-core-sdk/cron"
	sdkdisk "github.com/opslang/opslang/pkg/ops-core-sdk/disk"
	sdkdocker "github.com/opslang/opslang/pkg/ops-core-sdk/docker"
	"github.com/opslang/opslang/pkg/ops-core-sdk/file"
	sdkfirewalld "github.com/opslang/opslang/pkg/ops-core-sdk/firewalld"
	opsgit "github.com/opslang/opslang/pkg/ops-core-sdk/git"
	opsgrp "github.com/opslang/opslang/pkg/ops-core-sdk/group"
	opshosts "github.com/opslang/opslang/pkg/ops-core-sdk/hosts"
	opsjson "github.com/opslang/opslang/pkg/ops-core-sdk/json"
	sdkkernel "github.com/opslang/opslang/pkg/ops-core-sdk/kernel"
	sdkknownhosts "github.com/opslang/opslang/pkg/ops-core-sdk/known_hosts"
	sdklimits "github.com/opslang/opslang/pkg/ops-core-sdk/limits"
	sdklocale "github.com/opslang/opslang/pkg/ops-core-sdk/locale"
	sdklogrotate "github.com/opslang/opslang/pkg/ops-core-sdk/logrotate"
	sdklvg "github.com/opslang/opslang/pkg/ops-core-sdk/lvg"
	opsnet "github.com/opslang/opslang/pkg/ops-core-sdk/net"
	sdkntp "github.com/opslang/opslang/pkg/ops-core-sdk/ntp"
	sdkpip "github.com/opslang/opslang/pkg/ops-core-sdk/pip"
	opspkg "github.com/opslang/opslang/pkg/ops-core-sdk/pkg"
	"github.com/opslang/opslang/pkg/ops-core-sdk/process"
	sdkresolv "github.com/opslang/opslang/pkg/ops-core-sdk/resolv"
	sdksnap "github.com/opslang/opslang/pkg/ops-core-sdk/snap"
	sdkselinux "github.com/opslang/opslang/pkg/ops-core-sdk/selinux"
	"github.com/opslang/opslang/pkg/ops-core-sdk/service"
	sdkssh "github.com/opslang/opslang/pkg/ops-core-sdk/ssh"
	"github.com/opslang/opslang/pkg/ops-core-sdk/sys"
	sdksysctl "github.com/opslang/opslang/pkg/ops-core-sdk/sysctl"
	optime "github.com/opslang/opslang/pkg/ops-core-sdk/time"
	opsuser "github.com/opslang/opslang/pkg/ops-core-sdk/user"
	opsyaml "github.com/opslang/opslang/pkg/ops-core-sdk/yaml"
	sdkyumrepo "github.com/opslang/opslang/pkg/ops-core-sdk/yum_repo"
	sdkufw "github.com/opslang/opslang/pkg/ops-core-sdk/ufw"
	sdkinifile "github.com/opslang/opslang/pkg/ops-core-sdk/ini_file"
	sdkmount "github.com/opslang/opslang/pkg/ops-core-sdk/mount"
	sdkhostname "github.com/opslang/opslang/pkg/ops-core-sdk/hostname"
	sdkiptables "github.com/opslang/opslang/pkg/ops-core-sdk/iptables"
	sdknpm "github.com/opslang/opslang/pkg/ops-core-sdk/npm"
	sdkmysql "github.com/opslang/opslang/pkg/ops-core-sdk/mysql"
	sdknginx "github.com/opslang/opslang/pkg/ops-core-sdk/nginx"
	sdkmodprobe "github.com/opslang/opslang/pkg/ops-core-sdk/modprobe"
	sdkalternatives "github.com/opslang/opslang/pkg/ops-core-sdk/alternatives"
	sdkblockdev "github.com/opslang/opslang/pkg/ops-core-sdk/blockdev"
	sdkat "github.com/opslang/opslang/pkg/ops-core-sdk/at"
	sdkpostgresql "github.com/opslang/opslang/pkg/ops-core-sdk/postgresql"
	sdkapache2 "github.com/opslang/opslang/pkg/ops-core-sdk/apache2"
	sdkfilesystem "github.com/opslang/opslang/pkg/ops-core-sdk/filesystem"
	sdkparted "github.com/opslang/opslang/pkg/ops-core-sdk/parted"
	sdkacl "github.com/opslang/opslang/pkg/ops-core-sdk/acl"
	sdkwaitfor "github.com/opslang/opslang/pkg/ops-core-sdk/wait_for"
	sdklvol "github.com/opslang/opslang/pkg/ops-core-sdk/lvol"
	sdksync "github.com/opslang/opslang/pkg/ops-core-sdk/synchronize"
	sdkfetch "github.com/opslang/opslang/pkg/ops-core-sdk/fetch"
	sdksebool "github.com/opslang/opslang/pkg/ops-core-sdk/seboolean"
	sdktimezone "github.com/opslang/opslang/pkg/ops-core-sdk/timezone"
	sdkuri "github.com/opslang/opslang/pkg/ops-core-sdk/uri"
	sdklineinfile "github.com/opslang/opslang/pkg/ops-core-sdk/lineinfile"
	sdkreplace "github.com/opslang/opslang/pkg/ops-core-sdk/replace"
	sdkxml "github.com/opslang/opslang/pkg/ops-core-sdk/xml"
	sdksystemd "github.com/opslang/opslang/pkg/ops-core-sdk/systemd"
	sdkpatch "github.com/opslang/opslang/pkg/ops-core-sdk/patch"
	sdkxattr "github.com/opslang/opslang/pkg/ops-core-sdk/xattr"
	sdkfirewalldzone "github.com/opslang/opslang/pkg/ops-core-sdk/firewalld_zone"
	sdkgeturl "github.com/opslang/opslang/pkg/ops-core-sdk/get_url"
	sdkseport "github.com/opslang/opslang/pkg/ops-core-sdk/seport"
	sdksefcontext "github.com/opslang/opslang/pkg/ops-core-sdk/sefcontext"
	sdkflatpak "github.com/opslang/opslang/pkg/ops-core-sdk/flatpak"
	sdkzfs "github.com/opslang/opslang/pkg/ops-core-sdk/zfs"
	sdknmcli "github.com/opslang/opslang/pkg/ops-core-sdk/nmcli"
	sdkcrypttab "github.com/opslang/opslang/pkg/ops-core-sdk/crypttab"
	sdksysfs "github.com/opslang/opslang/pkg/ops-core-sdk/sysfs"
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
	r.registerSelinuxOps()
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

	// snap operations
	r.Register("snap.install", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		channel, _ := argString(args, "channel")
		if channel == "" {
			channel = "stable"
		}
		classic, _ := args["classic"].(bool)
		return sdksnap.Install(name, channel, classic)
	})
	r.Register("snap.remove", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdksnap.Remove(name)
	})
	r.Register("snap.refresh", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		channel, _ := argString(args, "channel")
		return sdksnap.Refresh(name, channel)
	})
	r.Register("snap.list", func(_ map[string]interface{}) (interface{}, error) {
		return sdksnap.List()
	})
	r.Register("snap.get", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdksnap.Get(name)
	})
	r.Register("snap.enable", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdksnap.Enable(name)
	})
	r.Register("snap.disable", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdksnap.Disable(name)
	})
	r.Register("snap.switch", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		channel, _ := argString(args, "channel")
		return sdksnap.Switch(name, channel)
	})
	r.Register("snap.changes", func(_ map[string]interface{}) (interface{}, error) {
		return sdksnap.Changes()
	})
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
// selinux operations
// ============================================================

func (r *Registry) registerSelinuxOps() {
	r.Register("selinux.get", func(args map[string]interface{}) (interface{}, error) {
		return sdkselinux.Get()
	})
	r.Register("selinux.set", func(args map[string]interface{}) (interface{}, error) {
		mode, err := argString(args, "mode")
		if err != nil {
			return nil, fmt.Errorf("selinux.set: %w", err)
		}
		return sdkselinux.Set(mode)
	})
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

	// ── firewalld ────────────────────────────────────────────────────────
	r.Register("firewalld.get", func(args map[string]interface{}) (interface{}, error) {
		return sdkfirewalld.Get()
	})
	r.Register("firewalld.start", func(args map[string]interface{}) (interface{}, error) {
		return sdkfirewalld.Start()
	})
	r.Register("firewalld.stop", func(args map[string]interface{}) (interface{}, error) {
		return sdkfirewalld.Stop()
	})
	r.Register("firewalld.restart", func(args map[string]interface{}) (interface{}, error) {
		return sdkfirewalld.Restart()
	})
	r.Register("firewalld.enable", func(args map[string]interface{}) (interface{}, error) {
		return sdkfirewalld.Enable()
	})
	r.Register("firewalld.disable", func(args map[string]interface{}) (interface{}, error) {
		return sdkfirewalld.Disable()
	})
	r.Register("firewalld.list_zones", func(args map[string]interface{}) (interface{}, error) {
		return sdkfirewalld.ListZones()
	})
	r.Register("firewalld.reload", func(args map[string]interface{}) (interface{}, error) {
		return sdkfirewalld.Reload()
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

	// ── apt_repo.* ──────────────────────────────────────────────────────
	r.Register("apt_repo.list", func(_ map[string]interface{}) (interface{}, error) {
		return sdkaptrepo.List()
	})
	r.Register("apt_repo.exists", func(args map[string]interface{}) (interface{}, error) {
		uri, err := argString(args, "uri")
		if err != nil {
			return nil, fmt.Errorf("apt_repo.exists: %w", err)
		}
		return sdkaptrepo.Exists(uri)
	})
	r.Register("apt_repo.add", func(args map[string]interface{}) (interface{}, error) {
		uri, err := argString(args, "uri")
		if err != nil {
			return nil, fmt.Errorf("apt_repo.add: %w", err)
		}
		dist, _ := args["dist"].(string)
		comps, _ := args["components"].(string)
		return sdkaptrepo.Add(uri, dist, comps)
	})
	r.Register("apt_repo.remove", func(args map[string]interface{}) (interface{}, error) {
		uri, err := argString(args, "uri")
		if err != nil {
			return nil, fmt.Errorf("apt_repo.remove: %w", err)
		}
		return sdkaptrepo.Remove(uri)
	})
	r.Register("apt_repo.update", func(_ map[string]interface{}) (interface{}, error) {
		return sdkaptrepo.Update()
	})

	// ── logrotate.* ─────────────────────────────────────────────────────
	r.Register("logrotate.list", func(_ map[string]interface{}) (interface{}, error) {
		return sdklogrotate.List()
	})
	r.Register("logrotate.get", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("logrotate.get: %w", err)
		}
		return sdklogrotate.Get(name)
	})
	r.Register("logrotate.set", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("logrotate.set: %w", err)
		}
		pattern, _ := args["pattern"].(string)
		freq, _ := args["frequency"].(string)
		rotate := 7
		if v, ok := args["rotate"].(float64); ok {
			rotate = int(v)
		}
		compress := getBoolArg(args, "compress", false)
		postRotate, _ := args["post_rotate"].(string)
		return sdklogrotate.Set(name, pattern, freq, rotate, compress, postRotate)
	})
	r.Register("logrotate.remove", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("logrotate.remove: %w", err)
		}
		return sdklogrotate.Remove(name)
	})

	// ── lvg.* ─────────────────────────────────────────────────────────────
	r.Register("lvg.create", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		pvsRaw, _ := args["pvs"].([]interface{})
		pvs := make([]string, len(pvsRaw))
		for i, pv := range pvsRaw {
			pvs[i], _ = pv.(string)
		}
		return sdklvg.Create(name, pvs)
	})
	r.Register("lvg.remove", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdklvg.Remove(name)
	})
	r.Register("lvg.extend", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		pvsRaw, _ := args["pvs"].([]interface{})
		pvs := make([]string, len(pvsRaw))
		for i, pv := range pvsRaw {
			pvs[i], _ = pv.(string)
		}
		return sdklvg.Extend(name, pvs)
	})
	r.Register("lvg.reduce", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		pvsRaw, _ := args["pvs"].([]interface{})
		pvs := make([]string, len(pvsRaw))
		for i, pv := range pvsRaw {
			pvs[i], _ = pv.(string)
		}
		return sdklvg.Reduce(name, pvs)
	})
	r.Register("lvg.activate", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdklvg.Activate(name)
	})
	r.Register("lvg.deactivate", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdklvg.Deactivate(name)
	})
	r.Register("lvg.list", func(_ map[string]interface{}) (interface{}, error) {
		return sdklvg.List()
	})
	r.Register("lvg.get", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdklvg.Get(name)
	})

	// ── resolv.* ────────────────────────────────────────────────────────
	r.Register("resolv.get", func(_ map[string]interface{}) (interface{}, error) {
		return sdkresolv.Get()
	})
	r.Register("resolv.set", func(args map[string]interface{}) (interface{}, error) {
		var nameservers, search, options []string
		if ns, ok := args["nameservers"].([]interface{}); ok {
			for _, v := range ns {
				if s, ok := v.(string); ok {
					nameservers = append(nameservers, s)
				}
			}
		}
		if s, ok := args["search"].([]interface{}); ok {
			for _, v := range s {
				if str, ok := v.(string); ok {
					search = append(search, str)
				}
			}
		}
		if o, ok := args["options"].([]interface{}); ok {
			for _, v := range o {
				if str, ok := v.(string); ok {
					options = append(options, str)
				}
			}
		}
		domain, _ := args["domain"].(string)
		return sdkresolv.Set(nameservers, search, options, domain)
	})
	r.Register("resolv.add_nameserver", func(args map[string]interface{}) (interface{}, error) {
		ns, err := argString(args, "nameserver")
		if err != nil {
			return nil, fmt.Errorf("resolv.add_nameserver: %w", err)
		}
		return sdkresolv.AddNameserver(ns)
	})
	r.Register("resolv.remove_nameserver", func(args map[string]interface{}) (interface{}, error) {
		ns, err := argString(args, "nameserver")
		if err != nil {
			return nil, fmt.Errorf("resolv.remove_nameserver: %w", err)
		}
		return sdkresolv.RemoveNameserver(ns)
	})

	// ── yum_repo.* ──────────────────────────────────────────────────────
	r.Register("yum_repo.list", func(_ map[string]interface{}) (interface{}, error) {
		return sdkyumrepo.List()
	})
	r.Register("yum_repo.exists", func(args map[string]interface{}) (interface{}, error) {
		id, err := argString(args, "id")
		if err != nil {
			return nil, fmt.Errorf("yum_repo.exists: %w", err)
		}
		return sdkyumrepo.Exists(id)
	})
	r.Register("yum_repo.add", func(args map[string]interface{}) (interface{}, error) {
		id, err := argString(args, "id")
		if err != nil {
			return nil, fmt.Errorf("yum_repo.add: %w", err)
		}
		name, _ := args["name"].(string)
		baseURL, _ := args["base_url"].(string)
		gpgCheck := getBoolArg(args, "gpg_check", false)
		gpgKey, _ := args["gpg_key"].(string)
		return sdkyumrepo.Add(id, name, baseURL, gpgCheck, gpgKey)
	})
	r.Register("yum_repo.remove", func(args map[string]interface{}) (interface{}, error) {
		id, err := argString(args, "id")
		if err != nil {
			return nil, fmt.Errorf("yum_repo.remove: %w", err)
		}
		return sdkyumrepo.Remove(id)
	})

	// ── known_hosts ──────────────────────────────────────────────────────
	r.Register("known_hosts.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkknownhosts.List()
	})
	r.Register("known_hosts.check", func(args map[string]interface{}) (interface{}, error) {
		host, err := argString(args, "host")
		if err != nil {
			return nil, fmt.Errorf("known_hosts.check: %w", err)
		}
		return sdkknownhosts.Check(host)
	})
	r.Register("known_hosts.add", func(args map[string]interface{}) (interface{}, error) {
		host, err := argString(args, "host")
		if err != nil {
			return nil, fmt.Errorf("known_hosts.add: %w", err)
		}
		return sdkknownhosts.Add(host)
	})
	r.Register("known_hosts.remove", func(args map[string]interface{}) (interface{}, error) {
		host, err := argString(args, "host")
		if err != nil {
			return nil, fmt.Errorf("known_hosts.remove: %w", err)
		}
		return sdkknownhosts.Remove(host)
	})

	// ── limits ───────────────────────────────────────────────────────────
	r.Register("limits.list", func(args map[string]interface{}) (interface{}, error) {
		return sdklimits.List()
	})
	r.Register("limits.get", func(args map[string]interface{}) (interface{}, error) {
		domain, err := argString(args, "domain")
		if err != nil {
			return nil, fmt.Errorf("limits.get: %w", err)
		}
		return sdklimits.Get(domain)
	})
	r.Register("limits.set", func(args map[string]interface{}) (interface{}, error) {
		domain, _ := argString(args, "domain")
		typ, _ := argString(args, "type")
		item, _ := argString(args, "item")
		value, _ := argString(args, "value")
		return sdklimits.Set(domain, typ, item, value)
	})
	r.Register("limits.remove", func(args map[string]interface{}) (interface{}, error) {
		domain, err := argString(args, "domain")
		if err != nil {
			return nil, fmt.Errorf("limits.remove: %w", err)
		}
		return sdklimits.Remove(domain)
	})

	// ── ntp ──────────────────────────────────────────────────────────────
	r.Register("ntp.get", func(args map[string]interface{}) (interface{}, error) {
		return sdkntp.Get()
	})
	r.Register("ntp.set", func(args map[string]interface{}) (interface{}, error) {
		server, err := argString(args, "server")
		if err != nil {
			return nil, fmt.Errorf("ntp.set: %w", err)
		}
		return sdkntp.Set(server)
	})

	// ── ufw ──────────────────────────────────────────────────────────────
	r.Register("ufw.status", func(args map[string]interface{}) (interface{}, error) {
		return sdkufw.Status()
	})
	r.Register("ufw.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkufw.List()
	})
	r.Register("ufw.enable", func(args map[string]interface{}) (interface{}, error) {
		return sdkufw.Enable()
	})
	r.Register("ufw.disable", func(args map[string]interface{}) (interface{}, error) {
		return sdkufw.Disable()
	})
	r.Register("ufw.allow", func(args map[string]interface{}) (interface{}, error) {
		port, err := argString(args, "port")
		if err != nil {
			return nil, fmt.Errorf("ufw.allow: %w", err)
		}
		proto, _ := argString(args, "proto")
		if proto == "" {
			proto = "tcp"
		}
		return sdkufw.Allow(port, proto)
	})
	r.Register("ufw.deny", func(args map[string]interface{}) (interface{}, error) {
		port, err := argString(args, "port")
		if err != nil {
			return nil, fmt.Errorf("ufw.deny: %w", err)
		}
		proto, _ := argString(args, "proto")
		if proto == "" {
			proto = "tcp"
		}
		return sdkufw.Deny(port, proto)
	})
	r.Register("ufw.delete", func(args map[string]interface{}) (interface{}, error) {
		num, err := argInt(args, "number")
		if err != nil {
			return nil, fmt.Errorf("ufw.delete: %w", err)
		}
		return sdkufw.Delete(num)
	})
	r.Register("ufw.reset", func(args map[string]interface{}) (interface{}, error) {
		return sdkufw.Reset()
	})
	r.Register("ufw.reload", func(args map[string]interface{}) (interface{}, error) {
		return sdkufw.Reload()
	})

	// ── ini_file ─────────────────────────────────────────────────────────
	r.Register("ini_file.sections", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("ini_file.sections: %w", err)
		}
		return sdkinifile.Sections(path)
	})
	r.Register("ini_file.get", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("ini_file.get: %w", err)
		}
		section, err := argString(args, "section")
		if err != nil {
			return nil, fmt.Errorf("ini_file.get: %w", err)
		}
		key, err := argString(args, "key")
		if err != nil {
			return nil, fmt.Errorf("ini_file.get: %w", err)
		}
		return sdkinifile.Get(path, section, key)
	})
	r.Register("ini_file.set", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("ini_file.set: %w", err)
		}
		section, err := argString(args, "section")
		if err != nil {
			return nil, fmt.Errorf("ini_file.set: %w", err)
		}
		key, err := argString(args, "key")
		if err != nil {
			return nil, fmt.Errorf("ini_file.set: %w", err)
		}
		value, err := argString(args, "value")
		if err != nil {
			return nil, fmt.Errorf("ini_file.set: %w", err)
		}
		return sdkinifile.Set(path, section, key, value)
	})
	r.Register("ini_file.remove", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("ini_file.remove: %w", err)
		}
		section, err := argString(args, "section")
		if err != nil {
			return nil, fmt.Errorf("ini_file.remove: %w", err)
		}
		key, err := argString(args, "key")
		if err != nil {
			return nil, fmt.Errorf("ini_file.remove: %w", err)
		}
		return sdkinifile.Remove(path, section, key)
	})
	r.Register("ini_file.remove_section", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("ini_file.remove_section: %w", err)
		}
		section, err := argString(args, "section")
		if err != nil {
			return nil, fmt.Errorf("ini_file.remove_section: %w", err)
		}
		return sdkinifile.RemoveSection(path, section)
	})

	// ── mount ────────────────────────────────────────────────────────────
	r.Register("mount.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkmount.List()
	})
	r.Register("mount.mount", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("mount.mount: %w", err)
		}
		mountpoint, err := argString(args, "mountpoint")
		if err != nil {
			return nil, fmt.Errorf("mount.mount: %w", err)
		}
		fstype, _ := argString(args, "fstype")
		options, _ := argString(args, "options")
		return sdkmount.Mount(device, mountpoint, fstype, options)
	})
	r.Register("mount.umount", func(args map[string]interface{}) (interface{}, error) {
		mountpoint, err := argString(args, "mountpoint")
		if err != nil {
			return nil, fmt.Errorf("mount.umount: %w", err)
		}
		return sdkmount.Unmount(mountpoint)
	})
	r.Register("mount.fstab", func(args map[string]interface{}) (interface{}, error) {
		return sdkmount.Fstab()
	})
	r.Register("mount.add_fstab", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("mount.add_fstab: %w", err)
		}
		mountpoint, err := argString(args, "mountpoint")
		if err != nil {
			return nil, fmt.Errorf("mount.add_fstab: %w", err)
		}
		fstype, err := argString(args, "fstype")
		if err != nil {
			return nil, fmt.Errorf("mount.add_fstab: %w", err)
		}
		options, _ := argString(args, "options")
		return sdkmount.AddFstab(device, mountpoint, fstype, options)
	})
	r.Register("mount.remove_fstab", func(args map[string]interface{}) (interface{}, error) {
		target, err := argString(args, "target")
		if err != nil {
			return nil, fmt.Errorf("mount.remove_fstab: %w", err)
		}
		return sdkmount.RemoveFstab(target)
	})

	// ── hostname ─────────────────────────────────────────────────────────
	r.Register("hostname.get", func(args map[string]interface{}) (interface{}, error) {
		return sdkhostname.Get()
	})
	r.Register("hostname.set", func(args map[string]interface{}) (interface{}, error) {
		hostname, err := argString(args, "hostname")
		if err != nil {
			return nil, fmt.Errorf("hostname.set: %w", err)
		}
		return sdkhostname.Set(hostname)
	})
	r.Register("hostname.set_fqdn", func(args map[string]interface{}) (interface{}, error) {
		fqdn, err := argString(args, "fqdn")
		if err != nil {
			return nil, fmt.Errorf("hostname.set_fqdn: %w", err)
		}
		return sdkhostname.SetFQDN(fqdn)
	})

	// ── timezone ─────────────────────────────────────────────────────────
	r.Register("timezone.get", func(args map[string]interface{}) (interface{}, error) {
		return sdktimezone.Get()
	})
	r.Register("timezone.set", func(args map[string]interface{}) (interface{}, error) {
		timezone, err := argString(args, "timezone")
		if err != nil {
			return nil, fmt.Errorf("timezone.set: %w", err)
		}
		return sdktimezone.Set(timezone)
	})
	r.Register("timezone.list", func(args map[string]interface{}) (interface{}, error) {
		return sdktimezone.List()
	})

	// ── iptables ──────────────────────────────────────────────────────
	r.Register("iptables.list", func(args map[string]interface{}) (interface{}, error) {
		chain := getStringArg(args, "chain", "")
		return sdkiptables.List(chain)
	})
	r.Register("iptables.flush", func(args map[string]interface{}) (interface{}, error) {
		table := getStringArg(args, "table", "")
		return sdkiptables.Flush(table)
	})
	r.Register("iptables.add_rule", func(args map[string]interface{}) (interface{}, error) {
		chain, err := argString(args, "chain")
		if err != nil {
			return nil, fmt.Errorf("iptables.add_rule: %w", err)
		}
		ruleSpec, err := argString(args, "rule_spec")
		if err != nil {
			return nil, fmt.Errorf("iptables.add_rule: %w", err)
		}
		return sdkiptables.AddRule(chain, ruleSpec)
	})
	r.Register("iptables.delete_rule", func(args map[string]interface{}) (interface{}, error) {
		chain, err := argString(args, "chain")
		if err != nil {
			return nil, fmt.Errorf("iptables.delete_rule: %w", err)
		}
		num, err := argInt(args, "number")
		if err != nil {
			return nil, fmt.Errorf("iptables.delete_rule: %w", err)
		}
		return sdkiptables.DeleteRule(chain, num)
	})
	r.Register("iptables.save", func(args map[string]interface{}) (interface{}, error) {
		return sdkiptables.Save()
	})
	r.Register("iptables.list_chains", func(args map[string]interface{}) (interface{}, error) {
		return sdkiptables.ListChains()
	})

	// ── npm ───────────────────────────────────────────────────────────
	r.Register("npm.list", func(args map[string]interface{}) (interface{}, error) {
		global, _ := argBool(args, "global")
		return sdknpm.List(global)
	})
	r.Register("npm.install", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("npm.install: %w", err)
		}
		global, _ := argBool(args, "global")
		return sdknpm.Install(name, global)
	})
	r.Register("npm.uninstall", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("npm.uninstall: %w", err)
		}
		global, _ := argBool(args, "global")
		return sdknpm.Uninstall(name, global)
	})
	r.Register("npm.outdated", func(args map[string]interface{}) (interface{}, error) {
		global, _ := argBool(args, "global")
		return sdknpm.Outdated(global)
	})

	// ── mysql ─────────────────────────────────────────────────────────
	r.Register("mysql.databases", func(args map[string]interface{}) (interface{}, error) {
		return sdkmysql.Databases()
	})
	r.Register("mysql.create_database", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("mysql.create_database: %w", err)
		}
		return sdkmysql.CreateDatabase(name)
	})
	r.Register("mysql.drop_database", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("mysql.drop_database: %w", err)
		}
		return sdkmysql.DropDatabase(name)
	})
	r.Register("mysql.users", func(args map[string]interface{}) (interface{}, error) {
		return sdkmysql.Users()
	})
	r.Register("mysql.create_user", func(args map[string]interface{}) (interface{}, error) {
		user, err := argString(args, "user")
		if err != nil {
			return nil, fmt.Errorf("mysql.create_user: %w", err)
		}
		host, err := argString(args, "host")
		if err != nil {
			return nil, fmt.Errorf("mysql.create_user: %w", err)
		}
		password, err := argString(args, "password")
		if err != nil {
			return nil, fmt.Errorf("mysql.create_user: %w", err)
		}
		return sdkmysql.CreateUser(user, host, password)
	})
	r.Register("mysql.drop_user", func(args map[string]interface{}) (interface{}, error) {
		user, err := argString(args, "user")
		if err != nil {
			return nil, fmt.Errorf("mysql.drop_user: %w", err)
		}
		host, err := argString(args, "host")
		if err != nil {
			return nil, fmt.Errorf("mysql.drop_user: %w", err)
		}
		return sdkmysql.DropUser(user, host)
	})
	r.Register("mysql.grant", func(args map[string]interface{}) (interface{}, error) {
		privileges, err := argString(args, "privileges")
		if err != nil {
			return nil, fmt.Errorf("mysql.grant: %w", err)
		}
		database, err := argString(args, "database")
		if err != nil {
			return nil, fmt.Errorf("mysql.grant: %w", err)
		}
		user, err := argString(args, "user")
		if err != nil {
			return nil, fmt.Errorf("mysql.grant: %w", err)
		}
		host, err := argString(args, "host")
		if err != nil {
			return nil, fmt.Errorf("mysql.grant: %w", err)
		}
		return sdkmysql.Grant(privileges, database, user, host)
	})

	// ── nginx ─────────────────────────────────────────────────────────
	r.Register("nginx.config_test", func(args map[string]interface{}) (interface{}, error) {
		return sdknginx.ConfigTest()
	})
	r.Register("nginx.reload", func(args map[string]interface{}) (interface{}, error) {
		return sdknginx.Reload()
	})
	r.Register("nginx.sites_list", func(args map[string]interface{}) (interface{}, error) {
		return sdknginx.SitesList()
	})
	r.Register("nginx.site_enable", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("nginx.site_enable: %w", err)
		}
		return sdknginx.SiteEnable(name)
	})
	r.Register("nginx.site_disable", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("nginx.site_disable: %w", err)
		}
		return sdknginx.SiteDisable(name)
	})

	// ── modprobe ──────────────────────────────────────────────────────
	r.Register("modprobe.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkmodprobe.List()
	})
	r.Register("modprobe.load", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("modprobe.load: %w", err)
		}
		return sdkmodprobe.Load(name)
	})
	r.Register("modprobe.unload", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("modprobe.unload: %w", err)
		}
		return sdkmodprobe.Unload(name)
	})
	r.Register("modprobe.is_loaded", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("modprobe.is_loaded: %w", err)
		}
		return sdkmodprobe.IsLoaded(name)
	})

	// ── alternatives ──────────────────────────────────────────────────
	r.Register("alternatives.list", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("alternatives.list: %w", err)
		}
		return sdkalternatives.List(name)
	})
	r.Register("alternatives.display", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("alternatives.display: %w", err)
		}
		return sdkalternatives.Display(name)
	})
	r.Register("alternatives.set", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("alternatives.set: %w", err)
		}
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("alternatives.set: %w", err)
		}
		return sdkalternatives.Set(name, path)
	})
	r.Register("alternatives.install", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("alternatives.install: %w", err)
		}
		link, err := argString(args, "link")
		if err != nil {
			return nil, fmt.Errorf("alternatives.install: %w", err)
		}
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("alternatives.install: %w", err)
		}
		priority, err := argInt(args, "priority")
		if err != nil {
			return nil, fmt.Errorf("alternatives.install: %w", err)
		}
		return sdkalternatives.Install(name, link, path, priority)
	})
	r.Register("alternatives.remove", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("alternatives.remove: %w", err)
		}
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("alternatives.remove: %w", err)
		}
		return sdkalternatives.Remove(name, path)
	})

	// ── blockdev ──────────────────────────────────────────────────────
	r.Register("blockdev.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkblockdev.List()
	})
	r.Register("blockdev.info", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("blockdev.info: %w", err)
		}
		return sdkblockdev.Info(device)
	})
	r.Register("blockdev.flush_buffers", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("blockdev.flush_buffers: %w", err)
		}
		return sdkblockdev.FlushBuffers(device)
	})
	r.Register("blockdev.set_readahead", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("blockdev.set_readahead: %w", err)
		}
		value, err := argInt(args, "value")
		if err != nil {
			return nil, fmt.Errorf("blockdev.set_readahead: %w", err)
		}
		return sdkblockdev.SetReadahead(device, value)
	})

	// ── at ────────────────────────────────────────────────────────────
	r.Register("at.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkat.List()
	})
	r.Register("at.schedule", func(args map[string]interface{}) (interface{}, error) {
		command, err := argString(args, "command")
		if err != nil {
			return nil, fmt.Errorf("at.schedule: %w", err)
		}
		timeSpec, err := argString(args, "time_spec")
		if err != nil {
			return nil, fmt.Errorf("at.schedule: %w", err)
		}
		return sdkat.Schedule(command, timeSpec)
	})
	r.Register("at.remove", func(args map[string]interface{}) (interface{}, error) {
		jobID, err := argString(args, "job_id")
		if err != nil {
			return nil, fmt.Errorf("at.remove: %w", err)
		}
		return sdkat.Remove(jobID)
	})

	// ── postgresql ─────────────────────────────────────────────────────
	r.Register("postgresql.databases", func(args map[string]interface{}) (interface{}, error) {
		return sdkpostgresql.Databases()
	})
	r.Register("postgresql.create_database", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("postgresql.create_database: %w", err)
		}
		return sdkpostgresql.CreateDatabase(name)
	})
	r.Register("postgresql.drop_database", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("postgresql.drop_database: %w", err)
		}
		return sdkpostgresql.DropDatabase(name)
	})
	r.Register("postgresql.users", func(args map[string]interface{}) (interface{}, error) {
		return sdkpostgresql.Users()
	})
	r.Register("postgresql.create_user", func(args map[string]interface{}) (interface{}, error) {
		user, err := argString(args, "user")
		if err != nil {
			return nil, fmt.Errorf("postgresql.create_user: %w", err)
		}
		password, err := argString(args, "password")
		if err != nil {
			return nil, fmt.Errorf("postgresql.create_user: %w", err)
		}
		return sdkpostgresql.CreateUser(user, password)
	})
	r.Register("postgresql.drop_user", func(args map[string]interface{}) (interface{}, error) {
		user, err := argString(args, "user")
		if err != nil {
			return nil, fmt.Errorf("postgresql.drop_user: %w", err)
		}
		return sdkpostgresql.DropUser(user)
	})
	r.Register("postgresql.grant", func(args map[string]interface{}) (interface{}, error) {
		privileges, err := argString(args, "privileges")
		if err != nil {
			return nil, fmt.Errorf("postgresql.grant: %w", err)
		}
		database, err := argString(args, "database")
		if err != nil {
			return nil, fmt.Errorf("postgresql.grant: %w", err)
		}
		user, err := argString(args, "user")
		if err != nil {
			return nil, fmt.Errorf("postgresql.grant: %w", err)
		}
		return sdkpostgresql.Grant(privileges, database, user)
	})

	// ── apache2 ────────────────────────────────────────────────────────
	r.Register("apache2.config_test", func(args map[string]interface{}) (interface{}, error) {
		return sdkapache2.ConfigTest()
	})
	r.Register("apache2.reload", func(args map[string]interface{}) (interface{}, error) {
		return sdkapache2.Reload()
	})
	r.Register("apache2.sites_list", func(args map[string]interface{}) (interface{}, error) {
		return sdkapache2.SitesList()
	})
	r.Register("apache2.site_enable", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("apache2.site_enable: %w", err)
		}
		return sdkapache2.SiteEnable(name)
	})
	r.Register("apache2.site_disable", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("apache2.site_disable: %w", err)
		}
		return sdkapache2.SiteDisable(name)
	})
	r.Register("apache2.modules_list", func(args map[string]interface{}) (interface{}, error) {
		return sdkapache2.ModulesList()
	})
	r.Register("apache2.module_enable", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("apache2.module_enable: %w", err)
		}
		return sdkapache2.ModuleEnable(name)
	})
	r.Register("apache2.module_disable", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("apache2.module_disable: %w", err)
		}
		return sdkapache2.ModuleDisable(name)
	})

	// ── filesystem ─────────────────────────────────────────────────────
	r.Register("filesystem.mkfs", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("filesystem.mkfs: %w", err)
		}
		fsType, err := argString(args, "fstype")
		if err != nil {
			return nil, fmt.Errorf("filesystem.mkfs: %w", err)
		}
		label, _ := args["label"].(string)
		return sdkfilesystem.Mkfs(device, fsType, label)
	})
	r.Register("filesystem.resize_ext4", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("filesystem.resize_ext4: %w", err)
		}
		return sdkfilesystem.ResizeExt4(device)
	})
	r.Register("filesystem.resize_xfs", func(args map[string]interface{}) (interface{}, error) {
		mountpoint, err := argString(args, "mountpoint")
		if err != nil {
			return nil, fmt.Errorf("filesystem.resize_xfs: %w", err)
		}
		return sdkfilesystem.ResizeXFS(mountpoint)
	})
	r.Register("filesystem.check", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("filesystem.check: %w", err)
		}
		return sdkfilesystem.Check(device)
	})

	// ── parted ─────────────────────────────────────────────────────────
	r.Register("parted.list", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("parted.list: %w", err)
		}
		return sdkparted.List(device)
	})
	r.Register("parted.mklabel", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("parted.mklabel: %w", err)
		}
		labelType, _ := args["label_type"].(string)
		if labelType == "" {
			labelType = "gpt"
		}
		return sdkparted.MkLabel(device, labelType)
	})
	r.Register("parted.mkpart", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("parted.mkpart: %w", err)
		}
		partType, _ := args["part_type"].(string)
		fsType, _ := args["fstype"].(string)
		start, err := argString(args, "start")
		if err != nil {
			return nil, fmt.Errorf("parted.mkpart: %w", err)
		}
		end, err := argString(args, "end")
		if err != nil {
			return nil, fmt.Errorf("parted.mkpart: %w", err)
		}
		return sdkparted.MkPart(device, partType, fsType, start, end)
	})
	r.Register("parted.rm", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("parted.rm: %w", err)
		}
		number, err := argInt(args, "number")
		if err != nil {
			return nil, fmt.Errorf("parted.rm: %w", err)
		}
		return sdkparted.Rm(device, number)
	})

	// ── acl ────────────────────────────────────────────────────────────
	r.Register("acl.get", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("acl.get: %w", err)
		}
		return sdkacl.Get(path)
	})
	r.Register("acl.set", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("acl.set: %w", err)
		}
		entry, err := argString(args, "entry")
		if err != nil {
			return nil, fmt.Errorf("acl.set: %w", err)
		}
		recursive, _ := argBool(args, "recursive")
		return sdkacl.Set(path, entry, recursive)
	})
	r.Register("acl.remove", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("acl.remove: %w", err)
		}
		entry, err := argString(args, "entry")
		if err != nil {
			return nil, fmt.Errorf("acl.remove: %w", err)
		}
		recursive, _ := argBool(args, "recursive")
		return sdkacl.Remove(path, entry, recursive)
	})
	r.Register("acl.remove_all", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("acl.remove_all: %w", err)
		}
		recursive, _ := argBool(args, "recursive")
		return sdkacl.RemoveAll(path, recursive)
	})

	// ── wait_for ───────────────────────────────────────────────────────
	r.Register("wait_for.port", func(args map[string]interface{}) (interface{}, error) {
		host, err := argString(args, "host")
		if err != nil {
			return nil, fmt.Errorf("wait_for.port: %w", err)
		}
		port, err := argInt(args, "port")
		if err != nil {
			return nil, fmt.Errorf("wait_for.port: %w", err)
		}
		timeoutMs := 30000
		if v, ok := args["timeout_ms"]; ok {
			if t, e := argInt(map[string]interface{}{"timeout_ms": v}, "timeout_ms"); e == nil {
				timeoutMs = t
			}
		}
		return sdkwaitfor.Port(host, port, timeoutMs)
	})
	r.Register("wait_for.file", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("wait_for.file: %w", err)
		}
		timeoutMs := 30000
		if v, ok := args["timeout_ms"]; ok {
			if t, e := argInt(map[string]interface{}{"timeout_ms": v}, "timeout_ms"); e == nil {
				timeoutMs = t
			}
		}
		return sdkwaitfor.File(path, timeoutMs)
	})
	r.Register("wait_for.url", func(args map[string]interface{}) (interface{}, error) {
		url, err := argString(args, "url")
		if err != nil {
			return nil, fmt.Errorf("wait_for.url: %w", err)
		}
		timeoutMs := 30000
		if v, ok := args["timeout_ms"]; ok {
			if t, e := argInt(map[string]interface{}{"timeout_ms": v}, "timeout_ms"); e == nil {
				timeoutMs = t
			}
		}
		return sdkwaitfor.URL(url, timeoutMs)
	})

	// ── lvol ───────────────────────────────────────────────────────────
	r.Register("lvol.list", func(args map[string]interface{}) (interface{}, error) {
		return sdklvol.List()
	})
	r.Register("lvol.vg_list", func(args map[string]interface{}) (interface{}, error) {
		return sdklvol.VGList()
	})
	r.Register("lvol.create", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("lvol.create: %w", err)
		}
		vg, err := argString(args, "vg")
		if err != nil {
			return nil, fmt.Errorf("lvol.create: %w", err)
		}
		size, err := argString(args, "size")
		if err != nil {
			return nil, fmt.Errorf("lvol.create: %w", err)
		}
		return sdklvol.Create(name, vg, size)
	})
	r.Register("lvol.remove", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("lvol.remove: %w", err)
		}
		vg, err := argString(args, "vg")
		if err != nil {
			return nil, fmt.Errorf("lvol.remove: %w", err)
		}
		return sdklvol.Remove(name, vg)
	})
	r.Register("lvol.resize", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("lvol.resize: %w", err)
		}
		vg, err := argString(args, "vg")
		if err != nil {
			return nil, fmt.Errorf("lvol.resize: %w", err)
		}
		size, err := argString(args, "size")
		if err != nil {
			return nil, fmt.Errorf("lvol.resize: %w", err)
		}
		return sdklvol.Resize(name, vg, size)
	})

	// ── synchronize ────────────────────────────────────────────────────
	r.Register("synchronize.sync", func(args map[string]interface{}) (interface{}, error) {
		source, err := argString(args, "source")
		if err != nil {
			return nil, fmt.Errorf("synchronize.sync: %w", err)
		}
		dest, err := argString(args, "dest")
		if err != nil {
			return nil, fmt.Errorf("synchronize.sync: %w", err)
		}
		del, _ := argBool(args, "delete")
		compress, _ := argBool(args, "compress")
		return sdksync.Sync(source, dest, del, compress)
	})

	// ── fetch ──────────────────────────────────────────────────────────
	r.Register("fetch.file", func(args map[string]interface{}) (interface{}, error) {
		source, err := argString(args, "source")
		if err != nil {
			return nil, fmt.Errorf("fetch.file: %w", err)
		}
		dest, err := argString(args, "dest")
		if err != nil {
			return nil, fmt.Errorf("fetch.file: %w", err)
		}
		return sdkfetch.File(source, dest)
	})
	r.Register("fetch.url", func(args map[string]interface{}) (interface{}, error) {
		url, err := argString(args, "url")
		if err != nil {
			return nil, fmt.Errorf("fetch.url: %w", err)
		}
		dest, err := argString(args, "dest")
		if err != nil {
			return nil, fmt.Errorf("fetch.url: %w", err)
		}
		return sdkfetch.URL(url, dest)
	})

	// ── seboolean ──────────────────────────────────────────────────────
	r.Register("seboolean.list", func(args map[string]interface{}) (interface{}, error) {
		return sdksebool.List()
	})
	r.Register("seboolean.get", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("seboolean.get: %w", err)
		}
		return sdksebool.Get(name)
	})
	r.Register("seboolean.set", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("seboolean.set: %w", err)
		}
		state, err := argBool(args, "state")
		if err != nil {
			return nil, fmt.Errorf("seboolean.set: %w", err)
		}
		persistent, _ := argBool(args, "persistent")
		return sdksebool.Set(name, state, persistent)
	})

	// ── uri ──────────────────────────────────────────────────────────────
	r.Register("uri.do", func(args map[string]interface{}) (interface{}, error) {
		url, err := argString(args, "url")
		if err != nil {
			return nil, fmt.Errorf("uri.do: %w", err)
		}
		method := getStringArg(args, "method", "GET")
		headers := toStringMapArg(args, "headers")
		body := getStringArg(args, "body", "")
		timeoutMs, _ := argInt(args, "timeout_ms")
		if timeoutMs <= 0 {
			timeoutMs = 30000
		}
		return sdkuri.Do(url, method, headers, body, timeoutMs)
	})
	r.Register("uri.get", func(args map[string]interface{}) (interface{}, error) {
		url, err := argString(args, "url")
		if err != nil {
			return nil, fmt.Errorf("uri.get: %w", err)
		}
		return sdkuri.Get(url)
	})
	r.Register("uri.post", func(args map[string]interface{}) (interface{}, error) {
		url, err := argString(args, "url")
		if err != nil {
			return nil, fmt.Errorf("uri.post: %w", err)
		}
		return sdkuri.Post(url, args["body"])
	})
	r.Register("uri.put", func(args map[string]interface{}) (interface{}, error) {
		url, err := argString(args, "url")
		if err != nil {
			return nil, fmt.Errorf("uri.put: %w", err)
		}
		return sdkuri.Put(url, args["body"])
	})
	r.Register("uri.delete", func(args map[string]interface{}) (interface{}, error) {
		url, err := argString(args, "url")
		if err != nil {
			return nil, fmt.Errorf("uri.delete: %w", err)
		}
		return sdkuri.Delete(url)
	})
	r.Register("uri.download", func(args map[string]interface{}) (interface{}, error) {
		url, err := argString(args, "url")
		if err != nil {
			return nil, fmt.Errorf("uri.download: %w", err)
		}
		dest, err := argString(args, "dest")
		if err != nil {
			return nil, fmt.Errorf("uri.download: %w", err)
		}
		return sdkuri.Download(url, dest)
	})

	// ── lineinfile ───────────────────────────────────────────────────────
	r.Register("lineinfile.present", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("lineinfile.ensure: %w", err)
		}
		line, err := argString(args, "line")
		if err != nil {
			return nil, fmt.Errorf("lineinfile.ensure: %w", err)
		}
		re := getStringArg(args, "regexp", "")
		create, _ := argBool(args, "create")
		return sdklineinfile.Ensure(path, line, re, create)
	})
	r.Register("lineinfile.absent", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("lineinfile.absent: %w", err)
		}
		re, err := argString(args, "regexp")
		if err != nil {
			return nil, fmt.Errorf("lineinfile.absent: %w", err)
		}
		return sdklineinfile.Absent(path, re)
	})

	// ── replace ──────────────────────────────────────────────────────────
	r.Register("replace.replace", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("replace.replace: %w", err)
		}
		pattern, err := argString(args, "pattern")
		if err != nil {
			return nil, fmt.Errorf("replace.replace: %w", err)
		}
		replacement, err := argString(args, "replacement")
		if err != nil {
			return nil, fmt.Errorf("replace.replace: %w", err)
		}
		regexpMode, _ := argBool(args, "regexp_mode")
		return sdkreplace.Replace(path, pattern, replacement, regexpMode)
	})

	// ── xml ──────────────────────────────────────────────────────────────
	r.Register("xml.get_element", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("xml.get_element: %w", err)
		}
		element, err := argString(args, "element")
		if err != nil {
			return nil, fmt.Errorf("xml.get_element: %w", err)
		}
		return sdkxml.GetElement(path, element)
	})
	r.Register("xml.set_element", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("xml.set_element: %w", err)
		}
		element, err := argString(args, "element")
		if err != nil {
			return nil, fmt.Errorf("xml.set_element: %w", err)
		}
		value, err := argString(args, "value")
		if err != nil {
			return nil, fmt.Errorf("xml.set_element: %w", err)
		}
		return sdkxml.SetElement(path, element, value)
	})

	// ── systemd ─────────────────────────────────────────────────────────────
	r.Register("systemd.is_active", func(args map[string]interface{}) (interface{}, error) {
		unit, err := argString(args, "unit")
		if err != nil {
			return nil, fmt.Errorf("systemd.is_active: %w", err)
		}
		return sdksystemd.IsActive(unit)
	})
	r.Register("systemd.is_enabled", func(args map[string]interface{}) (interface{}, error) {
		unit, err := argString(args, "unit")
		if err != nil {
			return nil, fmt.Errorf("systemd.is_enabled: %w", err)
		}
		return sdksystemd.IsEnabled(unit)
	})
	r.Register("systemd.enable", func(args map[string]interface{}) (interface{}, error) {
		unit, err := argString(args, "unit")
		if err != nil {
			return nil, fmt.Errorf("systemd.enable: %w", err)
		}
		return sdksystemd.Enable(unit)
	})
	r.Register("systemd.disable", func(args map[string]interface{}) (interface{}, error) {
		unit, err := argString(args, "unit")
		if err != nil {
			return nil, fmt.Errorf("systemd.disable: %w", err)
		}
		return sdksystemd.Disable(unit)
	})
	r.Register("systemd.start", func(args map[string]interface{}) (interface{}, error) {
		unit, err := argString(args, "unit")
		if err != nil {
			return nil, fmt.Errorf("systemd.start: %w", err)
		}
		return sdksystemd.Start(unit)
	})
	r.Register("systemd.stop", func(args map[string]interface{}) (interface{}, error) {
		unit, err := argString(args, "unit")
		if err != nil {
			return nil, fmt.Errorf("systemd.stop: %w", err)
		}
		return sdksystemd.Stop(unit)
	})
	r.Register("systemd.restart", func(args map[string]interface{}) (interface{}, error) {
		unit, err := argString(args, "unit")
		if err != nil {
			return nil, fmt.Errorf("systemd.restart: %w", err)
		}
		return sdksystemd.Restart(unit)
	})
	r.Register("systemd.reload", func(args map[string]interface{}) (interface{}, error) {
		unit, err := argString(args, "unit")
		if err != nil {
			return nil, fmt.Errorf("systemd.reload: %w", err)
		}
		return sdksystemd.Reload(unit)
	})
	r.Register("systemd.daemon_reload", func(args map[string]interface{}) (interface{}, error) {
		return sdksystemd.DaemonReload()
	})
	r.Register("systemd.mask", func(args map[string]interface{}) (interface{}, error) {
		unit, err := argString(args, "unit")
		if err != nil {
			return nil, fmt.Errorf("systemd.mask: %w", err)
		}
		return sdksystemd.Mask(unit)
	})
	r.Register("systemd.unmask", func(args map[string]interface{}) (interface{}, error) {
		unit, err := argString(args, "unit")
		if err != nil {
			return nil, fmt.Errorf("systemd.unmask: %w", err)
		}
		return sdksystemd.Unmask(unit)
	})
	r.Register("systemd.show", func(args map[string]interface{}) (interface{}, error) {
		unit, err := argString(args, "unit")
		if err != nil {
			return nil, fmt.Errorf("systemd.show: %w", err)
		}
		return sdksystemd.Show(unit)
	})
	r.Register("systemd.list", func(args map[string]interface{}) (interface{}, error) {
		unitType := getStringArg(args, "unit_type", "")
		return sdksystemd.List(unitType)
	})

	// ── patch ───────────────────────────────────────────────────────────────
	r.Register("patch.apply", func(args map[string]interface{}) (interface{}, error) {
		patchContent, err := argString(args, "patch_content")
		if err != nil {
			return nil, fmt.Errorf("patch.apply: %w", err)
		}
		reverse, _ := argBool(args, "reverse")
		return sdkpatch.Apply(patchContent, reverse)
	})
	r.Register("patch.dry_run", func(args map[string]interface{}) (interface{}, error) {
		patchContent, err := argString(args, "patch_content")
		if err != nil {
			return nil, fmt.Errorf("patch.dry_run: %w", err)
		}
		return sdkpatch.DryRun(patchContent)
	})

	// ── xattr ───────────────────────────────────────────────────────────────
	r.Register("xattr.get", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("xattr.get: %w", err)
		}
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("xattr.get: %w", err)
		}
		return sdkxattr.Get(path, name)
	})
	r.Register("xattr.set", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("xattr.set: %w", err)
		}
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("xattr.set: %w", err)
		}
		value, err := argString(args, "value")
		if err != nil {
			return nil, fmt.Errorf("xattr.set: %w", err)
		}
		return sdkxattr.Set(path, name, value)
	})
	r.Register("xattr.remove", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("xattr.remove: %w", err)
		}
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("xattr.remove: %w", err)
		}
		return sdkxattr.Remove(path, name)
	})
	r.Register("xattr.list", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("xattr.list: %w", err)
		}
		return sdkxattr.List(path)
	})

	// ── firewalld_zone ──────────────────────────────────────────────────────
	r.Register("firewalld_zone.get_default", func(args map[string]interface{}) (interface{}, error) {
		return sdkfirewalldzone.GetDefaultZone()
	})
	r.Register("firewalld_zone.set_default", func(args map[string]interface{}) (interface{}, error) {
		zone, err := argString(args, "zone")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.set_default: %w", err)
		}
		return sdkfirewalldzone.SetDefaultZone(zone)
	})
	r.Register("firewalld_zone.add_zone", func(args map[string]interface{}) (interface{}, error) {
		zone, err := argString(args, "zone")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.add_zone: %w", err)
		}
		return sdkfirewalldzone.AddZone(zone)
	})
	r.Register("firewalld_zone.remove_zone", func(args map[string]interface{}) (interface{}, error) {
		zone, err := argString(args, "zone")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.remove_zone: %w", err)
		}
		return sdkfirewalldzone.RemoveZone(zone)
	})
	r.Register("firewalld_zone.add_service", func(args map[string]interface{}) (interface{}, error) {
		zone, err := argString(args, "zone")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.add_service: %w", err)
		}
		svc, err := argString(args, "service")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.add_service: %w", err)
		}
		return sdkfirewalldzone.AddService(zone, svc)
	})
	r.Register("firewalld_zone.remove_service", func(args map[string]interface{}) (interface{}, error) {
		zone, err := argString(args, "zone")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.remove_service: %w", err)
		}
		svc, err := argString(args, "service")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.remove_service: %w", err)
		}
		return sdkfirewalldzone.RemoveService(zone, svc)
	})
	r.Register("firewalld_zone.add_port", func(args map[string]interface{}) (interface{}, error) {
		zone, err := argString(args, "zone")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.add_port: %w", err)
		}
		pp, err := argString(args, "port_protocol")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.add_port: %w", err)
		}
		return sdkfirewalldzone.AddPort(zone, pp)
	})
	r.Register("firewalld_zone.remove_port", func(args map[string]interface{}) (interface{}, error) {
		zone, err := argString(args, "zone")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.remove_port: %w", err)
		}
		pp, err := argString(args, "port_protocol")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.remove_port: %w", err)
		}
		return sdkfirewalldzone.RemovePort(zone, pp)
	})
	r.Register("firewalld_zone.add_rich_rule", func(args map[string]interface{}) (interface{}, error) {
		zone, err := argString(args, "zone")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.add_rich_rule: %w", err)
		}
		rule, err := argString(args, "rule")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.add_rich_rule: %w", err)
		}
		return sdkfirewalldzone.AddRichRule(zone, rule)
	})
	r.Register("firewalld_zone.remove_rich_rule", func(args map[string]interface{}) (interface{}, error) {
		zone, err := argString(args, "zone")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.remove_rich_rule: %w", err)
		}
		rule, err := argString(args, "rule")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.remove_rich_rule: %w", err)
		}
		return sdkfirewalldzone.RemoveRichRule(zone, rule)
	})
	r.Register("firewalld_zone.info", func(args map[string]interface{}) (interface{}, error) {
		zone, err := argString(args, "zone")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.info: %w", err)
		}
		return sdkfirewalldzone.Info(zone)
	})
	r.Register("firewalld_zone.list_zones", func(args map[string]interface{}) (interface{}, error) {
		return sdkfirewalldzone.ListZones()
	})

	// ── get_url ─────────────────────────────────────────────────────────────
	r.Register("get_url.download", func(args map[string]interface{}) (interface{}, error) {
		url, err := argString(args, "url")
		if err != nil {
			return nil, fmt.Errorf("get_url.download: %w", err)
		}
		dest, err := argString(args, "dest")
		if err != nil {
			return nil, fmt.Errorf("get_url.download: %w", err)
		}
		checksum := getStringArg(args, "checksum", "")
		force, _ := argBool(args, "force")
		return sdkgeturl.Download(url, dest, checksum, force)
	})

	// ── sys utilities ───────────────────────────────────────────────────────
	r.Register("sys.uuid", func(args map[string]interface{}) (interface{}, error) {
		return sys.UUID()
	})
	r.Register("sys.random_password", func(args map[string]interface{}) (interface{}, error) {
		length := 16
		if v, ok := args["length"]; ok {
			switch n := v.(type) {
			case int:
				length = n
			case float64:
				length = int(n)
			default:
				if parsed, err := argInt(args, "length"); err == nil {
					length = parsed
				}
			}
		}
		useSpecial := getBoolArg(args, "use_special", true)
		useNumbers := getBoolArg(args, "use_numbers", true)
		useUppercase := getBoolArg(args, "use_uppercase", true)
		return sys.RandomPassword(length, useSpecial, useNumbers, useUppercase)
	})

	// ── sys mac_address ──────────────────────────────────────────────────────
	r.Register("sys.mac_address", func(args map[string]interface{}) (interface{}, error) {
		iface, _ := args["interface"].(string)
		return sys.MACAddress(iface)
	})
	r.Register("sys.mac_addresses", func(args map[string]interface{}) (interface{}, error) {
		return sys.MACAddresses()
	})
	r.Register("sys.dmidecode", func(args map[string]interface{}) (interface{}, error) {
		return sys.Dmidecode()
	})
	r.Register("sys.lspci", func(args map[string]interface{}) (interface{}, error) {
		return sys.LsPci()
	})
	r.Register("sys.lsblk", func(args map[string]interface{}) (interface{}, error) {
		return sys.LsBlk()
	})
	r.Register("sys.lsusb", func(args map[string]interface{}) (interface{}, error) {
		return sys.LsUsb()
	})
	r.Register("sys.ip_route", func(args map[string]interface{}) (interface{}, error) {
		return sys.IpRoute()
	})
	r.Register("sys.ethtool", func(args map[string]interface{}) (interface{}, error) {
		iface, _ := argString(args, "iface")
		return sys.Ethtool(iface)
	})

	// ── modprobe.set_boot ──────────────────────────────────────────────────────
	r.Register("modprobe.set_boot", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		present := true
		if v, ok := args["present"]; ok {
			if b, ok := v.(bool); ok {
				present = b
			}
		}
		return sdkmodprobe.SetBoot(name, present)
	})

	// ── seport ─────────────────────────────────────────────────────────────────
	r.Register("seport.add", func(args map[string]interface{}) (interface{}, error) {
		seportType, _ := argString(args, "seport_type")
		proto, _ := argString(args, "protocol")
		port, _ := argString(args, "port")
		return sdkseport.Add(seportType, proto, port)
	})
	r.Register("seport.remove", func(args map[string]interface{}) (interface{}, error) {
		proto, _ := argString(args, "protocol")
		port, _ := argString(args, "port")
		return sdkseport.Remove(proto, port)
	})
	r.Register("seport.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkseport.List()
	})
	r.Register("seport.get", func(args map[string]interface{}) (interface{}, error) {
		proto, _ := argString(args, "protocol")
		port, _ := argString(args, "port")
		return sdkseport.Get(proto, port)
	})

	// ── sefcontext ─────────────────────────────────────────────────────────────
	r.Register("sefcontext.add", func(args map[string]interface{}) (interface{}, error) {
		filespec, _ := argString(args, "filespec")
		seType, _ := argString(args, "se_type")
		return sdksefcontext.Add(filespec, seType)
	})
	r.Register("sefcontext.modify", func(args map[string]interface{}) (interface{}, error) {
		filespec, _ := argString(args, "filespec")
		seType, _ := argString(args, "se_type")
		return sdksefcontext.Modify(filespec, seType)
	})
	r.Register("sefcontext.remove", func(args map[string]interface{}) (interface{}, error) {
		filespec, _ := argString(args, "filespec")
		return sdksefcontext.Remove(filespec)
	})
	r.Register("sefcontext.list", func(args map[string]interface{}) (interface{}, error) {
		return sdksefcontext.List()
	})
	r.Register("sefcontext.apply", func(args map[string]interface{}) (interface{}, error) {
		recursive := false
		if v, ok := args["recursive"]; ok {
			if b, ok := v.(bool); ok {
				recursive = b
			}
		}
		filespec, _ := argString(args, "filespec")
		return sdksefcontext.Apply(filespec, recursive)
	})

	// ── flatpak ─────────────────────────────────────────────────────────────
	r.Register("flatpak.install", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		from, _ := argString(args, "from")
		user, _ := argBool(args, "user")
		return sdkflatpak.Install(name, from, user)
	})
	r.Register("flatpak.remove", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		user, _ := argBool(args, "user")
		return sdkflatpak.Remove(name, user)
	})
	r.Register("flatpak.update", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		user, _ := argBool(args, "user")
		return sdkflatpak.Update(name, user)
	})
	r.Register("flatpak.list", func(args map[string]interface{}) (interface{}, error) {
		user, _ := argBool(args, "user")
		return sdkflatpak.List(user)
	})
	r.Register("flatpak.info", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		user, _ := argBool(args, "user")
		return sdkflatpak.Info(name, user)
	})
	r.Register("flatpak.run", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		var runArgs []string
		if argsList, ok := args["args"]; ok && argsList != nil {
			if list, ok := argsList.([]interface{}); ok {
				for _, item := range list {
					if s, ok := item.(string); ok {
						runArgs = append(runArgs, s)
					}
				}
			}
		}
		user, _ := argBool(args, "user")
		return sdkflatpak.Run(name, runArgs, user)
	})
	r.Register("flatpak.repair", func(args map[string]interface{}) (interface{}, error) {
		user, _ := argBool(args, "user")
		return sdkflatpak.Repair(user)
	})

	// ── zfs ─────────────────────────────────────────────────────────────
	r.Register("zfs.create", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		var props map[string]string
		if propsRaw, ok := args["properties"]; ok && propsRaw != nil {
			if m, ok := propsRaw.(map[string]interface{}); ok {
				props = make(map[string]string)
				for k, v := range m {
					if s, ok := v.(string); ok {
						props[k] = s
					}
				}
			}
		}
		return sdkzfs.Create(name, props)
	})
	r.Register("zfs.destroy", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		recursive, _ := argBool(args, "recursive")
		return sdkzfs.Destroy(name, recursive)
	})
	r.Register("zfs.set", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		property, _ := argString(args, "property")
		value, _ := argString(args, "value")
		return sdkzfs.Set(name, property, value)
	})
	r.Register("zfs.get", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		property, _ := argString(args, "property")
		return sdkzfs.Get(name, property)
	})
	r.Register("zfs.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkzfs.List()
	})
	r.Register("zfs.exists", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		exists, err := sdkzfs.Exists(name)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"exists": exists}, nil
	})
	r.Register("zfs.list_pools", func(args map[string]interface{}) (interface{}, error) {
		return sdkzfs.ListPools()
	})
	r.Register("zfs.get_pool_status", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdkzfs.GetPoolStatus(name)
	})
	r.Register("zfs.snapshot", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		snapName, _ := argString(args, "snapshot_name")
		return sdkzfs.Snapshot(name, snapName)
	})
	r.Register("zfs.destroy_snapshot", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		snapName, _ := argString(args, "snapshot_name")
		return sdkzfs.DestroySnapshot(name, snapName)
	})

	// ── nmcli ─────────────────────────────────────────────────────────────
	r.Register("nmcli.add", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		connType, _ := argString(args, "type")
		var settings map[string]string
		if settingsRaw, ok := args["settings"]; ok && settingsRaw != nil {
			if m, ok := settingsRaw.(map[string]interface{}); ok {
				settings = make(map[string]string)
				for k, v := range m {
					if s, ok := v.(string); ok {
						settings[k] = s
					}
				}
			}
		}
		return sdknmcli.Add(name, connType, settings)
	})
	r.Register("nmcli.modify", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		var settings map[string]string
		if settingsRaw, ok := args["settings"]; ok && settingsRaw != nil {
			if m, ok := settingsRaw.(map[string]interface{}); ok {
				settings = make(map[string]string)
				for k, v := range m {
					if s, ok := v.(string); ok {
						settings[k] = s
					}
				}
			}
		}
		return sdknmcli.Modify(name, settings)
	})
	r.Register("nmcli.delete", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdknmcli.Delete(name)
	})
	r.Register("nmcli.up", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdknmcli.Up(name)
	})
	r.Register("nmcli.down", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdknmcli.Down(name)
	})
	r.Register("nmcli.list", func(args map[string]interface{}) (interface{}, error) {
		return sdknmcli.List()
	})
	r.Register("nmcli.show", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdknmcli.Show(name)
	})
	r.Register("nmcli.list_devices", func(args map[string]interface{}) (interface{}, error) {
		return sdknmcli.ListDevices()
	})
	r.Register("nmcli.reload", func(args map[string]interface{}) (interface{}, error) {
		return sdknmcli.Reload()
	})
	r.Register("nmcli.get_general_status", func(args map[string]interface{}) (interface{}, error) {
		return sdknmcli.GetGeneralStatus()
	})

	// ── crypttab ──────────────────────────────────────────────────────────
	r.Register("crypttab.add", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		device, _ := argString(args, "device")
		keyFile, _ := argString(args, "key_file")
		options, _ := argString(args, "options")
		return sdkcrypttab.Add(name, device, keyFile, options)
	})
	r.Register("crypttab.remove", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdkcrypttab.Remove(name)
	})
	r.Register("crypttab.modify", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		device, _ := argString(args, "device")
		keyFile, _ := argString(args, "key_file")
		options, _ := argString(args, "options")
		return sdkcrypttab.Modify(name, device, keyFile, options)
	})
	r.Register("crypttab.get", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdkcrypttab.Get(name)
	})
	r.Register("crypttab.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkcrypttab.List()
	})
	r.Register("crypttab.exists", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		exists, err := sdkcrypttab.Exists(name)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"exists": exists}, nil
	})
	r.Register("crypttab.validate", func(args map[string]interface{}) (interface{}, error) {
		return sdkcrypttab.Validate()
	})
	r.Register("crypttab.backup", func(args map[string]interface{}) (interface{}, error) {
		backupDir, _ := argString(args, "backup_dir")
		return sdkcrypttab.Backup(backupDir)
	})

	// ── sysfs ─────────────────────────────────────────────────────────────
	r.Register("sysfs.read", func(args map[string]interface{}) (interface{}, error) {
		path, _ := argString(args, "path")
		value, err := sdksysfs.Read(path)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"value": value}, nil
	})
	r.Register("sysfs.write", func(args map[string]interface{}) (interface{}, error) {
		path, _ := argString(args, "path")
		value, _ := argString(args, "value")
		return sdksysfs.Write(path, value)
	})
	r.Register("sysfs.exists", func(args map[string]interface{}) (interface{}, error) {
		path, _ := argString(args, "path")
		exists, err := sdksysfs.Exists(path)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"exists": exists}, nil
	})
	r.Register("sysfs.get", func(args map[string]interface{}) (interface{}, error) {
		path, _ := argString(args, "path")
		return sdksysfs.Get(path)
	})
	r.Register("sysfs.list", func(args map[string]interface{}) (interface{}, error) {
		dirPath, _ := argString(args, "dir_path")
		return sdksysfs.List(dirPath)
	})
	r.Register("sysfs.set_device_power", func(args map[string]interface{}) (interface{}, error) {
		devicePath, _ := argString(args, "device_path")
		state, _ := argString(args, "state")
		return sdksysfs.SetDevicePower(devicePath, state)
	})
	r.Register("sysfs.get_device_power", func(args map[string]interface{}) (interface{}, error) {
		devicePath, _ := argString(args, "device_path")
		state, err := sdksysfs.GetDevicePower(devicePath)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"state": state}, nil
	})
	r.Register("sysfs.set_kernel_parameter", func(args map[string]interface{}) (interface{}, error) {
		param, _ := argString(args, "param")
		value, _ := argString(args, "value")
		return sdksysfs.SetKernelParameter(param, value)
	})
	r.Register("sysfs.get_kernel_parameter", func(args map[string]interface{}) (interface{}, error) {
		param, _ := argString(args, "param")
		value, err := sdksysfs.GetKernelParameter(param)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"value": value}, nil
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
