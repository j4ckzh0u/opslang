// Package vault provides HashiCorp Vault client operations.
// Supports reading, writing, listing, and deleting secrets from Vault KV v2.
package vault

import (
	"fmt"
	"time"

	vault "github.com/hashicorp/vault/api"
)

// VaultResult represents the result of Vault operations.
type VaultResult struct {
	Success  bool                   `json:"success"`
	Path     string                 `json:"path,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
	Keys     []string               `json:"keys,omitempty"`
	Changed  bool                   `json:"changed,omitempty"`
	Error    string                 `json:"error,omitempty"`
	Duration int64                  `json:"duration_ms"`
}

// Config represents Vault connection configuration.
type Config struct {
	Address string
	Token   string
	Timeout time.Duration
}

// Connect establishes a connection to Vault.
func Connect(cfg Config) (*vault.Client, error) {
	if cfg.Address == "" {
		cfg.Address = "http://127.0.0.1:8200"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}

	config := vault.DefaultConfig()
	config.Address = cfg.Address
	config.Timeout = cfg.Timeout

	client, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	if cfg.Token != "" {
		client.SetToken(cfg.Token)
	}

	return client, nil
}

// Read retrieves a secret from Vault KV v2.
func Read(path, token, address string) VaultResult {
	start := time.Now()

	client, err := Connect(Config{
		Address: address,
		Token:   token,
	})
	if err != nil {
		return VaultResult{
			Success:  false,
			Error:    err.Error(),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	secret, err := client.Logical().Read(path)
	if err != nil {
		return VaultResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to read secret: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	if secret == nil || secret.Data == nil {
		return VaultResult{
			Success:  false,
			Error:    "secret not found",
			Duration: time.Since(start).Milliseconds(),
		}
	}

	// KV v2 returns data in a nested structure
	data := secret.Data
	if v2Data, ok := secret.Data["data"].(map[string]interface{}); ok {
		data = v2Data
	}

	return VaultResult{
		Success:  true,
		Path:     path,
		Data:     data,
		Duration: time.Since(start).Milliseconds(),
	}
}

// Write writes a secret to Vault KV v2.
func Write(path, token, address string, data map[string]interface{}) VaultResult {
	start := time.Now()

	client, err := Connect(Config{
		Address: address,
		Token:   token,
	})
	if err != nil {
		return VaultResult{
			Success:  false,
			Error:    err.Error(),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	// KV v2 wraps data in a "data" field
	payload := map[string]interface{}{
		"data": data,
	}

	_, err = client.Logical().Write(path, payload)
	if err != nil {
		return VaultResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to write secret: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return VaultResult{
		Success:  true,
		Path:     path,
		Changed:  true,
		Duration: time.Since(start).Milliseconds(),
	}
}

// Delete deletes a secret from Vault KV v2.
func Delete(path, token, address string) VaultResult {
	start := time.Now()

	client, err := Connect(Config{
		Address: address,
		Token:   token,
	})
	if err != nil {
		return VaultResult{
			Success:  false,
			Error:    err.Error(),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	_, err = client.Logical().Delete(path)
	if err != nil {
		return VaultResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to delete secret: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return VaultResult{
		Success:  true,
		Path:     path,
		Changed:  true,
		Duration: time.Since(start).Milliseconds(),
	}
}

// List lists secrets at a path in Vault KV v2.
func List(path, token, address string) VaultResult {
	start := time.Now()

	client, err := Connect(Config{
		Address: address,
		Token:   token,
	})
	if err != nil {
		return VaultResult{
			Success:  false,
			Error:    err.Error(),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	secret, err := client.Logical().List(path)
	if err != nil {
		return VaultResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to list secrets: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	if secret == nil || secret.Data == nil {
		return VaultResult{
			Success:  true,
			Path:     path,
			Keys:     []string{},
			Duration: time.Since(start).Milliseconds(),
		}
	}

	var keys []string
	if k, ok := secret.Data["keys"].([]interface{}); ok {
		for _, key := range k {
			if s, ok := key.(string); ok {
				keys = append(keys, s)
			}
		}
	}

	return VaultResult{
		Success:  true,
		Path:     path,
		Keys:     keys,
		Duration: time.Since(start).Milliseconds(),
	}
}
