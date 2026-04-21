#!/usr/bin/env bash
# Smoke test for jarvis-memory MCP server.
#
# Exercises the full flow against the chosen embedding provider:
#   initialize → remember x3 (different scopes) → recall x2 → reflect
#
# Requires an embedding provider to be configured via env. Priority:
#   1. VOYAGE_API_KEY        (preferred)
#   2. JARVIS_USE_OPENAI=true + OPENAI_API_KEY
#   3. Ollama running locally with the embedding model pulled
#
# Loads repo_root/.env if present. .env is gitignored. NEVER commit it.

set -euo pipefail

repo_root="$(cd "$(dirname "$0")/.." && pwd)"

bin="${repo_root}/bin/jarvis-memory"
[ -f "${bin}.exe" ] && bin="${bin}.exe"

if [ -f "${repo_root}/.env" ]; then
  set -a
  # shellcheck disable=SC1091
  source "${repo_root}/.env"
  set +a
fi

provider="ollama (default)"
if [ -n "${VOYAGE_API_KEY:-}" ]; then
  provider="voyage (${JARVIS_VOYAGE_MODEL:-voyage-3.5-lite})"
elif [ "${JARVIS_USE_OPENAI:-}" = "true" ] && [ -n "${OPENAI_API_KEY:-}" ]; then
  provider="openai"
fi

if [ ! -x "${bin}" ]; then
  echo "error: binary not found at ${bin}" >&2
  echo "       run ./scripts/build-jarvis-memory.sh first" >&2
  exit 1
fi

export JARVIS_DATA_DIR="${repo_root}/.jarvis-smoke"
rm -rf "${JARVIS_DATA_DIR}"

echo "==> jarvis-memory smoke test"
echo "    binary:   ${bin}"
echo "    provider: ${provider}"
echo "    data:     ${JARVIS_DATA_DIR}"
echo

# We run the test in two phases because the Go MCP SDK dispatches requests
# concurrently: if we stream writes and reads in the same session, the faster
# recalls can race ahead of the slower remembers and see an empty collection.
# A real MCP client (Claude Code) waits for responses, so this is a test-only
# concern. Splitting phases also validates on-disk persistence.

write_init='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"smoke","version":"0.0.1"}}}
{"jsonrpc":"2.0","method":"notifications/initialized"}'

writes='{"jsonrpc":"2.0","id":10,"method":"tools/call","params":{"name":"remember","arguments":{"content":"El usuario prefiere respuestas en espanol casual.","scope":"user"}}}
{"jsonrpc":"2.0","id":11,"method":"tools/call","params":{"name":"remember","arguments":{"content":"Jarvis usa Voyage AI para embeddings con asymmetric retrieval (document/query).","scope":"project","tags":["jarvis","embeddings"]}}}
{"jsonrpc":"2.0","id":12,"method":"tools/call","params":{"name":"remember","arguments":{"content":"Nunca pegar API keys en el chat. Usar variables de entorno o archivo .env gitignored.","scope":"feedback"}}}'

reads='{"jsonrpc":"2.0","id":20,"method":"tools/call","params":{"name":"recall","arguments":{"query":"en que idioma tengo que hablar?","k":2}}}
{"jsonrpc":"2.0","id":21,"method":"tools/call","params":{"name":"recall","arguments":{"query":"que vector database usa el proyecto","k":2,"scope":"project"}}}
{"jsonrpc":"2.0","id":22,"method":"tools/call","params":{"name":"reflect","arguments":{"topic":"practicas de seguridad con credenciales","k":2}}}'

stdout1=$(mktemp); stderr1=$(mktemp)
stdout2=$(mktemp); stderr2=$(mktemp)
trap 'rm -f "${stdout1}" "${stderr1}" "${stdout2}" "${stderr2}"' EXIT

echo "---- phase 1: remember x3 ----"
(printf '%s\n%s\n' "${write_init}" "${writes}"; sleep 8) | "${bin}" >"${stdout1}" 2>"${stderr1}" || true
if command -v jq >/dev/null 2>&1; then
  while IFS= read -r line; do
    [ -z "${line}" ] && continue
    echo "${line}" | jq -c 'select(.id != null) | {id, error: (.error // null), result: (.result.structuredContent // .result)}'
  done <"${stdout1}"
else
  cat "${stdout1}"
fi
echo

echo "---- phase 2: recall x2 + reflect ----"
(printf '%s\n%s\n' "${write_init}" "${reads}"; sleep 8) | "${bin}" >"${stdout2}" 2>"${stderr2}" || true
if command -v jq >/dev/null 2>&1; then
  while IFS= read -r line; do
    [ -z "${line}" ] && continue
    echo "${line}" | jq '.result.structuredContent // .error // .result | select(. != null)'
  done <"${stdout2}"
else
  cat "${stdout2}"
fi
echo

echo "---- server stderr (phase 2) ----"
cat "${stderr2}"
echo

rm -rf "${JARVIS_DATA_DIR}"
echo "==> Done. Data dir cleaned."
