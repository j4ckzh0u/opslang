// Package stat provides Ansible stat module equivalent.
package stat

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"os/user"
	"strconv"
	"syscall"
	"time"
)

// StatResult is returned by Stat.
type StatResult struct {
	Exists     bool   `json:"exists"`
	Path       string `json:"path"`
	IsDir      bool   `json:"isdir"`
	IsLink     bool   `json:"islnk"`
	IsFIFO     bool   `json:"isfifo"`
	IsBlock    bool   `json:"isblk"`
	IsChar     bool   `json:"ischr"`
	IsSocket   bool   `json:"issock"`
	Mode       string `json:"mode"`
	UID        int    `json:"uid"`
	GID        int    `json:"gid"`
	Owner      string `json:"pw_name"`
	Group      string `json:"gr_name"`
	Size       int64  `json:"size"`
	Inode      uint64 `json:"inode"`
	Dev        uint64 `json:"dev"`
	NLink      uint64 `json:"nlink"`
	ModTime    string `json:"mtime"`
	AccessTime string `json:"atime"`
	ChangeTime string `json:"ctime"`
	MD5        string `json:"md5,omitempty"`
	SHA1       string `json:"sha1,omitempty"`
	SHA256     string `json:"sha256,omitempty"`
	Error      string `json:"error,omitempty"`
}

// Stat returns detailed file information.
func Stat(path string, getChecksum bool, checksumAlgo string) StatResult {
	if path == "" {
		return StatResult{Error: "path is required"}
	}
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return StatResult{Exists: false, Path: path}
		}
		return StatResult{Error: err.Error()}
	}

	result := StatResult{
		Exists:  true,
		Path:    path,
		IsDir:   info.IsDir(),
		IsLink:  info.Mode()&os.ModeSymlink != 0,
		IsFIFO:  info.Mode()&os.ModeNamedPipe != 0,
		IsBlock: info.Mode()&os.ModeDevice != 0,
		IsChar:  info.Mode()&os.ModeCharDevice != 0,
		IsSocket: info.Mode()&os.ModeSocket != 0,
		Mode:    info.Mode().String(),
		Size:    info.Size(),
		ModTime: info.ModTime().Format(time.RFC3339),
	}

	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		result.UID = int(stat.Uid)
		result.GID = int(stat.Gid)
		result.Inode = stat.Ino
		result.Dev = uint64(stat.Dev)
		result.NLink = uint64(stat.Nlink)
		// Use info.ModTime() for access/change as portable fallback
		result.AccessTime = info.ModTime().Format(time.RFC3339)
		result.ChangeTime = info.ModTime().Format(time.RFC3339)
		if u, err := user.LookupId(strconv.Itoa(result.UID)); err == nil {
			result.Owner = u.Username
		}
	}

	if getChecksum && !info.IsDir() {
		f, err := os.Open(path)
		if err == nil {
			defer f.Close()
			switch checksumAlgo {
			case "sha256", "":
				h := sha256.New()
				io.Copy(h, f)
				result.SHA256 = hex.EncodeToString(h.Sum(nil))
			case "md5":
				h := md5.New()
				io.Copy(h, f)
				result.MD5 = hex.EncodeToString(h.Sum(nil))
			case "sha1":
				h := sha1.New()
				io.Copy(h, f)
				result.SHA1 = hex.EncodeToString(h.Sum(nil))
			}
		}
	}
	return result
}
