package os

import (
	"fmt"
	"strings"

	"github.com/opslang/opslang/pkg/vm"
	"github.com/opslang/opslang/stdlib/net/ssh"
)

// makeCPUInfo 创建 sys.cpu_info 函数。
// 用法: sys.cpu_info(host, [user], [password])
// 返回 {cores, model}。
func makeCPUInfo(pool *ssh.Pool) vm.Value {
	return vm.Value{
		Type: vm.TypeFunction,
		Fn: &vm.FuncValue{
			Name:      "sys.cpu_info",
			IsBuiltin: true,
			Builtin: func(args []vm.Value) (vm.Value, error) {
				if len(args) < 1 || args[0].Type != vm.TypeString {
					return vm.Value{}, &vm.RuntimeError{
						Message: "sys.cpu_info(host, [user], [password]) 需要至少 1 个参数",
					}
				}
				host := args[0].Str
				user := getStrArg(args, 1, "root")
				password := getStrArg(args, 2, "")

				cmd := `nproc && grep "model name" /proc/cpuinfo | head -1`
				output, exitCode, err := remoteExec(pool, host, user, password, cmd)
				if err != nil && exitCode == 255 {
					return vm.Value{Type: vm.TypeMap, Map: map[string]vm.Value{
						"error": {Type: vm.TypeString, Str: fmt.Sprintf("SSH 执行失败: %v", err)},
					}}, nil
				}

				return parseCPUOutput(output), nil
			},
		},
	}
}

// parseCPUOutput 解析 nproc + cpuinfo 输出。
func parseCPUOutput(output string) vm.Value {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	cores := int64(0)
	model := ""

	if len(lines) >= 1 {
		cores = parseInt64(lines[0])
	}
	if len(lines) >= 2 {
		// 格式: "model name\t: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz"
		parts := strings.SplitN(lines[1], ":", 2)
		if len(parts) == 2 {
			model = strings.TrimSpace(parts[1])
		}
	}

	return vm.Value{Type: vm.TypeMap, Map: map[string]vm.Value{
		"cores": {Type: vm.TypeInt, Int: cores},
		"model": {Type: vm.TypeString, Str: model},
	}}
}

// makeLoadAvg 创建 sys.load_average 函数。
// 用法: sys.load_average(host, [user], [password])
// 返回 {load1, load5, load15, cores}。
func makeLoadAvg(pool *ssh.Pool) vm.Value {
	return vm.Value{
		Type: vm.TypeFunction,
		Fn: &vm.FuncValue{
			Name:      "sys.load_average",
			IsBuiltin: true,
			Builtin: func(args []vm.Value) (vm.Value, error) {
				if len(args) < 1 || args[0].Type != vm.TypeString {
					return vm.Value{}, &vm.RuntimeError{
						Message: "sys.load_average(host, [user], [password]) 需要至少 1 个参数",
					}
				}
				host := args[0].Str
				user := getStrArg(args, 1, "root")
				password := getStrArg(args, 2, "")

				cmd := `cat /proc/loadavg && nproc`
				output, exitCode, err := remoteExec(pool, host, user, password, cmd)
				if err != nil && exitCode == 255 {
					return vm.Value{Type: vm.TypeMap, Map: map[string]vm.Value{
						"error": {Type: vm.TypeString, Str: fmt.Sprintf("SSH 执行失败: %v", err)},
					}}, nil
				}

				return parseLoadOutput(output), nil
			},
		},
	}
}

// parseLoadOutput 解析 /proc/loadavg + nproc 输出。
func parseLoadOutput(output string) vm.Value {
	lines := strings.Split(strings.TrimSpace(output), "\n")

	load1, load5, load15 := 0.0, 0.0, 0.0
	cores := int64(0)

	if len(lines) >= 1 {
		fields := strings.Fields(lines[0])
		if len(fields) >= 3 {
			load1 = parseFloat64(fields[0])
			load5 = parseFloat64(fields[1])
			load15 = parseFloat64(fields[2])
		}
	}
	if len(lines) >= 2 {
		cores = parseInt64(lines[1])
	}

	return vm.Value{Type: vm.TypeMap, Map: map[string]vm.Value{
		"load1":  {Type: vm.TypeFloat, Float: load1},
		"load5":  {Type: vm.TypeFloat, Float: load5},
		"load15": {Type: vm.TypeFloat, Float: load15},
		"cores":  {Type: vm.TypeInt, Int: cores},
	}}
}
