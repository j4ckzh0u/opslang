package os

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/opslang/opslang/pkg/vm"
	"github.com/opslang/opslang/stdlib/net/ssh"
)

// validServiceName matches allowed service name characters (alphanumeric, -, _, ., @)
var validServiceName = regexp.MustCompile(`^[a-zA-Z0-9_\-\.@]+$`)

// makeSvcStatus 创建 sys.service_status 函数。
// 用法: sys.service_status(host, name, [user], [password])
// 远程执行 systemctl status，返回 {name, active, enabled, status_text}。
func makeSvcStatus(pool *ssh.Pool) vm.Value {
	return vm.Value{
		Type: vm.TypeFunction,
		Fn: &vm.FuncValue{
			Name:      "sys.service_status",
			IsBuiltin: true,
			Builtin: func(args []vm.Value) (vm.Value, error) {
				if len(args) < 2 || args[0].Type != vm.TypeString || args[1].Type != vm.TypeString {
					return vm.Value{}, &vm.RuntimeError{
						Message: "sys.service_status(host, name, [user], [password]) 需要至少 2 个参数",
					}
				}
				host := args[0].Str
				name := args[1].Str
				if !validServiceName.MatchString(name) {
					return vm.Value{}, &vm.RuntimeError{
						Message: fmt.Sprintf("sys.service_status: 非法服务名 %q，仅允许字母数字和 -_.@", name),
					}
				}
				user := getStrArg(args, 2, "root")
				password := getStrArg(args, 3, "")

				cmd := fmt.Sprintf("systemctl status %s --no-pager 2>&1; echo '---EXIT:' $?", name)
				output, _, _ := remoteExec(pool, host, user, password, cmd)

				return parseServiceOutput(name, output), nil
			},
		},
	}
}

// parseServiceOutput 解析 systemctl status 输出。
func parseServiceOutput(name string, output string) vm.Value {
	active := false
	enabled := false
	statusText := "unknown"

	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Active: active (running) since ...
		if strings.HasPrefix(trimmed, "Active:") {
			active = strings.Contains(trimmed, "active (running)")
			// 提取状态文本，如 "active (running)" 或 "inactive (dead)"
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				statusText = strings.TrimSpace(parts[1])
				// 去掉 "since ..." 部分
				if idx := strings.Index(statusText, " since "); idx > 0 {
					statusText = statusText[:idx]
				}
			}
		}

		// Loaded: loaded; enabled; ...
		if strings.HasPrefix(trimmed, "Loaded:") {
			enabled = strings.Contains(trimmed, "enabled")
		}
	}

	return vm.Value{Type: vm.TypeMap, Map: map[string]vm.Value{
		"name":        {Type: vm.TypeString, Str: name},
		"active":      {Type: vm.TypeBool, Bool: active},
		"enabled":     {Type: vm.TypeBool, Bool: enabled},
		"status_text": {Type: vm.TypeString, Str: statusText},
	}}
}
