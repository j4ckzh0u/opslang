package webhook

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendValidation(t *testing.T) {
	result := Send(WebhookConfig{})
	if result.Success {
		t.Error("Send() should fail with empty URL")
	}
	if !strings.Contains(result.Error, "URL") {
		t.Error("error should mention URL")
	}
}

func TestSendSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected application/json, got %s", ct)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	result := Send(WebhookConfig{
		URL:  server.URL,
		Body: map[string]string{"key": "value"},
	})

	if !result.Success {
		t.Errorf("Send() should succeed: %v", result.Error)
	}
	if result.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", result.StatusCode)
	}
	if result.Duration == 0 {
		t.Error("duration should be recorded")
	}
}

func TestSendFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	result := Send(WebhookConfig{
		URL:  server.URL,
		Body: map[string]string{"key": "value"},
	})

	if result.Success {
		t.Error("Send() should fail with 500 response")
	}
	if result.StatusCode != 500 {
		t.Errorf("expected status 500, got %d", result.StatusCode)
	}
}

func TestSendCustomMethod(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	result := Send(WebhookConfig{
		URL:    server.URL,
		Method: "PUT",
		Body:   map[string]string{"key": "value"},
	})

	if !result.Success {
		t.Errorf("Send() should succeed: %v", result.Error)
	}
}

func TestSendCustomHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth := r.Header.Get("Authorization"); auth != "Bearer token123" {
			t.Errorf("expected Bearer token123, got %s", auth)
		}
		if custom := r.Header.Get("X-Custom"); custom != "value" {
			t.Errorf("expected custom header, got %s", custom)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	result := Send(WebhookConfig{
		URL:    server.URL,
		Method: "POST",
		Headers: map[string]string{
			"Authorization": "Bearer token123",
			"X-Custom":      "value",
		},
		Body: map[string]string{"key": "value"},
	})

	if !result.Success {
		t.Errorf("Send() should succeed: %v", result.Error)
	}
}

func TestSendSlack(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]string
		json.NewDecoder(r.Body).Decode(&payload)
		if payload["text"] != "Hello Slack" {
			t.Errorf("expected 'Hello Slack', got %s", payload["text"])
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	result := SendSlack(server.URL, "Hello Slack")
	if !result.Success {
		t.Errorf("SendSlack() should succeed: %v", result.Error)
	}
}

func TestSendSlackRich(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		json.NewDecoder(r.Body).Decode(&payload)
		if payload["channel"] != "#general" {
			t.Errorf("expected #general, got %v", payload["channel"])
		}
		if payload["username"] != "Bot" {
			t.Errorf("expected Bot, got %v", payload["username"])
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	attachments := []map[string]interface{}{
		{"text": "Attachment 1"},
	}
	result := SendSlackRich(server.URL, "Hello", "#general", "Bot", ":robot:", attachments)
	if !result.Success {
		t.Errorf("SendSlackRich() should succeed: %v", result.Error)
	}
}

func TestSendDiscord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]string
		json.NewDecoder(r.Body).Decode(&payload)
		if payload["content"] != "Hello Discord" {
			t.Errorf("expected 'Hello Discord', got %s", payload["content"])
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	result := SendDiscord(server.URL, "Hello Discord")
	if !result.Success {
		t.Errorf("SendDiscord() should succeed: %v", result.Error)
	}
}

func TestSendDiscordEmbed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		json.NewDecoder(r.Body).Decode(&payload)
		embeds, ok := payload["embeds"].([]interface{})
		if !ok || len(embeds) == 0 {
			t.Error("expected embeds array")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	fields := []map[string]interface{}{
		{"name": "Field 1", "value": "Value 1", "inline": true},
	}
	result := SendDiscordEmbed(server.URL, "Title", "Description", "0xFF0000", fields)
	if !result.Success {
		t.Errorf("SendDiscordEmbed() should succeed: %v", result.Error)
	}
}

func TestSendTeams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		json.NewDecoder(r.Body).Decode(&payload)
		if payload["@type"] != "MessageCard" {
			t.Errorf("expected MessageCard type, got %v", payload["@type"])
		}
		sections, ok := payload["sections"].([]interface{})
		if !ok || len(sections) == 0 {
			t.Error("expected sections array")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	result := SendTeams(server.URL, "Alert", "Something happened")
	if !result.Success {
		t.Errorf("SendTeams() should succeed: %v", result.Error)
	}
}

func TestSendMattermost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]string
		json.NewDecoder(r.Body).Decode(&payload)
		if payload["channel"] != "town-square" {
			t.Errorf("expected town-square, got %s", payload["channel"])
		}
		if payload["username"] != "AlertBot" {
			t.Errorf("expected AlertBot, got %s", payload["username"])
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	result := SendMattermost(server.URL, "Hello", "town-square", "AlertBot", "")
	if !result.Success {
		t.Errorf("SendMattermost() should succeed: %v", result.Error)
	}
}

func TestSendRocketChat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]string
		json.NewDecoder(r.Body).Decode(&payload)
		if payload["alias"] != "Notifier" {
			t.Errorf("expected Notifier, got %s", payload["alias"])
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	result := SendRocketChat(server.URL, "Hello", "Notifier", ":bell:", "")
	if !result.Success {
		t.Errorf("SendRocketChat() should succeed: %v", result.Error)
	}
}

func TestSendGeneric(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.Header.Get("X-API-Key") != "secret" {
			t.Error("missing X-API-Key header")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	headers := map[string]string{"X-API-Key": "secret"}
	body := map[string]interface{}{"event": "test", "data": map[string]int{"count": 42}}
	result := SendGeneric(server.URL, "PATCH", headers, body)
	if !result.Success {
		t.Errorf("SendGeneric() should succeed: %v", result.Error)
	}
}

func TestSendNoBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength != 0 {
			t.Errorf("expected no body, got %d bytes", r.ContentLength)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	result := Send(WebhookConfig{
		URL:    server.URL,
		Method: "GET",
	})
	if !result.Success {
		t.Errorf("Send() should succeed: %v", result.Error)
	}
}
