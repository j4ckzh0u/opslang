// Append and Template file operations.
package file

import (
	"fmt"
	"os"
	"strings"
)

// AppendResult is returned by Append, confirming the bytes appended.
type AppendResult struct {
	Path string `json:"path"`
	Size int64  `json:"size"` // total file size after append
}

// Append writes content to the end of the file at path, creating it with
// mode 0644 if it does not exist. It never truncates existing content.
func Append(path string, content string) (AppendResult, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return AppendResult{}, fmt.Errorf("file.Append: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		return AppendResult{}, fmt.Errorf("file.Append: %w", err)
	}

	info, err := f.Stat()
	if err != nil {
		return AppendResult{}, fmt.Errorf("file.Append: %w", err)
	}

	return AppendResult{
		Path: path,
		Size: info.Size(),
	}, nil
}

// TemplateResult is returned by Template, holding the rendered content.
type TemplateResult struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Size    int64  `json:"size"` // rendered content length in bytes
}

// Template reads the file at path and replaces every {{key}} placeholder
// with the corresponding value from vars (formatted via fmt.Sprint for
// non-string values). Unknown placeholders are left untouched so that
// escaping ({{literal}}) stays possible. The file itself is not modified.
func Template(path string, vars map[string]interface{}) (TemplateResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return TemplateResult{}, fmt.Errorf("file.Template: %w", err)
	}

	if vars == nil {
		vars = map[string]interface{}{}
	}

	rendered := replacePlaceholders(string(data), vars)

	return TemplateResult{
		Path:    path,
		Content: rendered,
		Size:    int64(len(rendered)),
	}, nil
}

// replacePlaceholders substitutes {{key}} occurrences using vars.
func replacePlaceholders(text string, vars map[string]interface{}) string {
	var b strings.Builder
	i := 0
	for {
		open := strings.Index(text[i:], "{{")
		if open < 0 {
			b.WriteString(text[i:])
			return b.String()
		}
		open += i
		closeIdx := strings.Index(text[open+2:], "}}")
		if closeIdx < 0 {
			b.WriteString(text[i:])
			return b.String()
		}
		closeIdx += open + 2

		b.WriteString(text[i:open])
		key := strings.TrimSpace(text[open+2 : closeIdx])
		if v, ok := vars[key]; ok {
			b.WriteString(fmt.Sprintf("%v", v))
		} else {
			// Unknown key: keep the placeholder verbatim.
			b.WriteString(text[open : closeIdx+2])
		}
		i = closeIdx + 2
	}
}
