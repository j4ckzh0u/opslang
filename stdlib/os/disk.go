package os

import (
	"fmt"
	"strings"

	"github.com/opslang/opslang/pkg/vm"
	"github.com/opslang/opslang/stdlib/net/ssh"
)

// makeDiskUsage 创建 sys.disk_usage 函数。
// 用法: sys.disk_usage(host, [user], [password], [mount])
// 远程执行 df -h 并解析，返回磁盘使用率列表。
func makeDiskUsage(pool *ssh.Pool) vm.Value {
	return vm.Value{
		Type: vm.TypeFunction,
		Fn: &vm.FuncValue{
			Name:      "sys.disk_usage",
			IsBuiltin: true,
			Builtin: func(args []vm.Value) (vm.Value, error) {
				if len(args) < 1 || args[0].Type != vm.TypeString {
					return vm.Value{}, &vm.RuntimeError{
						Message: "sys.disk_usage(host, [user], [password], [mount]) 需要至少 1 个参数",
					}
				}
				host := args[0].Str
				user := getStrArg(args, 1, "root")
				password := getStrArg(args, 2, "")
				mountFilter := getStrArg(args, 3, "")

				cmd := "df -h -x tmpfs -x devtmpfs -x squashfs 2>/dev/null || df -h"
				output, exitCode, err := remoteExec(pool, host, user, password, cmd)
				if err != nil && exitCode == 255 {
					return vm.Value{Type: vm.TypeMap, Map: map[string]vm.Value{
						"error": {Type: vm.TypeString, Str: fmt.Sprintf("SSH 执行失败: %v", err)},
					}}, nil
				}

				disks := parseDiskOutput(output, mountFilter)
				return vm.Value{Type: vm.TypeArray, Arr: disks}, nil
			},
		},
	}
}

// parseDiskOutput 解析 df -h 输出为 Value 数组。
func parseDiskOutput(output string, mountFilter string) []vm.Value {
	var disks []vm.Value
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for i, line := range lines {
		if i == 0 {
			continue // skip header
		}
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		mountpoint := fields[5]
		if mountFilter != "" && mountpoint != mountFilter {
			continue
		}

		usePctStr := strings.TrimSuffix(fields[4], "%")
		disks = append(disks, vm.Value{
			Type: vm.TypeMap,
			Map: map[string]vm.Value{
				"device":     {Type: vm.TypeString, Str: fields[0]},
				"size":       {Type: vm.TypeString, Str: fields[1]},
				"used":       {Type: vm.TypeString, Str: fields[2]},
				"avail":      {Type: vm.TypeString, Str: fields[3]},
				"use_pct":    {Type: vm.TypeInt, Int: parseInt64(usePctStr)},
				"mountpoint": {Type: vm.TypeString, Str: mountpoint},
			},
		})
	}
	return disks
}
