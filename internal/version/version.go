// Package version 提供版本信息
package version

import "fmt"

// 编译时注入的变量
var (
	Version   = "0.1.0-dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// Full 返回完整版本字符串
func Full() string {
	return fmt.Sprintf("OpsLang %s (build: %s, commit: %s)", Version, BuildTime, GitCommit)
}
