package securityscan

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func scanManifests(ctx context.Context, root string, options Options) ([]SBOMComponent, []ScanError) {
	components := make([]SBOMComponent, 0)
	errors := make([]ScanError, 0)
	walkErr := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			errors = append(errors, ScanError{Path: path, Code: "walk_error", Message: err.Error()})
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if entry.IsDir() {
			if options.MaxDepth > 0 && relativeDepth(root, path) >= options.MaxDepth {
				return filepath.SkipDir
			}
			if path != root && shouldIgnore(path, root, options.IgnorePaths) {
				return filepath.SkipDir
			}
			return nil
		}
		if shouldIgnore(path, root, options.IgnorePaths) || !isManifest(path) {
			return nil
		}
		if options.MaxFileSize > 0 {
			info, statErr := entry.Info()
			if statErr != nil {
				errors = append(errors, ScanError{Path: path, Code: "stat_error", Message: statErr.Error()})
				return nil
			}
			if info.Size() > options.MaxFileSize {
				errors = append(errors, ScanError{Path: path, Code: "limit_exceeded", Message: fmt.Sprintf("file size %d exceeds limit %d", info.Size(), options.MaxFileSize)})
				return nil
			}
		}
		parsed, parseErr := parseManifest(path)
		if parseErr != nil {
			errors = append(errors, ScanError{Path: path, Code: "parse_error", Message: parseErr.Error()})
			return nil
		}
		components = append(components, parsed...)
		return nil
	})
	if walkErr != nil {
		code := "walk_error"
		if ctx.Err() == context.Canceled {
			code = "cancelled"
		} else if ctx.Err() == context.DeadlineExceeded {
			code = "timeout"
		}
		errors = append(errors, ScanError{Path: root, Code: code, Message: walkErr.Error()})
	}
	return mergeComponents(components), errors
}

func relativeDepth(root, path string) int {
	relative, err := filepath.Rel(root, path)
	if err != nil || relative == "." {
		return 0
	}
	return len(strings.Split(filepath.ToSlash(relative), "/"))
}

func isManifest(path string) bool {
	base := filepath.Base(path)
	switch base {
	case "go.mod", "package.json", "requirements.txt", "pyproject.toml":
		return true
	default:
		return false
	}
}

func shouldIgnore(path, root string, configured []string) bool {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	for _, part := range strings.Split(filepath.ToSlash(relative), "/") {
		switch part {
		case ".git", "node_modules", ".venv", "vendor", "dist", "build", "target", "__pycache__":
			return true
		}
	}
	for _, ignored := range configured {
		ignored = filepath.Clean(strings.TrimSpace(ignored))
		if ignored == "" {
			continue
		}
		if relative == ignored || strings.HasPrefix(relative, ignored+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

func parseManifest(path string) ([]SBOMComponent, error) {
	switch filepath.Base(path) {
	case "package.json":
		return parsePackageJSON(path)
	case "go.mod":
		return parseGoMod(path)
	case "requirements.txt":
		return parseRequirements(path)
	case "pyproject.toml":
		return parsePyproject(path)
	default:
		return nil, nil
	}
}

func parsePackageJSON(path string) ([]SBOMComponent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var document struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &document); err != nil {
		return nil, err
	}
	components := make([]SBOMComponent, 0, len(document.Dependencies)+len(document.DevDependencies))
	for _, dependencies := range []map[string]string{document.Dependencies, document.DevDependencies} {
		for name, version := range dependencies {
			components = append(components, manifestComponent("npm", name, version, path))
		}
	}
	return components, nil
}

func parseGoMod(path string) ([]SBOMComponent, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	components := make([]SBOMComponent, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 3 && fields[0] == "require" && fields[1] != "(" {
			components = append(components, manifestComponent("go", fields[1], fields[2], path))
			continue
		}
		if len(fields) >= 2 && !strings.HasPrefix(scanner.Text(), " ") && fields[0] != "module" && fields[0] != "go" {
			continue
		}
		if len(fields) >= 2 && strings.HasPrefix(scanner.Text(), "\t") {
			components = append(components, manifestComponent("go", fields[0], fields[1], path))
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return components, nil
}

func parseRequirements(path string) ([]SBOMComponent, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	components := make([]SBOMComponent, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(strings.SplitN(scanner.Text(), "#", 2)[0])
		if line == "" || strings.HasPrefix(line, "-") {
			continue
		}
		name, version, ok := splitDependency(line)
		if ok {
			components = append(components, manifestComponent("pypi", name, version, path))
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return components, nil
}

func parsePyproject(path string) ([]SBOMComponent, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	components := make([]SBOMComponent, 0)
	insideDependencies := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "[") {
			insideDependencies = strings.Contains(line, "dependencies")
			continue
		}
		if !insideDependencies || !strings.Contains(line, "=") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		name := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"")
		if name != "" && value != "" && !strings.HasPrefix(value, "[") {
			if parsedName, version, ok := splitDependency(name + value); ok {
				components = append(components, manifestComponent("pypi", parsedName, version, path))
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return components, nil
}

func splitDependency(value string) (string, string, bool) {
	for _, operator := range []string{"==", ">=", "<=", "~=", "!=", ">", "<"} {
		if index := strings.Index(value, operator); index > 0 {
			name := strings.TrimSpace(value[:index])
			version := strings.TrimSpace(value[index+len(operator):])
			return name, version, name != "" && version != ""
		}
	}
	return "", "", false
}

func manifestComponent(ecosystem, name, version, path string) SBOMComponent {
	return SBOMComponent{ID: ComponentID(ecosystem, name, version), Name: name, Version: version, Ecosystem: ecosystem, Locations: []string{path}, Source: filepath.Base(path)}
}

func mergeComponents(components []SBOMComponent) []SBOMComponent {
	merged := make(map[string]SBOMComponent, len(components))
	for _, component := range components {
		current := merged[component.ID]
		if current.ID == "" {
			current = component
		}
		current.Locations = appendUnique(current.Locations, component.Locations...)
		merged[component.ID] = current
	}
	result := make([]SBOMComponent, 0, len(merged))
	for _, component := range merged {
		result = append(result, component)
	}
	StableSortComponents(result)
	return result
}
