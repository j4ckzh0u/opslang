// Package zookeeper provides ZooKeeper client operations.
// Supports get, set, delete, create, list, and exists operations on ZooKeeper nodes.
package zookeeper

import (
	"fmt"
	"time"

	"github.com/go-zookeeper/zk"
)

// ZookeeperResult represents the result of ZooKeeper operations.
type ZookeeperResult struct {
	Success  bool              `json:"success"`
	Path     string            `json:"path,omitempty"`
	Value    string            `json:"value,omitempty"`
	Children []string          `json:"children,omitempty"`
	Exists   bool              `json:"exists,omitempty"`
	Changed  bool              `json:"changed,omitempty"`
	Error    string            `json:"error,omitempty"`
	Duration int64             `json:"duration_ms"`
}

// Connect establishes a connection to ZooKeeper.
func Connect(servers []string, timeout time.Duration) (*zk.Conn, error) {
	if len(servers) == 0 {
		servers = []string{"localhost:2181"}
	}
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	conn, _, err := zk.Connect(servers, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return conn, nil
}

// Get retrieves the value of a ZooKeeper node.
func Get(path string, servers []string) ZookeeperResult {
	start := time.Now()

	conn, err := Connect(servers, 5*time.Second)
	if err != nil {
		return ZookeeperResult{
			Success:  false,
			Error:    err.Error(),
			Duration: time.Since(start).Milliseconds(),
		}
	}
	defer conn.Close()

	data, _, err := conn.Get(path)
	if err != nil {
		return ZookeeperResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to get node: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return ZookeeperResult{
		Success:  true,
		Path:     path,
		Value:    string(data),
		Duration: time.Since(start).Milliseconds(),
	}
}

// Set sets the value of a ZooKeeper node.
func Set(path, value string, servers []string) ZookeeperResult {
	start := time.Now()

	conn, err := Connect(servers, 5*time.Second)
	if err != nil {
		return ZookeeperResult{
			Success:  false,
			Error:    err.Error(),
			Duration: time.Since(start).Milliseconds(),
		}
	}
	defer conn.Close()

	_, err = conn.Set(path, []byte(value), -1)
	if err != nil {
		return ZookeeperResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to set node: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return ZookeeperResult{
		Success:  true,
		Path:     path,
		Value:    value,
		Changed:  true,
		Duration: time.Since(start).Milliseconds(),
	}
}

// Create creates a new ZooKeeper node.
func Create(path, value string, ephemeral bool, servers []string) ZookeeperResult {
	start := time.Now()

	conn, err := Connect(servers, 5*time.Second)
	if err != nil {
		return ZookeeperResult{
			Success:  false,
			Error:    err.Error(),
			Duration: time.Since(start).Milliseconds(),
		}
	}
	defer conn.Close()

	flags := int32(0)
	if ephemeral {
		flags = zk.FlagEphemeral
	}

	createdPath, err := conn.Create(path, []byte(value), flags, zk.WorldACL(zk.PermAll))
	if err != nil {
		return ZookeeperResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to create node: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return ZookeeperResult{
		Success:  true,
		Path:     createdPath,
		Value:    value,
		Changed:  true,
		Duration: time.Since(start).Milliseconds(),
	}
}

// Delete removes a ZooKeeper node.
func Delete(path string, servers []string) ZookeeperResult {
	start := time.Now()

	conn, err := Connect(servers, 5*time.Second)
	if err != nil {
		return ZookeeperResult{
			Success:  false,
			Error:    err.Error(),
			Duration: time.Since(start).Milliseconds(),
		}
	}
	defer conn.Close()

	err = conn.Delete(path, -1)
	if err != nil {
		return ZookeeperResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to delete node: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return ZookeeperResult{
		Success:  true,
		Path:     path,
		Changed:  true,
		Duration: time.Since(start).Milliseconds(),
	}
}

// List retrieves the children of a ZooKeeper node.
func List(path string, servers []string) ZookeeperResult {
	start := time.Now()

	conn, err := Connect(servers, 5*time.Second)
	if err != nil {
		return ZookeeperResult{
			Success:  false,
			Error:    err.Error(),
			Duration: time.Since(start).Milliseconds(),
		}
	}
	defer conn.Close()

	children, _, err := conn.Children(path)
	if err != nil {
		return ZookeeperResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to list children: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return ZookeeperResult{
		Success:  true,
		Path:     path,
		Children: children,
		Duration: time.Since(start).Milliseconds(),
	}
}

// Exists checks if a ZooKeeper node exists.
func Exists(path string, servers []string) ZookeeperResult {
	start := time.Now()

	conn, err := Connect(servers, 5*time.Second)
	if err != nil {
		return ZookeeperResult{
			Success:  false,
			Error:    err.Error(),
			Duration: time.Since(start).Milliseconds(),
		}
	}
	defer conn.Close()

	exists, _, err := conn.Exists(path)
	if err != nil {
		return ZookeeperResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to check existence: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return ZookeeperResult{
		Success:  true,
		Path:     path,
		Exists:   exists,
		Duration: time.Since(start).Milliseconds(),
	}
}
