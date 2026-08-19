// Package dmidecode provides hardware information via DMI tables.
package dmidecode

import (
	"fmt"
	"os/exec"
	"strings"
)

// SystemInfo represents DMI system information.
type SystemInfo struct {
	Manufacturer string `json:"manufacturer,omitempty"`
	ProductName  string `json:"product_name,omitempty"`
	SerialNumber string `json:"serial_number,omitempty"`
	UUID         string `json:"uuid,omitempty"`
	Family       string `json:"family,omitempty"`
	SKU          string `json:"sku,omitempty"`
	Error        string `json:"error,omitempty"`
}

// BIOSInfo represents BIOS information.
type BIOSInfo struct {
	Vendor      string `json:"vendor,omitempty"`
	Version     string `json:"version,omitempty"`
	ReleaseDate string `json:"release_date,omitempty"`
	Address     string `json:"address,omitempty"`
	RuntimeSize string `json:"runtime_size,omitempty"`
	Error       string `json:"error,omitempty"`
}

// ChassisInfo represents chassis information.
type ChassisInfo struct {
	Manufacturer string `json:"manufacturer,omitempty"`
	Type         string `json:"type,omitempty"`
	SerialNumber string `json:"serial_number,omitempty"`
	AssetTag     string `json:"asset_tag,omitempty"`
	Error        string `json:"error,omitempty"`
}

// CPUInfo represents processor information.
type CPUInfo struct {
	SocketDesignation string `json:"socket_designation,omitempty"`
	Manufacturer      string `json:"manufacturer,omitempty"`
	Version           string `json:"version,omitempty"`
	MaxSpeed          string `json:"max_speed,omitempty"`
	CurrentSpeed      string `json:"current_speed,omitempty"`
	CoreCount         string `json:"core_count,omitempty"`
	ThreadCount       string `json:"thread_count,omitempty"`
	Error             string `json:"error,omitempty"`
}

func dmidecode(args ...string) (string, error) {
	cmd := exec.Command("dmidecode", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func getValue(output, key string) string {
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, key+":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

// System returns system information (type 1).
func System() SystemInfo {
	out, err := dmidecode("-t", "1")
	if err != nil {
		return SystemInfo{Error: fmt.Sprintf("dmidecode failed: %s: %s", err, out)}
	}
	return SystemInfo{
		Manufacturer: getValue(out, "Manufacturer"),
		ProductName:  getValue(out, "Product Name"),
		SerialNumber: getValue(out, "Serial Number"),
		UUID:         getValue(out, "UUID"),
		Family:       getValue(out, "Family"),
		SKU:          getValue(out, "SKU Number"),
	}
}

// BIOS returns BIOS information (type 0).
func BIOS() BIOSInfo {
	out, err := dmidecode("-t", "0")
	if err != nil {
		return BIOSInfo{Error: fmt.Sprintf("dmidecode failed: %s: %s", err, out)}
	}
	return BIOSInfo{
		Vendor:      getValue(out, "Vendor"),
		Version:     getValue(out, "Version"),
		ReleaseDate: getValue(out, "Release Date"),
		Address:     getValue(out, "Address"),
		RuntimeSize: getValue(out, "Runtime Size"),
	}
}

// Chassis returns chassis information (type 3).
func Chassis() ChassisInfo {
	out, err := dmidecode("-t", "3")
	if err != nil {
		return ChassisInfo{Error: fmt.Sprintf("dmidecode failed: %s: %s", err, out)}
	}
	return ChassisInfo{
		Manufacturer: getValue(out, "Manufacturer"),
		Type:         getValue(out, "Type"),
		SerialNumber: getValue(out, "Serial Number"),
		AssetTag:     getValue(out, "Asset Tag"),
	}
}

// Processor returns CPU information (type 4).
func Processor() CPUInfo {
	out, err := dmidecode("-t", "4")
	if err != nil {
		return CPUInfo{Error: fmt.Sprintf("dmidecode failed: %s: %s", err, out)}
	}
	return CPUInfo{
		SocketDesignation: getValue(out, "Socket Designation"),
		Manufacturer:      getValue(out, "Manufacturer"),
		Version:           getValue(out, "Version"),
		MaxSpeed:          getValue(out, "Max Speed"),
		CurrentSpeed:      getValue(out, "Current Speed"),
		CoreCount:         getValue(out, "Core Count"),
		ThreadCount:       getValue(out, "Thread Count"),
	}
}

// Keyword reads a specific DMI keyword directly.
func Keyword(keyword string) (string, error) {
	if keyword == "" {
		return "", fmt.Errorf("keyword is required")
	}
	out, err := dmidecode("-s", keyword)
	if err != nil {
		return "", fmt.Errorf("dmidecode keyword failed: %w: %s", err, out)
	}
	return strings.TrimSpace(out), nil
}
