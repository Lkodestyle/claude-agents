#!/bin/bash
# scripts/setup-obsidian.sh
# Setup del vault de Obsidian para usar con Claude Code via Local REST API plugin
#
# Uso:
#   ./scripts/setup-obsidian.sh                    # Setup interactivo
#   ./scripts/setup-obsidian.sh /path/to/vault     # Solo crear estructura del vault

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'
BOLD='\033[1m'

log_info() { echo -e "${GREEN}[✓]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[!]${NC} $1"; }
log_error() { echo -e "${RED}[✗]${NC} $1"; }
log_step() { echo -e "${BLUE}[→]${NC} $1"; }

echo ""
echo -e "${BOLD}${CYAN}╔══════════════════════════════════════════╗${NC}"
echo -e "${BOLD}${CYAN}║   Obsidian Vault Setup                   ║${NC}"
echo -e "${BOLD}${CYAN}╚══════════════════════════════════════════╝${NC}"
echo ""

# -- STEP 1: Plugin & API key --
echo -e "${BOLD}Step 1 — Install the Local REST API plugin in Obsidian${NC}"
echo ""
echo "  1. Open Obsidian"
echo "  2. Settings → Community Plugins → Browse"
echo "  3. Search: Local REST API → Install & Enable"
echo "  4. Settings → Local REST API → enable 'Enable HTTP server'"
echo "  5. Copy the API key shown on that screen"
echo ""
read -r -p "Paste your API key here (or press Enter to skip): " API_KEY

if [[ -n "$API_KEY" ]]; then
    # Test connectivity
    log_step "Testing connection to Obsidian Local REST API..."
    if command -v curl &> /dev/null; then
        HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 5 \
            -H "Authorization: Bearer $API_KEY" \
            "http://127.0.0.1:27123/" 2>/dev/null || echo "000")
        if [[ "$HTTP_CODE" == "200" || "$HTTP_CODE" == "204" ]]; then
            log_info "Connected to Obsidian Local REST API"
        else
            log_warn "Could not connect (HTTP $HTTP_CODE). Make sure Obsidian is open and HTTP server is enabled."
        fi
    fi

    # Offer to persist in shell config
    echo ""
    echo -e "${BOLD}Add OBSIDIAN_API_KEY to your shell config?${NC}"
    SHELL_RC=""
    if [[ -f "$HOME/.zshrc" ]]; then
        SHELL_RC="$HOME/.zshrc"
    elif [[ -f "$HOME/.bashrc" ]]; then
        SHELL_RC="$HOME/.bashrc"
    fi

    if [[ -n "$SHELL_RC" ]]; then
        read -r -p "Add to $SHELL_RC? (y/N) " add_to_shell
        if [[ "$add_to_shell" == "y" || "$add_to_shell" == "Y" ]]; then
            if grep -q "OBSIDIAN_API_KEY" "$SHELL_RC" 2>/dev/null; then
                log_warn "OBSIDIAN_API_KEY already in $SHELL_RC, skipping"
            else
                echo "" >> "$SHELL_RC"
                echo "# Obsidian Local REST API" >> "$SHELL_RC"
                echo "export OBSIDIAN_API_KEY=\"$API_KEY\"" >> "$SHELL_RC"
                log_info "Added to $SHELL_RC — run 'source $SHELL_RC' to apply"
            fi
        fi
    fi
fi

# -- STEP 2: Vault structure (optional) --
echo ""
echo -e "${BOLD}Step 2 — Create vault folder structure (optional)${NC}"
echo "  Creates inbox, projects, areas, resources, templates, etc."
echo ""

VAULT_PATH="${1:-}"

if [[ -z "$VAULT_PATH" ]]; then
    read -r -p "Enter vault path to create structure (or press Enter to skip): " VAULT_PATH
fi

if [[ -z "$VAULT_PATH" ]]; then
    log_info "Skipping vault structure creation"
else
    VAULT_PATH="${VAULT_PATH/#\~/$HOME}"

    if [[ ! -d "$VAULT_PATH" ]]; then
        log_error "Directory not found: $VAULT_PATH"
        exit 1
    fi

    if [[ ! -d "$VAULT_PATH/.obsidian" ]]; then
        log_warn "No .obsidian directory found — may not be an Obsidian vault"
        read -r -p "Continue anyway? (y/N) " confirm
        [[ "$confirm" != "y" && "$confirm" != "Y" ]] && exit 0
    fi

    log_step "Creating vault structure in $VAULT_PATH..."

    mkdir -p "$VAULT_PATH"/{00-inbox,01-daily,02-projects,03-areas,04-resources,05-archive,06-meetings,07-prompts,08-snippets,_templates,_assets}

    create_readme() {
        local filepath="$VAULT_PATH/$1/README.md"
        [[ ! -f "$filepath" ]] && echo "$2" > "$filepath"
    }

    create_readme "00-inbox" "# Inbox

Quick capture. Everything enters here first. Process weekly."

    create_readme "01-daily" "# Daily Notes

One note per day. Name format: \`YYYY-MM-DD.md\`"

    create_readme "02-projects" "# Projects

One subfolder per active project. Move to \`05-archive/\` when done."

    create_readme "03-areas" "# Areas

Ongoing responsibilities without an end date (career, health, finances)."

    create_readme "04-resources" "# Resources

Processed knowledge: tutorials, references, how-tos."

    create_readme "05-archive" "# Archive

Completed projects and inactive notes. Move here instead of deleting."

    create_readme "06-meetings" "# Meetings

Meeting notes and action items. Format: \`YYYY-MM-DD-topic.md\`"

    create_readme "07-prompts" "# Prompts

Personal library of prompts that work well."

    create_readme "08-snippets" "# Snippets

Reusable code fragments organized by language."

    log_info "Folder structure created"
    log_info "Templates folder ready: $VAULT_PATH/_templates/"
fi

# -- Summary --
echo ""
echo -e "${BOLD}${CYAN}╔══════════════════════════════════════════╗${NC}"
echo -e "${BOLD}${CYAN}║   Done!                                  ║${NC}"
echo -e "${BOLD}${CYAN}╚══════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BOLD}How the MCP connection works:${NC}"
echo "  - Plugin: Local REST API (running inside Obsidian)"
echo "  - Endpoint: http://127.0.0.1:27123/mcp/"
echo "  - Auth: Bearer token (OBSIDIAN_API_KEY)"
echo "  - Obsidian must be open for the server to run"
echo ""
echo -e "${BOLD}Claude Code config (.mcp.json):${NC}"
echo '  "obsidian": {'
echo '    "type": "http",'
echo '    "url": "http://127.0.0.1:27123/mcp/",'
echo '    "headers": { "Authorization": "Bearer YOUR_KEY" }'
echo '  }'
echo ""
log_info "Restart Claude Code to activate the obsidian MCP server"
echo ""
