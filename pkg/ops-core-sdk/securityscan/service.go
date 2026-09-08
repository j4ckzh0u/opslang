package securityscan

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// ServiceConfig describes a controller-side scanner service. The service
// owns the vulnerability database; clients only submit scan metadata.
type ServiceConfig struct {
	URL             string        `json:"url"`
	Token           string        `json:"token"`
	DatabaseName    string        `json:"database_name,omitempty"`
	DatabaseVersion string        `json:"database_version,omitempty"`
	DatabaseSHA256  string        `json:"database_sha256,omitempty"`
	CA              []byte        `json:"ca,omitempty"`
	Timeout         time.Duration `json:"timeout,omitempty"`
}

func ValidateServiceConfig(config *ServiceConfig) error {
	if config == nil {
		return fmt.Errorf("scanner service configuration is nil")
	}
	parsed, err := url.Parse(strings.TrimSpace(config.URL))
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return fmt.Errorf("scanner service URL must be an HTTPS URL")
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("scanner service URL must not contain credentials, query, or fragment")
	}
	if strings.TrimSpace(config.Token) == "" {
		return fmt.Errorf("scanner service token is required")
	}
	if config.Timeout < 0 {
		return fmt.Errorf("scanner service timeout must not be negative")
	}
	if config.DatabaseSHA256 != "" {
		digest, decodeErr := hex.DecodeString(config.DatabaseSHA256)
		if decodeErr != nil || len(digest) != sha256.Size {
			return fmt.Errorf("scanner service database sha256 must be a 64-character hexadecimal digest")
		}
	}
	return nil
}

func (config *ServiceConfig) Clone() (*ServiceConfig, error) {
	if err := ValidateServiceConfig(config); err != nil {
		return nil, err
	}
	clone := *config
	clone.CA = append([]byte(nil), config.CA...)
	return &clone, nil
}
