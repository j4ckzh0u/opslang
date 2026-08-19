// Package slurp provides base64 encoding of file content.
// Equivalent to ansible.builtin.slurp module.
package slurp

import (
	"encoding/base64"
	"fmt"
	"os"
)

// Result is returned by Encode.
type Result struct {
	Status   string `json:"status"`
	Content  string `json:"content"`            // base64 encoded
	Encoding string `json:"encoding"`           // "base64"
	Size     int64  `json:"size"`               // bytes
	Path     string `json:"path"`
	Error    string `json:"error,omitempty"`
}

// Encode reads a file and returns its base64-encoded content.
func Encode(path string) Result {
	if path == "" {
		return Result{Status: "failed", Error: "path is required"}
	}

	info, err := os.Stat(path)
	if err != nil {
		return Result{Status: "failed", Path: path, Error: fmt.Sprintf("stat: %v", err)}
	}
	if info.IsDir() {
		return Result{Status: "failed", Path: path, Error: "path is a directory"}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Result{Status: "failed", Path: path, Error: fmt.Sprintf("read: %v", err)}
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	return Result{
		Status:   "success",
		Content:  encoded,
		Encoding: "base64",
		Size:     info.Size(),
		Path:     path,
	}
}

// Decode decodes a base64 string and optionally writes it to a file.
func Decode(encoded, destPath string) Result {
	if encoded == "" {
		return Result{Status: "failed", Error: "encoded content is required"}
	}

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("decode: %v", err)}
	}

	if destPath != "" {
		if err := os.WriteFile(destPath, data, 0644); err != nil {
			return Result{Status: "failed", Path: destPath, Error: fmt.Sprintf("write: %v", err)}
		}
		return Result{
			Status:   "success",
			Content:  string(data),
			Encoding: "utf-8",
			Size:     int64(len(data)),
			Path:     destPath,
		}
	}

	return Result{
		Status:   "success",
		Content:  string(data),
		Encoding: "utf-8",
		Size:     int64(len(data)),
	}
}
