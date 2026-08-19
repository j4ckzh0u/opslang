// Package memcached provides Memcached operations via network protocol.
package memcached

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// StatsResult is returned by stats.
type StatsResult struct {
	Version      string `json:"version"`
	CurrItems    int64  `json:"curr_items"`
	CurrConns    int64  `json:"curr_connections"`
	TotalItems   int64  `json:"total_items"`
	GetHits      int64  `json:"get_hits"`
	GetMisses    int64  `json:"get_misses"`
	Uptime       int64  `json:"uptime"`
	Error        string `json:"error,omitempty"`
}

// GetResult is returned by get operations.
type GetResult struct {
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
	Found bool   `json:"found"`
	Error string `json:"error,omitempty"`
}

// SetResult is returned by set operations.
type SetResult struct {
	Key     string `json:"key"`
	Success bool   `json:"success"`
	Stored  bool   `json:"stored"`
	Error   string `json:"error,omitempty"`
}

// DeleteResult is returned by delete operations.
type DeleteResult struct {
	Key     string `json:"key"`
	Success bool   `json:"success"`
	Deleted bool   `json:"deleted"`
	Error   string `json:"error,omitempty"`
}

// FlushResult is returned by flush operations.
type FlushResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// VersionResult is returned by version check.
type VersionResult struct {
	Version string `json:"version"`
	Error   string `json:"error,omitempty"`
}

func dial(host string, port int) (net.Conn, error) {
	if host == "" {
		host = "127.0.0.1"
	}
	if port == 0 {
		port = 11211
	}
	return net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 5*time.Second)
}

func sendCmd(conn net.Conn, cmd string) (string, error) {
	fmt.Fprintf(conn, "%s\r\n", cmd)
	reader := bufio.NewReader(conn)
	var result strings.Builder
	for {
		line, err := reader.ReadString('\n')
		result.WriteString(line)
		if err != nil {
			return result.String(), err
		}
		// End of response markers
		trimmed := strings.TrimSpace(line)
		if trimmed == "END" || trimmed == "STORED" || trimmed == "DELETED" || trimmed == "NOT_FOUND" || trimmed == "OK" {
			break
		}
		if strings.HasPrefix(trimmed, "ERROR") || strings.HasPrefix(trimmed, "CLIENT_ERROR") || strings.HasPrefix(trimmed, "SERVER_ERROR") {
			break
		}
	}
	return result.String(), nil
}

// Get retrieves a value from Memcached.
func Get(key, host string, port int) GetResult {
	if key == "" {
		return GetResult{Error: "key is required"}
	}
	conn, err := dial(host, port)
	if err != nil {
		return GetResult{Key: key, Error: fmt.Sprintf("connect failed: %v", err)}
	}
	defer conn.Close()
	out, err := sendCmd(conn, "get "+key)
	if err != nil {
		return GetResult{Key: key, Error: fmt.Sprintf("get failed: %v", err)}
	}
	// Parse VALUE <key> <flags> <bytes>\r\n<data>\r\nEND
	if strings.Contains(out, "VALUE") {
		lines := strings.Split(out, "\r\n")
		if len(lines) >= 3 {
			return GetResult{Key: key, Value: lines[1], Found: true}
		}
	}
	return GetResult{Key: key, Found: false}
}

// Set stores a value in Memcached.
func Set(key, value, host string, port int, expiry int) SetResult {
	if key == "" {
		return SetResult{Error: "key is required"}
	}
	conn, err := dial(host, port)
	if err != nil {
		return SetResult{Key: key, Error: fmt.Sprintf("connect failed: %v", err)}
	}
	defer conn.Close()
	cmd := fmt.Sprintf("set %s 0 %d %d\r\n%s", key, expiry, len(value), value)
	out, err := sendCmd(conn, cmd)
	if err != nil {
		return SetResult{Key: key, Error: fmt.Sprintf("set failed: %v", err)}
	}
	if strings.Contains(out, "STORED") {
		return SetResult{Key: key, Success: true, Stored: true}
	}
	return SetResult{Key: key, Success: false}
}

// Delete removes a key from Memcached.
func Delete(key, host string, port int) DeleteResult {
	if key == "" {
		return DeleteResult{Error: "key is required"}
	}
	conn, err := dial(host, port)
	if err != nil {
		return DeleteResult{Key: key, Error: fmt.Sprintf("connect failed: %v", err)}
	}
	defer conn.Close()
	out, err := sendCmd(conn, "delete "+key)
	if err != nil {
		return DeleteResult{Key: key, Error: fmt.Sprintf("delete failed: %v", err)}
	}
	if strings.Contains(out, "DELETED") {
		return DeleteResult{Key: key, Success: true, Deleted: true}
	}
	return DeleteResult{Key: key, Success: false, Deleted: false}
}

// FlushAll clears all keys.
func FlushAll(host string, port int) FlushResult {
	conn, err := dial(host, port)
	if err != nil {
		return FlushResult{Error: fmt.Sprintf("connect failed: %v", err)}
	}
	defer conn.Close()
	out, err := sendCmd(conn, "flush_all")
	if err != nil {
		return FlushResult{Error: fmt.Sprintf("flush_all failed: %v", err)}
	}
	if strings.Contains(out, "OK") {
		return FlushResult{Success: true}
	}
	return FlushResult{Success: false}
}

// Stats returns server statistics.
func Stats(host string, port int) StatsResult {
	conn, err := dial(host, port)
	if err != nil {
		return StatsResult{Error: fmt.Sprintf("connect failed: %v", err)}
	}
	defer conn.Close()
	out, err := sendCmd(conn, "stats")
	if err != nil {
		return StatsResult{Error: fmt.Sprintf("stats failed: %v", err)}
	}
	result := StatsResult{}
	for _, line := range strings.Split(out, "\r\n") {
		if strings.HasPrefix(line, "STAT ") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				switch parts[1] {
				case "version":
					result.Version = parts[2]
				case "curr_items":
					result.CurrItems, _ = strconv.ParseInt(parts[2], 10, 64)
				case "curr_connections":
					result.CurrConns, _ = strconv.ParseInt(parts[2], 10, 64)
				case "total_items":
					result.TotalItems, _ = strconv.ParseInt(parts[2], 10, 64)
				case "get_hits":
					result.GetHits, _ = strconv.ParseInt(parts[2], 10, 64)
				case "get_misses":
					result.GetMisses, _ = strconv.ParseInt(parts[2], 10, 64)
				case "uptime":
					result.Uptime, _ = strconv.ParseInt(parts[2], 10, 64)
				}
			}
		}
	}
	return result
}

// Version returns the server version.
func Version(host string, port int) VersionResult {
	conn, err := dial(host, port)
	if err != nil {
		return VersionResult{Error: fmt.Sprintf("connect failed: %v", err)}
	}
	defer conn.Close()
	out, err := sendCmd(conn, "version")
	if err != nil {
		return VersionResult{Error: fmt.Sprintf("version failed: %v", err)}
	}
	if strings.HasPrefix(out, "VERSION ") {
		return VersionResult{Version: strings.TrimPrefix(strings.TrimSpace(out), "VERSION ")}
	}
	return VersionResult{Error: "unexpected response: " + out}
}
