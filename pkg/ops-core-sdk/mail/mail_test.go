package mail

import (
	"strings"
	"testing"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     MailConfig
		wantErr string
	}{
		{
			name:    "empty host",
			cfg:     MailConfig{From: "a@b.com", To: []string{"c@d.com"}, Subject: "test"},
			wantErr: "SMTP host is required",
		},
		{
			name:    "empty from",
			cfg:     MailConfig{Host: "smtp.example.com", To: []string{"c@d.com"}, Subject: "test"},
			wantErr: "from address is required",
		},
		{
			name:    "no recipients",
			cfg:     MailConfig{Host: "smtp.example.com", From: "a@b.com", Subject: "test"},
			wantErr: "at least one recipient",
		},
		{
			name:    "empty subject",
			cfg:     MailConfig{Host: "smtp.example.com", From: "a@b.com", To: []string{"c@d.com"}},
			wantErr: "subject is required",
		},
		{
			name:    "valid config",
			cfg:     MailConfig{Host: "smtp.example.com", From: "a@b.com", To: []string{"c@d.com"}, Subject: "test"},
			wantErr: "",
		},
		{
			name:    "CC only",
			cfg:     MailConfig{Host: "smtp.example.com", From: "a@b.com", CC: []string{"c@d.com"}, Subject: "test"},
			wantErr: "",
		},
		{
			name:    "BCC only",
			cfg:     MailConfig{Host: "smtp.example.com", From: "a@b.com", BCC: []string{"c@d.com"}, Subject: "test"},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.cfg)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("validateConfig() unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("validateConfig() expected error containing %q, got nil", tt.wantErr)
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("validateConfig() error = %v, want containing %q", err, tt.wantErr)
				}
			}
		})
	}
}

func TestSendValidation(t *testing.T) {
	// Test that Send returns error for invalid config without trying to connect
	result := Send(MailConfig{})
	if result.Success {
		t.Error("Send() should fail with empty config")
	}
	if result.Error == "" {
		t.Error("Send() should have error message")
	}
	// Duration may be 0 for immediate validation failures
}

func TestBuildMessage(t *testing.T) {
	cfg := MailConfig{
		Host:    "smtp.example.com",
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Test Subject",
		Body:    "Test body content",
	}

	msg, err := buildMessage(cfg)
	if err != nil {
		t.Fatalf("buildMessage() error: %v", err)
	}

	msgStr := string(msg)

	// Check headers
	if !strings.Contains(msgStr, "From: sender@example.com") {
		t.Error("missing From header")
	}
	if !strings.Contains(msgStr, "To: recipient@example.com") {
		t.Error("missing To header")
	}
	if !strings.Contains(msgStr, "Subject: Test Subject") {
		t.Error("missing Subject header")
	}
	if !strings.Contains(msgStr, "MIME-Version: 1.0") {
		t.Error("missing MIME-Version header")
	}
	if !strings.Contains(msgStr, "text/plain") {
		t.Error("should be text/plain for non-HTML")
	}
}

func TestBuildMessageHTML(t *testing.T) {
	cfg := MailConfig{
		Host:    "smtp.example.com",
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "HTML Test",
		Body:    "<h1>Hello</h1>",
		HTML:    true,
	}

	msg, err := buildMessage(cfg)
	if err != nil {
		t.Fatalf("buildMessage() error: %v", err)
	}

	if !strings.Contains(string(msg), "text/html") {
		t.Error("should be text/html for HTML email")
	}
}

func TestBuildMessageWithCC(t *testing.T) {
	cfg := MailConfig{
		Host:    "smtp.example.com",
		From:    "sender@example.com",
		To:      []string{"to@example.com"},
		CC:      []string{"cc1@example.com", "cc2@example.com"},
		Subject: "CC Test",
		Body:    "test",
	}

	msg, err := buildMessage(cfg)
	if err != nil {
		t.Fatalf("buildMessage() error: %v", err)
	}

	msgStr := string(msg)
	if !strings.Contains(msgStr, "Cc: cc1@example.com, cc2@example.com") {
		t.Error("missing CC header")
	}
}

func TestEncodeBase64(t *testing.T) {
	input := "Hello, World!"
	encoded := encodeBase64(input)

	// Should be base64 encoded
	if encoded == input {
		t.Error("should be encoded")
	}

	// Should not contain raw input
	if strings.Contains(encoded, "Hello") {
		t.Error("should not contain raw text")
	}
}

func TestEncodeSubject(t *testing.T) {
	tests := []struct {
		input string
		check func(string) bool
	}{
		{"ASCII Subject", func(s string) bool { return strings.Contains(s, "ASCII Subject") }},
		{"中文主题", func(s string) bool { return strings.Contains(s, "UTF-8") }},
		{"日本語", func(s string) bool { return strings.Contains(s, "UTF-8") }},
	}

	for _, tt := range tests {
		encoded := encodeSubject(tt.input)
		if !tt.check(encoded) {
			t.Errorf("encodeSubject(%q) = %q, check failed", tt.input, encoded)
		}
	}
}

func TestGenerateMessageID(t *testing.T) {
	id1 := generateMessageID()
	id2 := generateMessageID()

	if id1 == "" {
		t.Error("message ID should not be empty")
	}
	if id1 == id2 {
		t.Error("message IDs should be unique")
	}
	if !strings.Contains(id1, "@opslang") {
		t.Error("message ID should contain @opslang")
	}
}

func TestGenerateBoundary(t *testing.T) {
	b1 := generateBoundary()
	b2 := generateBoundary()

	if b1 == "" {
		t.Error("boundary should not be empty")
	}
	if b1 == b2 {
		t.Error("boundaries should be unique")
	}
	if !strings.HasPrefix(b1, "==") {
		t.Error("boundary should start with ==")
	}
}

func TestSendSimple(t *testing.T) {
	// This will fail to connect but should validate params
	result := SendSimple("", 0, "", nil, "", "")
	if result.Success {
		t.Error("SendSimple() should fail with empty params")
	}
}

func TestSendWithAuth(t *testing.T) {
	// This will fail to connect but should validate params
	result := SendWithAuth("", 0, "", "", "", nil, "", "", false)
	if result.Success {
		t.Error("SendWithAuth() should fail with empty params")
	}
}
