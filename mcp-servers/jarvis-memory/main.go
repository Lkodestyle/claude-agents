package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/philippgille/chromem-go"
)

const (
	serverName    = "jarvis-memory"
	serverVersion = "0.1.0"
)

func main() {
	log.SetOutput(os.Stderr) // stdio transport owns stdout; never log there.

	dataDir := resolveDataDir()
	docEmbed, queryEmbed, embedLabel := resolveEmbedFunc()

	store, err := OpenStore(dataDir, docEmbed, queryEmbed)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}

	srv := &Server{store: store}
	server := mcp.NewServer(&mcp.Implementation{
		Name:    serverName,
		Version: serverVersion,
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "remember",
		Description: "Persist a memory (scope: user | feedback | project | reference | session) into Jarvis' long-term vector store.",
	}, srv.handleRemember)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "recall",
		Description: "Semantic search across stored memories. Optional scope filter.",
	}, srv.handleRecall)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "forget",
		Description: "Delete a memory by id.",
	}, srv.handleForget)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "reflect",
		Description: "Sample top-K memories across scopes for a topic so the model can synthesize a situational reflection.",
	}, srv.handleReflect)

	log.Printf("jarvis-memory v%s starting (data=%s, embed=%s)", serverVersion, dataDir, embedLabel)
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}

func resolveDataDir() string {
	if env := os.Getenv("JARVIS_DATA_DIR"); env != "" {
		return env
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".jarvis", "memory")
	}
	return filepath.Join(home, ".jarvis", "memory")
}

// resolveEmbedFunc picks a provider in priority order:
//   1. Voyage AI  — if VOYAGE_API_KEY is set (asymmetric doc/query embeddings)
//   2. OpenAI     — if JARVIS_USE_OPENAI=true and OPENAI_API_KEY is set
//   3. Ollama     — default, local
//
// Returns (docEmbed, queryEmbed, label). queryEmbed is nil unless the provider
// benefits from asymmetric retrieval.
func resolveEmbedFunc() (chromem.EmbeddingFunc, chromem.EmbeddingFunc, string) {
	if key := os.Getenv("VOYAGE_API_KEY"); key != "" {
		model := os.Getenv("JARVIS_VOYAGE_MODEL")
		c := NewVoyageClient(key, model, os.Getenv("JARVIS_VOYAGE_URL"))
		return c.DocumentFunc(), c.QueryFunc(), "voyage:" + c.model
	}

	if os.Getenv("JARVIS_USE_OPENAI") == "true" && os.Getenv("OPENAI_API_KEY") != "" {
		return chromem.NewEmbeddingFuncOpenAI(
			os.Getenv("OPENAI_API_KEY"),
			chromem.EmbeddingModelOpenAI3Small,
		), nil, "openai:text-embedding-3-small"
	}

	model := os.Getenv("JARVIS_EMBED_MODEL")
	if model == "" {
		model = "nomic-embed-text"
	}
	baseURL := os.Getenv("JARVIS_OLLAMA_URL") // empty = default http://localhost:11434/api
	return chromem.NewEmbeddingFuncOllama(model, baseURL), nil, "ollama:" + model
}
