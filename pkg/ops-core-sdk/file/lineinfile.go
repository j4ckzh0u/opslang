// LineInFile file operations.
package file

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// LineInFileResult is returned by LineInFile, reporting whether the file changed.
type LineInFileResult struct {
	Changed bool   `json:"changed"`
	Line    string `json:"line"`
	Error   string `json:"error,omitempty"`
}

// LineInFile ensures a line exists (or doesn't) in a file.
// If present=true and the line doesn't exist, append it.
// If present=false and the line exists, remove it.
// If regexp is provided with present=true, replace matching line instead of appending.
func LineInFile(path string, line string, present bool, rx string) (LineInFileResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return LineInFileResult{}, fmt.Errorf("file.LineInFile: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	// If the file ended with a newline, Split produces a trailing empty element.
	// Preserve that trailing newline on write-back.
	trailingNL := len(data) > 0 && data[len(data)-1] == '\n'
	if trailingNL && len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	changed := false
	var compiled *regexp.Regexp
	if rx != "" {
		c, err := regexp.Compile(rx)
		if err != nil {
			return LineInFileResult{}, fmt.Errorf("file.LineInFile: %w", err)
		}
		compiled = c
	}

	var out []string
	if present {
		if compiled != nil {
			matched := false
			for _, l := range lines {
				if compiled.MatchString(l) {
					out = append(out, line)
					matched = true
					if l != line {
						changed = true
					}
				} else {
					out = append(out, l)
				}
			}
			if !matched {
				out = append(out, line)
				changed = true
			}
		} else {
			found := false
			for _, l := range lines {
				if l == line {
					found = true
				}
				out = append(out, l)
			}
			if !found {
				out = append(out, line)
				changed = true
			}
		}
	} else {
		// absent
		for _, l := range lines {
			if compiled != nil {
				if compiled.MatchString(l) {
					changed = true
					continue
				}
			} else if l == line {
				changed = true
				continue
			}
			out = append(out, l)
		}
	}

	if !changed {
		return LineInFileResult{Changed: false, Line: line}, nil
	}

	var content string
	if trailingNL || present {
		content = strings.Join(out, "\n") + "\n"
	} else {
		content = strings.Join(out, "\n")
		if len(out) > 0 {
			content += "\n"
		}
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return LineInFileResult{}, fmt.Errorf("file.LineInFile: %w", err)
	}
	return LineInFileResult{Changed: true, Line: line}, nil
}
