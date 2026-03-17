#!/bin/bash
# scripts/setup-obsidian.sh
# Script para configurar el MCP server de Obsidian Vault en Claude Code
#
# Uso:
#   ./scripts/setup-obsidian.sh                    # Setup interactivo
#   ./scripts/setup-obsidian.sh /path/to/vault     # Setup con path directo

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color
BOLD='\033[1m'

log_info() { echo -e "${GREEN}[✓]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[!]${NC} $1"; }
log_error() { echo -e "${RED}[✗]${NC} $1"; }
log_step() { echo -e "${BLUE}[→]${NC} $1"; }

echo ""
echo -e "${BOLD}${CYAN}╔══════════════════════════════════════════╗${NC}"
echo -e "${BOLD}${CYAN}║   Obsidian Vault MCP Setup               ║${NC}"
echo -e "${BOLD}${CYAN}╚══════════════════════════════════════════╝${NC}"
echo ""

# Check prerequisites
if ! command -v npx &> /dev/null; then
    log_error "npx not found. Install Node.js first: https://nodejs.org"
    exit 1
fi

# Get vault path
VAULT_PATH="${1:-}"

if [[ -z "$VAULT_PATH" ]]; then
    echo -e "${BOLD}Enter the path to your Obsidian vault:${NC}"
    echo -e "  Example: ~/brain-vault"
    echo -e "  Example: /mnt/c/Users/you/Documents/my-vault"
    echo ""
    read -r -p "> " VAULT_PATH
fi

# Expand tilde
VAULT_PATH="${VAULT_PATH/#\~/$HOME}"

# Validate vault path
if [[ ! -d "$VAULT_PATH" ]]; then
    log_error "Directory not found: $VAULT_PATH"
    exit 1
fi

if [[ ! -d "$VAULT_PATH/.obsidian" ]]; then
    log_warn "No .obsidian directory found. This may not be an Obsidian vault."
    read -r -p "Continue anyway? (y/N) " confirm
    if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
        log_info "Aborted."
        exit 0
    fi
fi

log_info "Vault found: $VAULT_PATH"

# Determine MCP config location
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"
MCP_CONFIG="$REPO_DIR/.mcp.json"

if [[ ! -f "$MCP_CONFIG" ]]; then
    log_error "MCP config not found: $MCP_CONFIG"
    exit 1
fi

# Update .mcp.json
log_step "Updating MCP configuration..."

python3 -c "
import json
import sys

config_path = '$MCP_CONFIG'
vault_path = '$VAULT_PATH'

with open(config_path, 'r') as f:
    config = json.load(f)

servers = config.get('mcpServers', {})

servers['obsidian-vault'] = {
    'type': 'stdio',
    'command': 'npx',
    'args': ['@bitbonsai/mcpvault@latest', vault_path]
}

config['mcpServers'] = servers

with open(config_path, 'w') as f:
    json.dump(config, f, indent=2)
    f.write('\n')

print('OK')
" || {
    log_error "Failed to update MCP config"
    exit 1
}

log_info "MCP config updated: $MCP_CONFIG"

# Also update global config if requested
echo ""
echo -e "${BOLD}Also install to global config (~/.mcp.json)?${NC}"
read -r -p "(y/N) " install_global

if [[ "$install_global" == "y" || "$install_global" == "Y" ]]; then
    GLOBAL_MCP="$HOME/.mcp.json"

    if [[ -f "$GLOBAL_MCP" ]]; then
        python3 -c "
import json

config_path = '$GLOBAL_MCP'
vault_path = '$VAULT_PATH'

with open(config_path, 'r') as f:
    config = json.load(f)

servers = config.get('mcpServers', {})

servers['obsidian-vault'] = {
    'type': 'stdio',
    'command': 'npx',
    'args': ['@bitbonsai/mcpvault@latest', vault_path]
}

config['mcpServers'] = servers

with open(config_path, 'w') as f:
    json.dump(config, f, indent=2)
    f.write('\n')

print('OK')
"
        log_info "Global MCP config updated: $GLOBAL_MCP"
    else
        python3 -c "
import json

vault_path = '$VAULT_PATH'
config = {
    'mcpServers': {
        'obsidian-vault': {
            'type': 'stdio',
            'command': 'npx',
            'args': ['@bitbonsai/mcpvault@latest', vault_path]
        }
    }
}

with open('$GLOBAL_MCP', 'w') as f:
    json.dump(config, f, indent=2)
    f.write('\n')

print('OK')
"
        log_info "Global MCP config created: $GLOBAL_MCP"
    fi
fi

# Optional: create base vault structure
echo ""
echo -e "${BOLD}Create recommended vault structure for projects?${NC}"
echo -e "  This adds: 02-projects/, _templates/, CLAUDE.md (if they don't exist)"
read -r -p "(y/N) " create_structure

if [[ "$create_structure" == "y" || "$create_structure" == "Y" ]]; then
    log_step "Creating vault structure..."

    mkdir -p "$VAULT_PATH"/{00-inbox,01-daily,02-projects,03-areas,04-resources,05-archive,_templates}

    if [[ ! -f "$VAULT_PATH/CLAUDE.md" ]]; then
        cat > "$VAULT_PATH/CLAUDE.md" << 'CLAUDEEOF'
# CLAUDE.md — Obsidian Vault

> This file is read automatically by Claude Code at session start.

## Vault Structure

```
vault/
├── 00-inbox/        → Quick capture
├── 01-daily/        → Daily notes
├── 02-projects/     → Active projects (one subfolder each)
├── 03-areas/        → Ongoing responsibilities
├── 04-resources/    → Processed knowledge
├── 05-archive/      → Completed projects
└── _templates/      → Note templates
```

## Conventions

- Use `[[wikilinks]]` for internal links
- Use `#tags` for categorization
- Frontmatter YAML required: `date`, `tags`, `status`
- File names in kebab-case

## Behavior

1. Respect folder structure
2. Use wikilinks for references
3. Include frontmatter in new notes
4. Capture in 00-inbox/ if unsure where something goes
5. Do not modify existing notes without confirmation
CLAUDEEOF
        log_info "Created CLAUDE.md in vault"
    fi

    log_info "Vault structure ready"
fi

# Summary
echo ""
echo -e "${BOLD}${CYAN}╔══════════════════════════════════════════╗${NC}"
echo -e "${BOLD}${CYAN}║   Setup Complete!                        ║${NC}"
echo -e "${BOLD}${CYAN}╚══════════════════════════════════════════╝${NC}"
echo ""
log_info "Vault path: $VAULT_PATH"
log_info "MCP server: @bitbonsai/mcpvault@latest"
echo ""
echo -e "${BOLD}Available tools in Claude Code:${NC}"
echo "  - read_note / write_note — Read and write notes"
echo "  - search_notes — Search vault by content"
echo "  - list_directory — Browse vault structure"
echo "  - get_vault_stats — Vault statistics"
echo "  - manage_tags — Tag management"
echo ""
log_info "Restart Claude Code to activate the MCP server"
echo ""
