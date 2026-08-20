package certbot

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// CertInfo represents certificate information
type CertInfo struct {
	Domain     string   `json:"domain"`
	CertPath   string   `json:"cert_path"`
	KeyPath    string   `json:"key_path"`
	ExpiryDate string   `json:"expiry_date"`
	Status     string   `json:"status"` // valid/expiring/expired
	Serial     string   `json:"serial"`
	Issuer     string   `json:"issuer"`
}

// CertbotResult represents certbot operation result
type CertbotResult struct {
	Changed    bool   `json:"changed"`
	Domain     string `json:"domain,omitempty"`
	CertPath   string `json:"cert_path,omitempty"`
	KeyPath    string `json:"key_path,omitempty"`
	Message    string `json:"message,omitempty"`
	DurationMs int64  `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
}

// Certificates lists all managed certificates
func Certificates() ([]CertInfo, error) {
	start := time.Now()
	cmd := exec.Command("certbot", "certificates", "--non-interactive")
	output, err := cmd.CombinedOutput()
	_ = time.Since(start).Milliseconds()

	if err != nil {
		return nil, fmt.Errorf("certbot certificates failed: %v, output: %s", err, string(output))
	}

	// Parse output (simplified)
	var certs []CertInfo
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Certificate Name:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				certs = append(certs, CertInfo{
					Domain: strings.TrimSpace(parts[1]),
					Status: "valid",
				})
			}
		}
	}

	return certs, nil
}

// Obtain obtains a new certificate
func Obtain(domains []string, email, webroot string, standalone bool) (CertbotResult, error) {
	start := time.Now()

	if len(domains) == 0 {
		return CertbotResult{DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("at least one domain required")
	}

	args := []string{"certonly", "--non-interactive", "--agree-tos"}

	if email != "" {
		args = append(args, "--email", email)
	} else {
		args = append(args, "--register-unsafely-without-email")
	}

	if standalone {
		args = append(args, "--standalone")
	} else if webroot != "" {
		args = append(args, "--webroot", "-w", webroot)
	} else {
		return CertbotResult{DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("either standalone or webroot required")
	}

	for _, domain := range domains {
		args = append(args, "-d", domain)
	}

	cmd := exec.Command("certbot", args...)
	output, err := cmd.CombinedOutput()
	duration := time.Since(start).Milliseconds()

	result := CertbotResult{
		Changed:    true,
		Domain:     strings.Join(domains, ","),
		DurationMs: duration,
	}

	if err != nil {
		result.Error = fmt.Sprintf("certbot obtain failed: %v, output: %s", err, string(output))
		return result, fmt.Errorf("%s", result.Error)
	}

	result.Message = "Certificate obtained successfully"
	return result, nil
}

// Renew renews certificates
func Renew(force bool) (CertbotResult, error) {
	start := time.Now()

	args := []string{"renew", "--non-interactive"}
	if force {
		args = append(args, "--force-renewal")
	}

	cmd := exec.Command("certbot", args...)
	output, err := cmd.CombinedOutput()
	duration := time.Since(start).Milliseconds()

	result := CertbotResult{
		Changed:    true,
		DurationMs: duration,
	}

	if err != nil {
		result.Error = fmt.Sprintf("certbot renew failed: %v, output: %s", err, string(output))
		return result, fmt.Errorf("%s", result.Error)
	}

	result.Message = "Certificates renewed successfully"
	return result, nil
}

// Delete deletes a certificate
func Delete(domain string) (CertbotResult, error) {
	start := time.Now()

	if domain == "" {
		return CertbotResult{DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("domain required")
	}

	cmd := exec.Command("certbot", "delete", "--non-interactive", "--cert-name", domain)
	output, err := cmd.CombinedOutput()
	duration := time.Since(start).Milliseconds()

	result := CertbotResult{
		Changed:    true,
		Domain:     domain,
		DurationMs: duration,
	}

	if err != nil {
		result.Error = fmt.Sprintf("certbot delete failed: %v, output: %s", err, string(output))
		return result, fmt.Errorf("%s", result.Error)
	}

	result.Message = "Certificate deleted successfully"
	return result, nil
}
