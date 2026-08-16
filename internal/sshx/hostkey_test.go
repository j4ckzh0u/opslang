package sshx

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// TestTOFUKeyChangeRejected: a host whose key differs from the recorded one
// must be refused. This is the MITM guard that InsecureIgnoreHostKey used
// to disable.
func TestTOFUKeyChangeRejected(t *testing.T) {
	server, err := newMockSSHServer("testpass")
	if err != nil {
		t.Fatal(err)
	}
	defer server.listener.Close()

	khFile := filepath.Join(t.TempDir(), "known_hosts")

	// Record an unrelated key for the server's address.
	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	otherSigner, err := ssh.ParsePrivateKey(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(otherKey),
	}))
	if err != nil {
		t.Fatal(err)
	}
	_, port, _ := net.SplitHostPort(server.Addr())
	hostport := net.JoinHostPort("127.0.0.1", port)
	line := knownhosts.Line([]string{knownhosts.Normalize(hostport)}, otherSigner.PublicKey())
	if err := os.WriteFile(khFile, []byte(line+"\n"), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{
		Host:           "127.0.0.1",
		Port:           22,
		User:           "root",
		Password:       "testpass",
		Timeout:        5 * time.Second,
		Retries:        0,
		KnownHostsFile: khFile,
	}
	if p, err := fmt.Sscanf(port, "%d", &cfg.Port); err != nil || p != 1 {
		t.Fatalf("bad port %q", port)
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	err = client.Connect(context.Background())
	if err == nil {
		client.Close()
		t.Fatal("connection with a CHANGED host key must be rejected (possible MITM)")
	}
}

// TestTOFUIgnoredWhenInsecure: the explicit insecure flag still disables
// verification for lab use.
func TestTOFUIgnoredWhenInsecure(t *testing.T) {
	server, err := newMockSSHServer("testpass")
	if err != nil {
		t.Fatal(err)
	}
	defer server.listener.Close()

	_, port, _ := net.SplitHostPort(server.Addr())

	cfg := &Config{
		Host:                      "127.0.0.1",
		User:                      "root",
		Password:                  "testpass",
		Timeout:                   5 * time.Second,
		Retries:                   0,
		InsecureSkipHostKeyVerify: true,
		KnownHostsFile:            filepath.Join(t.TempDir(), "empty_known_hosts"),
	}
	if _, err := fmt.Sscanf(port, "%d", &cfg.Port); err != nil {
		t.Fatal(err)
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Connect(context.Background()); err != nil {
		t.Fatalf("connect with insecure flag should succeed: %v", err)
	}
	client.Close()
}
