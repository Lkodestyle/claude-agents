package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// Deepgram pre-recorded transcription endpoint. We use the simplest possible
// surface: POST raw audio bytes, get a JSON transcript back. No streaming, no
// SDK — keeps deps light.
//
// Docs: https://developers.deepgram.com/reference/listen-file

const deepgramURL = "https://api.deepgram.com/v1/listen"

type deepgramResponse struct {
	Results struct {
		Channels []struct {
			Alternatives []struct {
				Transcript string  `json:"transcript"`
				Confidence float64 `json:"confidence"`
			} `json:"alternatives"`
		} `json:"channels"`
	} `json:"results"`
}

func deepgramTranscribe(ctx context.Context, apiKey, wavPath string) (string, error) {
	f, err := os.Open(wavPath)
	if err != nil {
		return "", fmt.Errorf("open wav: %w", err)
	}
	defer f.Close()

	q := url.Values{}
	q.Set("model", "nova-3")      // current best general model
	q.Set("smart_format", "true") // capitalization, punctuation, numbers
	q.Set("punctuate", "true")
	q.Set("language", "multi") // auto-detect; works for ES + EN code-switching
	endpoint := deepgramURL + "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, f)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Token "+apiKey)
	req.Header.Set("Content-Type", "audio/wav")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("deepgram request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("deepgram %d: %s", resp.StatusCode, string(body))
	}

	var dr deepgramResponse
	if err := json.Unmarshal(body, &dr); err != nil {
		return "", fmt.Errorf("parse deepgram resp: %w", err)
	}
	if len(dr.Results.Channels) == 0 || len(dr.Results.Channels[0].Alternatives) == 0 {
		return "", nil
	}
	return dr.Results.Channels[0].Alternatives[0].Transcript, nil
}
