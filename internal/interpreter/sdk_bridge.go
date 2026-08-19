// Package interpreter - SDK bridge: registers ops-core-sdk functions as interpreter builtins.
package interpreter

import (
	"encoding/json"
	"fmt"

	"github.com/opslang/opslang/internal/opsspec"
	sdkaptrepo "github.com/opslang/opslang/pkg/ops-core-sdk/apt_repo"
	sdkarchive "github.com/opslang/opslang/pkg/ops-core-sdk/archive"
	sdkcron "github.com/opslang/opslang/pkg/ops-core-sdk/cron"
	sdkdisk "github.com/opslang/opslang/pkg/ops-core-sdk/disk"
	sdkdocker "github.com/opslang/opslang/pkg/ops-core-sdk/docker"
	sdkfile "github.com/opslang/opslang/pkg/ops-core-sdk/file"
	sdkfirewalld "github.com/opslang/opslang/pkg/ops-core-sdk/firewalld"
	sdkgit "github.com/opslang/opslang/pkg/ops-core-sdk/git"
	sdkgroup "github.com/opslang/opslang/pkg/ops-core-sdk/group"
	sdkhosts "github.com/opslang/opslang/pkg/ops-core-sdk/hosts"
	sdkjson "github.com/opslang/opslang/pkg/ops-core-sdk/json"
	sdkkernel "github.com/opslang/opslang/pkg/ops-core-sdk/kernel"
	sdkknownhosts "github.com/opslang/opslang/pkg/ops-core-sdk/known_hosts"
	sdklimits "github.com/opslang/opslang/pkg/ops-core-sdk/limits"
	sdklocale "github.com/opslang/opslang/pkg/ops-core-sdk/locale"
	sdklogrotate "github.com/opslang/opslang/pkg/ops-core-sdk/logrotate"
	sdknet "github.com/opslang/opslang/pkg/ops-core-sdk/net"
	sdkntp "github.com/opslang/opslang/pkg/ops-core-sdk/ntp"
	sdkpip "github.com/opslang/opslang/pkg/ops-core-sdk/pip"
	opspkg "github.com/opslang/opslang/pkg/ops-core-sdk/pkg"
	sdkprocess "github.com/opslang/opslang/pkg/ops-core-sdk/process"
	sdkresolv "github.com/opslang/opslang/pkg/ops-core-sdk/resolv"
	sdkselinux "github.com/opslang/opslang/pkg/ops-core-sdk/selinux"
	sdkservice "github.com/opslang/opslang/pkg/ops-core-sdk/service"
	sdkssh "github.com/opslang/opslang/pkg/ops-core-sdk/ssh"
	sdksys "github.com/opslang/opslang/pkg/ops-core-sdk/sys"
	sdksysctl "github.com/opslang/opslang/pkg/ops-core-sdk/sysctl"
	sdktime "github.com/opslang/opslang/pkg/ops-core-sdk/time"
	sdkuser "github.com/opslang/opslang/pkg/ops-core-sdk/user"
	sdkyaml "github.com/opslang/opslang/pkg/ops-core-sdk/yaml"
	sdkyumrepo "github.com/opslang/opslang/pkg/ops-core-sdk/yum_repo"
)

// SDKBuiltinNames returns every SDK function name registered by
// RegisterSDKBuiltins. Used by cross-engine consistency tests.
func SDKBuiltinNames() []string {
	interp := New(nil)
	RegisterSDKBuiltins(interp)
	names := make([]string, 0, len(interp.builtins))
	for name := range interp.builtins {
		names = append(names, name)
	}
	return names
}

// structToMap converts any struct to map[string]interface{} via JSON roundtrip.
// This is needed because the interpreter's member access only supports map[string]interface{}.
func structToMap(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("structToMap marshal: %w", err)
	}
	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("structToMap unmarshal: %w", err)
	}
	return result, nil
}

// RegisterSDKBuiltins registers all ops-core-sdk functions into the interpreter.
func RegisterSDKBuiltins(interp *Interpreter) {
	// ── sys.* ──────────────────────────────────────────────────────────
	interp.builtins["sys.hostname"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.Hostname()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["sys.cpu.usage"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.GetCPUUsage()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["sys.cpu.info"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.GetCPUInfo()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["sys.cpu.count"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.GetCPUCount()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["sys.memory.info"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.GetMemoryInfo()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["sys.disk.usage"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("sys.disk.usage() requires 1 argument (path)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("sys.disk.usage(): argument must be string")
		}
		r, err := sdksys.GetDiskUsage(path)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["sys.disk.partitions"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.GetDiskPartitions()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["sys.load"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.GetLoadAvg()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["sys.os"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.GetHostInfo()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["sys.uptime"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.Uptime()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["sys.users"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.Users()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["sys.net.interfaces"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.GetNetInterfaces()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── file.* ─────────────────────────────────────────────────────────
	interp.builtins["file.read"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("file.read() requires 1 argument (path)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.read(): argument must be string")
		}
		r, err := sdkfile.Read(path)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["file.write"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("file.write() requires 2 arguments (path, content)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.write(): path must be string")
		}
		content, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("file.write(): content must be string")
		}
		r, err := sdkfile.Write(path, content)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["file.exists"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("file.exists() requires 1 argument (path)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.exists(): argument must be string")
		}
		r, err := sdkfile.Exists(path)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["file.copy"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("file.copy() requires 2 arguments (src, dst)")
		}
		src, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.copy(): src must be string")
		}
		dst, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("file.copy(): dst must be string")
		}
		r, err := sdkfile.Copy(src, dst)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["file.move"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("file.move() requires 2 arguments (src, dst)")
		}
		src, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.move(): src must be string")
		}
		dst, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("file.move(): dst must be string")
		}
		r, err := sdkfile.Move(src, dst)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["file.delete"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("file.delete() requires 1 argument (path)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.delete(): argument must be string")
		}
		r, err := sdkfile.Delete(path)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["file.stat"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("file.stat() requires 1 argument (path)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.stat(): argument must be string")
		}
		r, err := sdkfile.Stat(path)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["file.list"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("file.list() requires 1 argument (dir)")
		}
		dir, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.list(): argument must be string")
		}
		r, err := sdkfile.List(dir)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["file.mkdir"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("file.mkdir() requires 1 argument (path)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.mkdir(): argument must be string")
		}
		r, err := sdkfile.Mkdir(path)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["file.checksum"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("file.checksum() requires 2 arguments (path, algo)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.checksum(): path must be string")
		}
		algo, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("file.checksum(): algo must be string")
		}
		r, err := sdkfile.Checksum(path, algo)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["file.distribute"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("file.distribute() requires at least 2 arguments (source, targets)")
		}
		source, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.distribute(): source must be string")
		}
		targetsRaw, ok := args[1].([]interface{})
		if !ok {
			return nil, fmt.Errorf("file.distribute(): targets must be a list")
		}
		var targets []sdkfile.DistributeTarget
		for i, item := range targetsRaw {
			m, ok := item.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("file.distribute(): target %d must be a dict", i)
			}
			t := sdkfile.DistributeTarget{}
			if h, ok := m["host"].(string); ok {
				t.Host = h
			}
			if p, ok := m["port"].(float64); ok {
				t.Port = int(p)
			}
			if u, ok := m["user"].(string); ok {
				t.User = u
			}
			if d, ok := m["dest"].(string); ok {
				t.Dest = d
			}
			targets = append(targets, t)
		}

		opts := sdkfile.DistributeOptions{}
		if len(args) >= 3 {
			if optsMap, ok := args[2].(map[string]interface{}); ok {
				if v, ok := optsMap["checksum"].(bool); ok {
					opts.Checksum = v
				}
				if v, ok := optsMap["mode"].(string); ok {
					opts.Mode = v
				}
				if v, ok := optsMap["parallel"].(float64); ok {
					opts.Parallel = int(v)
				}
				if v, ok := optsMap["retries"].(float64); ok {
					opts.Retries = int(v)
				}
			}
		}

		r, err := sdkfile.Distribute(source, targets, opts)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["file.collect"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("file.collect() requires at least 2 arguments (source, targets)")
		}
		source, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.collect(): source must be string")
		}
		targetsRaw, ok := args[1].([]interface{})
		if !ok {
			return nil, fmt.Errorf("file.collect(): targets must be a list")
		}
		var targets []sdkfile.CollectTarget
		for i, item := range targetsRaw {
			m, ok := item.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("file.collect(): target %d must be a dict", i)
			}
			t := sdkfile.CollectTarget{}
			if h, ok := m["host"].(string); ok {
				t.Host = h
			}
			if p, ok := m["port"].(float64); ok {
				t.Port = int(p)
			}
			if u, ok := m["user"].(string); ok {
				t.User = u
			}
			if s, ok := m["source"].(string); ok {
				t.Source = s
			}
			targets = append(targets, t)
		}

		opts := sdkfile.CollectOptions{}
		if len(args) >= 3 {
			if optsMap, ok := args[2].(map[string]interface{}); ok {
				if v, ok := optsMap["dest_dir"].(string); ok {
					opts.DestDir = v
				}
				if v, ok := optsMap["parallel"].(float64); ok {
					opts.Parallel = int(v)
				}
				if v, ok := optsMap["retries"].(float64); ok {
					opts.Retries = int(v)
				}
			}
		}

		r, err := sdkfile.Collect(source, targets, opts)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── net.* ──────────────────────────────────────────────────────────
	interp.builtins["net.http_get"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("net.http_get() requires 1 argument (url)")
		}
		url, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("net.http_get(): argument must be string")
		}
		r, err := sdknet.HTTPGet(url)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["net.http_post"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("net.http_post() requires 2 arguments (url, body)")
		}
		url, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("net.http_post(): url must be string")
		}
		body, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("net.http_post(): body must be string")
		}
		r, err := sdknet.HTTPPost(url, body)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["net.tcp_check"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("net.tcp_check() requires 2 arguments (host, port)")
		}
		host, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("net.tcp_check(): host must be string")
		}
		portF, err := toFloat(args[1])
		if err != nil {
			return nil, fmt.Errorf("net.tcp_check(): port must be number")
		}
		r, err := sdknet.TCPConnect(host, int(portF))
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["net.dns_lookup"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("net.dns_lookup() requires 1 argument (domain)")
		}
		domain, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("net.dns_lookup(): argument must be string")
		}
		r, err := sdknet.DNSLookup(domain)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["net.interfaces"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdknet.Interfaces()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── process.* ──────────────────────────────────────────────────────
	interp.builtins["process.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkprocess.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["process.find_by_name"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("process.find_by_name() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("process.find_by_name(): argument must be string")
		}
		r, err := sdkprocess.FindByName(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["process.find_by_port"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("process.find_by_port() requires 1 argument (port)")
		}
		portF, err := toFloat(args[0])
		if err != nil {
			return nil, fmt.Errorf("process.find_by_port(): argument must be number")
		}
		r, err := sdkprocess.FindByPort(int(portF))
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["process.exec"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("process.exec() requires at least 1 argument (command)")
		}
		cmd, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("process.exec(): command must be string")
		}
		var cmdArgs []string
		for _, a := range args[1:] {
			cmdArgs = append(cmdArgs, fmt.Sprintf("%v", a))
		}
		r, err := sdkprocess.Exec(cmd, cmdArgs)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── service.* ──────────────────────────────────────────────────────
	interp.builtins["service.status"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("service.status() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("service.status(): argument must be string")
		}
		r, err := sdkservice.Status(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["service.start"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("service.start() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("service.start(): argument must be string")
		}
		r, err := sdkservice.Start(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["service.stop"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("service.stop() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("service.stop(): argument must be string")
		}
		r, err := sdkservice.Stop(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["service.restart"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("service.restart() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("service.restart(): argument must be string")
		}
		r, err := sdkservice.Restart(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["service.enable"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("service.enable() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("service.enable(): argument must be string")
		}
		r, err := sdkservice.Enable(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── selinux.* ────────────────────────────────────────────────────────
	interp.builtins["selinux.get"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkselinux.Get()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["selinux.set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("selinux.set() requires 1 argument (mode)")
		}
		mode, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("selinux.set(): argument must be string")
		}
		r, err := sdkselinux.Set(mode)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── time.* ─────────────────────────────────────────────────────────
	interp.builtins["time.now"] = func(args ...interface{}) (interface{}, error) {
		r := sdktime.Now()
		return structToMap(r)
	}

	interp.builtins["time.format"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("time.format() requires 2 arguments (unix, layout)")
		}
		unixF, err := toFloat(args[0])
		if err != nil {
			return nil, fmt.Errorf("time.format(): unix must be number")
		}
		layout, strOk := args[1].(string)
		if !strOk {
			return nil, fmt.Errorf("time.format(): layout must be string")
		}
		r, err := sdktime.Format(int64(unixF), layout)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["time.since"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("time.since() requires 1 argument (unix)")
		}
		unixF, err := toFloat(args[0])
		if err != nil {
			return nil, fmt.Errorf("time.since(): argument must be number")
		}
		r := sdktime.Since(int64(unixF))
		return structToMap(r)
	}

	interp.builtins["time.sleep"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("time.sleep() requires 1 argument (ms)")
		}
		msF, err := toFloat(args[0])
		if err != nil {
			return nil, fmt.Errorf("time.sleep(): argument must be number")
		}
		r, err := sdktime.Sleep(int(msF))
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── json.* ─────────────────────────────────────────────────────────
	interp.builtins["json.encode"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("json.encode() requires 1 argument")
		}
		r, err := sdkjson.Encode(args[0])
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["json.decode"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("json.decode() requires 1 argument (string)")
		}
		input, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("json.decode(): argument must be string")
		}
		r, err := sdkjson.Decode(input)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── known_hosts.* ────────────────────────────────────────────────────
	interp.builtins["known_hosts.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkknownhosts.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["known_hosts.check"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("known_hosts.check() requires 1 argument (host)")
		}
		host, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("known_hosts.check(): argument must be string")
		}
		r, err := sdkknownhosts.Check(host)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["known_hosts.add"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("known_hosts.add() requires 1 argument (host)")
		}
		host, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("known_hosts.add(): argument must be string")
		}
		r, err := sdkknownhosts.Add(host)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["known_hosts.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("known_hosts.remove() requires 1 argument (host)")
		}
		host, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("known_hosts.remove(): argument must be string")
		}
		r, err := sdkknownhosts.Remove(host)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── yaml.* ─────────────────────────────────────────────────────────
	interp.builtins["yaml.encode"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("yaml.encode() requires 1 argument")
		}
		r, err := sdkyaml.Encode(args[0])
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["yaml.decode"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("yaml.decode() requires 1 argument (string)")
		}
		input, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("yaml.decode(): argument must be string")
		}
		r, err := sdkyaml.Decode(input)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── file.* (additions) ────────────────────────────────────────────
	interp.builtins["file.append"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("file.append() requires 2 arguments (path, content)")
		}
		path, _ := args[0].(string)
		content, _ := args[1].(string)
		if path == "" {
			return nil, fmt.Errorf("file.append(): path must be string")
		}
		r, err := sdkfile.Append(path, content)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["file.chmod"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("file.chmod() requires 2 arguments (path, mode)")
		}
		path, _ := args[0].(string)
		modeStr, _ := args[1].(string)
		var mode uint64
		if _, err := fmt.Sscanf(modeStr, "%o", &mode); err != nil {
			return nil, fmt.Errorf("file.chmod(): mode must be an octal string like \"0755\"")
		}
		r, err := sdkfile.Chmod(path, uint32(mode))
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["file.template"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("file.template() requires at least 1 argument (path)")
		}
		path, _ := args[0].(string)
		vars := map[string]interface{}{}
		if len(args) >= 2 {
			m, ok := args[1].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("file.template(): vars must be a dict")
			}
			vars = m
		}
		r, err := sdkfile.Template(path, vars)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── process.kill ─────────────────────────────────────────────────
	interp.builtins["process.kill"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("process.kill() requires at least 1 argument (pid)")
		}
		pidF, err := toFloat(args[0])
		if err != nil {
			return nil, fmt.Errorf("process.kill(): pid must be number")
		}
		signal := "TERM"
		if len(args) >= 2 {
			if s, ok := args[1].(string); ok && s != "" {
				signal = s
			}
		}
		r, err := sdkprocess.Kill(int(pidF), signal)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── service.disable ──────────────────────────────────────────────
	interp.builtins["service.disable"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("service.disable() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("service.disable(): argument must be string")
		}
		r, err := sdkservice.Disable(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── pkg.* ────────────────────────────────────────────────────────
	interp.builtins["pkg.install"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("pkg.install() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("pkg.install(): argument must be string")
		}
		r, _ := opspkg.Install(name)
		return structToMap(r)
	}

	interp.builtins["pkg.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("pkg.remove() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("pkg.remove(): argument must be string")
		}
		r, _ := opspkg.Remove(name)
		return structToMap(r)
	}

	interp.builtins["pkg.info"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("pkg.info() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("pkg.info(): argument must be string")
		}
		r, err := opspkg.Info(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["pkg.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := opspkg.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── time.parse / time.diff ───────────────────────────────────────
	interp.builtins["time.parse"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("time.parse() requires 2 arguments (layout, value)")
		}
		layout, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("time.parse(): layout must be string")
		}
		value, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("time.parse(): value must be string")
		}
		r, err := sdktime.Parse(layout, value)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["time.diff"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("time.diff() requires 2 arguments (t1, t2)")
		}
		t1F, err := toFloat(args[0])
		if err != nil {
			return nil, fmt.Errorf("time.diff(): t1 must be number")
		}
		t2F, err := toFloat(args[1])
		if err != nil {
			return nil, fmt.Errorf("time.diff(): t2 must be number")
		}
		r := sdktime.Diff(int64(t1F), int64(t2F))
		return structToMap(r)
	}

	// ── user.* ─────────────────────────────────────────────────────────
	interp.builtins["user.info"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("user.info() requires 1 argument (username)")
		}
		username, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("user.info(): username must be string")
		}
		r, err := sdkuser.Info(username)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["user.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkuser.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["user.add"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("user.add() requires at least 1 argument (username)")
		}
		username, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("user.add(): username must be string")
		}
		opts := toStringMap(args, 1)
		r, err := sdkuser.Add(username, opts)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["user.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("user.remove() requires at least 1 argument (username)")
		}
		username, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("user.remove(): username must be string")
		}
		removeHome := false
		if len(args) > 1 {
			removeHome, _ = args[1].(bool)
		}
		r, err := sdkuser.Remove(username, removeHome)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["user.modify"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("user.modify() requires at least 1 argument (username)")
		}
		username, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("user.modify(): username must be string")
		}
		opts := toStringMap(args, 1)
		r, err := sdkuser.Modify(username, opts)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["user.exists"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("user.exists() requires 1 argument (username)")
		}
		username, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("user.exists(): username must be string")
		}
		r, err := sdkuser.Exists(username)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── group.* ────────────────────────────────────────────────────────
	interp.builtins["group.info"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("group.info() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("group.info(): name must be string")
		}
		r, err := sdkgroup.Info(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["group.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkgroup.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["group.add"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("group.add() requires at least 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("group.add(): name must be string")
		}
		opts := toStringMap(args, 1)
		r, err := sdkgroup.Add(name, opts)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["group.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("group.remove() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("group.remove(): name must be string")
		}
		r, err := sdkgroup.Remove(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["group.exists"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("group.exists() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("group.exists(): name must be string")
		}
		r, err := sdkgroup.Exists(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── cron.* ─────────────────────────────────────────────────────────
	interp.builtins["cron.list"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("cron.list() requires 1 argument (user)")
		}
		user, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("cron.list(): user must be string")
		}
		r, err := sdkcron.List(user)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["cron.add"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("cron.add() requires 2 arguments (user, entry)")
		}
		user, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("cron.add(): user must be string")
		}
		entryMap, ok := args[1].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("cron.add(): entry must be a dict")
		}
		entry := sdkcron.CronEntry{
			Minute:     mapStr(entryMap, "minute", "*"),
			Hour:       mapStr(entryMap, "hour", "*"),
			DayOfMonth: mapStr(entryMap, "day_of_month", "*"),
			Month:      mapStr(entryMap, "month", "*"),
			DayOfWeek:  mapStr(entryMap, "day_of_week", "*"),
			Command:    mapStr(entryMap, "command", ""),
		}
		r, err := sdkcron.Add(user, entry)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["cron.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("cron.remove() requires 2 arguments (user, line_match)")
		}
		user, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("cron.remove(): user must be string")
		}
		lineMatch, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("cron.remove(): line_match must be string")
		}
		r, err := sdkcron.Remove(user, lineMatch)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── sysctl.* ───────────────────────────────────────────────────────
	interp.builtins["sysctl.get"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("sysctl.get() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("sysctl.get(): name must be string")
		}
		r, err := sdksysctl.Get(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sysctl.set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("sysctl.set() requires 2 arguments (name, value)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("sysctl.set(): name must be string")
		}
		value, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("sysctl.set(): value must be string")
		}
		r, err := sdksysctl.Set(name, value)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sysctl.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksysctl.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── git.* ──────────────────────────────────────────────────────────
	interp.builtins["git.clone"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("git.clone() requires at least 2 arguments (url, dest)")
		}
		url, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("git.clone(): url must be string")
		}
		dest, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("git.clone(): dest must be string")
		}
		opts := toStringMap(args, 2)
		r, err := sdkgit.Clone(url, dest, opts)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["git.pull"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("git.pull() requires at least 1 argument (repo_path)")
		}
		repoPath, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("git.pull(): repo_path must be string")
		}
		remote := getStringArgBridge(args, 1, "origin")
		branch := getStringArgBridge(args, 2, "")
		r, err := sdkgit.Pull(repoPath, remote, branch)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── file.lineinfile ────────────────────────────────────────────────
	interp.builtins["file.lineinfile"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("file.lineinfile() requires at least 2 arguments (path, line)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.lineinfile(): path must be string")
		}
		line, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("file.lineinfile(): line must be string")
		}
		present := true
		if len(args) > 2 {
			present, _ = args[2].(bool)
		}
		rx := ""
		if len(args) > 3 {
			rx, _ = args[3].(string)
		}
		r, err := sdkfile.LineInFile(path, line, present, rx)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── net.wait_for ───────────────────────────────────────────────────
	interp.builtins["net.wait_for"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("net.wait_for() requires at least 2 arguments (host, port)")
		}
		host, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("net.wait_for(): host must be string")
		}
		portF, err := toFloat(args[1])
		if err != nil {
			return nil, fmt.Errorf("net.wait_for(): port must be number")
		}
		timeout := 30
		if len(args) > 2 {
			tF, err := toFloat(args[2])
			if err != nil {
				return nil, fmt.Errorf("net.wait_for(): timeout must be number")
			}
			timeout = int(tF)
		}
		r, err := sdknet.WaitFor(host, int(portF), timeout)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── sys.mount / sys.unmount / sys.list_mounts ──────────────────────
	interp.builtins["sys.mount"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("sys.mount() requires at least 3 arguments (device, mountpoint, fs_type)")
		}
		device, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("sys.mount(): device must be string")
		}
		mountpoint, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("sys.mount(): mountpoint must be string")
		}
		fsType, ok := args[2].(string)
		if !ok {
			return nil, fmt.Errorf("sys.mount(): fs_type must be string")
		}
		opts := toStringMap(args, 3)
		r, err := sdksys.Mount(device, mountpoint, fsType, opts)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sys.unmount"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("sys.unmount() requires 1 argument (mountpoint)")
		}
		mountpoint, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("sys.unmount(): mountpoint must be string")
		}
		r, err := sdksys.Unmount(mountpoint)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sys.list_mounts"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.ListMounts()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── sys.hostname_set ───────────────────────────────────────────────
	interp.builtins["sys.hostname_set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("sys.hostname_set() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("sys.hostname_set(): name must be string")
		}
		r, err := sdksys.HostnameSet(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── firewall.rule ──────────────────────────────────────────────────
	interp.builtins["firewall.rule"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("firewall.rule() requires at least 2 arguments (action, protocol)")
		}
		action, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("firewall.rule(): action must be string")
		}
		protocol, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("firewall.rule(): protocol must be string")
		}
		port := 0
		if len(args) > 2 {
			pF, err := toFloat(args[2])
			if err != nil {
				return nil, fmt.Errorf("firewall.rule(): port must be number")
			}
			port = int(pF)
		}
		source := ""
		if len(args) > 3 {
			source, _ = args[3].(string)
		}
		r, err := sdksys.FirewallRule(action, protocol, port, source)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── firewalld ────────────────────────────────────────────────────────
	interp.builtins["firewalld.get"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkfirewalld.Get()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["firewalld.start"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkfirewalld.Start()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["firewalld.stop"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkfirewalld.Stop()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["firewalld.restart"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkfirewalld.Restart()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["firewalld.enable"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkfirewalld.Enable()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["firewalld.disable"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkfirewalld.Disable()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["firewalld.list_zones"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkfirewalld.ListZones()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["firewalld.reload"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkfirewalld.Reload()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── file.find ────────────────────────────────────────────────────────
	interp.builtins["file.find"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("file.find() requires at least 1 argument (paths)")
		}
		opts := sdkfile.FindOptions{}
		// paths: string or []string
		switch v := args[0].(type) {
		case string:
			opts.Paths = []string{v}
		case []interface{}:
			for _, p := range v {
				if s, ok := p.(string); ok {
					opts.Paths = append(opts.Paths, s)
				}
			}
		case []string:
			opts.Paths = v
		}
		// patterns: optional string or []string
		if len(args) > 1 {
			switch v := args[1].(type) {
			case string:
				if v != "" {
					opts.Patterns = []string{v}
				}
			case []interface{}:
				for _, p := range v {
					if s, ok := p.(string); ok {
						opts.Patterns = append(opts.Patterns, s)
					}
				}
			case []string:
				opts.Patterns = v
			}
		}
		// regex
		if len(args) > 2 {
			if s, ok := args[2].(string); ok {
				opts.Regex = s
			}
		}
		// file_type
		if len(args) > 3 {
			if s, ok := args[3].(string); ok {
				opts.FileType = s
			}
		}
		// max_depth
		if len(args) > 4 {
			if f, err := toFloat(args[4]); err == nil {
				opts.MaxDepth = int(f)
			}
		}
		// age
		if len(args) > 5 {
			if f, err := toFloat(args[5]); err == nil {
				opts.Age = int64(f)
			}
		}
		// size
		if len(args) > 6 {
			if f, err := toFloat(args[6]); err == nil {
				opts.Size = int64(f)
			}
		}
		r, err := sdkfile.Find(opts)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── file.replace ─────────────────────────────────────────────────────
	interp.builtins["file.replace"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("file.replace() requires at least 3 arguments (path, pattern, replacement)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.replace(): path must be string")
		}
		pattern, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("file.replace(): pattern must be string")
		}
		replacement, ok := args[2].(string)
		if !ok {
			return nil, fmt.Errorf("file.replace(): replacement must be string")
		}
		after := getStringArgBridge(args, 3, "")
		before := getStringArgBridge(args, 4, "")
		r, err := sdkfile.Replace(path, pattern, replacement, after, before)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── file.blockinfile ─────────────────────────────────────────────────
	interp.builtins["file.blockinfile"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("file.blockinfile() requires at least 3 arguments (path, marker, content)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.blockinfile(): path must be string")
		}
		marker, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("file.blockinfile(): marker must be string")
		}
		content, ok := args[2].(string)
		if !ok {
			return nil, fmt.Errorf("file.blockinfile(): content must be string")
		}
		present := true
		if len(args) > 3 {
			present = opsBool(args[3])
		}
		insertAfter := getStringArgBridge(args, 4, "")
		insertBefore := getStringArgBridge(args, 5, "")
		r, err := sdkfile.BlockInFile(path, marker, content, present, insertAfter, insertBefore)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── file.ini_get ─────────────────────────────────────────────────────
	interp.builtins["file.ini_get"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("file.ini_get() requires 3 arguments (path, section, key)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.ini_get(): path must be string")
		}
		section, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("file.ini_get(): section must be string")
		}
		key, ok := args[2].(string)
		if !ok {
			return nil, fmt.Errorf("file.ini_get(): key must be string")
		}
		r, err := sdkfile.IniGet(path, section, key)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── file.ini_set ─────────────────────────────────────────────────────
	interp.builtins["file.ini_set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("file.ini_set() requires 4 arguments (path, section, key, value)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.ini_set(): path must be string")
		}
		section, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("file.ini_set(): section must be string")
		}
		key, ok := args[2].(string)
		if !ok {
			return nil, fmt.Errorf("file.ini_set(): key must be string")
		}
		value, ok := args[3].(string)
		if !ok {
			return nil, fmt.Errorf("file.ini_set(): value must be string")
		}
		r, err := sdkfile.IniSet(path, section, key, value)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── archive.create ───────────────────────────────────────────────────
	interp.builtins["archive.create"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("archive.create() requires 2 arguments (dest, sources)")
		}
		dest, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("archive.create(): dest must be string")
		}
		var sources []string
		switch v := args[1].(type) {
		case string:
			sources = []string{v}
		case []interface{}:
			for _, s := range v {
				if str, ok := s.(string); ok {
					sources = append(sources, str)
				}
			}
		case []string:
			sources = v
		default:
			return nil, fmt.Errorf("archive.create(): sources must be string or list")
		}
		r, err := sdkarchive.Create(dest, sources)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── archive.extract ──────────────────────────────────────────────────
	interp.builtins["archive.extract"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("archive.extract() requires 2 arguments (src, dest)")
		}
		src, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("archive.extract(): src must be string")
		}
		dest, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("archive.extract(): dest must be string")
		}
		r, err := sdkarchive.Extract(src, dest)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── net.download ─────────────────────────────────────────────────────
	interp.builtins["net.download"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("net.download() requires at least 2 arguments (url, dest)")
		}
		url, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("net.download(): url must be string")
		}
		dest, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("net.download(): dest must be string")
		}
		algo := getStringArgBridge(args, 2, "")
		expected := getStringArgBridge(args, 3, "")
		r, err := sdknet.Download(url, dest, algo, expected)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── net.wait_for_connection ──────────────────────────────────────────
	interp.builtins["net.wait_for_connection"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("net.wait_for_connection() requires at least 2 arguments (host, port)")
		}
		host, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("net.wait_for_connection(): host must be string")
		}
		portF, err := toFloat(args[1])
		if err != nil {
			return nil, fmt.Errorf("net.wait_for_connection(): port must be number")
		}
		port := int(portF)
		timeout := 30
		if len(args) > 2 {
			tF, err := toFloat(args[2])
			if err != nil {
				return nil, fmt.Errorf("net.wait_for_connection(): timeout must be number")
			}
			timeout = int(tF)
		}
		r, err := sdknet.WaitForConnection(host, port, timeout)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── ntp.* ────────────────────────────────────────────────────────────
	interp.builtins["ntp.get"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkntp.Get()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["ntp.set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("ntp.set() requires 1 argument (server)")
		}
		server, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("ntp.set(): argument must be string")
		}
		r, err := sdkntp.Set(server)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── sys.timezone_get ─────────────────────────────────────────────────
	interp.builtins["sys.timezone_get"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.TimezoneGet()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── sys.timezone_set ─────────────────────────────────────────────────
	interp.builtins["sys.timezone_set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("sys.timezone_set() requires 1 argument (timezone)")
		}
		tz, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("sys.timezone_set(): timezone must be string")
		}
		r, err := sdksys.TimezoneSet(tz)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── sys.reboot ───────────────────────────────────────────────────────
	interp.builtins["sys.reboot"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.Reboot()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── ssh.authorized_key_add ───────────────────────────────────────────
	interp.builtins["ssh.authorized_key_add"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("ssh.authorized_key_add() requires at least 2 arguments (user, key)")
		}
		user, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("ssh.authorized_key_add(): user must be string")
		}
		key, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("ssh.authorized_key_add(): key must be string")
		}
		exclusive := false
		if len(args) > 2 {
			exclusive = opsBool(args[2])
		}
		r, err := sdkssh.AuthorizedKeyAdd(user, key, exclusive)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── ssh.authorized_key_remove ────────────────────────────────────────
	interp.builtins["ssh.authorized_key_remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("ssh.authorized_key_remove() requires 2 arguments (user, key)")
		}
		user, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("ssh.authorized_key_remove(): user must be string")
		}
		key, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("ssh.authorized_key_remove(): key must be string")
		}
		r, err := sdkssh.AuthorizedKeyRemove(user, key)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── ssh.authorized_key_list ──────────────────────────────────────────
	interp.builtins["ssh.authorized_key_list"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("ssh.authorized_key_list() requires 1 argument (user)")
		}
		user, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("ssh.authorized_key_list(): user must be string")
		}
		r, err := sdkssh.AuthorizedKeyList(user)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── kernel.module_list ───────────────────────────────────────────────
	interp.builtins["kernel.module_list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkkernel.ModuleList()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── kernel.module_load ───────────────────────────────────────────────
	interp.builtins["kernel.module_load"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("kernel.module_load() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("kernel.module_load(): name must be string")
		}
		r, err := sdkkernel.ModuleLoad(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── kernel.module_unload ─────────────────────────────────────────────
	interp.builtins["kernel.module_unload"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("kernel.module_unload() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("kernel.module_unload(): name must be string")
		}
		r, err := sdkkernel.ModuleUnload(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── limits.* ─────────────────────────────────────────────────────────
	interp.builtins["limits.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdklimits.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["limits.get"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("limits.get() requires 1 argument (domain)")
		}
		domain, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("limits.get(): argument must be string")
		}
		r, err := sdklimits.Get(domain)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["limits.set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("limits.set() requires 4 arguments (domain, type, item, value)")
		}
		domain, _ := args[0].(string)
		typ, _ := args[1].(string)
		item, _ := args[2].(string)
		value, _ := args[3].(string)
		r, err := sdklimits.Set(domain, typ, item, value)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["limits.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("limits.remove() requires 1 argument (domain)")
		}
		domain, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("limits.remove(): argument must be string")
		}
		r, err := sdklimits.Remove(domain)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── disk.filesystem ──────────────────────────────────────────────────
	interp.builtins["disk.filesystem"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("disk.filesystem() requires at least 1 argument (device)")
		}
		device, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("disk.filesystem(): device must be string")
		}
		fsType := getStringArgBridge(args, 1, "ext4")
		r, err := sdkdisk.FilesystemCreate(device, fsType)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── disk.part_list ───────────────────────────────────────────────────
	interp.builtins["disk.part_list"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("disk.part_list() requires 1 argument (device)")
		}
		device, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("disk.part_list(): device must be string")
		}
		r, err := sdkdisk.PartList(device)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── docker.* ────────────────────────────────────────────────────────
	interp.builtins["docker.container_list"] = func(args ...interface{}) (interface{}, error) {
		all := false
		if len(args) > 0 {
			all, _ = args[0].(bool)
		}
		r, err := sdkdocker.ContainerList(all)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["docker.container_exists"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("docker.container_exists() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("docker.container_exists(): name must be string")
		}
		r, err := sdkdocker.ContainerExists(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["docker.container_run"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("docker.container_run() requires at least 2 arguments (name, image)")
		}
		name, _ := args[0].(string)
		image, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("docker.container_run(): image must be string")
		}
		opts := toStringMap(args, 2)
		r, err := sdkdocker.ContainerRun(name, image, opts)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["docker.container_stop"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("docker.container_stop() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("docker.container_stop(): name must be string")
		}
		r, err := sdkdocker.ContainerStop(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["docker.container_remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("docker.container_remove() requires at least 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("docker.container_remove(): name must be string")
		}
		force := false
		if len(args) > 1 {
			force, _ = args[1].(bool)
		}
		r, err := sdkdocker.ContainerRemove(name, force)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["docker.image_list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkdocker.ImageList()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["docker.image_pull"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("docker.image_pull() requires 1 argument (image)")
		}
		image, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("docker.image_pull(): image must be string")
		}
		r, err := sdkdocker.ImagePull(image)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["docker.image_remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("docker.image_remove() requires at least 1 argument (image)")
		}
		image, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("docker.image_remove(): image must be string")
		}
		force := false
		if len(args) > 1 {
			force, _ = args[1].(bool)
		}
		r, err := sdkdocker.ImageRemove(image, force)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── hosts.* ─────────────────────────────────────────────────────────
	interp.builtins["hosts.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkhosts.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["hosts.exists"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("hosts.exists() requires 1 argument (hostname)")
		}
		hostname, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("hosts.exists(): hostname must be string")
		}
		r, err := sdkhosts.Exists(hostname)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["hosts.add"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("hosts.add() requires 2 arguments (ip, hostnames)")
		}
		ip, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("hosts.add(): ip must be string")
		}
		hostnamesRaw, ok := args[1].([]interface{})
		if !ok {
			return nil, fmt.Errorf("hosts.add(): hostnames must be array")
		}
		hostnames := make([]string, len(hostnamesRaw))
		for i, h := range hostnamesRaw {
			hostnames[i], _ = h.(string)
		}
		r, err := sdkhosts.Add(ip, hostnames)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["hosts.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("hosts.remove() requires 1 argument (hostnames)")
		}
		hostnamesRaw, ok := args[0].([]interface{})
		if !ok {
			return nil, fmt.Errorf("hosts.remove(): hostnames must be array")
		}
		hostnames := make([]string, len(hostnamesRaw))
		for i, h := range hostnamesRaw {
			hostnames[i], _ = h.(string)
		}
		r, err := sdkhosts.Remove(hostnames)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── locale.* ────────────────────────────────────────────────────────
	interp.builtins["locale.get"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdklocale.Get()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["locale.available"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdklocale.Available()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["locale.set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("locale.set() requires 1 argument (locale)")
		}
		locale, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("locale.set(): locale must be string")
		}
		r, err := sdklocale.Set(locale)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── pip.* ───────────────────────────────────────────────────────────
	interp.builtins["pip.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkpip.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["pip.exists"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("pip.exists() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("pip.exists(): name must be string")
		}
		r, err := sdkpip.Exists(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["pip.install"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("pip.install() requires at least 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("pip.install(): name must be string")
		}
		version := ""
		if len(args) > 1 {
			version, _ = args[1].(string)
		}
		r, err := sdkpip.Install(name, version)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["pip.uninstall"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("pip.uninstall() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("pip.uninstall(): name must be string")
		}
		r, err := sdkpip.Uninstall(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── apt_repo.* ──────────────────────────────────────────────────────
	interp.builtins["apt_repo.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkaptrepo.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apt_repo.exists"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("apt_repo.exists() requires 1 argument (uri)")
		}
		uri, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("apt_repo.exists(): uri must be string")
		}
		r, err := sdkaptrepo.Exists(uri)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apt_repo.add"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("apt_repo.add() requires 3 arguments (uri, dist, components)")
		}
		uri, _ := args[0].(string)
		dist, _ := args[1].(string)
		comps, _ := args[2].(string)
		r, err := sdkaptrepo.Add(uri, dist, comps)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apt_repo.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("apt_repo.remove() requires 1 argument (uri)")
		}
		uri, _ := args[0].(string)
		r, err := sdkaptrepo.Remove(uri)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apt_repo.update"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkaptrepo.Update()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── logrotate.* ─────────────────────────────────────────────────────
	interp.builtins["logrotate.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdklogrotate.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["logrotate.get"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("logrotate.get() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdklogrotate.Get(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["logrotate.set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("logrotate.set() requires at least 4 arguments (name, pattern, frequency, rotate)")
		}
		name, _ := args[0].(string)
		pattern, _ := args[1].(string)
		freq, _ := args[2].(string)
		rotate := int(opsFloat(args, 3))
		compress := false
		if len(args) > 4 {
			compress = opsBool(args[4])
		}
		postRotate := ""
		if len(args) > 5 {
			postRotate, _ = args[5].(string)
		}
		r, err := sdklogrotate.Set(name, pattern, freq, rotate, compress, postRotate)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["logrotate.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("logrotate.remove() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdklogrotate.Remove(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── resolv.* ────────────────────────────────────────────────────────
	interp.builtins["resolv.get"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkresolv.Get()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["resolv.set"] = func(args ...interface{}) (interface{}, error) {
		var nameservers, search, options []string
		domain := ""
		if len(args) > 0 {
			if l, ok := args[0].([]interface{}); ok {
				for _, v := range l {
					if s, ok := v.(string); ok {
						nameservers = append(nameservers, s)
					}
				}
			}
		}
		if len(args) > 1 {
			if l, ok := args[1].([]interface{}); ok {
				for _, v := range l {
					if s, ok := v.(string); ok {
						search = append(search, s)
					}
				}
			}
		}
		if len(args) > 2 {
			if l, ok := args[2].([]interface{}); ok {
				for _, v := range l {
					if s, ok := v.(string); ok {
						options = append(options, s)
					}
				}
			}
		}
		if len(args) > 3 {
			domain, _ = args[3].(string)
		}
		r, err := sdkresolv.Set(nameservers, search, options, domain)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["resolv.add_nameserver"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("resolv.add_nameserver() requires 1 argument (nameserver)")
		}
		ns, _ := args[0].(string)
		r, err := sdkresolv.AddNameserver(ns)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["resolv.remove_nameserver"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("resolv.remove_nameserver() requires 1 argument (nameserver)")
		}
		ns, _ := args[0].(string)
		r, err := sdkresolv.RemoveNameserver(ns)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── yum_repo.* ──────────────────────────────────────────────────────
	interp.builtins["yum_repo.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkyumrepo.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["yum_repo.exists"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("yum_repo.exists() requires 1 argument (id)")
		}
		id, _ := args[0].(string)
		r, err := sdkyumrepo.Exists(id)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["yum_repo.add"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("yum_repo.add() requires at least 3 arguments (id, name, base_url)")
		}
		id, _ := args[0].(string)
		name, _ := args[1].(string)
		baseURL, _ := args[2].(string)
		gpgCheck := false
		if len(args) > 3 {
			gpgCheck = opsBool(args[3])
		}
		gpgKey := ""
		if len(args) > 4 {
			gpgKey, _ = args[4].(string)
		}
		r, err := sdkyumrepo.Add(id, name, baseURL, gpgCheck, gpgKey)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["yum_repo.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("yum_repo.remove() requires 1 argument (id)")
		}
		id, _ := args[0].(string)
		r, err := sdkyumrepo.Remove(id)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
}

// toStringMap extracts a map[string]string from args starting at the given index.
// If the arg at idx is a map[string]interface{}, values are converted to strings.
// Returns an empty map if no arg is present at idx.
func toStringMap(args []interface{}, idx int) map[string]string {
	result := make(map[string]string)
	if idx >= len(args) {
		return result
	}
	m, ok := args[idx].(map[string]interface{})
	if !ok {
		return result
	}
	for k, v := range m {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result
}

// mapStr extracts a string value from a map with a default fallback.
func mapStr(m map[string]interface{}, key string, def string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
		return fmt.Sprintf("%v", v)
	}
	return def
}

// getStringArgBridge extracts a string argument at the given index with a default.
func getStringArgBridge(args []interface{}, idx int, def string) string {
	if idx >= len(args) {
		return def
	}
	if s, ok := args[idx].(string); ok {
		return s
	}
	return def
}

// verifyBridgeCoverage is a self-check that every function the canonical
// opsspec table promises for the controller (interpreter) is registered.
// It panics at init if the bridge and the spec drift apart — the two used
// to disagree silently, which made docs lie.
func init() {
	registered := make(map[string]bool)
	for _, name := range SDKBuiltinNames() {
		registered[name] = true
	}
	for _, f := range opsspec.Funcs {
		if !registered[f.Name] {
			panic(fmt.Sprintf("opsspec/interpreter mismatch: %s is in the spec but not registered in the interpreter bridge", f.Name))
		}
	}
}

// opsBool extracts a bool from an interface{} value.
func opsBool(v interface{}) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

// opsFloat extracts a float64 from an interface{} value at the given index.
func opsFloat(args []interface{}, idx int) float64 {
	if idx >= len(args) {
		return 0
	}
	if f, ok := args[idx].(float64); ok {
		return f
	}
	return 0
}
