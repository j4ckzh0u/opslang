// Package redis provides Redis key-value store management.
// Uses exec.Command to invoke redis-cli binary.
package redis

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// InfoResult is returned by Info.
type InfoResult struct {
	Version    string `json:"version"`
	Uptime     int64  `json:"uptime_seconds"`
	Connected  int    `json:"connected_clients"`
	UsedMemory int64  `json:"used_memory"`
	Keys       int64  `json:"total_keys"`
	Raw        string `json:"raw,omitempty"`
}

// GetResult is returned by Get.
type GetResult struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Found bool   `json:"found"`
}

// SetResult is returned by Set.
type SetResult struct {
	Key      string `json:"key"`
	Success  bool   `json:"success"`
	Duration int64  `json:"duration_ms"`
}

// DelResult is returned by Del.
type DelResult struct {
	Keys    []string `json:"keys"`
	Deleted int      `json:"deleted"`
	Success bool     `json:"success"`
}

// ListResult is returned by Keys.
type ListResult struct {
	Pattern string   `json:"pattern"`
	Keys    []string `json:"keys"`
	Count   int      `json:"count"`
}

// PingResult is returned by Ping.
type PingResult struct {
	Up       bool   `json:"up"`
	Response string `json:"response"`
}

// FlushResult is returned by FlushDB/FlushAll.
type FlushResult struct {
	Success  bool   `json:"success"`
	Duration int64  `json:"duration_ms"`
}

type connOpts struct {
	host string
	port int
	auth string
}

func defaultOpts() connOpts {
	return connOpts{host: "127.0.0.1", port: 6379}
}

func redisCmd(opts connOpts, args ...string) *exec.Cmd {
	base := []string{}
	if opts.host != "" && opts.host != "127.0.0.1" {
		base = append(base, "-h", opts.host)
	}
	if opts.port > 0 && opts.port != 6379 {
		base = append(base, "-p", strconv.Itoa(opts.port))
	}
	if opts.auth != "" {
		base = append(base, "-a", opts.auth)
	}
	base = append(base, args...)
	return exec.Command("redis-cli", base...)
}

func parseOpts(host string, port int, auth string) connOpts {
	o := defaultOpts()
	if host != "" {
		o.host = host
	}
	if port > 0 {
		o.port = port
	}
	if auth != "" {
		o.auth = auth
	}
	return o
}

// Ping checks if Redis is responding.
func Ping(host string, port int, auth string) (PingResult, error) {
	o := parseOpts(host, port, auth)
	out, err := redisCmd(o, "PING").Output()
	if err != nil {
		return PingResult{Up: false}, nil
	}
	resp := strings.TrimSpace(string(out))
	return PingResult{Up: resp == "PONG", Response: resp}, nil
}

// Get retrieves a value by key.
func Get(key, host string, port int, auth string) (GetResult, error) {
	if key == "" {
		return GetResult{}, fmt.Errorf("key is required")
	}
	o := parseOpts(host, port, auth)
	out, err := redisCmd(o, "GET", key).Output()
	if err != nil {
		return GetResult{Key: key}, err
	}
	val := strings.TrimSpace(string(out))
	if val == "" {
		return GetResult{Key: key, Found: false}, nil
	}
	return GetResult{Key: key, Value: val, Found: true}, nil
}

// Set sets a key-value pair with optional expiry.
func Set(key, value, host string, port int, auth string, expirySec int) (SetResult, error) {
	start := time.Now()
	if key == "" {
		return SetResult{}, fmt.Errorf("key is required")
	}
	o := parseOpts(host, port, auth)
	args := []string{"SET", key, value}
	if expirySec > 0 {
		args = append(args, "EX", strconv.Itoa(expirySec))
	}
	out, err := redisCmd(o, args...).Output()
	if err != nil {
		return SetResult{Key: key, Duration: time.Since(start).Milliseconds()}, err
	}
	return SetResult{Key: key, Success: strings.TrimSpace(string(out)) == "OK", Duration: time.Since(start).Milliseconds()}, nil
}

// Del deletes one or more keys.
func Del(keys []string, host string, port int, auth string) (DelResult, error) {
	if len(keys) == 0 {
		return DelResult{}, fmt.Errorf("at least one key is required")
	}
	o := parseOpts(host, port, auth)
	args := append([]string{"DEL"}, keys...)
	out, err := redisCmd(o, args...).Output()
	if err != nil {
		return DelResult{Keys: keys}, err
	}
	deleted, _ := strconv.Atoi(strings.TrimSpace(string(out)))
	return DelResult{Keys: keys, Deleted: deleted, Success: true}, nil
}

// Keys lists keys matching a pattern.
func Keys(pattern, host string, port int, auth string) (ListResult, error) {
	if pattern == "" {
		pattern = "*"
	}
	o := parseOpts(host, port, auth)
	out, err := redisCmd(o, "KEYS", pattern).Output()
	if err != nil {
		return ListResult{Pattern: pattern}, err
	}
	var keys []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			keys = append(keys, line)
		}
	}
	return ListResult{Pattern: pattern, Keys: keys, Count: len(keys)}, nil
}

// Info returns Redis server information.
func Info(host string, port int, auth string) (InfoResult, error) {
	o := parseOpts(host, port, auth)
	out, err := redisCmd(o, "INFO").Output()
	if err != nil {
		return InfoResult{}, fmt.Errorf("failed to get redis info: %w", err)
	}
	raw := string(out)
	info := InfoResult{Raw: raw}

	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "redis_version:") {
			info.Version = strings.TrimPrefix(line, "redis_version:")
		} else if strings.HasPrefix(line, "uptime_in_seconds:") {
			info.Uptime, _ = strconv.ParseInt(strings.TrimPrefix(line, "uptime_in_seconds:"), 10, 64)
		} else if strings.HasPrefix(line, "connected_clients:") {
			info.Connected, _ = strconv.Atoi(strings.TrimPrefix(line, "connected_clients:"))
		} else if strings.HasPrefix(line, "used_memory:") {
			info.UsedMemory, _ = strconv.ParseInt(strings.TrimPrefix(line, "used_memory:"), 10, 64)
		} else if strings.HasPrefix(line, "db0:keys=") {
			// Parse "db0:keys=123,expires=..."
			parts := strings.SplitN(strings.TrimPrefix(line, "db0:keys="), ",", 2)
			if len(parts) > 0 {
				info.Keys, _ = strconv.ParseInt(parts[0], 10, 64)
			}
		}
	}
	return info, nil
}

// FlushDB flushes the current database.
func FlushDB(host string, port int, auth string) (FlushResult, error) {
	start := time.Now()
	o := parseOpts(host, port, auth)
	_, err := redisCmd(o, "FLUSHDB").Output()
	if err != nil {
		return FlushResult{Duration: time.Since(start).Milliseconds()}, err
	}
	return FlushResult{Success: true, Duration: time.Since(start).Milliseconds()}, nil
}
