#!/usr/bin/env bash
# Build the jarvis-memory MCP server into ./bin/.
# Pure-Go, no CGO. Produces a single static binary.

set -euo pipefail

repo_root="$(cd "$(dirname "$0")/.." && pwd)"
src_dir="${repo_root}/mcp-servers/jarvis-memory"
bin_dir="${repo_root}/bin"

if ! command -v go >/dev/null 2>&1; then
  echo "error: go toolchain not found. Install Go 1.25+ from https://go.dev/dl/" >&2
  exit 1
fi

mkdir -p "${bin_dir}"

bin_name="jarvis-memory"
case "${OS:-}" in
  Windows_NT) bin_name="jarvis-memory.exe" ;;
esac

echo "==> Building ${bin_name}"
pushd "${src_dir}" >/dev/null
go mod tidy
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "${bin_dir}/${bin_name}" .
popd >/dev/null

echo "==> Built ${bin_dir}/${bin_name} ($(du -h "${bin_dir}/${bin_name}" | cut -f1))"
echo
echo "Next steps:"
echo "  1. Install Ollama + pull the embedding model:"
echo "       ollama pull nomic-embed-text"
echo "  2. Enable the server in .mcp.json (flip \"disabled\" to false)"
echo "  3. Restart Claude Code and run /mcp to confirm jarvis-memory is listed"
