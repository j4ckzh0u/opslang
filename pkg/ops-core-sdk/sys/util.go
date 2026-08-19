package sys

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"os"
	"os/exec"
	"strings"
)

// MACAddressResult represents the result of getting a MAC address.
type MACAddressResult struct {
	Interface string `json:"interface"`
	MAC       string `json:"mac"`
}

// MACListResult represents all non-loopback MAC addresses.
type MACListResult struct {
	Addresses []MACAddressResult `json:"addresses"`
}

// UUIDResult represents the result of UUID generation.
type UUIDResult struct {
	UUID string `json:"uuid"`
}

// PasswordResult represents the result of password generation.
type PasswordResult struct {
	Password string `json:"password"`
	Length   int    `json:"length"`
}

// UUID generates a random UUID (v4).
func UUID() (UUIDResult, error) {
	var uuid [16]byte
	_, err := rand.Read(uuid[:])
	if err != nil {
		return UUIDResult{}, fmt.Errorf("sys.UUID: failed to generate UUID: %w", err)
	}
	// Set version 4
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	// Set variant 10
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	return UUIDResult{
		UUID: fmt.Sprintf("%s-%s-%s-%s-%s",
			hex.EncodeToString(uuid[0:4]),
			hex.EncodeToString(uuid[4:6]),
			hex.EncodeToString(uuid[6:8]),
			hex.EncodeToString(uuid[8:10]),
			hex.EncodeToString(uuid[10:16])),
	}, nil
}

// RandomPassword generates a cryptographically secure random password.
// length specifies the password length (minimum 8).
// useSpecial, useNumbers, useUppercase control character set inclusion.
func RandomPassword(length int, useSpecial bool, useNumbers bool, useUppercase bool) (PasswordResult, error) {
	if length < 8 {
		length = 8
	}

	const (
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		numbers   = "0123456789"
		special   = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	)

	charset := lowercase
	if useUppercase {
		charset += uppercase
	}
	if useNumbers {
		charset += numbers
	}
	if useSpecial {
		charset += special
	}

	// Ensure at least one character from each selected set
	var password strings.Builder
	if useUppercase {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(uppercase))))
		password.WriteByte(uppercase[idx.Int64()])
	}
	if useNumbers {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(numbers))))
		password.WriteByte(numbers[idx.Int64()])
	}
	if useSpecial {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(special))))
		password.WriteByte(special[idx.Int64()])
	}

	// Fill remaining length with random characters from the full charset
	remaining := length - password.Len()
	for i := 0; i < remaining; i++ {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		password.WriteByte(charset[idx.Int64()])
	}

	// Shuffle the password to avoid predictable positions
	runes := []rune(password.String())
	for i := len(runes) - 1; i > 0; i-- {
		jBig, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		j := int(jBig.Int64())
		runes[i], runes[j] = runes[j], runes[i]
	}

	return PasswordResult{
		Password: string(runes),
		Length:   len(runes),
	}, nil
}

// MACAddress returns the MAC address of a specific network interface.
// If iface is empty, returns the first non-loopback interface's MAC.
func MACAddress(iface string) (MACAddressResult, error) {
	if iface != "" {
		ifi, err := net.InterfaceByName(iface)
		if err != nil {
			return MACAddressResult{}, fmt.Errorf("interface %s not found: %w", iface, err)
		}
		if len(ifi.HardwareAddr) == 0 {
			return MACAddressResult{}, fmt.Errorf("interface %s has no MAC address", iface)
		}
		return MACAddressResult{
			Interface: ifi.Name,
			MAC:       ifi.HardwareAddr.String(),
		}, nil
	}

	// Find first non-loopback interface with a MAC
	ifaces, err := net.Interfaces()
	if err != nil {
		return MACAddressResult{}, fmt.Errorf("failed to list interfaces: %w", err)
	}
	for _, ifi := range ifaces {
		if ifi.Flags&net.FlagLoopback != 0 {
			continue
		}
		if len(ifi.HardwareAddr) > 0 {
			return MACAddressResult{
				Interface: ifi.Name,
				MAC:       ifi.HardwareAddr.String(),
			}, nil
		}
	}
	return MACAddressResult{}, fmt.Errorf("no interface with MAC address found")
}

// MACAddresses returns MAC addresses for all non-loopback interfaces.
func MACAddresses() (MACListResult, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return MACListResult{}, fmt.Errorf("failed to list interfaces: %w", err)
	}
	result := MACListResult{Addresses: make([]MACAddressResult, 0)}
	for _, ifi := range ifaces {
		if ifi.Flags&net.FlagLoopback != 0 {
			continue
		}
		if len(ifi.HardwareAddr) > 0 {
			result.Addresses = append(result.Addresses, MACAddressResult{
				Interface: ifi.Name,
				MAC:       ifi.HardwareAddr.String(),
			})
		}
	}
	return result, nil
}

// DmidecodeResult represents SMBIOS/DMI hardware information.
type DmidecodeResult struct {
	BiosVendor     string `json:"bios_vendor"`
	BiosVersion    string `json:"bios_version"`
	SystemVendor   string `json:"system_vendor"`
	ProductName    string `json:"product_name"`
	SerialNumber   string `json:"serial_number"`
	BoardVendor    string `json:"board_vendor"`
	BoardName      string `json:"board_name"`
	ChassisType    string `json:"chassis_type"`
}

// PciDevice represents a PCI device.
type PciDevice struct {
	Slot        string `json:"slot"`
	Class       string `json:"class"`
	Vendor      string `json:"vendor"`
	Device      string `json:"device"`
	SVendor     string `json:"svendor,omitempty"`
	SDevice     string `json:"sdevice,omitempty"`
	Revision    string `json:"revision,omitempty"`
}

// LsPciResult represents the result of listing PCI devices.
type LsPciResult struct {
	Devices []PciDevice `json:"devices"`
}

// BlkDevice represents a block device.
type BlkDevice struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Size       string `json:"size"`
	MountPoint string `json:"mountpoint,omitempty"`
	FsType     string `json:"fstype,omitempty"`
	Label      string `json:"label,omitempty"`
	UUID       string `json:"uuid,omitempty"`
	Model      string `json:"model,omitempty"`
	Children   []BlkDevice `json:"children,omitempty"`
}

// LsBlkResult represents the result of listing block devices.
type LsBlkResult struct {
	Devices []BlkDevice `json:"devices"`
}

// Dmidecode reads hardware information from /sys/class/dmi/id/ files.
// Falls back to running dmidecode command if available.
func Dmidecode() (DmidecodeResult, error) {
	result := DmidecodeResult{}
	readDMIField := func(path string) string {
		data, err := os.ReadFile(path)
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(data))
	}

	result.BiosVendor = readDMIField("/sys/class/dmi/id/bios_vendor")
	result.BiosVersion = readDMIField("/sys/class/dmi/id/bios_version")
	result.SystemVendor = readDMIField("/sys/class/dmi/id/sys_vendor")
	result.ProductName = readDMIField("/sys/class/dmi/id/product_name")
	result.SerialNumber = readDMIField("/sys/class/dmi/id/product_serial")
	result.BoardVendor = readDMIField("/sys/class/dmi/id/board_vendor")
	result.BoardName = readDMIField("/sys/class/dmi/id/board_name")
	result.ChassisType = readDMIField("/sys/class/dmi/id/chassis_type")

	return result, nil
}

// LsPci lists PCI devices using the lspci command.
func LsPci() (LsPciResult, error) {
	out, err := exec.Command("lspci", "-mm").CombinedOutput()
	if err != nil {
		return LsPciResult{}, fmt.Errorf("lspci failed: %w (output: %s)", err, string(out))
	}

	result := LsPciResult{Devices: make([]PciDevice, 0)}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// lspci -mm outputs: "slot" "class" "vendor" "device" "svendor" "sdevice" "revision"
		dev := PciDevice{}
		// Parse quoted fields
		quoted := strings.Split(line, "\"")
		if len(quoted) >= 7 {
			dev.Slot = quoted[1]
			dev.Class = quoted[3]
			dev.Vendor = quoted[5]
			if len(quoted) >= 9 {
				dev.Device = quoted[7]
			}
			if len(quoted) >= 11 {
				dev.SVendor = quoted[9]
			}
			if len(quoted) >= 13 {
				dev.SDevice = quoted[11]
			}
			if len(quoted) >= 15 {
				dev.Revision = quoted[13]
			}
		} else {
			// Fallback for non-quoted output
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				dev.Slot = parts[0]
				dev.Class = strings.Join(parts[1:], " ")
			}
		}
		result.Devices = append(result.Devices, dev)
	}
	return result, nil
}

// LsBlk lists block devices using the lsblk command with JSON output.
func LsBlk() (LsBlkResult, error) {
	out, err := exec.Command("lsblk", "--json", "-o", "NAME,TYPE,SIZE,MOUNTPOINT,FSTYPE,LABEL,UUID,MODEL").CombinedOutput()
	if err != nil {
		return LsBlkResult{}, fmt.Errorf("lsblk failed: %w (output: %s)", err, string(out))
	}

	// Parse JSON output from lsblk
	var lsblkOut struct {
		BlockDevices []struct {
			Name       string `json:"name"`
			Type       string `json:"type"`
			Size       string `json:"size"`
			MountPoint string `json:"mountpoint"`
			FsType     string `json:"fstype"`
			Label      string `json:"label"`
			UUID       string `json:"uuid"`
			Model      string `json:"model"`
			Children   []struct {
				Name       string `json:"name"`
				Type       string `json:"type"`
				Size       string `json:"size"`
				MountPoint string `json:"mountpoint"`
				FsType     string `json:"fstype"`
				Label      string `json:"label"`
				UUID       string `json:"uuid"`
				Model      string `json:"model"`
			} `json:"children"`
		} `json:"blockdevices"`
	}

	if err := json.Unmarshal(out, &lsblkOut); err != nil {
		return LsBlkResult{}, fmt.Errorf("failed to parse lsblk output: %w", err)
	}

	result := LsBlkResult{Devices: make([]BlkDevice, 0, len(lsblkOut.BlockDevices))}
	for _, bd := range lsblkOut.BlockDevices {
		dev := BlkDevice{
			Name:       bd.Name,
			Type:       bd.Type,
			Size:       bd.Size,
			MountPoint: bd.MountPoint,
			FsType:     bd.FsType,
			Label:      bd.Label,
			UUID:       bd.UUID,
			Model:      bd.Model,
		}
		for _, child := range bd.Children {
			dev.Children = append(dev.Children, BlkDevice{
				Name:       child.Name,
				Type:       child.Type,
				Size:       child.Size,
				MountPoint: child.MountPoint,
				FsType:     child.FsType,
				Label:      child.Label,
				UUID:       child.UUID,
				Model:      child.Model,
			})
		}
		result.Devices = append(result.Devices, dev)
	}
	return result, nil
}
