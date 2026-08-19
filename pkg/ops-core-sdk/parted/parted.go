// Package parted provides disk partition management operations.
package parted

import (
	"fmt"
	"os/exec"
	"strings"
)

// Partition represents a partition.
type Partition struct {
	Number int    `json:"number"`
	Start  string `json:"start"`
	End    string `json:"end"`
	Size   string `json:"size"`
	FSType string `json:"fstype"`
	Name   string `json:"name"`
	Flags  string `json:"flags"`
}

// ListResult represents the result of listing partitions.
type ListResult struct {
	Device     string      `json:"device"`
	Model      string      `json:"model"`
	Size       string      `json:"size"`
	Partitions []Partition `json:"partitions"`
}

// ActionResult represents the result of a partition action.
type ActionResult struct {
	Changed bool   `json:"changed"`
	Message string `json:"message"`
}

// runParted executes a parted command.
func runParted(args ...string) (string, error) {
	cmd := exec.Command("parted", append([]string{"-s", "-m"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("parted failed: %w (output: %s)", err, string(out))
	}
	return string(out), nil
}

// List returns partitions on a device.
func List(device string) (*ListResult, error) {
	out, err := runParted(device, "unit", "s", "print")
	if err != nil {
		return nil, err
	}

	result := &ListResult{
		Device:     device,
		Partitions: make([]Partition, 0),
	}

	lines := strings.Split(out, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "BYT;") {
			continue
		}

		fields := strings.Split(line, ":")
		if len(fields) >= 7 {
			// Check if this is the device header line (e.g., "/dev/sda:...:...")
			if strings.HasPrefix(fields[0], "/dev/") && fields[1] == "" {
				result.Model = fields[2]
				result.Size = fields[3]
				continue
			}

			// Partition line: number:start:end:size:fstype:name:flags
			var num int
			fmt.Sscanf(fields[0], "%d", &num)
			if num > 0 {
				result.Partitions = append(result.Partitions, Partition{
					Number: num,
					Start:  fields[1],
					End:    fields[2],
					Size:   fields[3],
					FSType: fields[4],
					Name:   fields[5],
					Flags:  fields[6],
				})
			}
		}
	}
	return result, nil
}

// MkLabel creates a new partition table (disklabel).
func MkLabel(device string, labelType string) (*ActionResult, error) {
	if labelType == "" {
		labelType = "gpt"
	}
	_, err := runParted(device, "mklabel", labelType)
	if err != nil {
		return nil, err
	}
	return &ActionResult{Changed: true, Message: fmt.Sprintf("Created %s partition table on %s", labelType, device)}, nil
}

// MkPart creates a new partition.
func MkPart(device string, partType string, fsType string, start string, end string) (*ActionResult, error) {
	args := []string{device, "mkpart"}
	if partType != "" {
		args = append(args, partType)
	}
	if fsType != "" {
		args = append(args, fsType)
	}
	args = append(args, start, end)

	_, err := runParted(args...)
	if err != nil {
		return nil, err
	}
	return &ActionResult{Changed: true, Message: fmt.Sprintf("Created partition on %s (%s-%s)", device, start, end)}, nil
}

// Rm removes a partition.
func Rm(device string, number int) (*ActionResult, error) {
	_, err := runParted(device, "rm", fmt.Sprintf("%d", number))
	if err != nil {
		return nil, err
	}
	return &ActionResult{Changed: true, Message: fmt.Sprintf("Removed partition %d from %s", number, device)}, nil
}
