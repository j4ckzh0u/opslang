// Package nmcli provides NetworkManager connection management operations.
package nmcli

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// ActionResult represents the result of a nmcli operation.
type ActionResult struct {
	Name    string `json:"name"`
	Changed bool   `json:"changed"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// ConnectionInfo represents information about a NetworkManager connection.
type ConnectionInfo struct {
	Name    string `json:"name"`
	UUID    string `json:"uuid"`
	Type    string `json:"type"`
	Device  string `json:"device"`
	State   string `json:"state"`
}

// ListResult represents the result of listing connections.
type ListResult struct {
	Connections []ConnectionInfo `json:"connections"`
}

// DeviceInfo represents information about a network device.
type DeviceInfo struct {
	Device  string `json:"device"`
	Type    string `json:"type"`
	State   string `json:"state"`
	Connection string `json:"connection"`
}

// DeviceListResult represents the result of listing devices.
type DeviceListResult struct {
	Devices []DeviceInfo `json:"devices"`
}

// ConnectionSettings represents connection settings.
type ConnectionSettings struct {
	Settings map[string]map[string]interface{} `json:"settings"`
}

// Add creates a new NetworkManager connection.
func Add(name string, connType string, settings map[string]string) (ActionResult, error) {
	if name == "" || connType == "" {
		return ActionResult{}, fmt.Errorf("connection name and type are required")
	}

	args := []string{"connection", "add", "type", connType, "con-name", name, "save", "yes"}
	for k, v := range settings {
		args = append(args, fmt.Sprintf("%s", k), v)
	}

	out, err := exec.Command("nmcli", args...).CombinedOutput()
	if err != nil {
		return ActionResult{Name: name, Success: false, Error: string(out)}, fmt.Errorf("nmcli connection add failed: %w (output: %s)", err, string(out))
	}

	return ActionResult{Name: name, Changed: true, Success: true}, nil
}

// Modify modifies an existing NetworkManager connection.
func Modify(name string, settings map[string]string) (ActionResult, error) {
	if name == "" {
		return ActionResult{}, fmt.Errorf("connection name is required")
	}

	args := []string{"connection", "modify", name}
	for k, v := range settings {
		args = append(args, k, v)
	}

	out, err := exec.Command("nmcli", args...).CombinedOutput()
	if err != nil {
		return ActionResult{Name: name, Success: false, Error: string(out)}, fmt.Errorf("nmcli connection modify failed: %w (output: %s)", err, string(out))
	}

	return ActionResult{Name: name, Changed: true, Success: true}, nil
}

// Delete deletes a NetworkManager connection.
func Delete(name string) (ActionResult, error) {
	if name == "" {
		return ActionResult{}, fmt.Errorf("connection name is required")
	}

	out, err := exec.Command("nmcli", "connection", "delete", name).CombinedOutput()
	if err != nil {
		return ActionResult{Name: name, Success: false, Error: string(out)}, fmt.Errorf("nmcli connection delete failed: %w (output: %s)", err, string(out))
	}

	return ActionResult{Name: name, Changed: true, Success: true}, nil
}

// Up activates a NetworkManager connection.
func Up(name string) (ActionResult, error) {
	if name == "" {
		return ActionResult{}, fmt.Errorf("connection name is required")
	}

	out, err := exec.Command("nmcli", "connection", "up", name).CombinedOutput()
	if err != nil {
		return ActionResult{Name: name, Success: false, Error: string(out)}, fmt.Errorf("nmcli connection up failed: %w (output: %s)", err, string(out))
	}

	return ActionResult{Name: name, Changed: true, Success: true}, nil
}

// Down deactivates a NetworkManager connection.
func Down(name string) (ActionResult, error) {
	if name == "" {
		return ActionResult{}, fmt.Errorf("connection name is required")
	}

	out, err := exec.Command("nmcli", "connection", "down", name).CombinedOutput()
	if err != nil {
		return ActionResult{Name: name, Success: false, Error: string(out)}, fmt.Errorf("nmcli connection down failed: %w (output: %s)", err, string(out))
	}

	return ActionResult{Name: name, Changed: true, Success: true}, nil
}

// List lists all NetworkManager connections.
func List() (ListResult, error) {
	out, err := exec.Command("nmcli", "connection", "show").CombinedOutput()
	if err != nil {
		return ListResult{}, fmt.Errorf("nmcli connection show failed: %w (output: %s)", err, string(out))
	}

	result := ListResult{Connections: make([]ConnectionInfo, 0)}
	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		return result, nil
	}

	// Skip header line
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		// Parse fixed-width columns
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			conn := ConnectionInfo{
				Name: fields[0],
				UUID: fields[1],
				Type: fields[2],
			}
			if len(fields) >= 5 {
				conn.Device = fields[3]
				conn.State = strings.Join(fields[4:], " ")
			}
			result.Connections = append(result.Connections, conn)
		}
	}
	return result, nil
}

// Show shows detailed information about a connection.
func Show(name string) (ConnectionSettings, error) {
	if name == "" {
		return ConnectionSettings{}, fmt.Errorf("connection name is required")
	}

	out, err := exec.Command("nmcli", "connection", "show", name).CombinedOutput()
	if err != nil {
		return ConnectionSettings{}, fmt.Errorf("nmcli connection show failed: %w (output: %s)", err, string(out))
	}

	settings := ConnectionSettings{Settings: make(map[string]map[string]interface{})}
	currentSection := ""
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Contains(line, ":") && !strings.Contains(line, ".") {
			// Section header
			currentSection = strings.Split(line, ":")[0]
			settings.Settings[currentSection] = make(map[string]interface{})
		} else if strings.Contains(line, ":") && currentSection != "" {
			// Key-value pair
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				settings.Settings[currentSection][key] = value
			}
		}
	}
	return settings, nil
}

// ListDevices lists all network devices.
func ListDevices() (DeviceListResult, error) {
	out, err := exec.Command("nmcli", "device", "status").CombinedOutput()
	if err != nil {
		return DeviceListResult{}, fmt.Errorf("nmcli device status failed: %w (output: %s)", err, string(out))
	}

	result := DeviceListResult{Devices: make([]DeviceInfo, 0)}
	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		return result, nil
	}

	// Skip header line
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			dev := DeviceInfo{
				Device: fields[0],
				Type:   fields[1],
				State:  fields[2],
			}
			if len(fields) >= 4 {
				dev.Connection = strings.Join(fields[3:], " ")
			}
			result.Devices = append(result.Devices, dev)
		}
	}
	return result, nil
}

// Reload reloads NetworkManager configuration.
func Reload() (ActionResult, error) {
	out, err := exec.Command("nmcli", "general", "reload").CombinedOutput()
	if err != nil {
		return ActionResult{Success: false, Error: string(out)}, fmt.Errorf("nmcli general reload failed: %w (output: %s)", err, string(out))
	}

	return ActionResult{Changed: true, Success: true}, nil
}

// GetGeneralStatus gets general NetworkManager status as JSON.
func GetGeneralStatus() (map[string]interface{}, error) {
	out, err := exec.Command("nmcli", "-t", "-f", "all", "general", "status").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("nmcli general status failed: %w (output: %s)", err, string(out))
	}

	result := make(map[string]interface{})
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				result[key] = value
			}
		}
	}

	// Try to get JSON status
	jsonOut, err := exec.Command("nmcli", "-p", "-t", "general", "status").CombinedOutput()
	if err == nil {
		var jsonResult map[string]interface{}
		if err := json.Unmarshal(jsonOut, &jsonResult); err == nil {
			return jsonResult, nil
		}
	}

	return result, nil
}
