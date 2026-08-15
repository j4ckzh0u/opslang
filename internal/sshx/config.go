package sshx

import (
	"fmt"
	"time"
)

// Config holds SSH connection configuration.
type Config struct {
	// Host is the target hostname or IP address.
	Host string
	// Port is the SSH port (default: 22).
	Port int
	// User is the username for authentication (default: root).
	User string
	// Password for password-based authentication.
	Password string
	// KeyFile path to private key for key-based authentication.
	KeyFile string
	// Timeout for connection and operations (default: 30s).
	Timeout time.Duration
	// Retries is the number of connection retry attempts (default: 3).
	Retries int
	// MaxConnections limits concurrent connections in pool (default: 5).
	MaxConnections int
}

// Validate checks the configuration and returns an error if invalid.
func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.Port < 0 || c.Port > 65535 {
		return fmt.Errorf("port must be between 0 and 65535")
	}
	if c.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}
	if c.Retries < 0 {
		return fmt.Errorf("retries cannot be negative")
	}
	if c.MaxConnections < 0 {
		return fmt.Errorf("max connections cannot be negative")
	}
	if c.Password == "" && c.KeyFile == "" {
		return fmt.Errorf("either password or key file must be provided")
	}
	return nil
}

// SetDefaults applies default values to zero-valued fields.
func (c *Config) SetDefaults() {
	if c.Port == 0 {
		c.Port = 22
	}
	if c.User == "" {
		c.User = "root"
	}
	if c.Timeout == 0 {
		c.Timeout = 30 * time.Second
	}
	if c.Retries == 0 {
		c.Retries = 3
	}
	if c.MaxConnections == 0 {
		c.MaxConnections = 5
	}
}
