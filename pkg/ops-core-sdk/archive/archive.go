// Package archive provides archive create and extract operations.
// Supports tar, tar.gz, and zip formats. Pure Go stdlib, no shell calls.
package archive

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CreateResult is returned by Create.
type CreateResult struct {
	Dest    string   `json:"dest"`
	Format  string   `json:"format"`
	Files   []string `json:"files"`
	Size    int64    `json:"size"`
}

// Create creates an archive at dest from the given source paths.
// Format is determined by dest extension: .tar, .tar.gz/.tgz, .zip.
func Create(dest string, sources []string) (CreateResult, error) {
	result := CreateResult{Dest: dest, Files: sources}
	format := detectFormat(dest)
	result.Format = format

	out, err := os.Create(dest)
	if err != nil {
		return result, fmt.Errorf("archive.Create: %w", err)
	}
	defer out.Close()

	switch format {
	case "tar.gz":
		gw := gzip.NewWriter(out)
		defer gw.Close()
		tw := tar.NewWriter(gw)
		defer tw.Close()
		for _, src := range sources {
			if err := addToTar(tw, src); err != nil {
				return result, fmt.Errorf("archive.Create: %w", err)
			}
		}
	case "tar":
		tw := tar.NewWriter(out)
		defer tw.Close()
		for _, src := range sources {
			if err := addToTar(tw, src); err != nil {
				return result, fmt.Errorf("archive.Create: %w", err)
			}
		}
	case "zip":
		zw := zip.NewWriter(out)
		defer zw.Close()
		for _, src := range sources {
			if err := addToZip(zw, src); err != nil {
				return result, fmt.Errorf("archive.Create: %w", err)
			}
		}
	default:
		return result, fmt.Errorf("archive.Create: unsupported format for %q", dest)
	}

	info, err := os.Stat(dest)
	if err == nil {
		result.Size = info.Size()
	}
	return result, nil
}

// ExtractResult is returned by Extract.
type ExtractResult struct {
	Dest  string   `json:"dest"`
	Files []string `json:"files"`
	Count int      `json:"count"`
}

// Extract extracts an archive to dest directory.
func Extract(src string, dest string) (ExtractResult, error) {
	result := ExtractResult{Dest: dest, Files: make([]string, 0)}

	if err := os.MkdirAll(dest, 0755); err != nil {
		return result, fmt.Errorf("archive.Extract: %w", err)
	}

	format := detectFormat(src)
	f, err := os.Open(src)
	if err != nil {
		return result, fmt.Errorf("archive.Extract: %w", err)
	}
	defer f.Close()

	switch format {
	case "tar.gz":
		gr, err := gzip.NewReader(f)
		if err != nil {
			return result, fmt.Errorf("archive.Extract: %w", err)
		}
		defer gr.Close()
		tr := tar.NewReader(gr)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return result, fmt.Errorf("archive.Extract: %w", err)
			}
			target := filepath.Join(dest, hdr.Name)
			// Prevent path traversal
			if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(dest)+string(os.PathSeparator)) && filepath.Clean(target) != filepath.Clean(dest) {
				continue
			}
			if err := extractTarEntry(tr, hdr, target); err != nil {
				return result, fmt.Errorf("archive.Extract: %w", err)
			}
			result.Files = append(result.Files, hdr.Name)
		}
	case "tar":
		tr := tar.NewReader(f)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return result, fmt.Errorf("archive.Extract: %w", err)
			}
			target := filepath.Join(dest, hdr.Name)
			if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(dest)+string(os.PathSeparator)) && filepath.Clean(target) != filepath.Clean(dest) {
				continue
			}
			if err := extractTarEntry(tr, hdr, target); err != nil {
				return result, fmt.Errorf("archive.Extract: %w", err)
			}
			result.Files = append(result.Files, hdr.Name)
		}
	case "zip":
		// zip needs Seekable reader
		info, err := f.Stat()
		if err != nil {
			return result, fmt.Errorf("archive.Extract: %w", err)
		}
		zr, err := zip.NewReader(f, info.Size())
		if err != nil {
			return result, fmt.Errorf("archive.Extract: %w", err)
		}
		for _, zf := range zr.File {
			target := filepath.Join(dest, zf.Name)
			if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(dest)+string(os.PathSeparator)) && filepath.Clean(target) != filepath.Clean(dest) {
				continue
			}
			if err := extractZipEntry(zf, target); err != nil {
				return result, fmt.Errorf("archive.Extract: %w", err)
			}
			result.Files = append(result.Files, zf.Name)
		}
	default:
		return result, fmt.Errorf("archive.Extract: unsupported format for %q", src)
	}

	result.Count = len(result.Files)
	return result, nil
}

func detectFormat(path string) string {
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz"):
		return "tar.gz"
	case strings.HasSuffix(lower, ".tar"):
		return "tar"
	case strings.HasSuffix(lower, ".zip"):
		return "zip"
	default:
		return "unknown"
	}
}

func addToTar(tw *tar.Writer, src string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = path
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if !info.IsDir() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := io.Copy(tw, f); err != nil {
				return err
			}
		}
		return nil
	})
}

func addToZip(zw *zip.Writer, src string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = path
		header.Method = zip.Deflate
		w, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := io.Copy(w, f); err != nil {
				return err
			}
		}
		return nil
	})
}

func extractTarEntry(tr *tar.Reader, hdr *tar.Header, target string) error {
	switch hdr.Typeflag {
	case tar.TypeDir:
		return os.MkdirAll(target, os.FileMode(hdr.Mode))
	case tar.TypeReg:
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(f, tr)
		return err
	default:
		return nil // skip symlinks, etc.
	}
}

func extractZipEntry(zf *zip.File, target string) error {
	if zf.FileInfo().IsDir() {
		return os.MkdirAll(target, 0755)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	rc, err := zf.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, zf.Mode())
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, rc)
	return err
}
