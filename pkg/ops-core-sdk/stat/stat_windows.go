//go:build windows

package stat

import (
	"os"
	"time"
)

func populatePlatformStat(result *StatResult, info os.FileInfo) {
	result.AccessTime = info.ModTime().Format(time.RFC3339)
	result.ChangeTime = info.ModTime().Format(time.RFC3339)
}
