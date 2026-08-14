package os

import (
	"fmt"
	"strings"

	"github.com/opslang/opslang/pkg/vm"
	"github.com/opslang/opslang/stdlib/net/ssh"
)

// makeTopProcs 创建 sys.top_processes 函数。
// 用法: sys.top_processes(host, [user], [password], [n])
// 远程执行 ps aux --sort=-%mem，返回 Top N 进程列表。
func makeTopProcs(pool *ssh.Pool) vm.Value {
	return vm.Value{
		Type: vm.TypeFunction,
		Fn: &vm.FuncValue{
			Name:      "sys.top_processes",
			IsBuiltin: true,
			Builtin: func(args []vm.Value) (vm.Value, error) {
				if len(args) < 1 || args[0].Type != vm.TypeString {
					return vm.Value{}, &vm.RuntimeError{
						Message: "sys.top_processes(host, [user], [password], [n]) 需要至少 1 个参数",
					}
				}
				host := args[0].Str
				user := getStrArg(args, 1, "root")
				password := getStrArg(args, 2, "")

				n := int64(10)
				if len(args) >= 4 && args[3].Type == vm.TypeInt {
					n = args[3].Int
					if n <= 0 {
						n = 10
					}
				}

				cmd := fmt.Sprintf("ps aux --sort=-%%mem | head -%d", n+1)
				output, exitCode, err := remoteExec(pool, host, user, password, cmd)
				if err != nil && exitCode == 255 {
					return vm.Value{Type: vm.TypeMap, Map: map[string]vm.Value{
						"error": {Type: vm.TypeString, Str: fmt.Sprintf("SSH 执行失败: %v", err)},
					}}, nil
				}

				return parsePsOutput(output), nil
			},
		},
	}
}

// parsePsOutput 解析 ps aux 输出为进程列表。
// 期望字段: USER PID %CPU %MEM VSZ RSS TTY STAT START TIME COMMAND
func parsePsOutput(output string) vm.Value {
	var procs []vm.Value
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for i, line := range lines {
		if i == 0 {
			continue // skip header
		}
		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}

		// 从第 10 个字段开始拼接 command（command 可能包含空格）
		command := strings.Join(fields[10:], " ")

		procs = append(procs, vm.Value{
			Type: vm.TypeMap,
			Map: map[string]vm.Value{
				"pid":     {Type: vm.TypeString, Str: fields[1]},
				"user":    {Type: vm.TypeString, Str: fields[0]},
				"cpu_pct": {Type: vm.TypeFloat, Float: parseFloat64(fields[2])},
				"mem_pct": {Type: vm.TypeFloat, Float: parseFloat64(fields[3])},
				"command": {Type: vm.TypeString, Str: command},
			},
		})
	}

	return vm.Value{Type: vm.TypeArray, Arr: procs}
}
