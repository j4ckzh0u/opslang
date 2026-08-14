package os

import (
	"fmt"
	"net"
	"time"

	"github.com/opslang/opslang/pkg/vm"
)

// makePortCheck 创建 sys.port_check 函数（本地执行）。
// 用法: sys.port_check(host, port)
// 使用 Go net.Dial 检查端口连通性，返回 {host, port, open}。
func makePortCheck() vm.Value {
	return vm.Value{
		Type: vm.TypeFunction,
		Fn: &vm.FuncValue{
			Name:      "sys.port_check",
			IsBuiltin: true,
			Builtin: func(args []vm.Value) (vm.Value, error) {
				if len(args) < 2 || args[0].Type != vm.TypeString {
					return vm.Value{}, &vm.RuntimeError{
						Message: "sys.port_check(host, port) 需要 2 个参数",
					}
				}
				host := args[0].Str

				var port int64
				if args[1].Type == vm.TypeInt {
					port = args[1].Int
				} else if args[1].Type == vm.TypeString {
					port = parseInt64(args[1].Str)
				} else {
					return vm.Value{}, &vm.RuntimeError{
						Message: "sys.port_check: port 必须是整数或字符串",
					}
				}

				addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
				conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
				open := false
				if err == nil {
					open = true
					conn.Close()
				}

				return vm.Value{Type: vm.TypeMap, Map: map[string]vm.Value{
					"host": {Type: vm.TypeString, Str: host},
					"port": {Type: vm.TypeInt, Int: port},
					"open": {Type: vm.TypeBool, Bool: open},
				}}, nil
			},
		},
	}
}

// makePing 创建 sys.ping 函数（本地执行）。
// 用法: sys.ping(host, [count])
// 使用 TCP 连接测量延迟，返回 {host, reachable, latency_ms}。
func makePing() vm.Value {
	return vm.Value{
		Type: vm.TypeFunction,
		Fn: &vm.FuncValue{
			Name:      "sys.ping",
			IsBuiltin: true,
			Builtin: func(args []vm.Value) (vm.Value, error) {
				if len(args) < 1 || args[0].Type != vm.TypeString {
					return vm.Value{}, &vm.RuntimeError{
						Message: "sys.ping(host, [count]) 需要至少 1 个参数",
					}
				}
				host := args[0].Str

				count := int64(3)
				if len(args) >= 2 && args[1].Type == vm.TypeInt {
					count = args[1].Int
					if count <= 0 {
						count = 3
					}
				}

				// 尝试默认 SSH 端口 22 进行 TCP ping
				addr := net.JoinHostPort(host, "22")
				var totalLatency float64
				successes := int64(0)

				for i := int64(0); i < count; i++ {
					start := time.Now()
					conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
					if err == nil {
						latency := time.Since(start).Seconds() * 1000
						totalLatency += latency
						successes++
						conn.Close()
					}
				}

				reachable := successes > 0
				avgLatency := 0.0
				if successes > 0 {
					avgLatency = totalLatency / float64(successes)
				}

				// 四舍五入到 2 位小数
				avgLatency = float64(int64(avgLatency*100+0.5)) / 100

				return vm.Value{Type: vm.TypeMap, Map: map[string]vm.Value{
					"host":       {Type: vm.TypeString, Str: host},
					"reachable":  {Type: vm.TypeBool, Bool: reachable},
					"latency_ms": {Type: vm.TypeFloat, Float: avgLatency},
				}}, nil
			},
		},
	}
}
