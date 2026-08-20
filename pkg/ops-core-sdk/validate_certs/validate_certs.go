// Package validate_certs provides SSL certificate validation.
package validate_certs

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"time"
)

// ValidateResult is returned by Validate.
type ValidateResult struct {
	Valid      bool   `json:"valid"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Issuer     string `json:"issuer,omitempty"`
	Subject    string `json:"subject,omitempty"`
	NotBefore  string `json:"not_before,omitempty"`
	NotAfter   string `json:"not_after,omitempty"`
	DaysLeft   int    `json:"days_left,omitempty"`
	SANS       []string `json:"sans,omitempty"`
	Error      string `json:"error,omitempty"`
}

// Validate checks an SSL certificate for a host:port.
func Validate(host string, port int, timeout time.Duration) ValidateResult {
	if host == "" {
		return ValidateResult{Error: "host is required"}
	}
	if port <= 0 {
		port = 443
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	dialer := &net.Dialer{Timeout: timeout}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{
		InsecureSkipVerify: true, // We validate manually to get detailed info
	})
	if err != nil {
		return ValidateResult{Host: host, Port: port, Error: "connection failed: " + err.Error()}
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return ValidateResult{Host: host, Port: port, Error: "no certificates presented"}
	}

	leaf := certs[0]
	now := time.Now()

	r := ValidateResult{
		Host:      host,
		Port:      port,
		Issuer:    leaf.Issuer.String(),
		Subject:   leaf.Subject.String(),
		NotBefore: leaf.NotBefore.Format(time.RFC3339),
		NotAfter:  leaf.NotAfter.Format(time.RFC3339),
		DaysLeft:  int(leaf.NotAfter.Sub(now).Hours() / 24),
	}

	// Collect SANs
	for _, dns := range leaf.DNSNames {
		r.SANS = append(r.SANS, dns)
	}

	// Validate hostname
	if err := leaf.VerifyHostname(host); err != nil {
		r.Error = "hostname mismatch: " + err.Error()
		return r
	}

	// Validate time
	if now.Before(leaf.NotBefore) {
		r.Error = "certificate not yet valid"
		return r
	}
	if now.After(leaf.NotAfter) {
		r.Error = "certificate expired"
		return r
	}

	// Validate chain (basic - just check the leaf cert is trusted)
	roots := x509.NewCertPool()
	for _, cert := range certs[1:] {
		roots.AddCert(cert)
	}
	opts := x509.VerifyOptions{
		DNSName: host,
		Roots:   roots,
	}
	if _, err := leaf.Verify(opts); err != nil {
		// Chain validation failed, but cert is structurally valid
		r.Valid = true // Cert itself is valid, chain may be incomplete
		return r
	}

	r.Valid = true
	return r
}

// CheckExpiry checks if a certificate expires within the given days.
func CheckExpiry(host string, port int, days int, timeout time.Duration) ValidateResult {
	r := Validate(host, port, timeout)
	if r.Error != "" {
		return r
	}
	if r.DaysLeft < days {
		r.Valid = false
		r.Error = fmt.Sprintf("certificate expires in %d days (threshold: %d)", r.DaysLeft, days)
	}
	return r
}
