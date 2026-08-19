// File find operations — Ansible find module equivalent.
package file

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// FindMatch represents a single matched file entry.
type FindMatch struct {
	Path  string `json:"path"`
	Size  int64  `json:"size"`
	Mode  string `json:"mode"`
	IsDir bool   `json:"is_dir"`
	MTime int64  `json:"mtime"`
}

// FindResult is returned by Find, listing matched files.
type FindResult struct {
	Files   []FindMatch `json:"files"`
	Matched int         `json:"matched"`
	Examined int        `json:"examined"`
}

// FindOptions controls the search behaviour.
type FindOptions struct {
	Paths    []string `json:"paths"`     // directories to search
	Patterns []string `json:"patterns"`  // glob patterns (match filename)
	Regex    string   `json:"regex"`     // regex pattern (match full path)
	FileType string   `json:"file_type"` // "file", "directory", or "" (both)
	MaxDepth int      `json:"max_depth"` // 0 = unlimited
	Age      int64    `json:"age"`       // seconds; positive = older than, negative = newer than
	Size     int64    `json:"size"`      // bytes; positive = larger than, negative = smaller than
}

// Find searches for files matching the given options.
func Find(opts FindOptions) (FindResult, error) {
	result := FindResult{Files: make([]FindMatch, 0)}
	if len(opts.Paths) == 0 {
		return result, nil
	}

	var rx *regexp.Regexp
	if opts.Regex != "" {
		var err error
		rx, err = regexp.Compile(opts.Regex)
		if err != nil {
			return result, fmt.Errorf("file.Find: invalid regex: %w", err)
		}
	}

	cutoff := time.Now().Add(-time.Duration(opts.Age) * time.Second)
	now := time.Now()

	for _, dir := range opts.Paths {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // skip errors
			}
			if path == dir {
				return nil
			}

			// depth check
			if opts.MaxDepth > 0 {
				rel, _ := filepath.Rel(dir, path)
				depth := strings.Count(rel, string(filepath.Separator)) + 1
				if depth > opts.MaxDepth {
					if info.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
			}

			result.Examined++

			// file type filter
			if opts.FileType == "file" && info.IsDir() {
				return nil
			}
			if opts.FileType == "directory" && !info.IsDir() {
				return nil
			}

			// pattern filter
			if len(opts.Patterns) > 0 {
				matched := false
				for _, p := range opts.Patterns {
					if ok, _ := filepath.Match(p, info.Name()); ok {
						matched = true
						break
					}
				}
				if !matched {
					return nil
				}
			}

			// regex filter
			if rx != nil && !rx.MatchString(path) {
				return nil
			}

			// age filter
			if opts.Age != 0 {
				if opts.Age > 0 && info.ModTime().After(cutoff) {
					return nil // too new
				}
				if opts.Age < 0 {
					newerCutoff := now.Add(time.Duration(opts.Age) * time.Second) // Age is negative
					if info.ModTime().Before(newerCutoff) {
						return nil // too old
					}
				}
			}

			// size filter
			if opts.Size != 0 {
				if opts.Size > 0 && info.Size() < opts.Size {
					return nil // too small
				}
				if opts.Size < 0 && info.Size() > -opts.Size {
					return nil // too large
				}
			}

			result.Files = append(result.Files, FindMatch{
				Path:  path,
				Size:  info.Size(),
				Mode:  fmt.Sprintf("%04o", info.Mode().Perm()),
				IsDir: info.IsDir(),
				MTime: info.ModTime().Unix(),
			})
			return nil
		})
		if err != nil {
			return result, fmt.Errorf("file.Find: %w", err)
		}
	}

	result.Matched = len(result.Files)
	return result, nil
}

// FindFromArgs is a wrapper that takes individual arguments instead of a struct.
// Used by the AOT code generator which cannot easily construct structs.
func FindFromArgs(paths []string, patterns []string, regex string, fileType string, maxDepth int, age int64, size int64) (FindResult, error) {
	return Find(FindOptions{
		Paths:    paths,
		Patterns: patterns,
		Regex:    regex,
		FileType: fileType,
		MaxDepth: maxDepth,
		Age:      age,
		Size:     size,
	})
}
