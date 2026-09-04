//go:build !windows

package stat

import (
	"os"
	"os/user"
	"strconv"
	"syscall"
	"time"
)

func populatePlatformStat(result *StatResult, info os.FileInfo) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return
	}
	result.UID = int(stat.Uid)
	result.GID = int(stat.Gid)
	result.Inode = stat.Ino
	result.Dev = uint64(stat.Dev)
	result.NLink = uint64(stat.Nlink)
	result.AccessTime = info.ModTime().Format(time.RFC3339)
	result.ChangeTime = info.ModTime().Format(time.RFC3339)
	if u, err := user.LookupId(strconv.Itoa(result.UID)); err == nil {
		result.Owner = u.Username
	}
}
