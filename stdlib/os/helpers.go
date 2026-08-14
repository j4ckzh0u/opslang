// Package os 实现运维原子操作 stdlib，注册为 sys 模块。
// 每个远程函数通过 SSH 执行命令并解析输出为结构化数据。
package os

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/opslang/opslang/pkg/vm"
	"github.com/opslang/opslang/stdlib/net/ssh"
)

// remoteExec 通过 SSH 连接池在远程主机上执行命令。
// 返回 (stdout, exitCode, error)。
func remoteExec(pool *ssh.Pool, host, user, password, cmd string) (string, int, error) {
	cfg := ssh.Config{
		Host:     host,
		User:     user,
		Password: password,
	}
	client, err := pool.Get(cfg)
	if err != nil {
		return fmt.Sprintf("SSH 连接失败: %v", err), 255, err
	}
	output, exitCode, runErr := client.CombinedOutput(cmd)
	if runErr != nil {
		// Any error (SIGPIPE/exit 141, connection reset/255, etc.):
		// evict the connection to avoid reusing a potentially dirty session.
		pool.Remove(cfg)
	}
	return output, exitCode, runErr
}

// localExec 在本地执行 shell 命令。
func localExec(command string) (string, int, error) {
	cmd := exec.Command("sh", "-c", command)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	exitCode := 0
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return "", 255, err
		}
	}
	return out.String(), exitCode, nil
}

// getStrArg 安全获取字符串参数，越界或类型不匹配时返回默认值。
func getStrArg(args []vm.Value, idx int, def string) string {
	if idx < len(args) && args[idx].Type == vm.TypeString {
		return args[idx].Str
	}
	return def
}

// parseInt64 将字符串解析为 int64，忽略错误。
func parseInt64(s string) int64 {
	n, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return n
}

// parseFloat64 将字符串解析为 float64，忽略错误。
func parseFloat64(s string) float64 {
	f, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return f
}
