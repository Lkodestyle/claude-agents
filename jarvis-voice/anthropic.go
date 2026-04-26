package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Minimal Anthropic Messages client. We deliberately don't pull in the
// official SDK for the MVP — the Messages endpoint is small and stable
// enough that a tight HTTP wrapper keeps deps + binary size low. We can
// switch to anthropic-sdk-go later if we need streaming, tool use, or
// extended thinking.
//
// Docs: https://docs.anthropic.com/en/api/messages

const (
	anthropicURL     = "https://api.anthropic.com/v1/messages"
	anthropicVersion = "2023-06-01"
)

// Message is a turn in the chat history. Roles are "user" or "assistant".
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type messagesReq struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system,omitempty"`
	Messages  []Message `json:"messages"`
}

type messagesResp struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func anthropicChat(ctx context.Context, cfg *Config, history []Message) (string, error) {
	body, err := json.Marshal(messagesReq{
		Model:     cfg.Model,
		MaxTokens: 400, // keep replies tight; voice playback time scales with this
		System:    cfg.SystemPrompt,
		Messages:  history,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("x-api-key", cfg.AnthropicKey)
	req.Header.Set("anthropic-version", anthropicVersion)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("anthropic request: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	var mr messagesResp
	_ = json.Unmarshal(raw, &mr)
	if mr.Error != nil {
		return "", fmt.Errorf("anthropic %s: %s", mr.Error.Type, mr.Error.Message)
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("anthropic %d: %s", resp.StatusCode, string(raw))
	}

	for _, c := range mr.Content {
		if c.Type == "text" {
			return c.Text, nil
		}
	}
	return "", fmt.Errorf("anthropic returned no text content")
}
