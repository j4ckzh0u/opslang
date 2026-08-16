package sshx_test

// End-to-end tests for controller-side file distribution over a real
// in-process SSH + SFTP server. These verify the actual transfer bytes -
// mocked transfer functions cannot catch a broken wiring.

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

	"github.com/opslang/opslang/pkg/ops-core-sdk/file"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// sftpTestServer is an SSH server that accepts password auth for "root"
// and serves an SFTP subsystem rooted at a temp directory.
type sftpTestServer struct {
	listener net.Listener
	root     string
	addr     string
	quit     chan struct{}
}

func newSFTPTtestServer(t *testing.T) *sftpTestServer {
	t.Helper()

	// Isolate TOFU state so tests never touch the user's known-hosts file.
	oldKH, hadKH := os.LookupEnv("OPSLANG_KNOWN_HOSTS")
	os.Setenv("OPSLANG_KNOWN_HOSTS", filepath.Join(t.TempDir(), "kh"))
	t.Cleanup(func() {
		if hadKH {
			os.Setenv("OPSLANG_KNOWN_HOSTS", oldKH)
		} else {
			os.Unsetenv("OPSLANG_KNOWN_HOSTS")
		}
	})

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	signer, err := ssh.ParsePrivateKey(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}))
	if err != nil {
		t.Fatalf("parse key: %v", err)
	}

	root := t.TempDir()
	cfg := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == "root" && string(pass) == "s3cret" {
				return nil, nil
			}
			return nil, fmt.Errorf("auth failed")
		},
	}
	cfg.AddHostKey(signer)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	srv := &sftpTestServer{
		listener: listener,
		root:     root,
		addr:     listener.Addr().String(),
		quit:     make(chan struct{}),
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go srv.handle(conn, cfg)
		}
	}()

	t.Cleanup(func() {
		close(srv.quit)
		listener.Close()
	})
	return srv
}

func (s *sftpTestServer) handle(netConn net.Conn, cfg *ssh.ServerConfig) {
	serverConn, chans, reqs, err := ssh.NewServerConn(netConn, cfg)
	if err != nil {
		netConn.Close()
		return
	}
	defer serverConn.Close()
	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unsupported")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			continue
		}
		go func(ch ssh.Channel, reqs <-chan *ssh.Request) {
			for req := range reqs {
				if req.Type == "subsystem" && string(req.Payload[4:]) == "sftp" {
					if req.WantReply {
						req.Reply(true, nil)
					}
					server, err := sftp.NewServer(ch, sftp.WithServerWorkingDirectory(s.root))
					if err != nil {
						ch.Close()
						return
					}
					_ = server.Serve()
					server.Close()
					return
				}
				if req.WantReply {
					req.Reply(false, nil)
				}
			}
			ch.Close()
		}(channel, requests)
	}
}

// setSSHPassword points the SDK's SSH transfer layer at the test server's
// credentials via the environment.
func setSSHPassword(t *testing.T, password string) {
	t.Helper()
	old, had := os.LookupEnv("OPSLANG_SSH_PASSWORD")
	os.Setenv("OPSLANG_SSH_PASSWORD", password)
	t.Cleanup(func() {
		if had {
			os.Setenv("OPSLANG_SSH_PASSWORD", old)
		} else {
			os.Unsetenv("OPSLANG_SSH_PASSWORD")
		}
	})
}

func TestDistributeRealSSHTransfer(t *testing.T) {
	server := newSFTPTtestServer(t)
	setSSHPassword(t, "s3cret")
	file.WireSSHTransfer()

	_, port, _ := net.SplitHostPort(server.addr)
	host, _, _ := net.SplitHostPort(server.addr)

	// Local source file with known content.
	dir := t.TempDir()
	src := filepath.Join(dir, "app.conf")
	content := []byte("listen={{port}}\n")
	if err := os.WriteFile(src, content, 0644); err != nil {
		t.Fatal(err)
	}

	// Absolute destination under the server's temp root: the SFTP server
	// serves the real filesystem, exactly like production.
	destAbs := filepath.Join(server.root, "etc", "app", "app.conf")
	result, err := file.Distribute(src, []file.DistributeTarget{
		{Host: host, Port: mustPort(t, port), User: "root", Dest: destAbs},
	}, file.DistributeOptions{
		Checksum: true,
		Mode:     "0600",
		Retries:  2,
	})
	if err != nil {
		t.Fatalf("Distribute: %v", err)
	}

	if result.Succeeded != 1 || result.Failed != 0 {
		t.Fatalf("result: %+v", result)
	}

	// Verify the file actually arrived with the right bytes, mode, and a
	// real checksum verification.
	remoteFile := destAbs
	got, err := os.ReadFile(remoteFile)
	if err != nil {
		t.Fatalf("remote file missing: %v (distribute reported success — this is exactly the lie we guard against)", err)
	}
	if string(got) != string(content) {
		t.Errorf("remote content = %q, want %q", got, content)
	}

	info, err := os.Stat(remoteFile)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("remote mode = %v, want 0600 (Mode option not applied)", info.Mode().Perm())
	}

	hr := result.Results[0]
	if hr.Status != "success" || !hr.Changed || hr.Checksum == "" {
		t.Errorf("host result: %+v", hr)
	}
}

func TestDistributeChecksumDetectsCorruption(t *testing.T) {
	// Verify the default verifier really compares remote content: point it
	// at a path whose content differs from the expected digest.
	server := newSFTPTtestServer(t)
	setSSHPassword(t, "s3cret")
	file.WireSSHTransfer()

	host, portStr, _ := net.SplitHostPort(server.addr)
	port := mustPort(t, portStr)

	// Place a decoy file with content that differs from the expected digest.
	decoy := filepath.Join(server.root, "etc", "app.conf")
	if err := os.MkdirAll(filepath.Dir(decoy), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(decoy, []byte("CORRUPTED"), 0644); err != nil {
		t.Fatal(err)
	}

	err := file.SSHVerifyChecksum(context.Background(),
		fmt.Sprintf("ssh://root@%s:%d%s", host, port, decoy),
		"0000000000000000000000000000000000000000000000000000000000000000")
	if err == nil {
		t.Fatal("checksum verification must fail when remote content differs")
	}
}

func TestDistributeAuthFailureIsReported(t *testing.T) {
	server := newSFTPTtestServer(t)
	setSSHPassword(t, "wrong-password") // never accepted by the server
	file.WireSSHTransfer()

	host, portStr, _ := net.SplitHostPort(server.addr)
	dir := t.TempDir()
	src := filepath.Join(dir, "f.txt")
	os.WriteFile(src, []byte("x"), 0644)

	result, err := file.Distribute(src, []file.DistributeTarget{
		{Host: host, Port: mustPort(t, portStr), User: "root", Dest: filepath.Join(server.root, "f.txt")},
	}, file.DistributeOptions{Retries: 1})
	if err != nil {
		t.Fatalf("Distribute: %v", err)
	}

	if result.Failed != 1 || result.Succeeded != 0 {
		t.Fatalf("auth failure must be reported per-host: %+v", result)
	}
	if result.Results[0].Error == "" {
		t.Error("failure reason missing from host result")
	}
}

func mustPort(t *testing.T, s string) int {
	t.Helper()
	var p int
	if _, err := fmt.Sscanf(s, "%d", &p); err != nil {
		t.Fatalf("bad port %q", s)
	}
	return p
}
