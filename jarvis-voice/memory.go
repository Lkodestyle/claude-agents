package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MemoryClient connects the voice loop to the jarvis-memory MCP server
// (Milestone B). The memory store keeps a single owner — the MCP server —
// and jarvis-voice is just another client of it, exactly like Claude Code.
// The whole feature is opt-in: when JARVIS_VOICE_MEMORY is off (the default)
// none of this code runs and the loop behaves like Milestone A.
type MemoryClient struct {
	session *mcp.ClientSession
}

// memoryHit is the subset of the server's RecallHit we consume.
type memoryHit struct {
	Content    string  `json:"content"`
	Similarity float32 `json:"similarity"`
	Scope      string  `json:"scope,omitempty"`
}

// connectMemory spawns the jarvis-memory binary and performs the MCP
// handshake over stdio. The caller owns Close().
func connectMemory(ctx context.Context, binPath string) (*MemoryClient, error) {
	bin, err := resolveMemoryBin(binPath)
	if err != nil {
		return nil, err
	}

	client := mcp.NewClient(&mcp.Implementation{
		Name:    appName,
		Version: appVersion,
	}, nil)

	session, err := client.Connect(ctx, &mcp.CommandTransport{Command: exec.Command(bin)}, nil)
	if err != nil {
		return nil, fmt.Errorf("connect to jarvis-memory (%s): %w", bin, err)
	}
	return &MemoryClient{session: session}, nil
}

// resolveMemoryBin finds the jarvis-memory executable: explicit path via
// JARVIS_MEMORY_BIN, then next to our own binary (both live in bin/ after
// build-jarvis-memory.sh), then PATH.
func resolveMemoryBin(explicit string) (string, error) {
	name := "jarvis-memory"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}

	if explicit != "" {
		if _, err := os.Stat(explicit); err != nil {
			return "", fmt.Errorf("JARVIS_MEMORY_BIN=%q: %w", explicit, err)
		}
		return explicit, nil
	}

	if exe, err := os.Executable(); err == nil {
		sibling := filepath.Join(filepath.Dir(exe), name)
		if _, err := os.Stat(sibling); err == nil {
			return sibling, nil
		}
	}

	if p, err := exec.LookPath(name); err == nil {
		return p, nil
	}

	return "", fmt.Errorf("jarvis-memory binary not found (set JARVIS_MEMORY_BIN or run scripts/build-jarvis-memory.sh)")
}

func (m *MemoryClient) Close() error {
	return m.session.Close()
}

// Recall runs a semantic search and returns hits above minSimilarity.
func (m *MemoryClient) Recall(ctx context.Context, query string, k int, minSimilarity float32) ([]memoryHit, error) {
	res, err := m.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "recall",
		Arguments: map[string]any{"query": query, "k": k},
	})
	if err != nil {
		return nil, err
	}

	var out struct {
		Hits []memoryHit `json:"hits"`
	}
	if err := decodeToolResult(res, &out); err != nil {
		return nil, err
	}

	hits := out.Hits[:0]
	for _, h := range out.Hits {
		if h.Similarity >= minSimilarity {
			hits = append(hits, h)
		}
	}
	return hits, nil
}

// Remember persists one voice exchange as a session-scoped memory.
func (m *MemoryClient) Remember(ctx context.Context, content string) error {
	res, err := m.session.CallTool(ctx, &mcp.CallToolParams{
		Name: "remember",
		Arguments: map[string]any{
			"content": content,
			"scope":   "session",
			"tags":    []string{"voice"},
			"source":  appName,
		},
	})
	if err != nil {
		return err
	}
	var out struct {
		ID string `json:"id"`
	}
	return decodeToolResult(res, &out)
}

// decodeToolResult unmarshals a tool result into v, preferring the typed
// StructuredContent the go-sdk emits for typed handlers and falling back to
// the JSON text content.
func decodeToolResult(res *mcp.CallToolResult, v any) error {
	if res.IsError {
		return fmt.Errorf("jarvis-memory: %s", firstText(res))
	}
	if res.StructuredContent != nil {
		raw, err := json.Marshal(res.StructuredContent)
		if err != nil {
			return err
		}
		return json.Unmarshal(raw, v)
	}
	if txt := firstText(res); txt != "" {
		return json.Unmarshal([]byte(txt), v)
	}
	return fmt.Errorf("jarvis-memory returned no content")
}

func firstText(res *mcp.CallToolResult) string {
	for _, c := range res.Content {
		if t, ok := c.(*mcp.TextContent); ok {
			return t.Text
		}
	}
	return ""
}

// memoryContextBlock renders recall hits as a system prompt suffix. Empty
// when there is nothing relevant, so the base prompt stays untouched.
func memoryContextBlock(hits []memoryHit) string {
	if len(hits) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n\nRecuerdos de conversaciones anteriores (usalos solo si vienen al caso, no los menciones si no aportan):\n")
	for _, h := range hits {
		b.WriteString("- ")
		b.WriteString(strings.TrimSpace(h.Content))
		b.WriteString("\n")
	}
	return b.String()
}
