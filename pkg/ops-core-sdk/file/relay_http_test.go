package file

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRelayHTTPServerAuthorizationRangeAndFingerprint(t *testing.T) {
	path := filepath.Join(t.TempDir(), "payload.bin")
	content := []byte("0123456789abcdef")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write payload: %v", err)
	}
	server, err := StartRelayHTTPServer(path, time.Minute, 2)
	if err != nil {
		t.Fatalf("start relay server: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if warnings := server.Stop(ctx); len(warnings) != 0 {
			t.Errorf("stop warnings: %v", warnings)
		}
	})

	client := relayTestClient(server.Info.CertFingerprint)
	unauthorized, err := client.Get(server.Info.URL)
	if err != nil {
		t.Fatalf("unauthorized request: %v", err)
	}
	unauthorized.Body.Close()
	if unauthorized.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d, want %d", unauthorized.StatusCode, http.StatusUnauthorized)
	}

	request, err := http.NewRequest(http.MethodGet, server.Info.URL, nil)
	if err != nil {
		t.Fatalf("create range request: %v", err)
	}
	request.Header.Set("Authorization", "Bearer "+server.Info.Token)
	request.Header.Set("Range", "bytes=4-9")
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("range request: %v", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read range response: %v", err)
	}
	if response.StatusCode != http.StatusPartialContent {
		t.Fatalf("range status = %d, want %d", response.StatusCode, http.StatusPartialContent)
	}
	if string(body) != "456789" {
		t.Fatalf("range body = %q, want %q", body, "456789")
	}
	if response.Header.Get("X-Content-SHA256") != server.Info.SHA256 {
		t.Fatalf("checksum header = %q, want %q", response.Header.Get("X-Content-SHA256"), server.Info.SHA256)
	}
	if server.Info.Size != int64(len(content)) {
		t.Fatalf("relay size = %d, want %d", server.Info.Size, len(content))
	}
}

func TestRelayHTTPServerRejectsWrongFingerprint(t *testing.T) {
	path := filepath.Join(t.TempDir(), "payload.bin")
	if err := os.WriteFile(path, []byte("payload"), 0o600); err != nil {
		t.Fatalf("write payload: %v", err)
	}
	server, err := StartRelayHTTPServer(path, time.Minute, 1)
	if err != nil {
		t.Fatalf("start relay server: %v", err)
	}
	defer stopRelayTestServer(t, server)

	request, err := http.NewRequest(http.MethodGet, server.Info.URL, nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	request.Header.Set("Authorization", "Bearer "+server.Info.Token)
	_, err = relayTestClient(strings.Repeat("0", sha256.Size*2)).Do(request)
	if err == nil || !strings.Contains(err.Error(), "fingerprint mismatch") {
		t.Fatalf("wrong fingerprint error = %v", err)
	}
}

func TestRelayHTTPServerRejectsInvalidRequests(t *testing.T) {
	path := createRelayTestFile(t)
	server, err := StartRelayHTTPServer(path, time.Minute, 1)
	if err != nil {
		t.Fatalf("start relay server: %v", err)
	}
	defer stopRelayTestServer(t, server)
	client := relayTestClient(server.Info.CertFingerprint)

	tests := []struct {
		name   string
		method string
		url    string
		token  string
		range_ string
		want   int
	}{
		{name: "wrong token", method: http.MethodGet, url: server.Info.URL, token: "wrong", want: http.StatusUnauthorized},
		{name: "unknown path", method: http.MethodGet, url: strings.TrimSuffix(server.Info.URL, relayFilePath) + "/../secret", token: server.Info.Token, want: http.StatusNotFound},
		{name: "method", method: http.MethodPost, url: server.Info.URL, token: server.Info.Token, want: http.StatusMethodNotAllowed},
		{name: "range past end", method: http.MethodGet, url: server.Info.URL, token: server.Info.Token, range_: "bytes=999-1000", want: http.StatusRequestedRangeNotSatisfiable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request, err := http.NewRequest(test.method, test.url, nil)
			if err != nil {
				t.Fatalf("create request: %v", err)
			}
			request.Header.Set("Authorization", "Bearer "+test.token)
			if test.range_ != "" {
				request.Header.Set("Range", test.range_)
			}
			response, err := client.Do(request)
			if err != nil {
				t.Fatalf("perform request: %v", err)
			}
			response.Body.Close()
			if response.StatusCode != test.want {
				t.Fatalf("status = %d, want %d", response.StatusCode, test.want)
			}
		})
	}
}

func TestRelayHTTPServerEnforcesConcurrencyLimit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "large-payload.bin")
	if err := os.WriteFile(path, make([]byte, 16<<20), 0o600); err != nil {
		t.Fatalf("write payload: %v", err)
	}
	server, err := StartRelayHTTPServer(path, time.Minute, 1)
	if err != nil {
		t.Fatalf("start relay server: %v", err)
	}
	defer stopRelayTestServer(t, server)
	client := relayTestClient(server.Info.CertFingerprint)

	first, err := http.NewRequest(http.MethodGet, server.Info.URL, nil)
	if err != nil {
		t.Fatalf("create first request: %v", err)
	}
	first.Header.Set("Authorization", "Bearer "+server.Info.Token)
	firstResponse, err := client.Do(first)
	if err != nil {
		t.Fatalf("perform first request: %v", err)
	}
	defer firstResponse.Body.Close()

	second, err := http.NewRequest(http.MethodGet, server.Info.URL, nil)
	if err != nil {
		t.Fatalf("create second request: %v", err)
	}
	second.Header.Set("Authorization", "Bearer "+server.Info.Token)
	secondResponse, err := client.Do(second)
	if err != nil {
		t.Fatalf("perform second request: %v", err)
	}
	secondResponse.Body.Close()
	if secondResponse.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("second status = %d, want %d", secondResponse.StatusCode, http.StatusTooManyRequests)
	}
}

func TestRelayHTTPServerExpires(t *testing.T) {
	path := filepath.Join(t.TempDir(), "payload.bin")
	if err := os.WriteFile(path, []byte("payload"), 0o600); err != nil {
		t.Fatalf("write payload: %v", err)
	}
	server, err := StartRelayHTTPServer(path, 20*time.Millisecond, 1)
	if err != nil {
		t.Fatalf("start relay server: %v", err)
	}
	defer stopRelayTestServer(t, server)
	time.Sleep(40 * time.Millisecond)

	request, err := http.NewRequest(http.MethodGet, server.Info.URL, nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	request.Header.Set("Authorization", "Bearer "+server.Info.Token)
	response, err := relayTestClient(server.Info.CertFingerprint).Do(request)
	if err != nil {
		t.Fatalf("expired request: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusGone {
		t.Fatalf("expired status = %d, want %d", response.StatusCode, http.StatusGone)
	}
}

func TestStartRelayHTTPServerValidatesInputs(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		ttl           time.Duration
		maxConcurrent int
	}{
		{name: "empty path", path: "", ttl: time.Minute, maxConcurrent: 1},
		{name: "missing path", path: filepath.Join(t.TempDir(), "missing"), ttl: time.Minute, maxConcurrent: 1},
		{name: "zero ttl", path: createRelayTestFile(t), ttl: 0, maxConcurrent: 1},
		{name: "zero concurrency", path: createRelayTestFile(t), ttl: time.Minute, maxConcurrent: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, err := StartRelayHTTPServer(test.path, test.ttl, test.maxConcurrent)
			if err == nil {
				stopRelayTestServer(t, server)
				t.Fatal("expected validation error")
			}
		})
	}
}

func relayTestClient(fingerprint string) *http.Client {
	transport := &http.Transport{TLSClientConfig: &tls.Config{
		MinVersion:         tls.VersionTLS13,
		InsecureSkipVerify: true, // The exact ephemeral certificate is pinned below.
		VerifyConnection: func(state tls.ConnectionState) error {
			if len(state.PeerCertificates) != 1 {
				return fmt.Errorf("unexpected relay certificate chain length %d", len(state.PeerCertificates))
			}
			digest := sha256.Sum256(state.PeerCertificates[0].Raw)
			if hex.EncodeToString(digest[:]) != fingerprint {
				return fmt.Errorf("relay certificate fingerprint mismatch")
			}
			return nil
		},
	}}
	return &http.Client{Transport: transport, Timeout: 2 * time.Second}
}

func createRelayTestFile(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "payload.bin")
	if err := os.WriteFile(path, []byte("payload"), 0o600); err != nil {
		t.Fatalf("write payload: %v", err)
	}
	return path
}

func stopRelayTestServer(t *testing.T, server *RelayHTTPServer) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if warnings := server.Stop(ctx); len(warnings) != 0 {
		t.Errorf("stop warnings: %v", warnings)
	}
}
