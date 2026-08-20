// Package find provides Ansible find module equivalent.
// Find files matching patterns in directories.
package find

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FindResult contains matched files.
type FindResult struct {
	Files   []FileMatch `json:"files"`
	Matched int         `json:"matched"`
	Examined int        `json:"examined"`
	Error   string      `json:"error,omitempty"`
}

// FileMatch is a matched file.
type FileMatch struct {
	Path     string    `json:"path"`
	Name     string    `json:"filename"`
	Mode     string    `json:"mode"`
	IsDir    bool      `json:"isdir"`
	IsLink   bool      `json:"islnk"`
	Size     int64     `json:"size"`
	UID      int       `json:"uid"`
	GID      int       `json:"gid"`
	ModTime  time.Time `json:"mtime"`
}

// FindOptions controls file matching.
type FindOptions struct {
	Paths      []string // directories to search
	Patterns   []string // glob patterns (file names)
	FileType   string   // "file", "directory", "any"
	Recurse    bool
	Age        string   // e.g. "1d", "2w"
	AgeOp      string   // "lt" or "gt"
	Size       string   // e.g. "10m"
	SizeOp     string   // "lt" or "gt"
	Contains   string   // grep content
	Depth      int      // max recursion depth
	FollowLinks bool
}

// Find returns files matching options.
func Find(opts FindOptions) FindResult {
	if len(opts.Paths) == 0 {
		return FindResult{Error: "paths is required"}
	}
	if opts.FileType == "" {
		opts.FileType = "any"
	}

	var files []FileMatch
	var examined int

	for _, dir := range opts.Paths {
		info, err := os.Stat(dir)
		if err != nil {
			return FindResult{Error: err.Error()}
		}
		if !info.IsDir() {
			return FindResult{Error: dir + " is not a directory"}
		}

		walkFn := func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			depth := strings.Count(strings.TrimPrefix(path, dir), string(os.PathSeparator))
			if !opts.Recurse && depth > 1 {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if opts.Depth > 0 && depth > opts.Depth {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			examined++
			if !matchesFilter(d, path, opts) {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}
			fm := FileMatch{
				Path:    path,
				Name:    d.Name(),
				Mode:    info.Mode().String(),
				IsDir:   d.IsDir(),
				IsLink:  info.Mode()&os.ModeSymlink != 0,
				Size:    info.Size(),
				ModTime: info.ModTime(),
			}
			files = append(files, fm)
			return nil
		}
		_ = filepath.WalkDir(dir, walkFn)
	}

	return FindResult{Files: files, Matched: len(files), Examined: examined}
}

func matchesFilter(d os.DirEntry, path string, opts FindOptions) bool {
	if opts.FileType == "file" && d.IsDir() {
		return false
	}
	if opts.FileType == "directory" && !d.IsDir() {
		return false
	}
	if len(opts.Patterns) > 0 {
		matched := false
		for _, p := range opts.Patterns {
			if ok, _ := filepath.Match(p, d.Name()); ok {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}
