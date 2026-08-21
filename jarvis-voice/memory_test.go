package main

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// TestMemoryClientHandshake spins up a real jarvis-memory server (path via
// JARVIS_MEMORY_BIN) and exercises the MCP handshake plus a recall on an
// empty store — the one call that needs no embedding provider. Skipped when
// the binary isn't available so `go test ./...` stays green on a bare clone.
func TestMemoryClientHandshake(t *testing.T) {
	bin := os.Getenv("JARVIS_MEMORY_BIN")
	if bin == "" {
		t.Skip("JARVIS_MEMORY_BIN not set; run scripts/build-jarvis-memory.sh first")
	}

	t.Setenv("JARVIS_DATA_DIR", t.TempDir())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mem, err := connectMemory(ctx, bin)
	if err != nil {
		t.Fatalf("connectMemory: %v", err)
	}
	defer mem.Close()

	hits, err := mem.Recall(ctx, "anything at all", 3, 0)
	if err != nil {
		t.Fatalf("Recall on empty store: %v", err)
	}
	if len(hits) != 0 {
		t.Fatalf("expected 0 hits on empty store, got %d", len(hits))
	}
}

// TestMemoryRememberRecallE2E exercises the full write→read cycle against
// real embeddings. Opt-in: needs both the server binary and a VOYAGE_API_KEY
// (the subprocess inherits our env), so it only runs when explicitly set up.
func TestMemoryRememberRecallE2E(t *testing.T) {
	bin := os.Getenv("JARVIS_MEMORY_BIN")
	if bin == "" || os.Getenv("VOYAGE_API_KEY") == "" {
		t.Skip("needs JARVIS_MEMORY_BIN and VOYAGE_API_KEY")
	}

	t.Setenv("JARVIS_DATA_DIR", t.TempDir())

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	mem, err := connectMemory(ctx, bin)
	if err != nil {
		t.Fatalf("connectMemory: %v", err)
	}
	defer mem.Close()

	fact := "El usuario dijo: mi perro se llama Rocco\nJarvis respondio: Buen nombre."
	if err := mem.Remember(ctx, fact); err != nil {
		t.Fatalf("Remember: %v", err)
	}

	hits, err := mem.Recall(ctx, "como se llama el perro del usuario?", 3, 0)
	if err != nil {
		t.Fatalf("Recall: %v", err)
	}
	if len(hits) == 0 {
		t.Fatal("expected at least one hit after Remember")
	}
	if !strings.Contains(hits[0].Content, "Rocco") {
		t.Fatalf("top hit does not contain the stored fact: %+v", hits[0])
	}
	t.Logf("top hit similarity=%.3f", hits[0].Similarity)
}

func TestMemoryContextBlock(t *testing.T) {
	if got := memoryContextBlock(nil); got != "" {
		t.Fatalf("expected empty block for no hits, got %q", got)
	}

	block := memoryContextBlock([]memoryHit{{Content: "El usuario prefiere espanol casual"}})
	if block == "" {
		t.Fatal("expected non-empty block")
	}
	if want := "- El usuario prefiere espanol casual\n"; !strings.Contains(block, want) {
		t.Fatalf("block missing hit line: %q", block)
	}
}
