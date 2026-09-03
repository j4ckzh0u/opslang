package exec

// Full-pipeline end-to-end tests: a real in-process SSH server with BOTH
// command execution and an SFTP subsystem. The tests verify the complete
// chain — connect, arch detection, runner upload, stdin piping, output
// parsing — against real bytes on the wire.
//
// The pipeline normally uploads a Linux runner; the test host may not be
// Linux. Commands therefore run locally on the test host, and the runner
// invocation is redirected to a real ops-runner binary built for the host
// platform. Everything else (SFTP upload, chmod, JSON protocol) is real.

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/j4ckzh0u/opslang/internal/compiler"
	"github.com/j4ckzh0u/opslang/internal/runner"
	"github.com/j4ckzh0u/opslang/internal/sshx"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// pipelineServer is an SSH server with password auth, an exec channel that
// runs commands locally, and an SFTP subsystem rooted at "/".
type pipelineServer struct {
	listener    net.Listener
	mu          sync.Mutex
	commands    []string
	connections int

	// hostRunner is executed whenever the pipeline asks to run an uploaded
	// ops-runner binary (which is built for the remote OS).
	hostRunner string
}

func newPipelineServer(t *testing.T) *pipelineServer {
	t.Helper()

	// Fresh remote cache per test: cache-hit paths skip chmod/upload, and
	// the assertions below expect a full cold deployment.
	t.Setenv("OPSLANG_REMOTE_CACHE_DIR", filepath.Join(t.TempDir(), "rcache"))

	// Per-test TOFU store: each test server gets a fresh host key, and a
	// shared ~/.ssh/opslang_known_hosts accumulates stale entries for
	// reused local ports — reruns then fail the hostkey check.
	t.Setenv("OPSLANG_KNOWN_HOSTS", filepath.Join(t.TempDir(), "known_hosts"))

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.ParsePrivateKey(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}))
	if err != nil {
		t.Fatal(err)
	}

	cfg := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == "root" && string(pass) == "pipetest" {
				return nil, nil
			}
			return nil, fmt.Errorf("auth failed")
		},
	}
	cfg.AddHostKey(signer)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	srv := &pipelineServer{listener: listener}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go srv.serve(conn, cfg)
		}
	}()
	t.Cleanup(func() { listener.Close() })
	return srv
}

func (s *pipelineServer) port() int {
	_, port, _ := net.SplitHostPort(s.listener.Addr().String())
	var p int
	fmt.Sscanf(port, "%d", &p)
	return p
}

func (s *pipelineServer) recordedCommands() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.commands...)
}

func (s *pipelineServer) connectionCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.connections
}

func (s *pipelineServer) serve(netConn net.Conn, cfg *ssh.ServerConfig) {
	serverConn, chans, reqs, err := ssh.NewServerConn(netConn, cfg)
	if err != nil {
		netConn.Close()
		return
	}
	defer serverConn.Close()
	s.mu.Lock()
	s.connections++
	s.mu.Unlock()
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
		go s.handleSession(channel, requests)
	}
}

func (s *pipelineServer) handleSession(channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()
	for req := range requests {
		switch req.Type {
		case "subsystem":
			if string(req.Payload[4:]) == "sftp" {
				if req.WantReply {
					req.Reply(true, nil)
				}
				server, err := sftp.NewServer(channel)
				if err != nil {
					return
				}
				_ = server.Serve()
				server.Close()
				return
			}
			if req.WantReply {
				req.Reply(false, nil)
			}
		case "exec":
			cmd := string(req.Payload[4:])
			if req.WantReply {
				req.Reply(true, nil)
			}
			s.mu.Lock()
			s.commands = append(s.commands, cmd)
			runnerPath := s.hostRunner
			s.mu.Unlock()

			// Redirect the uploaded (remote-OS) runner to the host build.
			// Only when the runner is the command itself - NOT for helper
			// commands like chmod that merely mention its path.
			actual := cmd
			fields := strings.Fields(cmd)
			if runnerPath != "" && len(fields) > 0 && strings.HasSuffix(fields[0], "/ops-runner") {
				fields[0] = runnerPath
				actual = strings.Join(fields, " ")
			}

			execCmd := exec.Command("sh", "-c", actual)
			execCmd.Stdin = channel
			execCmd.Stdout = channel
			execCmd.Stderr = channel.Stderr()
			err := execCmd.Run()

			status := uint32(0)
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					status = uint32(exitErr.ExitCode())
				} else {
					status = 1
				}
			}
			_, _ = channel.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{status}))
			return
		default:
			if req.WantReply {
				req.Reply(false, nil)
			}
		}
	}
}

// buildHostRunner compiles the real ops-runner for the test host platform.
func buildHostRunner(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "ops-runner-host")
	cmd := exec.Command("go", "build", "-o", bin, "github.com/j4ckzh0u/opslang/cmd/ops-runner")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build host runner: %v\n%s", err, out)
	}
	return bin
}

func TestFullPipelineRunnerMode(t *testing.T) {
	if testing.Short() {
		t.Skip("full pipeline test builds binaries")
	}

	server := newPipelineServer(t)
	server.hostRunner = buildHostRunner(t)

	e := &Executor{
		Instructions: &runner.InstructionPackage{
			Version: "1.0",
			TaskID:  "e2e-runner",
			Instructions: []runner.Instruction{
				{Op: "sys.cpu.usage", Assign: "cpu"},
				{Op: "report", Args: map[string]interface{}{"cpu": "$cpu"}},
			},
		},
		Password:   "pipetest",
		RunnerPath: server.hostRunner, // exists; the server redirects execution to the host build anyway
	}
	e.buildInFlight = map[string]*sync.Once{}
	e.buildResults = map[string]error{}
	e.runnerCache = newRunnerCache("")

	target := Target{Name: "pipe-host", Host: "127.0.0.1", Port: server.port(), User: "root"}
	result := e.executeOnHost(context.Background(), target)

	if result.Status != "success" {
		t.Fatalf("status = %q, error = %q, errors = %v", result.Status, result.Error, result.Errors)
	}
	cpu, ok := result.Data["cpu"].(map[string]interface{})
	if !ok {
		t.Fatalf("report missing cpu data: %+v", result.Data)
	}
	if _, ok := cpu["percent"]; !ok {
		t.Errorf("cpu result missing percent: %+v", cpu)
	}

	cmds := server.recordedCommands()
	if !containsCommand(cmds, "uname -m") {
		t.Errorf("arch detection (uname -m) not executed: %v", cmds)
	}
}

// TestFullPipelineAOTMode drives the complete AOT deploy path: compile the
// script for the host platform, upload it next to the runner, rewrite the
// placeholder, and execute it through binary.exec inside the real runner.
func TestFullPipelineAOTMode(t *testing.T) {
	if testing.Short() {
		t.Skip("full pipeline test builds binaries")
	}

	server := newPipelineServer(t)
	server.hostRunner = buildHostRunner(t)

	dir := t.TempDir()
	script := filepath.Join(dir, "app.ops")
	source := `
let cores = sys.cpu.count()
report { logical: cores.logical, physical: cores.physical }
`
	if err := os.WriteFile(script, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	c, err := compiler.NewCompiler()
	if err != nil {
		t.Fatal(err)
	}

	e := &Executor{
		Instructions: &runner.InstructionPackage{
			Version: "1.0",
			TaskID:  "e2e-aot",
			Instructions: []runner.Instruction{
				{Op: "binary.exec", Args: map[string]interface{}{"path": AppBinaryPlaceholder}},
			},
		},
		Password:   "pipetest",
		RunnerPath: server.hostRunner,
		AppBinary: func(goos, goarch string) (string, error) {
			bin := filepath.Join(dir, "app-host")
			if err := c.Compile(script, "", bin); err != nil {
				return "", err
			}
			return bin, nil
		},
	}
	e.buildInFlight = map[string]*sync.Once{}
	e.buildResults = map[string]error{}
	e.runnerCache = newRunnerCache("")

	target := Target{Name: "pipe-host", Host: "127.0.0.1", Port: server.port(), User: "root"}
	result := e.executeOnHost(context.Background(), target)

	if result.Status != "success" {
		t.Fatalf("status = %q, error = %q, errors = %v", result.Status, result.Error, result.Errors)
	}
	// binary.exec parses the AOT binary's JSON report (the printed _output).
	if _, ok := result.Data["logical"]; !ok {
		t.Logf("binary.exec result shape: %+v", result.Data)
	}

	cmds := server.recordedCommands()
	if !containsCommandPrefix(cmds, "chmod 0755") {
		t.Errorf("uploaded binaries were not chmodded: %v", cmds)
	}
}

// TestFullPipelineRunnerPartialFailurePropagates: a failing instruction
// must yield a non-success host result (the old code reported success).
func TestFullPipelineRunnerPartialFailurePropagates(t *testing.T) {
	if testing.Short() {
		t.Skip("full pipeline test builds binaries")
	}

	server := newPipelineServer(t)
	server.hostRunner = buildHostRunner(t)

	e := &Executor{
		Instructions: &runner.InstructionPackage{
			Version: "1.0",
			TaskID:  "e2e-fail",
			Instructions: []runner.Instruction{
				{Op: "file.read", Args: map[string]interface{}{"path": "/nonexistent/definitely/missing/file"}},
			},
		},
		Password:   "pipetest",
		RunnerPath: server.hostRunner,
	}
	e.buildInFlight = map[string]*sync.Once{}
	e.buildResults = map[string]error{}
	e.runnerCache = newRunnerCache("")

	target := Target{Name: "pipe-host", Host: "127.0.0.1", Port: server.port(), User: "root"}
	result := e.executeOnHost(context.Background(), target)

	if result.Status == "success" {
		t.Fatal("a failed instruction must not report host success")
	}
	if len(result.Errors) == 0 {
		t.Error("expected error details from the runner output")
	}
}

func containsCommand(cmds []string, want string) bool {
	for _, c := range cmds {
		if c == want {
			return true
		}
	}
	return false
}

func containsCommandPrefix(cmds []string, prefix string) bool {
	for _, c := range cmds {
		if strings.HasPrefix(c, prefix) {
			return true
		}
	}
	return false
}

// Ensure sshx import stays referenced when test shapes change.
var _ = sshx.Config{Timeout: time.Second}

// TestRemoteBinaryCacheSkipsReupload: the second deployment to the same
// host must reuse the cached runner - no upload, no chmod. This is the
// bandwidth guarantee the cache exists for; assert it, don't assume it.
func TestRemoteBinaryCacheSkipsReupload(t *testing.T) {
	if testing.Short() {
		t.Skip("full pipeline test builds binaries")
	}

	server := newPipelineServer(t)
	server.hostRunner = buildHostRunner(t)
	connectionPool, err := NewConnectionPool(1)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := connectionPool.Close(); err != nil {
			t.Errorf("close connection pool: %v", err)
		}
	}()

	newExecutor := func() *Executor {
		e := &Executor{
			Instructions: &runner.InstructionPackage{
				Version:      "1.0",
				TaskID:       "cache-test",
				Instructions: []runner.Instruction{{Op: "sys.cpu.usage", Assign: "cpu"}},
			},
			Password:       "pipetest",
			RunnerPath:     server.hostRunner,
			ConnectionPool: connectionPool,
		}
		e.buildInFlight = map[string]*sync.Once{}
		e.buildResults = map[string]error{}
		e.appPaths = map[string]string{}
		e.runnerCache = newRunnerCache("")
		return e
	}

	target := Target{Name: "pipe-host", Host: "127.0.0.1", Port: server.port(), User: "root"}

	r1 := newExecutor().executeOnHost(context.Background(), target)
	if r1.Status != "success" {
		t.Fatalf("first run failed: %+v", r1)
	}

	r2 := newExecutor().executeOnHost(context.Background(), target)
	if r2.Status != "success" {
		t.Fatalf("second run failed: %+v", r2)
	}

	cmds := server.recordedCommands()
	uploads := 0
	for _, c := range cmds {
		if strings.HasPrefix(c, "chmod 0755") {
			uploads++
		}
	}
	if uploads != 1 {
		t.Errorf("expected exactly 1 upload (first run), saw %d in %v - cache is not preventing re-uploads", uploads, cmds)
	}
	if got := server.connectionCount(); got != 1 {
		t.Errorf("SSH connections = %d, want 1 reused connection", got)
	}
}
