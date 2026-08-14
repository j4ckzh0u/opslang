package os

import (
	"github.com/opslang/opslang/pkg/vm"
	"github.com/opslang/opslang/stdlib/net/ssh"
)

// Register 将 sys 模块注册到 VM globals。
// sys 模块提供运维原子操作函数：磁盘、CPU、内存、进程、服务、网络检查。
func Register(globals map[string]vm.Value, sshPool *ssh.Pool) {
	globals["sys"] = vm.Value{
		Type: vm.TypeMap,
		Map: map[string]vm.Value{
			"disk_usage":     makeDiskUsage(sshPool),
			"cpu_info":       makeCPUInfo(sshPool),
			"load_average":   makeLoadAvg(sshPool),
			"mem_info":       makeMemInfo(sshPool),
			"top_processes":  makeTopProcs(sshPool),
			"service_status": makeSvcStatus(sshPool),
			"port_check":     makePortCheck(),
			"ping":           makePing(),
		},
	}
}
