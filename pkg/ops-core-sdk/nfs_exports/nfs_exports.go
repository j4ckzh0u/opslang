// Package nfs_exports manages NFS exports configuration.
package nfs_exports

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

const defaultExportsPath = "/etc/exports"

// ExportEntry represents a single NFS export entry.
type ExportEntry struct {
	Path    string   `json:"path"`
	Hosts   []string `json:"hosts"`
	Options []string `json:"options"`
	Line    string   `json:"line"` // Original line
}

// Result is the common return type for nfs_exports operations.
type Result struct {
	Export     *ExportEntry `json:"export,omitempty"`
	Changed    bool         `json:"changed"`
	Error      string       `json:"error,omitempty"`
	DurationMs int64        `json:"duration_ms"`
}

// ListResult is returned by List.
type ListResult struct {
	Exports    []ExportEntry `json:"exports"`
	DurationMs int64         `json:"duration_ms"`
}

// Present adds or updates an export entry (idempotent).
func Present(path, hosts, options string) (Result, error) {
	start := time.Now()
	if path == "" {
		return Result{Error: "path must not be empty", DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("path must not be empty")
	}
	if hosts == "" {
		return Result{Error: "hosts must not be empty", DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("hosts must not be empty")
	}

	newLine := buildLine(path, hosts, options)
	existing, idx := findExport(defaultExportsPath, path)
	if existing != nil && existing.Line == newLine {
		return Result{Export: existing, Changed: false, DurationMs: time.Since(start).Milliseconds()}, nil
	}

	if err := updateExport(defaultExportsPath, path, newLine, idx); err != nil {
		return Result{Error: err.Error(), DurationMs: time.Since(start).Milliseconds()}, err
	}

	entry := &ExportEntry{Path: path, Line: newLine}
	return Result{Export: entry, Changed: true, DurationMs: time.Since(start).Milliseconds()}, nil
}

// Absent removes an export entry (idempotent).
func Absent(path string) (Result, error) {
	start := time.Now()
	if path == "" {
		return Result{Error: "path must not be empty", DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("path must not be empty")
	}

	_, idx := findExport(defaultExportsPath, path)
	if idx < 0 {
		return Result{Changed: false, DurationMs: time.Since(start).Milliseconds()}, nil
	}

	if err := removeExport(defaultExportsPath, idx); err != nil {
		return Result{Error: err.Error(), DurationMs: time.Since(start).Milliseconds()}, err
	}

	return Result{Changed: true, DurationMs: time.Since(start).Milliseconds()}, nil
}

// List returns all export entries.
func List() (ListResult, error) {
	start := time.Now()
	entries, err := parseExports(defaultExportsPath)
	if err != nil {
		return ListResult{DurationMs: time.Since(start).Milliseconds()}, err
	}
	return ListResult{Exports: entries, DurationMs: time.Since(start).Milliseconds()}, nil
}

func buildLine(path, hosts, options string) string {
	if options != "" {
		return path + " " + hosts + "(" + options + ")"
	}
	return path + " " + hosts
}

func parseExports(path string) ([]ExportEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	var entries []ExportEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		entry := parseLine(line)
		if entry != nil {
			entries = append(entries, *entry)
		}
	}
	return entries, nil
}

func parseLine(line string) *ExportEntry {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return nil
	}
	path := fields[0]
	var hosts, options []string
	for _, f := range fields[1:] {
		if idx := strings.Index(f, "("); idx >= 0 {
			hosts = append(hosts, f[:idx])
			optStr := strings.TrimRight(f[idx+1:], ")")
			for _, o := range strings.Split(optStr, ",") {
				o = strings.TrimSpace(o)
				if o != "" {
					options = append(options, o)
				}
			}
		} else {
			hosts = append(hosts, f)
		}
	}
	return &ExportEntry{Path: path, Hosts: hosts, Options: options, Line: line}
}

func findExport(path, exportPath string) (*ExportEntry, int) {
	f, err := os.Open(path)
	if err != nil {
		return nil, -1
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	idx := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			idx++
			continue
		}
		entry := parseLine(line)
		if entry != nil && entry.Path == exportPath {
			entry.Line = line
			return entry, idx
		}
		idx++
	}
	return nil, -1
}

func updateExport(filePath, exportPath, newLine string, existingIdx int) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read %s: %w", filePath, err)
	}

	lines := strings.Split(string(data), "\n")
	if existingIdx >= 0 && existingIdx < len(lines) {
		lines[existingIdx] = newLine
	} else {
		lines = append(lines, newLine)
	}

	return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644)
}

func removeExport(filePath string, idx int) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read %s: %w", filePath, err)
	}

	lines := strings.Split(string(data), "\n")
	if idx < 0 || idx >= len(lines) {
		return nil
	}
	lines = append(lines[:idx], lines[idx+1:]...)

	return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644)
}
