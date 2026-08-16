// Package interpreter - SDK bridge: registers ops-core-sdk functions as interpreter builtins.
package interpreter

import (
	"encoding/json"
	"fmt"

	"github.com/opslang/opslang/internal/opsspec"
	sdkfile "github.com/opslang/opslang/pkg/ops-core-sdk/file"
	sdkjson "github.com/opslang/opslang/pkg/ops-core-sdk/json"
	sdknet "github.com/opslang/opslang/pkg/ops-core-sdk/net"
	opspkg "github.com/opslang/opslang/pkg/ops-core-sdk/pkg"
	sdkprocess "github.com/opslang/opslang/pkg/ops-core-sdk/process"
	sdkservice "github.com/opslang/opslang/pkg/ops-core-sdk/service"
	sdksys "github.com/opslang/opslang/pkg/ops-core-sdk/sys"
	sdktime "github.com/opslang/opslang/pkg/ops-core-sdk/time"
	sdkyaml "github.com/opslang/opslang/pkg/ops-core-sdk/yaml"
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
		r, err := opspkg.Install(name)
		if err != nil {
			return nil, err
		}
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
		r, err := opspkg.Remove(name)
		if err != nil {
			return nil, err
		}
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
