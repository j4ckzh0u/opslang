// Package ini_file provides INI configuration file management operations.
package ini_file

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// SectionResult represents the result of getting sections.
type SectionResult struct {
	Sections []string `json:"sections"`
}

// ValueResult represents the result of getting a value.
type ValueResult struct {
	Exists bool   `json:"exists"`
	Value  string `json:"value"`
}

// ActionResult represents the result of a modification.
type ActionResult struct {
	Changed bool   `json:"changed"`
	Message string `json:"message"`
}

// Sections returns all sections in an INI file.
func Sections(path string) (*SectionResult, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &SectionResult{Sections: []string{}}, nil
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	sections := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section := line[1 : len(line)-1]
			sections = append(sections, section)
		}
	}

	return &SectionResult{Sections: sections}, nil
}

// Get retrieves a value from an INI file.
func Get(path string, section string, key string) (*ValueResult, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ValueResult{Exists: false, Value: ""}, nil
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	inSection := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Check for section header
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection := line[1 : len(line)-1]
			inSection = (currentSection == section)
			continue
		}

		// If in the target section, look for the key
		if inSection {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				k := strings.TrimSpace(parts[0])
				if k == key {
					v := strings.TrimSpace(parts[1])
					// Remove inline comments
					if idx := strings.Index(v, " ;"); idx != -1 {
						v = strings.TrimSpace(v[:idx])
					}
					if idx := strings.Index(v, " #"); idx != -1 {
						v = strings.TrimSpace(v[:idx])
					}
					return &ValueResult{Exists: true, Value: v}, nil
				}
			}
		}
	}

	return &ValueResult{Exists: false, Value: ""}, nil
}

// Set sets a value in an INI file.
func Set(path string, section string, key string, value string) (*ActionResult, error) {
	// Read existing content
	lines, err := readLines(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		lines = []string{}
	}

	// Find or create section
	sectionStart := -1
	sectionEnd := len(lines)
	inSection := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			currentSection := trimmed[1 : len(trimmed)-1]
			if currentSection == section {
				sectionStart = i
				inSection = true
			} else if inSection {
				sectionEnd = i
				break
			}
		}
	}

	// Section not found, add it
	if sectionStart == -1 {
		if len(lines) > 0 && !strings.HasSuffix(lines[len(lines)-1], "\n") {
			lines = append(lines, "")
		}
		lines = append(lines, fmt.Sprintf("[%s]", section))
		lines = append(lines, fmt.Sprintf("%s = %s", key, value))
	} else {
		// Section exists, find or update key
		keyFound := false
		for i := sectionStart + 1; i < sectionEnd; i++ {
			line := lines[i]
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
				continue
			}
			if strings.HasPrefix(trimmed, "[") {
				break
			}

			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 && strings.TrimSpace(parts[0]) == key {
				lines[i] = fmt.Sprintf("%s = %s", key, value)
				keyFound = true
				break
			}
		}

		if !keyFound {
			// Add key at the end of the section
			lines = append(lines[:sectionEnd], append([]string{fmt.Sprintf("%s = %s", key, value)}, lines[sectionEnd:]...)...)
		}
	}

	if err := writeLines(path, lines); err != nil {
		return nil, err
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Set %s.%s = %s", section, key, value),
	}, nil
}

// Remove removes a key from an INI file.
func Remove(path string, section string, key string) (*ActionResult, error) {
	lines, err := readLines(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ActionResult{Changed: false, Message: "File does not exist"}, nil
		}
		return nil, err
	}

	sectionStart := -1
	sectionEnd := len(lines)
	inSection := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			currentSection := trimmed[1 : len(trimmed)-1]
			if currentSection == section {
				sectionStart = i
				inSection = true
			} else if inSection {
				sectionEnd = i
				break
			}
		}
	}

	if sectionStart == -1 {
		return &ActionResult{Changed: false, Message: "Section not found"}, nil
	}

	// Find and remove key
	for i := sectionStart + 1; i < sectionEnd; i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") {
			break
		}

		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == key {
			lines = append(lines[:i], lines[i+1:]...)
			if err := writeLines(path, lines); err != nil {
				return nil, err
			}
			return &ActionResult{
				Changed: true,
				Message: fmt.Sprintf("Removed %s.%s", section, key),
			}, nil
		}
	}

	return &ActionResult{Changed: false, Message: "Key not found"}, nil
}

// RemoveSection removes an entire section from an INI file.
func RemoveSection(path string, section string) (*ActionResult, error) {
	lines, err := readLines(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ActionResult{Changed: false, Message: "File does not exist"}, nil
		}
		return nil, err
	}

	sectionStart := -1
	sectionEnd := len(lines)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			currentSection := trimmed[1 : len(trimmed)-1]
			if currentSection == section {
				sectionStart = i
			} else if sectionStart != -1 {
				sectionEnd = i
				break
			}
		}
	}

	if sectionStart == -1 {
		return &ActionResult{Changed: false, Message: "Section not found"}, nil
	}

	lines = append(lines[:sectionStart], lines[sectionEnd:]...)
	if err := writeLines(path, lines); err != nil {
		return nil, err
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Removed section [%s]", section),
	}, nil
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func writeLines(path string, lines []string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}
	return writer.Flush()
}
