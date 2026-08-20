// Package webhook provides generic HTTP webhook notifications.
// Supports Slack, Discord, Teams, Mattermost, Rocket.Chat, and custom webhooks.
package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// WebhookResult represents the result of sending a webhook.
type WebhookResult struct {
	Success    bool   `json:"success"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Error      string `json:"error,omitempty"`
	Duration   int64  `json:"duration_ms"`
}

// WebhookConfig holds webhook configuration.
type WebhookConfig struct {
	URL      string
	Method   string
	Headers  map[string]string
	Body     interface{}
	Timeout  time.Duration
	Insecure bool
}

// Send sends a webhook with the given configuration.
func Send(cfg WebhookConfig) WebhookResult {
	start := time.Now()

	if cfg.URL == "" {
		return WebhookResult{
			Success: false,
			Message: "webhook URL is required",
			Error:   "empty URL",
		}
	}

	if cfg.Method == "" {
		cfg.Method = "POST"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	// Encode body
	var bodyReader io.Reader
	if cfg.Body != nil {
		bodyBytes, err := json.Marshal(cfg.Body)
		if err != nil {
			return WebhookResult{
				Success: false,
				Message: "failed to encode body",
				Error:   err.Error(),
			}
		}
		bodyReader = bytes.NewReader(bodyBytes)
		if cfg.Headers == nil {
			cfg.Headers = make(map[string]string)
		}
		if _, ok := cfg.Headers["Content-Type"]; !ok {
			cfg.Headers["Content-Type"] = "application/json"
		}
	}

	// Create request
	req, err := http.NewRequest(cfg.Method, cfg.URL, bodyReader)
	if err != nil {
		return WebhookResult{
			Success: false,
			Message: "failed to create request",
			Error:   err.Error(),
		}
	}

	// Add headers
	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}

	// Execute request
	client := &http.Client{Timeout: cfg.Timeout}
	resp, err := client.Do(req)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		return WebhookResult{
			Success:  false,
			Message:  "request failed",
			Error:    err.Error(),
			Duration: duration,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return WebhookResult{
			Success:    false,
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("webhook returned %s", resp.Status),
			Duration:   duration,
		}
	}

	return WebhookResult{
		Success:    true,
		StatusCode: resp.StatusCode,
		Message:    "webhook delivered",
		Duration:   duration,
	}
}

// SendSlack sends a message to Slack via incoming webhook.
func SendSlack(webhookURL, text string) WebhookResult {
	return Send(WebhookConfig{
		URL:  webhookURL,
		Body: map[string]string{"text": text},
	})
}

// SendSlackRich sends a rich message to Slack with attachments.
func SendSlackRich(webhookURL, text, channel, username, iconEmoji string, attachments []map[string]interface{}) WebhookResult {
	payload := map[string]interface{}{
		"text": text,
	}
	if channel != "" {
		payload["channel"] = channel
	}
	if username != "" {
		payload["username"] = username
	}
	if iconEmoji != "" {
		payload["icon_emoji"] = iconEmoji
	}
	if len(attachments) > 0 {
		payload["attachments"] = attachments
	}
	return Send(WebhookConfig{
		URL:  webhookURL,
		Body: payload,
	})
}

// SendDiscord sends a message to Discord via webhook.
func SendDiscord(webhookURL, content string) WebhookResult {
	return Send(WebhookConfig{
		URL:  webhookURL,
		Body: map[string]string{"content": content},
	})
}

// SendDiscordEmbed sends an embed message to Discord.
func SendDiscordEmbed(webhookURL, title, description, color string, fields []map[string]interface{}) WebhookResult {
	embed := map[string]interface{}{
		"title":       title,
		"description": description,
	}
	if color != "" {
		embed["color"] = color
	}
	if len(fields) > 0 {
		embed["fields"] = fields
	}
	return Send(WebhookConfig{
		URL: webhookURL,
		Body: map[string]interface{}{
			"embeds": []interface{}{embed},
		},
	})
}

// SendTeams sends a message to Microsoft Teams via webhook.
func SendTeams(webhookURL, title, text string) WebhookResult {
	return Send(WebhookConfig{
		URL: webhookURL,
		Body: map[string]interface{}{
			"@type":      "MessageCard",
			"@context":   "http://schema.org/extensions",
			"themeColor": "0076D7",
			"summary":    title,
			"sections": []map[string]interface{}{
				{
					"activityTitle": title,
					"text":          text,
					"markdown":      true,
				},
			},
		},
	})
}

// SendMattermost sends a message to Mattermost via webhook.
func SendMattermost(webhookURL, text, channel, username, iconURL string) WebhookResult {
	payload := map[string]string{"text": text}
	if channel != "" {
		payload["channel"] = channel
	}
	if username != "" {
		payload["username"] = username
	}
	if iconURL != "" {
		payload["icon_url"] = iconURL
	}
	return Send(WebhookConfig{
		URL:  webhookURL,
		Body: payload,
	})
}

// SendRocketChat sends a message to Rocket.Chat via webhook.
func SendRocketChat(webhookURL, text, alias, emoji, avatar string) WebhookResult {
	payload := map[string]string{"text": text}
	if alias != "" {
		payload["alias"] = alias
	}
	if emoji != "" {
		payload["emoji"] = emoji
	}
	if avatar != "" {
		payload["avatar"] = avatar
	}
	return Send(WebhookConfig{
		URL:  webhookURL,
		Body: payload,
	})
}

// SendGeneric sends a custom payload to any webhook URL.
func SendGeneric(url, method string, headers map[string]string, body interface{}) WebhookResult {
	return Send(WebhookConfig{
		URL:     url,
		Method:  method,
		Headers: headers,
		Body:    body,
	})
}
