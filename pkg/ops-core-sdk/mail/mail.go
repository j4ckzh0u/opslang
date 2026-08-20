// Package mail provides email sending capabilities for notifications.
// It supports SMTP with authentication, TLS, attachments, and HTML bodies.
package mail

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// MailResult represents the result of sending an email.
type MailResult struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Error    string `json:"error,omitempty"`
	Duration int64  `json:"duration_ms"`
}

// MailConfig holds SMTP configuration.
type MailConfig struct {
	Host       string
	Port       int
	Username   string
	Password   string
	From       string
	To         []string
	CC         []string
	BCC        []string
	Subject    string
	Body       string
	HTML       bool
	Attachments []string
	Timeout    time.Duration
	StartTLS   bool
	InsecureTLS bool
}

// Send sends an email with the given configuration.
func Send(cfg MailConfig) MailResult {
	start := time.Now()

	if err := validateConfig(cfg); err != nil {
		return MailResult{
			Success:  false,
			Message:  "validation failed",
			Error:    err.Error(),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.Port == 0 {
		cfg.Port = 587
	}

	err := sendMail(cfg)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		return MailResult{
			Success:  false,
			Message:  "failed to send email",
			Error:    err.Error(),
			Duration: duration,
		}
	}

	return MailResult{
		Success:  true,
		Message:  fmt.Sprintf("email sent to %d recipient(s)", len(cfg.To)+len(cfg.CC)+len(cfg.BCC)),
		Duration: duration,
	}
}

// SendSimple sends a plain text email with minimal configuration.
func SendSimple(host string, port int, from string, to []string, subject, body string) MailResult {
	return Send(MailConfig{
		Host:     host,
		Port:     port,
		From:     from,
		To:       to,
		Subject:  subject,
		Body:     body,
		StartTLS: true,
	})
}

// SendWithAuth sends an email with SMTP authentication.
func SendWithAuth(host string, port int, username, password, from string, to []string, subject, body string, html bool) MailResult {
	return Send(MailConfig{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
		To:       to,
		Subject:  subject,
		Body:     body,
		HTML:     html,
		StartTLS: true,
	})
}

func validateConfig(cfg MailConfig) error {
	if cfg.Host == "" {
		return fmt.Errorf("SMTP host is required")
	}
	if cfg.From == "" {
		return fmt.Errorf("from address is required")
	}
	if len(cfg.To) == 0 && len(cfg.CC) == 0 && len(cfg.BCC) == 0 {
		return fmt.Errorf("at least one recipient (to, cc, or bcc) is required")
	}
	if cfg.Subject == "" {
		return fmt.Errorf("subject is required")
	}
	return nil
}

func sendMail(cfg MailConfig) error {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	// Connect to SMTP server with timeout
	conn, err := net.DialTimeout("tcp", addr, cfg.Timeout)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	client, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Try STARTTLS if requested
	if cfg.StartTLS {
		tlsConfig := &tls.Config{
			ServerName:         cfg.Host,
			InsecureSkipVerify: cfg.InsecureTLS,
		}
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("STARTTLS failed: %w", err)
			}
		}
	}

	// Authenticate if credentials provided
	if cfg.Username != "" && cfg.Password != "" {
		auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
	}

	// Set sender
	if err := client.Mail(cfg.From); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}

	// Set all recipients
	allRecipients := append(append(cfg.To, cfg.CC...), cfg.BCC...)
	for _, rcpt := range allRecipients {
		if err := client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("RCPT TO (%s) failed: %w", rcpt, err)
		}
	}

	// Send message body
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA command failed: %w", err)
	}

	// Build message
	msg, err := buildMessage(cfg)
	if err != nil {
		wc.Close()
		return fmt.Errorf("failed to build message: %w", err)
	}

	if _, err := wc.Write(msg); err != nil {
		wc.Close()
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err := wc.Close(); err != nil {
		return fmt.Errorf("failed to close message: %w", err)
	}

	return client.Quit()
}

func buildMessage(cfg MailConfig) ([]byte, error) {
	var buf bytes.Buffer

	// Headers
	buf.WriteString(fmt.Sprintf("From: %s\r\n", cfg.From))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(cfg.To, ", ")))
	if len(cfg.CC) > 0 {
		buf.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(cfg.CC, ", ")))
	}
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", encodeSubject(cfg.Subject)))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	buf.WriteString("Message-ID: <" + generateMessageID() + ">\r\n")

	hasAttachments := len(cfg.Attachments) > 0

	if hasAttachments {
		// Multipart message with attachments
		boundary := generateBoundary()
		if cfg.HTML {
			buf.WriteString("Content-Type: multipart/mixed; boundary=\"" + boundary + "\"\r\n")
			buf.WriteString("\r\n")
			buf.WriteString("--" + boundary + "\r\n")
			buf.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
			buf.WriteString("Content-Transfer-Encoding: base64\r\n")
			buf.WriteString("\r\n")
			buf.WriteString(encodeBase64(cfg.Body))
			buf.WriteString("\r\n")
		} else {
			buf.WriteString("Content-Type: multipart/mixed; boundary=\"" + boundary + "\"\r\n")
			buf.WriteString("\r\n")
			buf.WriteString("--" + boundary + "\r\n")
			buf.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
			buf.WriteString("Content-Transfer-Encoding: base64\r\n")
			buf.WriteString("\r\n")
			buf.WriteString(encodeBase64(cfg.Body))
			buf.WriteString("\r\n")
		}

		// Add attachments
		for _, path := range cfg.Attachments {
			attachment, err := readAttachment(path, boundary)
			if err != nil {
				return nil, err
			}
			buf.Write(attachment)
		}

		buf.WriteString("--" + boundary + "--\r\n")
	} else {
		// Simple message without attachments
		if cfg.HTML {
			buf.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
			buf.WriteString("Content-Transfer-Encoding: base64\r\n")
			buf.WriteString("\r\n")
			buf.WriteString(encodeBase64(cfg.Body))
		} else {
			buf.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
			buf.WriteString("Content-Transfer-Encoding: base64\r\n")
			buf.WriteString("\r\n")
			buf.WriteString(encodeBase64(cfg.Body))
		}
	}

	return buf.Bytes(), nil
}

func readAttachment(path, boundary string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read attachment %s: %w", path, err)
	}

	filename := filepath.Base(path)
	contentType := mime.TypeByExtension(filepath.Ext(path))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	var buf bytes.Buffer
	buf.WriteString("--" + boundary + "\r\n")
	buf.WriteString(fmt.Sprintf("Content-Type: %s; name=\"%s\"\r\n", contentType, filename))
	buf.WriteString("Content-Transfer-Encoding: base64\r\n")
	buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", filename))
	buf.WriteString("\r\n")
	buf.WriteString(encodeBase64(string(data)))
	buf.WriteString("\r\n")

	return buf.Bytes(), nil
}

func encodeBase64(s string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(s))
	// Wrap lines at 76 characters per RFC 2045
	var lines []string
	for len(encoded) > 0 {
		end := 76
		if len(encoded) < end {
			end = len(encoded)
		}
		lines = append(lines, encoded[:end])
		encoded = encoded[end:]
	}
	return strings.Join(lines, "\r\n")
}

func encodeSubject(subject string) string {
	// Use RFC 2047 encoding for non-ASCII subjects
	return mime.QEncoding.Encode("UTF-8", subject)
}

func generateMessageID() string {
	return fmt.Sprintf("%d.%d@opslang", time.Now().UnixNano(), os.Getpid())
}

func generateBoundary() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("==%x%x==", time.Now().UnixNano(), b)
}
