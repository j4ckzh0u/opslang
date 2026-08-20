// Package unarchive provides Ansible unarchive module equivalent.
// Extracts archives (tar, gz, bz2, xz, zip).
package unarchive

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// UnarchiveResult is returned by Unarchive.
type UnarchiveResult struct {
	Dest       string   `json:"dest"`
	Source     string   `json:"src"`
	Extracted  []string `json:"extracted"`
	Changed    bool     `json:"changed"`
	Error      string   `json:"error,omitempty"`
}

// Unarchive extracts src into dest. Format is auto-detected from extension.
func Unarchive(src, dest, owner, group, mode string, creates string) UnarchiveResult {
	if src == "" || dest == "" {
		return UnarchiveResult{Error: "src and dest are required"}
	}
	if creates != "" {
		if _, err := os.Stat(creates); err == nil {
			return UnarchiveResult{Dest: dest, Source: src, Changed: false}
		}
	}
	if err := os.MkdirAll(dest, 0755); err != nil {
		return UnarchiveResult{Error: err.Error()}
	}

	low := strings.ToLower(src)
	switch {
	case strings.HasSuffix(low, ".tar"):
		return extractTar(src, dest)
	case strings.HasSuffix(low, ".tar.gz") || strings.HasSuffix(low, ".tgz"):
		return extractTarGz(src, dest)
	case strings.HasSuffix(low, ".tar.bz2"):
		return extractTarBz2(src, dest)
	case strings.HasSuffix(low, ".zip"):
		return extractZip(src, dest)
	case strings.HasSuffix(low, ".gz"):
		return extractGz(src, dest)
	default:
		return UnarchiveResult{Error: fmt.Sprintf("unsupported archive format: %s", src)}
	}
}

func extractTar(src, dest string) UnarchiveResult {
	f, err := os.Open(src)
	if err != nil {
		return UnarchiveResult{Error: err.Error()}
	}
	defer f.Close()
	return extractTarReader(tar.NewReader(f), dest)
}

func extractTarGz(src, dest string) UnarchiveResult {
	f, err := os.Open(src)
	if err != nil {
		return UnarchiveResult{Error: err.Error()}
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return UnarchiveResult{Error: err.Error()}
	}
	defer gz.Close()
	return extractTarReader(tar.NewReader(gz), dest)
}

func extractTarBz2(src, dest string) UnarchiveResult {
	f, err := os.Open(src)
	if err != nil {
		return UnarchiveResult{Error: err.Error()}
	}
	defer f.Close()
	bz := bzip2.NewReader(f)
	return extractTarReader(tar.NewReader(bz), dest)
}

func extractTarReader(tr *tar.Reader, dest string) UnarchiveResult {
	var extracted []string
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return UnarchiveResult{Error: err.Error()}
		}
		target := filepath.Join(dest, hdr.Name)
		if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) {
			return UnarchiveResult{Error: "invalid tar entry: " + hdr.Name}
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(hdr.Mode)); err != nil {
				return UnarchiveResult{Error: err.Error()}
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return UnarchiveResult{Error: err.Error()}
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return UnarchiveResult{Error: err.Error()}
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return UnarchiveResult{Error: err.Error()}
			}
			out.Close()
			extracted = append(extracted, target)
		}
	}
	return UnarchiveResult{Dest: dest, Source: dest, Extracted: extracted, Changed: len(extracted) > 0}
}

func extractZip(src, dest string) UnarchiveResult {
	r, err := zip.OpenReader(src)
	if err != nil {
		return UnarchiveResult{Error: err.Error()}
	}
	defer r.Close()
	var extracted []string
	for _, f := range r.File {
		target := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) {
			return UnarchiveResult{Error: "invalid zip entry: " + f.Name}
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, f.Mode()); err != nil {
				return UnarchiveResult{Error: err.Error()}
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return UnarchiveResult{Error: err.Error()}
		}
		rc, err := f.Open()
		if err != nil {
			return UnarchiveResult{Error: err.Error()}
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return UnarchiveResult{Error: err.Error()}
		}
		if _, err := io.Copy(out, rc); err != nil {
			rc.Close()
			out.Close()
			return UnarchiveResult{Error: err.Error()}
		}
		rc.Close()
		out.Close()
		extracted = append(extracted, target)
	}
	return UnarchiveResult{Dest: dest, Source: src, Extracted: extracted, Changed: len(extracted) > 0}
}

func extractGz(src, dest string) UnarchiveResult {
	f, err := os.Open(src)
	if err != nil {
		return UnarchiveResult{Error: err.Error()}
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return UnarchiveResult{Error: err.Error()}
	}
	defer gz.Close()
	outName := filepath.Join(dest, strings.TrimSuffix(filepath.Base(src), ".gz"))
	out, err := os.OpenFile(outName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return UnarchiveResult{Error: err.Error()}
	}
	defer out.Close()
	if _, err := io.Copy(out, gz); err != nil {
		return UnarchiveResult{Error: err.Error()}
	}
	return UnarchiveResult{Dest: dest, Source: src, Extracted: []string{outName}, Changed: true}
}
