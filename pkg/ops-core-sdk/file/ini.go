// INI file operations — Ansible ini_file module equivalent.
package file

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// IniGetResult is returned by IniGet.
type IniGetResult struct {
	Section string `json:"section"`
	Key     string `json:"key"`
	Value   string `json:"value"`
	Found   bool   `json:"found"`
}

// IniSetResult is returned by IniSet.
type IniSetResult struct {
	Changed bool   `json:"changed"`
	Section string `json:"section"`
	Key     string `json:"key"`
	Value   string `json:"value"`
}

// IniGet reads a value from an INI file.
func IniGet(path string, section string, key string) (IniGetResult, error) {
	result := IniGetResult{Section: section, Key: key}

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return result, fmt.Errorf("file.IniGet: %w", err)
	}
	defer f.Close()

	currentSection := ""
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.TrimSpace(line[1 : len(line)-1])
			continue
		}
		if currentSection == section {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 && strings.TrimSpace(parts[0]) == key {
				result.Value = strings.TrimSpace(parts[1])
				result.Found = true
				return result, nil
			}
		}
	}
	return result, nil
}

// IniSet sets a value in an INI file, creating sections/keys as needed.
func IniSet(path string, section string, key string, value string) (IniSetResult, error) {
	result := IniSetResult{Section: section, Key: key, Value: value}

	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return result, fmt.Errorf("file.IniSet: %w", err)
	}

	var lines []string
	if len(data) > 0 {
		lines = strings.Split(string(data), "\n")
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
	}

	// Find the section and key
	sectionStart := -1
	sectionEnd := -1 // line after last line in section
	keyLine := -1

	currentSection := ""
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
			continue
		}
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			if sectionStart >= 0 && sectionEnd < 0 {
				sectionEnd = i
			}
			currentSection = strings.TrimSpace(trimmed[1 : len(trimmed)-1])
			if currentSection == section {
				sectionStart = i
			}
			continue
		}
		if currentSection == section && sectionStart >= 0 {
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 && strings.TrimSpace(parts[0]) == key {
				keyLine = i
				if strings.TrimSpace(parts[1]) == value {
					return IniSetResult{Changed: false, Section: section, Key: key, Value: value}, nil
				}
			}
		}
	}
	if sectionStart >= 0 && sectionEnd < 0 {
		sectionEnd = len(lines)
	}

	if keyLine >= 0 {
		// Update existing key
		lines[keyLine] = fmt.Sprintf("%s = %s", key, value)
		result.Changed = true
	} else if sectionStart >= 0 {
		// Section exists, add key
		if sectionEnd > sectionStart {
			// Insert before section end
			newLines := make([]string, 0, len(lines)+1)
			newLines = append(newLines, lines[:sectionEnd]...)
			newLines = append(newLines, fmt.Sprintf("%s = %s", key, value))
			newLines = append(newLines, lines[sectionEnd:]...)
			lines = newLines
		} else {
			lines = append(lines, fmt.Sprintf("%s = %s", key, value))
		}
		result.Changed = true
	} else {
		// Create new section
		if len(lines) > 0 && lines[len(lines)-1] != "" {
			lines = append(lines, "")
		}
		lines = append(lines, fmt.Sprintf("[%s]", section))
		lines = append(lines, fmt.Sprintf("%s = %s", key, value))
		result.Changed = true
	}

	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return result, fmt.Errorf("file.IniSet: %w", err)
	}
	return result, nil
}
