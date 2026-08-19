// Package getent provides system database enumeration (users, groups, services, networks, protocols).
// Uses Go standard library to read /etc/passwd, /etc/group, /etc/services, /etc/protocols, /etc/shells.
package getent

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// UserEntry represents a passwd entry.
type UserEntry struct {
	Username string `json:"username"`
	UID      int    `json:"uid"`
	GID      int    `json:"gid"`
	Gecos    string `json:"gecos"`
	Home     string `json:"home"`
	Shell    string `json:"shell"`
}

// GroupEntry represents a group entry.
type GroupEntry struct {
	Name    string   `json:"name"`
	GID     int      `json:"gid"`
	Members []string `json:"members"`
}

// ServiceEntry represents a /etc/services entry.
type ServiceEntry struct {
	Name     string `json:"name"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Alias    string `json:"alias,omitempty"`
}

// NetworkEntry represents a /etc/networks entry.
type NetworkEntry struct {
	Name    string `json:"name"`
	Network string `json:"network"`
}

// ProtocolEntry represents a /etc/protocols entry.
type ProtocolEntry struct {
	Name     string `json:"name"`
	Number   int    `json:"number"`
	Alias    string `json:"alias,omitempty"`
}

// LookupResult is returned by all lookup functions.
type LookupResult struct {
	Database string      `json:"database"`
	Key      string      `json:"key"`
	Found    bool        `json:"found"`
	Entries  interface{} `json:"entries,omitempty"`
	Count    int         `json:"count"`
}

// ShellList is returned by Shells.
type ShellList struct {
	Shells []string `json:"shells"`
	Count  int      `json:"count"`
}

// GetPasswd returns all user entries from /etc/passwd.
func GetPasswd() ([]UserEntry, error) {
	f, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, fmt.Errorf("failed to open /etc/passwd: %w", err)
	}
	defer f.Close()

	var entries []UserEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, ":")
		if len(fields) < 7 {
			continue
		}
		uid, _ := strconv.Atoi(fields[2])
		gid, _ := strconv.Atoi(fields[3])
		entries = append(entries, UserEntry{
			Username: fields[0],
			UID:      uid,
			GID:      gid,
			Gecos:    fields[4],
			Home:     fields[5],
			Shell:    fields[6],
		})
	}
	return entries, scanner.Err()
}

// LookupUser finds a user by name or UID.
func LookupUser(key string) (LookupResult, error) {
	if key == "" {
		return LookupResult{}, fmt.Errorf("key is required")
	}
	entries, err := GetPasswd()
	if err != nil {
		return LookupResult{Database: "passwd", Key: key}, err
	}
	targetUID, uidErr := strconv.Atoi(key)
	for _, e := range entries {
		if e.Username == key || (uidErr == nil && e.UID == targetUID) {
			return LookupResult{Database: "passwd", Key: key, Found: true, Entries: []UserEntry{e}, Count: 1}, nil
		}
	}
	return LookupResult{Database: "passwd", Key: key, Found: false, Count: 0}, nil
}

// GetGroups returns all group entries from /etc/group.
func GetGroups() ([]GroupEntry, error) {
	f, err := os.Open("/etc/group")
	if err != nil {
		return nil, fmt.Errorf("failed to open /etc/group: %w", err)
	}
	defer f.Close()

	var entries []GroupEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, ":")
		if len(fields) < 4 {
			continue
		}
		gid, _ := strconv.Atoi(fields[2])
		var members []string
		if fields[3] != "" {
			members = strings.Split(fields[3], ",")
		}
		entries = append(entries, GroupEntry{
			Name:    fields[0],
			GID:     gid,
			Members: members,
		})
	}
	return entries, scanner.Err()
}

// LookupGroup finds a group by name or GID.
func LookupGroup(key string) (LookupResult, error) {
	if key == "" {
		return LookupResult{}, fmt.Errorf("key is required")
	}
	entries, err := GetGroups()
	if err != nil {
		return LookupResult{Database: "group", Key: key}, err
	}
	targetGID, gidErr := strconv.Atoi(key)
	for _, e := range entries {
		if e.Name == key || (gidErr == nil && e.GID == targetGID) {
			return LookupResult{Database: "group", Key: key, Found: true, Entries: []GroupEntry{e}, Count: 1}, nil
		}
	}
	return LookupResult{Database: "group", Key: key, Found: false, Count: 0}, nil
}

// GetServices returns all entries from /etc/services.
func GetServices() ([]ServiceEntry, error) {
	f, err := os.Open("/etc/services")
	if err != nil {
		return nil, fmt.Errorf("failed to open /etc/services: %w", err)
	}
	defer f.Close()

	var entries []ServiceEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Remove inline comments
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		portProto := strings.SplitN(fields[1], "/", 2)
		if len(portProto) != 2 {
			continue
		}
		port, err := strconv.Atoi(portProto[0])
		if err != nil {
			continue
		}
		s := ServiceEntry{Name: fields[0], Port: port, Protocol: portProto[1]}
		if len(fields) > 2 {
			s.Alias = fields[2]
		}
		entries = append(entries, s)
	}
	return entries, scanner.Err()
}

// LookupService finds a service by name or port/protocol.
func LookupService(key string) (LookupResult, error) {
	if key == "" {
		return LookupResult{}, fmt.Errorf("key is required")
	}
	entries, err := GetServices()
	if err != nil {
		return LookupResult{Database: "services", Key: key}, err
	}
	var found []ServiceEntry
	for _, e := range entries {
		if e.Name == key || fmt.Sprintf("%d/%s", e.Port, e.Protocol) == key {
			found = append(found, e)
		}
	}
	if len(found) == 0 {
		return LookupResult{Database: "services", Key: key, Found: false, Count: 0}, nil
	}
	return LookupResult{Database: "services", Key: key, Found: true, Entries: found, Count: len(found)}, nil
}

// GetProtocols returns all entries from /etc/protocols.
func GetProtocols() ([]ProtocolEntry, error) {
	f, err := os.Open("/etc/protocols")
	if err != nil {
		return nil, fmt.Errorf("failed to open /etc/protocols: %w", err)
	}
	defer f.Close()

	var entries []ProtocolEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		num, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		p := ProtocolEntry{Name: fields[0], Number: num}
		if len(fields) > 2 {
			p.Alias = fields[2]
		}
		entries = append(entries, p)
	}
	return entries, scanner.Err()
}

// LookupProtocol finds a protocol by name or number.
func LookupProtocol(key string) (LookupResult, error) {
	if key == "" {
		return LookupResult{}, fmt.Errorf("key is required")
	}
	entries, err := GetProtocols()
	if err != nil {
		return LookupResult{Database: "protocols", Key: key}, err
	}
	targetNum, numErr := strconv.Atoi(key)
	for _, e := range entries {
		if e.Name == key || (numErr == nil && e.Number == targetNum) {
			return LookupResult{Database: "protocols", Key: key, Found: true, Entries: []ProtocolEntry{e}, Count: 1}, nil
		}
	}
	return LookupResult{Database: "protocols", Key: key, Found: false, Count: 0}, nil
}

// Shells returns all valid login shells from /etc/shells.
func Shells() (ShellList, error) {
	f, err := os.Open("/etc/shells")
	if err != nil {
		return ShellList{}, fmt.Errorf("failed to open /etc/shells: %w", err)
	}
	defer f.Close()

	var shells []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		shells = append(shells, line)
	}
	if err := scanner.Err(); err != nil {
		return ShellList{}, err
	}
	return ShellList{Shells: shells, Count: len(shells)}, nil
}
