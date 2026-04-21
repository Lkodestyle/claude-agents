# jarvis-memory

Pure-Go MCP server that gives Claude Code (and any MCP-compatible client) a
portable, local-first **semantic memory** layer. Core of the optional Jarvis
mode: a single binary, no external DB, no CGO.

## Tools

| Tool | Purpose |
|------|---------|
| `remember` | Persist a memory with scope (user, feedback, project, reference, session), tags and optional source |
| `recall` | Semantic search (top-K) with optional scope filter |
| `forget` | Delete a memory by id |
| `reflect` | Sample top-K memories across scopes for a topic so Claude can synthesize context |

## Stack

- **MCP SDK:** [`modelcontextprotocol/go-sdk`](https://github.com/modelcontextprotocol/go-sdk), stdio transport
- **Vector DB:** [`chromem-go`](https://github.com/philippgille/chromem-go) — embedded, persistent, zero deps
- **Embeddings:** pluggable. **Voyage AI** (recommended by Anthropic) with asymmetric doc/query embeddings, **OpenAI**, or **Ollama** (local).

## Embedding providers

Priority order at startup (first match wins):

1. **Voyage AI** — if `VOYAGE_API_KEY` is set. Uses `input_type=document` for stores and `input_type=query` for searches, which measurably improves recall quality.
2. **OpenAI** — if `JARVIS_USE_OPENAI=true` and `OPENAI_API_KEY` is set.
3. **Ollama** — default fallback, fully local.

## Environment

| Var | Default | Notes |
|-----|---------|-------|
| `JARVIS_DATA_DIR` | `~/.jarvis/memory` | Where chromem persists collection files |
| `VOYAGE_API_KEY` | *(unset)* | Enables Voyage provider. Key format: `pa-...` |
| `JARVIS_VOYAGE_MODEL` | `voyage-3.5-lite` | Any Voyage embedding model |
| `JARVIS_VOYAGE_URL` | `https://api.voyageai.com/v1` | Override for proxies / self-hosted |
| `JARVIS_USE_OPENAI` | `false` | Set to `true` + `OPENAI_API_KEY` to use OpenAI |
| `JARVIS_EMBED_MODEL` | `nomic-embed-text` | Ollama embedding model |
| `JARVIS_OLLAMA_URL` | `http://localhost:11434/api` | Override if Ollama runs elsewhere |

## Build

One-liner from the repo root:

```bash
./scripts/build-jarvis-memory.sh
```

Or manually:

```bash
cd mcp-servers/jarvis-memory
go mod tidy
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o ../../bin/jarvis-memory .
```

Cross-compile (no CGO needed):

```bash
GOOS=linux   GOARCH=amd64 go build -o bin/jarvis-memory-linux-amd64 .
GOOS=darwin  GOARCH=arm64 go build -o bin/jarvis-memory-darwin-arm64 .
GOOS=windows GOARCH=amd64 go build -o bin/jarvis-memory-windows-amd64.exe .
```

## Wire into Claude Code

The repo's `.mcp.json` already has a `jarvis-memory` entry, disabled by default. To activate:

1. Build the binary (see above)
2. Set `VOYAGE_API_KEY` in your shell or `.env`
3. Flip `"disabled": false` in `.mcp.json`
4. Restart Claude Code and verify with `/mcp`

## Prerequisites

One of:
- A Voyage API key from [dash.voyageai.com](https://dash.voyageai.com/api-keys) (free tier: 200M tokens/mo).
- An OpenAI API key.
- [Ollama](https://ollama.com) running locally with `ollama pull nomic-embed-text`.

## Roadmap

- **Phase 1 (this)** — pure-Go local persistence, 4 core tools, 3 embedding providers.
- **Phase 2** — Turso/libSQL embedded replica for metadata sync across machines (accepts CGO).
- **Phase 3** — optional Qdrant backend for true remote agent deployments; swap `Store` implementation.
- **Phase 4** — `/jarvis on|off` toggle + personality layer via hook.
