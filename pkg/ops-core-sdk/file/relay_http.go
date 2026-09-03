package file

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const relayFilePath = "/file"

// RelayHTTPInfo contains the short-lived credentials required by a relay client.
type RelayHTTPInfo struct {
	URL             string    `json:"url"`
	Token           string    `json:"token"`
	CertFingerprint string    `json:"cert_fingerprint"`
	SHA256          string    `json:"sha256"`
	Size            int64     `json:"size"`
	ExpiresAt       time.Time `json:"expires_at"`
}

// RelayHTTPServerOptions controls binding and lifetime of a relay endpoint.
type RelayHTTPServerOptions struct {
	ListenAddress string
	AdvertiseHost string
	TTL           time.Duration
	MaxConcurrent int
}

// RelayHTTPServer serves one immutable file through a bounded HTTPS endpoint.
type RelayHTTPServer struct {
	Info     RelayHTTPInfo
	server   *http.Server
	listener net.Listener
	done     chan error
	stopOnce sync.Once
}

// StartRelayHTTPServer starts a loopback HTTPS server for one file.
func StartRelayHTTPServer(filePath string, ttl time.Duration, maxConcurrent int) (*RelayHTTPServer, error) {
	return StartRelayHTTPServerWithOptions(filePath, RelayHTTPServerOptions{
		ListenAddress: "127.0.0.1:0",
		TTL:           ttl,
		MaxConcurrent: maxConcurrent,
	})
}

// StartRelayHTTPServerWithOptions starts a relay on a caller-selected address.
func StartRelayHTTPServerWithOptions(filePath string, opts RelayHTTPServerOptions) (*RelayHTTPServer, error) {
	if strings.TrimSpace(filePath) == "" {
		return nil, fmt.Errorf("relay file path is empty")
	}
	if opts.TTL <= 0 {
		return nil, fmt.Errorf("relay TTL must be positive")
	}
	if opts.MaxConcurrent < 1 {
		return nil, fmt.Errorf("relay max concurrent requests must be at least 1")
	}
	listenAddress := strings.TrimSpace(opts.ListenAddress)
	if listenAddress == "" {
		listenAddress = "127.0.0.1:0"
	}
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("stat relay file %s: %w", filePath, err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("relay file %s is not a regular file", filePath)
	}
	checksum, err := computeFileChecksum(filePath)
	if err != nil {
		return nil, fmt.Errorf("checksum relay file: %w", err)
	}
	token, err := randomHex(32)
	if err != nil {
		return nil, err
	}
	certificate, fingerprint, err := relayCertificate()
	if err != nil {
		return nil, err
	}
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return nil, fmt.Errorf("listen for relay HTTPS: %w", err)
	}
	listenerHost, listenerPort, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		return nil, errors.Join(fmt.Errorf("parse relay listen address: %w", err), listener.Close())
	}
	advertiseHost := strings.TrimSpace(opts.AdvertiseHost)
	if advertiseHost == "" {
		advertiseHost = listenerHost
	}
	if advertiseHost == "" || net.ParseIP(advertiseHost).IsUnspecified() {
		return nil, errors.Join(fmt.Errorf("relay advertise host is required for an unspecified listen address"), listener.Close())
	}
	expiresAt := time.Now().UTC().Add(opts.TTL)
	semaphore := make(chan struct{}, opts.MaxConcurrent)
	handler := http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != relayFilePath {
			http.NotFound(response, request)
			return
		}
		if request.Method != http.MethodGet && request.Method != http.MethodHead {
			response.Header().Set("Allow", "GET, HEAD")
			http.Error(response, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if request.Header.Get("Authorization") != "Bearer "+token {
			http.Error(response, "unauthorized", http.StatusUnauthorized)
			return
		}
		if !time.Now().Before(expiresAt) {
			http.Error(response, "relay session expired", http.StatusGone)
			return
		}
		select {
		case semaphore <- struct{}{}:
			defer func() { <-semaphore }()
		default:
			http.Error(response, "relay concurrency limit reached", http.StatusTooManyRequests)
			return
		}
		opened, err := os.Open(filePath)
		if err != nil {
			http.Error(response, "relay file unavailable", http.StatusInternalServerError)
			return
		}
		defer opened.Close()
		response.Header().Set("X-Content-SHA256", checksum)
		http.ServeContent(response, request, info.Name(), info.ModTime(), opened)
	})
	server := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	tlsListener := tls.NewListener(listener, &tls.Config{
		Certificates: []tls.Certificate{certificate},
		MinVersion:   tls.VersionTLS13,
	})
	relay := &RelayHTTPServer{
		Info: RelayHTTPInfo{
			URL:             "https://" + net.JoinHostPort(advertiseHost, listenerPort) + relayFilePath,
			Token:           token,
			CertFingerprint: fingerprint,
			SHA256:          checksum,
			Size:            info.Size(),
			ExpiresAt:       expiresAt,
		},
		server:   server,
		listener: listener,
		done:     make(chan error, 1),
	}
	go func() {
		err := server.Serve(tlsListener)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		relay.done <- err
	}()
	return relay, nil
}

// Stop closes the listener and waits for active handlers until ctx expires.
func (s *RelayHTTPServer) Stop(ctx context.Context) []string {
	if s == nil {
		return []string{"relay server is nil"}
	}
	var warnings []string
	s.stopOnce.Do(func() {
		if err := s.server.Shutdown(ctx); err != nil {
			warnings = append(warnings, "shutdown relay HTTPS server: "+err.Error())
			if closeErr := s.server.Close(); closeErr != nil {
				warnings = append(warnings, "force close relay HTTPS server: "+closeErr.Error())
			}
		}
		select {
		case err := <-s.done:
			if err != nil {
				warnings = append(warnings, "relay HTTPS server stopped with error: "+err.Error())
			}
		case <-ctx.Done():
			warnings = append(warnings, "wait for relay HTTPS server: "+ctx.Err().Error())
		}
	})
	return warnings
}

func randomHex(byteCount int) (string, error) {
	if byteCount < 1 {
		return "", fmt.Errorf("random byte count must be positive")
	}
	content := make([]byte, byteCount)
	if _, err := rand.Read(content); err != nil {
		return "", fmt.Errorf("generate relay credential: %w", err)
	}
	return hex.EncodeToString(content), nil
}

func relayCertificate() (tls.Certificate, string, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return tls.Certificate{}, "", fmt.Errorf("generate relay TLS key: %w", err)
	}
	serialLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serial, err := rand.Int(rand.Reader, serialLimit)
	if err != nil {
		return tls.Certificate{}, "", fmt.Errorf("generate relay certificate serial: %w", err)
	}
	now := time.Now().UTC()
	template := x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "opslang-relay"},
		NotBefore:    now.Add(-time.Minute),
		NotAfter:     now.Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey, privateKey)
	if err != nil {
		return tls.Certificate{}, "", fmt.Errorf("create relay certificate: %w", err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return tls.Certificate{}, "", fmt.Errorf("marshal relay TLS key: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes})
	certificate, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return tls.Certificate{}, "", fmt.Errorf("load relay TLS key pair: %w", err)
	}
	digest := sha256.Sum256(der)
	return certificate, hex.EncodeToString(digest[:]), nil
}
