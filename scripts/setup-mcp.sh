#!/bin/bash
# scripts/setup-mcp.sh
# Script para configurar MCP servers en Claude Code
# Only adds servers that don't already exist (safe to re-run)

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() { echo -e "${GREEN}[✓]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[!]${NC} $1"; }
log_error() { echo -e "${RED}[✗]${NC} $1"; }
log_step() { echo -e "${BLUE}[→]${NC} $1"; }

echo ""
echo "=========================================="
echo "   Claude Code MCP Setup Script"
echo "=========================================="
echo ""

# Check if claude is installed
if ! command -v claude &> /dev/null; then
    log_error "Claude Code not installed. Install first:"
    echo "  npm install -g @anthropic-ai/claude-code"
    exit 1
fi

log_info "Claude Code detected: $(claude --version)"
echo ""

# Get list of existing MCP servers
EXISTING_SERVERS=$(claude mcp list 2>/dev/null || echo "")

# Helper: only add if not already configured
add_mcp_if_missing() {
    local name="$1"
    shift
    if echo "$EXISTING_SERVERS" | grep -q "$name"; then
        log_warn "$name already configured, skipping"
    else
        log_step "Adding: $name"
        claude mcp add "$@" 2>/dev/null && \
            log_info "$name added" || \
            log_warn "Could not add $name (may need manual setup)"
    fi
}

log_step "Configuring MCP servers (skipping existing ones)..."
echo ""

# Memory - Local persistent memory
add_mcp_if_missing "memory" memory -- npx -y @modelcontextprotocol/server-memory

# Context7 - Documentation
add_mcp_if_missing "context7" --transport http context7 https://mcp.context7.com/mcp

# Notion - Workspace
add_mcp_if_missing "notion" --transport http notion https://mcp.notion.com/mcp

# Supabase - Database
add_mcp_if_missing "supabase" --transport http supabase https://mcp.supabase.com/mcp

# Obsidian Vault (optional)
echo ""
if echo "$EXISTING_SERVERS" | grep -q "obsidian-vault"; then
    log_warn "obsidian-vault already configured, skipping"
else
    log_step "Obsidian Vault MCP (optional)"
    echo ""
    read -r -p "Do you have an Obsidian vault to connect? (y/N) " setup_obsidian

    if [[ "$setup_obsidian" == "y" || "$setup_obsidian" == "Y" ]]; then
        read -r -p "Enter vault path (e.g., ~/brain-vault): " vault_path
        vault_path="${vault_path/#\~/$HOME}"

        if [[ -d "$vault_path" ]]; then
            add_mcp_if_missing "obsidian-vault" obsidian-vault -- npx -y @bitbonsai/mcpvault@latest "$vault_path"
        else
            log_warn "Path not found: $vault_path (skipping)"
        fi
    fi
fi

echo ""
log_step "Current MCP servers:"
echo ""
claude mcp list

echo ""
echo "=========================================="
echo "   Setup complete!"
echo "=========================================="
echo ""
log_info "To authenticate Notion and Supabase:"
echo "  1. Run: claude"
echo "  2. Inside Claude, type: /mcp"
echo "  3. Follow the OAuth instructions"
echo ""
