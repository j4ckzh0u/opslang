package os

import (
	"fmt"
	"strings"

	"github.com/opslang/opslang/pkg/vm"
	"github.com/opslang/opslang/stdlib/net/ssh"
)

// makeMemInfo 创建 sys.mem_info 函数。
// 用法: sys.mem_info(host, [user], [password])
// 远程执行 free -m 并解析，返回 {total_mb, used_mb, free_mb, available_mb, use_pct}。
func makeMemInfo(pool *ssh.Pool) vm.Value {
	return vm.Value{
		Type: vm.TypeFunction,
		Fn: &vm.FuncValue{
			Name:      "sys.mem_info",
			IsBuiltin: true,
			Builtin: func(args []vm.Value) (vm.Value, error) {
				if len(args) < 1 || args[0].Type != vm.TypeString {
					return vm.Value{}, &vm.RuntimeError{
						Message: "sys.mem_info(host, [user], [password]) 需要至少 1 个参数",
					}
				}
				host := args[0].Str
				user := getStrArg(args, 1, "root")
				password := getStrArg(args, 2, "")

				cmd := "free -m"
				output, exitCode, err := remoteExec(pool, host, user, password, cmd)
				if err != nil && exitCode == 255 {
					return vm.Value{Type: vm.TypeMap, Map: map[string]vm.Value{
						"error": {Type: vm.TypeString, Str: fmt.Sprintf("SSH 执行失败: %v", err)},
					}}, nil
				}

				return parseMemOutput(output), nil
			},
		},
	}
}

// parseMemOutput 解析 free -m 输出。
// 期望格式（第一行为 header，第二行为 Mem:）:
//
//	              total        used        free      shared  buff/cache   available
//	Mem:           7982        3211        1024         312        3746        4201
func parseMemOutput(output string) vm.Value {
	lines := strings.Split(strings.TrimSpace(output), "\n")

	totalMB, usedMB, freeMB, availableMB := int64(0), int64(0), int64(0), int64(0)
	usePct := int64(0)

	for _, line := range lines {
		if !strings.HasPrefix(line, "Mem:") {
			continue
		}
		fields := strings.Fields(line)
		// fields: [Mem:, total, used, free, shared, buff/cache, available]
		if len(fields) >= 7 {
			totalMB = parseInt64(fields[1])
			usedMB = parseInt64(fields[2])
			freeMB = parseInt64(fields[3])
			availableMB = parseInt64(fields[6])
		} else if len(fields) >= 4 {
			// Fallback for older free versions with fewer columns
			totalMB = parseInt64(fields[1])
			usedMB = parseInt64(fields[2])
			freeMB = parseInt64(fields[3])
		}
		break
	}

	if totalMB > 0 {
		usePct = (usedMB * 100) / totalMB
	}

	return vm.Value{Type: vm.TypeMap, Map: map[string]vm.Value{
		"total_mb":     {Type: vm.TypeInt, Int: totalMB},
		"used_mb":      {Type: vm.TypeInt, Int: usedMB},
		"free_mb":      {Type: vm.TypeInt, Int: freeMB},
		"available_mb": {Type: vm.TypeInt, Int: availableMB},
		"use_pct":      {Type: vm.TypeInt, Int: usePct},
	}}
}
