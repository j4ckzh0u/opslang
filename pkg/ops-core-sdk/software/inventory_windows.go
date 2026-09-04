//go:build windows

package software

import (
	"fmt"
	"strings"

	"golang.org/x/sys/windows/registry"
)

var uninstallRoots = []struct {
	hive    registry.Key
	path    string
	manager string
}{
	{registry.LOCAL_MACHINE, `Software\Microsoft\Windows\CurrentVersion\Uninstall`, "registry"},
	{registry.LOCAL_MACHINE, `Software\Wow6432Node\Microsoft\Windows\CurrentVersion\Uninstall`, "registry32"},
	{registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Uninstall`, "registry_user"},
}

func collectPackages() ([]Package, []Error) {
	items := make([]Package, 0)
	errors := make([]Error, 0)
	seen := make(map[string]bool)
	for _, root := range uninstallRoots {
		key, err := registry.OpenKey(root.hive, root.path, registry.READ)
		if err != nil {
			errors = append(errors, Error{Scope: "windows_registry", Item: root.path, Message: err.Error()})
			continue
		}
		names, err := key.ReadSubKeyNames(-1)
		key.Close()
		if err != nil {
			errors = append(errors, Error{Scope: "windows_registry", Item: root.path, Message: err.Error()})
			continue
		}
		for _, name := range names {
			item, err := readUninstallEntry(root.hive, root.path, name, root.manager)
			if err != nil {
				errors = append(errors, Error{Scope: "windows_package", Item: name, Message: err.Error()})
				continue
			}
			identity := strings.ToLower(item.Name + "|" + item.Version + "|" + item.InstallLocation)
			if item.Name != "" && !seen[identity] {
				seen[identity] = true
				items = append(items, item)
			}
		}
	}
	return items, errors
}

func readUninstallEntry(hive registry.Key, root, name, manager string) (Package, error) {
	key, err := registry.OpenKey(hive, root+`\`+name, registry.READ)
	if err != nil {
		return Package{}, err
	}
	defer key.Close()
	read := func(value string) string { result, _, _ := key.GetStringValue(value); return strings.TrimSpace(result) }
	item := Package{Name: read("DisplayName"), Version: read("DisplayVersion"), Architecture: read("ReleaseType"), Manager: manager, InstallLocation: read("InstallLocation"), Publisher: read("Publisher"), UninstallCommand: read("UninstallString")}
	if item.Name == "" {
		return Package{}, fmt.Errorf("missing DisplayName")
	}
	return item, nil
}
