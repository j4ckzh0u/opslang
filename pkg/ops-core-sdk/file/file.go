// Package file provides file system operations for OpsLang.
// All functions return structured results with JSON tags, enabling
// easy serialization and downstream processing. No shell calls are
// used — everything is implemented with pure Go stdlib.
package file

import (
	"fmt"
	"io"
	"os"
)

// FileContent is returned by Read, holding the path and full text content.
type FileContent struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Size    int64  `json:"size"`
}

// WriteResult is returned by Write, confirming the bytes written.
type WriteResult struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
}

// CopyResult is returned by Copy, confirming source, destination, and size.
type CopyResult struct {
	Src  string `json:"src"`
	Dst  string `json:"dst"`
	Size int64  `json:"size"`
}

// MoveResult is returned by Move, confirming source and destination.
type MoveResult struct {
	Src string `json:"src"`
	Dst string `json:"dst"`
}

// DeleteResult is returned by Delete, reporting whether the path existed.
type DeleteResult struct {
	Path    string `json:"path"`
	Existed bool   `json:"existed"`
}

// ExistsResult is returned by Exists, reporting presence and whether it is a directory.
type ExistsResult struct {
	Path   string `json:"path"`
	Exists bool   `json:"exists"`
	IsDir  bool   `json:"is_dir"`
}

// FileInfo is returned by Stat, holding full metadata about a file.
type FileInfo struct {
	Path    string `json:"path"`
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	Mode    string `json:"mode"`
	ModTime int64  `json:"mod_time"`
	IsDir   bool   `json:"is_dir"`
}

// ChmodResult is returned by Chmod, reporting the new mode as an octal string.
type ChmodResult struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
}

// Read reads the entire contents of the file at path and returns a FileContent.
func Read(path string) (FileContent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return FileContent{}, fmt.Errorf("file.Read: %w", err)
	}
	return FileContent{
		Path:    path,
		Content: string(data),
		Size:    int64(len(data)),
	}, nil
}

// Write writes content to the file at path, creating or truncating it with mode 0644.
func Write(path string, content string) (WriteResult, error) {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return WriteResult{}, fmt.Errorf("file.Write: %w", err)
	}
	return WriteResult{
		Path: path,
		Size: int64(len(content)),
	}, nil
}

// Copy copies the file at src to dst using io.Copy, returning the number of bytes copied.
func Copy(src, dst string) (CopyResult, error) {
	in, err := os.Open(src)
	if err != nil {
		return CopyResult{}, fmt.Errorf("file.Copy open src: %w", err)
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return CopyResult{}, fmt.Errorf("file.Copy stat src: %w", err)
	}

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return CopyResult{}, fmt.Errorf("file.Copy open dst: %w", err)
	}
	defer func() {
		if cerr := out.Close(); cerr != nil {
			_ = fmt.Errorf("file.Copy close dst: %w", cerr)
		}
	}()

	n, err := io.Copy(out, in)
	if err != nil {
		return CopyResult{}, fmt.Errorf("file.Copy copy: %w", err)
	}

	return CopyResult{
		Src:  src,
		Dst:  dst,
		Size: n,
	}, nil
}

// Move renames src to dst using os.Rename.
func Move(src, dst string) (MoveResult, error) {
	if err := os.Rename(src, dst); err != nil {
		return MoveResult{}, fmt.Errorf("file.Move: %w", err)
	}
	return MoveResult{
		Src: src,
		Dst: dst,
	}, nil
}

// Delete removes the file at path. Existed indicates whether the file existed prior.
func Delete(path string) (DeleteResult, error) {
	existed := true
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			existed = false
		} else {
			return DeleteResult{}, fmt.Errorf("file.Delete stat: %w", err)
		}
	}

	if existed {
		if err := os.Remove(path); err != nil {
			return DeleteResult{}, fmt.Errorf("file.Delete remove: %w", err)
		}
	}

	return DeleteResult{
		Path:    path,
		Existed: existed,
	}, nil
}

// Exists checks whether a path exists and whether it is a directory.
func Exists(path string) (ExistsResult, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ExistsResult{
				Path:   path,
				Exists: false,
				IsDir:  false,
			}, nil
		}
		return ExistsResult{}, fmt.Errorf("file.Exists: %w", err)
	}
	return ExistsResult{
		Path:   path,
		Exists: true,
		IsDir:  info.IsDir(),
	}, nil
}

// Stat returns detailed metadata for the file at path. Mode is formatted as an octal string
// (e.g., "0755") and ModTime is the modification time as a Unix timestamp in seconds.
func Stat(path string) (FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return FileInfo{}, fmt.Errorf("file.Stat: %w", err)
	}
	return FileInfo{
		Path:    path,
		Name:    info.Name(),
		Size:    info.Size(),
		Mode:    fmt.Sprintf("%04o", info.Mode().Perm()),
		ModTime: info.ModTime().Unix(),
		IsDir:   info.IsDir(),
	}, nil
}

// Chmod changes the mode of the file at path to the given octal mode (e.g. 0755).
func Chmod(path string, mode uint32) (ChmodResult, error) {
	if err := os.Chmod(path, os.FileMode(mode)); err != nil {
		return ChmodResult{}, fmt.Errorf("file.Chmod: %w", err)
	}
	return ChmodResult{
		Path: path,
		Mode: fmt.Sprintf("%04o", mode),
	}, nil
}
