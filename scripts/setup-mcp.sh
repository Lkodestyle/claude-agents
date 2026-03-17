#!/bin/bash
# scripts/setup-mcp.sh
# Script para configurar MCP servers en Claude Code

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step() { echo -e "${BLUE}[STEP]${NC} $1"; }

echo ""
echo "=========================================="
echo "   Claude Code MCP Setup Script"
echo "=========================================="
echo ""

# Check if claude is installed
if ! command -v claude &> /dev/null; then
    log_error "Claude Code no está instalado. Instálalo primero:"
    echo "  npm install -g @anthropic-ai/claude-code"
    exit 1
fi

log_info "Claude Code detectado: $(claude --version)"
echo ""

# Remove existing MCP servers
log_step "Limpiando MCP servers existentes..."
claude mcp remove memory 2>/dev/null || true
claude mcp remove context7 2>/dev/null || true
claude mcp remove notion 2>/dev/null || true
claude mcp remove supabase 2>/dev/null || true

echo ""
log_step "Agregando MCP servers..."

# Memory - Local persistent memory
log_info "Agregando: memory (memoria persistente local)"
claude mcp add memory -- npx -y @whenmoon-afk/memory-mcp

# Context7 - Documentation
log_info "Agregando: context7 (documentación actualizada)"
claude mcp add --transport http context7 https://mcp.context7.com/mcp

# Notion - Workspace
log_info "Agregando: notion (workspace de Notion)"
claude mcp add --transport http notion https://mcp.notion.com/mcp

# Supabase - Database
log_info "Agregando: supabase (base de datos)"
claude mcp add --transport http supabase https://mcp.supabase.com/mcp

echo ""

# Obsidian Vault (optional)
echo ""
log_step "Obsidian Vault MCP (optional)"
echo ""
read -r -p "Do you have an Obsidian vault to connect? (y/N) " setup_obsidian

if [[ "$setup_obsidian" == "y" || "$setup_obsidian" == "Y" ]]; then
    read -r -p "Enter vault path (e.g., ~/brain-vault): " vault_path
    vault_path="${vault_path/#\~/$HOME}"

    if [[ -d "$vault_path" ]]; then
        log_info "Agregando: obsidian-vault (knowledge base)"
        claude mcp add obsidian-vault -- npx @bitbonsai/mcpvault@latest "$vault_path"
    else
        log_warn "Path not found: $vault_path (skipping)"
    fi
fi

echo ""
log_step "Verificando instalación..."
echo ""
claude mcp list

echo ""
echo "=========================================="
echo "   Setup completado!"
echo "=========================================="
echo ""
log_info "Para autenticar Notion y Supabase:"
echo "  1. Ejecuta: claude"
echo "  2. Dentro de Claude, escribe: /mcp"
echo "  3. Sigue las instrucciones de OAuth"
echo ""
log_info "Para usar los agentes en un proyecto:"
echo "  Agrega esto al CLAUDE.md de tu proyecto:"
echo ""
echo "  @~/claude-agents/agents/terraform.md"
echo "  @~/claude-agents/agents/aws.md"
echo "  @~/claude-agents/agents/cicd.md"
echo ""
