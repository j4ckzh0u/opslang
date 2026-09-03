package sshx_test

// End-to-end tests for controller-side file distribution over a real
// in-process SSH + SFTP server. These verify the actual transfer bytes -
// mocked transfer functions cannot catch a broken wiring.

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/file"
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
	exec     func(string) (string, string, uint32)
}

func writeResumeMetadata(t *testing.T, path string, size, confirmed int64, checksum string) {
	t.Helper()
	content, err := json.Marshal(map[string]interface{}{
		"version":        1,
		"session_id":     "0123456789abcdef0123456789abcdef",
		"source_size":    size,
		"source_sha256":  checksum,
		"confirmed_size": confirmed,
		"updated_at":     time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("marshal resume metadata: %v", err)
	}
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatalf("write resume metadata: %v", err)
	}
}

func checksumHex(content []byte) string {
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:])
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
				if req.Type == "exec" && s.exec != nil {
					var payload struct{ Command string }
					if err := ssh.Unmarshal(req.Payload, &payload); err != nil {
						if req.WantReply {
							req.Reply(false, nil)
						}
						ch.Close()
						return
					}
					if req.WantReply {
						req.Reply(true, nil)
					}
					stdout, stderr, status := s.exec(payload.Command)
					_, _ = ch.Write([]byte(stdout))
					_, _ = ch.Stderr().Write([]byte(stderr))
					_, _ = ch.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{status}))
					ch.Close()
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

func TestDistributeResumesAndSkipsRealSSHTransfer(t *testing.T) {
	server := newSFTPTtestServer(t)
	setSSHPassword(t, "s3cret")
	file.WireSSHTransfer()
	host, portText, _ := net.SplitHostPort(server.addr)

	content := bytes.Repeat([]byte("resume-upload-"), 90000)
	source := filepath.Join(t.TempDir(), "archive.bin")
	if err := os.WriteFile(source, content, 0600); err != nil {
		t.Fatalf("write source: %v", err)
	}
	destination := filepath.Join(server.root, "data", "archive.bin")
	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		t.Fatalf("create destination directory: %v", err)
	}
	confirmed := int64(1024 * 1024)
	if err := os.WriteFile(destination+".opslang.part", content[:confirmed], 0600); err != nil {
		t.Fatalf("write remote partial file: %v", err)
	}
	writeResumeMetadata(t, destination+".opslang.part.json", int64(len(content)), confirmed, checksumHex(content))

	result, err := file.Distribute(source, []file.DistributeTarget{{
		Host: host,
		Port: mustPort(t, portText),
		User: "root",
		Dest: destination,
	}}, file.DistributeOptions{Resume: true, Retries: 1})
	if err != nil {
		t.Fatalf("Distribute resumable: %v", err)
	}
	hostResult := result.Results[0]
	if hostResult.Status != "success" || hostResult.ResumedBytes != confirmed || hostResult.TransferredBytes != int64(len(content))-confirmed {
		t.Fatalf("resumable host result = %+v", hostResult)
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("read committed destination: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Fatal("committed destination content differs from source")
	}
	if _, err := os.Stat(destination + ".opslang.part.json"); !os.IsNotExist(err) {
		t.Fatalf("partial metadata remains after commit: %v", err)
	}

	skipped, err := file.Distribute(source, []file.DistributeTarget{{
		Host: host,
		Port: mustPort(t, portText),
		User: "root",
		Dest: destination,
	}}, file.DistributeOptions{Resume: true, Retries: 1})
	if err != nil {
		t.Fatalf("Distribute skipped: %v", err)
	}
	if skipped.Skipped != 1 || skipped.Results[0].Changed || skipped.Results[0].TransferredBytes != 0 {
		t.Fatalf("skipped result = %+v", skipped)
	}
}

func TestCollectResumesRealSSHTransfer(t *testing.T) {
	server := newSFTPTtestServer(t)
	setSSHPassword(t, "s3cret")
	file.WireSSHTransfer()
	host, portText, _ := net.SplitHostPort(server.addr)

	content := bytes.Repeat([]byte("resume-download-"), 80000)
	remoteSource := filepath.Join(server.root, "logs", "service.log")
	if err := os.MkdirAll(filepath.Dir(remoteSource), 0755); err != nil {
		t.Fatalf("create remote source directory: %v", err)
	}
	if err := os.WriteFile(remoteSource, content, 0600); err != nil {
		t.Fatalf("write remote source: %v", err)
	}
	destDir := t.TempDir()
	localDestination := filepath.Join(destDir, host, filepath.Base(remoteSource))
	if err := os.MkdirAll(filepath.Dir(localDestination), 0755); err != nil {
		t.Fatalf("create local destination directory: %v", err)
	}
	confirmed := int64(1024 * 1024)
	if err := os.WriteFile(localDestination+".opslang.part", content[:confirmed], 0600); err != nil {
		t.Fatalf("write local partial file: %v", err)
	}
	writeResumeMetadata(t, localDestination+".opslang.part.json", int64(len(content)), confirmed, checksumHex(content))

	result, err := file.Collect(remoteSource, []file.CollectTarget{{
		Host: host,
		Port: mustPort(t, portText),
		User: "root",
	}}, file.CollectOptions{DestDir: destDir, Resume: true, Retries: 1})
	if err != nil {
		t.Fatalf("Collect resumable: %v", err)
	}
	hostResult := result.Results[0]
	if hostResult.Status != "success" || hostResult.ResumedBytes != confirmed || hostResult.TransferredBytes != int64(len(content))-confirmed {
		t.Fatalf("resumable collect result = %+v", hostResult)
	}
	got, err := os.ReadFile(localDestination)
	if err != nil {
		t.Fatalf("read collected destination: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Fatal("collected destination content differs from source")
	}
}

func TestDistributeRelayRealSSHAndHTTPS(t *testing.T) {
	tests := []struct {
		name          string
		serveFailures int
		wantWarnings  int
		wantDirect    bool
	}{
		{name: "relay success"},
		{name: "candidate switch", serveFailures: 1, wantWarnings: 1},
		{name: "direct fallback", serveFailures: 10, wantWarnings: 3, wantDirect: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := newSFTPTtestServer(t)
			setSSHPassword(t, "s3cret")
			file.WireSSHTransfer()
			installRelayExecHandler(t, server, test.serveFailures)
			host, portText, _ := net.SplitHostPort(server.addr)
			content := bytes.Repeat([]byte("relay-e2e-"), 1000)
			source := filepath.Join(t.TempDir(), "payload.bin")
			if err := os.WriteFile(source, content, 0600); err != nil {
				t.Fatalf("write relay source: %v", err)
			}
			targets := make([]file.DistributeTarget, 3)
			for index := range targets {
				targets[index] = file.DistributeTarget{Host: host, Port: mustPort(t, portText), User: "root", Dest: filepath.Join(server.root, fmt.Sprintf("target-%d.bin", index))}
			}
			result, err := file.Distribute(source, targets, file.DistributeOptions{Relay: true, RelayThreshold: 2, Parallel: 3, Retries: 1, Timeout: 30 * time.Second})
			if err != nil {
				t.Fatalf("relay distribute: %v", err)
			}
			if result.Succeeded+result.Skipped != 3 || len(result.Results) != 3 {
				t.Fatalf("relay result: %+v", result)
			}
			for index, target := range targets {
				got, err := os.ReadFile(target.Dest)
				if err != nil || !bytes.Equal(got, content) {
					t.Fatalf("target %d content mismatch: %v", index, err)
				}
				if len(result.Results[index].Warnings) != test.wantWarnings {
					t.Fatalf("target %d warnings = %v", index, result.Results[index].Warnings)
				}
				if test.wantDirect && result.Results[index].TransferSource != "direct_sftp" {
					t.Fatalf("target %d source = %q", index, result.Results[index].TransferSource)
				}
			}
		})
	}
}

func installRelayExecHandler(t *testing.T, server *sftpTestServer, serveFailures int) {
	t.Helper()
	var mu sync.Mutex
	var sessions []*file.RelayHTTPServer
	server.exec = func(command string) (string, string, uint32) {
		arguments := relayCommandArguments(command)
		if len(arguments) < 3 {
			return "", "invalid relay command", 3
		}
		switch arguments[2] {
		case "serve":
			mu.Lock()
			if serveFailures > 0 {
				serveFailures--
				mu.Unlock()
				return "", "injected relay serve failure", 3
			}
			mu.Unlock()
			relay, err := file.StartRelayHTTPServerWithOptions(relayFlag(arguments, "--file"), file.RelayHTTPServerOptions{ListenAddress: "127.0.0.1:0", AdvertiseHost: "127.0.0.1", TTL: time.Minute, MaxConcurrent: 4})
			if err != nil {
				return "", err.Error(), 3
			}
			mu.Lock()
			sessions = append(sessions, relay)
			mu.Unlock()
			encoded, err := json.Marshal(relay.Info)
			if err != nil {
				return "", err.Error(), 3
			}
			return string(encoded), "", 0
		case "fetch":
			size, err := strconv.ParseInt(relayFlag(arguments, "--size"), 10, 64)
			if err != nil {
				return "", err.Error(), 3
			}
			outcome, err := file.RelayFetch(context.Background(), file.RelayFetchOptions{URL: relayFlag(arguments, "--url"), Token: relayFlag(arguments, "--token"), CertFingerprint: relayFlag(arguments, "--fingerprint"), SHA256: relayFlag(arguments, "--sha256"), Size: size, Dest: relayFlag(arguments, "--dest"), PartRetention: time.Hour, Timeout: 5 * time.Second})
			if err != nil {
				return "", err.Error(), 3
			}
			encoded, err := json.Marshal(outcome)
			if err != nil {
				return "", err.Error(), 3
			}
			return string(encoded), "", 0
		default:
			return "", "unknown relay operation", 3
		}
	}
	t.Cleanup(func() {
		mu.Lock()
		defer mu.Unlock()
		for _, session := range sessions {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			warnings := session.Stop(ctx)
			cancel()
			if len(warnings) > 0 {
				t.Errorf("relay cleanup warnings: %v", warnings)
			}
		}
	})
}

func relayCommandArguments(command string) []string {
	fields := strings.Fields(command)
	for index := range fields {
		fields[index] = strings.Trim(fields[index], "'")
	}
	return fields
}

func relayFlag(arguments []string, name string) string {
	for index := 0; index+1 < len(arguments); index++ {
		if arguments[index] == name {
			return arguments[index+1]
		}
	}
	return ""
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
