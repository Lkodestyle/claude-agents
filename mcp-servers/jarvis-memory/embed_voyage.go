package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/philippgille/chromem-go"
)

// Voyage AI embeddings client.
//
// Two things set Voyage apart for our use-case:
//   1. Anthropic's officially recommended embedding partner.
//   2. Asymmetric retrieval: different `input_type` for documents vs queries
//      improves recall quality noticeably. We expose that via DocumentFunc
//      and QueryFunc, which the Store then uses for Add vs Query paths.
//
// chromem-go doesn't ship a Voyage provider, so we roll our own against the
// public HTTP endpoint. It's a thin, dependency-free wrapper.

const (
	voyageDefaultBaseURL = "https://api.voyageai.com/v1"
	voyageDefaultModel   = "voyage-3.5-lite"
)

type VoyageClient struct {
	apiKey  string
	model   string
	baseURL string
	http    *http.Client
}

func NewVoyageClient(apiKey, model, baseURL string) *VoyageClient {
	if model == "" {
		model = voyageDefaultModel
	}
	if baseURL == "" {
		baseURL = voyageDefaultBaseURL
	}
	return &VoyageClient{
		apiKey:  apiKey,
		model:   model,
		baseURL: baseURL,
		http:    &http.Client{Timeout: 60 * time.Second},
	}
}

type voyageReq struct {
	Input     []string `json:"input"`
	Model     string   `json:"model"`
	InputType string   `json:"input_type,omitempty"`
}

type voyageResp struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

type voyageErrorResp struct {
	Detail string `json:"detail"`
}

func (c *VoyageClient) embedOne(ctx context.Context, text, inputType string) ([]float32, error) {
	body, err := json.Marshal(voyageReq{
		Input:     []string{text},
		Model:     c.model,
		InputType: inputType,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal voyage req: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build voyage req: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("voyage request: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		var ve voyageErrorResp
		_ = json.Unmarshal(raw, &ve)
		if ve.Detail != "" {
			return nil, fmt.Errorf("voyage %d: %s", resp.StatusCode, ve.Detail)
		}
		return nil, fmt.Errorf("voyage status %d: %s", resp.StatusCode, string(raw))
	}

	var out voyageResp
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("parse voyage resp: %w", err)
	}
	if len(out.Data) == 0 {
		return nil, errors.New("voyage returned no embeddings")
	}
	return out.Data[0].Embedding, nil
}

// DocumentFunc is the EmbeddingFunc used when storing memories.
func (c *VoyageClient) DocumentFunc() chromem.EmbeddingFunc {
	return func(ctx context.Context, text string) ([]float32, error) {
		return c.embedOne(ctx, text, "document")
	}
}

// QueryFunc is the EmbeddingFunc used when recalling memories.
func (c *VoyageClient) QueryFunc() chromem.EmbeddingFunc {
	return func(ctx context.Context, text string) ([]float32, error) {
		return c.embedOne(ctx, text, "query")
	}
}
