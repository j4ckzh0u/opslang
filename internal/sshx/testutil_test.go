package sshx

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"net"
	"sync"

	"golang.org/x/crypto/ssh"
)

// mockSSHServer provides a simple SSH server for testing.
type mockSSHServer struct {
	listener net.Listener
	config   *ssh.ServerConfig
	wg       sync.WaitGroup
	quit     chan struct{}
}

// newMockSSHServer creates and starts a mock SSH server.
func newMockSSHServer(password string) (*mockSSHServer, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	signer, err := ssh.ParsePrivateKey(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == "root" && string(pass) == password {
				return nil, nil
			}
			return nil, fmt.Errorf("authentication failed")
		},
	}
	config.AddHostKey(signer)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	server := &mockSSHServer{
		listener: listener,
		config:   config,
		quit:     make(chan struct{}),
	}

	server.wg.Add(1)
	go server.serve()

	return server, nil
}

func (s *mockSSHServer) serve() {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return
			default:
				continue
			}
		}

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handleConnection(conn)
		}()
	}
}

func (s *mockSSHServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	sshConn, chans, reqs, err := ssh.NewServerConn(conn, s.config)
	if err != nil {
		return
	}
	defer sshConn.Close()

	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			continue
		}

		go s.handleSession(channel, requests)
	}
}

func (s *mockSSHServer) handleSession(channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()

	for req := range requests {
		switch req.Type {
		case "exec":
			// Reply to exec request first
			if req.WantReply {
				req.Reply(true, nil)
			}

			// Parse command (4 bytes length prefix + command string)
			cmd := ""
			if len(req.Payload) > 4 {
				cmdLen := binary.BigEndian.Uint32(req.Payload[:4])
				if int(cmdLen)+4 <= len(req.Payload) {
					cmd = string(req.Payload[4 : 4+cmdLen])
				}
			}

			// Execute command
			exitCode := s.executeCommand(channel, cmd)

			// Send exit status
			status := make([]byte, 4)
			binary.BigEndian.PutUint32(status, uint32(exitCode))
			channel.SendRequest("exit-status", false, status)
			return
		default:
			if req.WantReply {
				req.Reply(false, nil)
			}
		}
	}
}

// executeCommand simulates command execution. Returns exit code.
func (s *mockSSHServer) executeCommand(channel ssh.Channel, cmd string) int {
	switch cmd {
	case "echo hello":
		channel.Write([]byte("hello\n"))
		return 0
	case "exit 1":
		return 1
	case "stderr test":
		channel.Stderr().Write([]byte("error message\n"))
		return 0
	default:
		channel.Write([]byte("command executed: " + cmd + "\n"))
		return 0
	}
}

func (s *mockSSHServer) Addr() string {
	return s.listener.Addr().String()
}

// Port returns the server's port number.
func (s *mockSSHServer) Port() int {
	_, portStr, _ := net.SplitHostPort(s.listener.Addr().String())
	port := 0
	for _, c := range portStr {
		port = port*10 + int(c-'0')
	}
	return port
}

func (s *mockSSHServer) Close() {
	close(s.quit)
	s.listener.Close()
	s.wg.Wait()
}
