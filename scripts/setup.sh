#!/bin/bash
# scripts/setup.sh
# Full onboarding: installs agents, configures MCP servers, sets up Obsidian vault,
# and shows recommended plugins. Run once after cloning the repo.
#
# Usage:
#   ./scripts/setup.sh              # Interactive full setup
#   ./scripts/setup.sh --skip-mcp   # Skip MCP server setup
#   ./scripts/setup.sh --skip-obsidian  # Skip Obsidian vault setup

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m'
BOLD='\033[1m'

log_info() { echo -e "${GREEN}[✓]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[!]${NC} $1"; }
log_error() { echo -e "${RED}[✗]${NC} $1"; }
log_step() { echo -e "${BLUE}[→]${NC} $1"; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"

SKIP_MCP=false
SKIP_OBSIDIAN=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-mcp) SKIP_MCP=true; shift ;;
        --skip-obsidian) SKIP_OBSIDIAN=true; shift ;;
        --help|-h)
            echo "Usage: ./scripts/setup.sh [--skip-mcp] [--skip-obsidian]"
            echo ""
            echo "Full onboarding for claude-agents:"
            echo "  1. Install agents, commands, skills, scripts"
            echo "  2. Configure MCP servers (memory, context7, etc.)"
            echo "  3. Setup Obsidian vault with full structure"
            echo "  4. Show recommended plugins"
            exit 0
            ;;
        *) log_error "Unknown option: $1"; exit 1 ;;
    esac
done

echo ""
echo -e "${BOLD}${MAGENTA}╔══════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}${MAGENTA}║   Claude Agents — Full Setup                 ║${NC}"
echo -e "${BOLD}${MAGENTA}╚══════════════════════════════════════════════╝${NC}"
echo ""

# ============================================================================
# STEP 1: Install agents, commands, skills, scripts
# ============================================================================

echo -e "${BOLD}${CYAN}━━━ Step 1/4: Installing agents, commands, and skills ━━━${NC}"
echo ""

echo -e "Installation mode:"
echo -e "  ${GREEN}1)${NC} Symlink (recommended — auto-updates with git pull)"
echo -e "  ${GREEN}2)${NC} Copy (independent of repo)"
read -r -p "Choose [1]: " install_mode
install_mode="${install_mode:-1}"

install_flag="--symlink"
[[ "$install_mode" == "2" ]] && install_flag="--copy"

bash "$SCRIPT_DIR/claude-agents-cli.sh" install "$install_flag"

# ============================================================================
# STEP 2: Configure MCP servers
# ============================================================================

if [[ "$SKIP_MCP" == false ]]; then
    echo ""
    echo -e "${BOLD}${CYAN}━━━ Step 2/4: Configuring MCP servers ━━━${NC}"
    echo ""

    if command -v claude &> /dev/null; then
        echo -e "Setup MCP servers via Claude CLI (memory, context7, notion, supabase)?"
        read -r -p "(y/N) " setup_mcp

        if [[ "$setup_mcp" == "y" || "$setup_mcp" == "Y" ]]; then
            bash "$SCRIPT_DIR/setup-mcp.sh"
        else
            log_info "Skipped MCP CLI setup. You can run it later: ./scripts/setup-mcp.sh"
            log_info "MCP config from .mcp.json will still be used by Claude Code."
        fi
    else
        log_warn "Claude CLI not found. Skipping MCP CLI setup."
        log_info "The .mcp.json in the repo will be used automatically."
        log_info "Install Claude Code: npm install -g @anthropic-ai/claude-code"
    fi
else
    echo ""
    log_info "Skipped MCP setup (--skip-mcp)"
fi

# ============================================================================
# STEP 3: Obsidian vault setup
# ============================================================================

if [[ "$SKIP_OBSIDIAN" == false ]]; then
    echo ""
    echo -e "${BOLD}${CYAN}━━━ Step 3/4: Obsidian Vault setup ━━━${NC}"
    echo ""

    echo -e "Do you want to connect an Obsidian vault as a knowledge base?"
    echo -e "  This configures the obsidian-vault MCP server and optionally"
    echo -e "  creates a complete vault structure (folders, templates, guides)."
    read -r -p "(y/N) " setup_obsidian

    if [[ "$setup_obsidian" == "y" || "$setup_obsidian" == "Y" ]]; then
        bash "$SCRIPT_DIR/setup-obsidian.sh"
    else
        log_info "Skipped Obsidian setup. You can run it later: ./scripts/setup-obsidian.sh"
    fi
else
    echo ""
    log_info "Skipped Obsidian setup (--skip-obsidian)"
fi

# ============================================================================
# STEP 4: Recommended plugins
# ============================================================================

echo ""
echo -e "${BOLD}${CYAN}━━━ Step 4/4: Recommended plugins ━━━${NC}"
echo ""

echo -e "${BOLD}Essential plugins to install in Claude Code:${NC}"
echo ""
printf "  ${GREEN}%-20s${NC} %s\n" "superpowers" "Brainstorming, TDD, debugging, plans, code review"
printf "  ${GREEN}%-20s${NC} %s\n" "skill-creator" "Create and test custom skills"
printf "  ${GREEN}%-20s${NC} %s\n" "context7" "Up-to-date library documentation (already in .mcp.json)"
echo ""

echo -e "Enable the superpowers plugin now?"
read -r -p "(y/N) " enable_superpowers

if [[ "$enable_superpowers" == "y" || "$enable_superpowers" == "Y" ]]; then
    if command -v claude &> /dev/null; then
        claude config set enabledPlugins.superpowers@claude-plugins-official true 2>/dev/null && \
            log_info "Superpowers plugin enabled" || \
            log_warn "Could not enable plugin. Run manually: claude config set enabledPlugins.superpowers@claude-plugins-official true"
    else
        log_warn "Claude CLI not found. Run this after installing Claude Code:"
        echo "  claude config set enabledPlugins.superpowers@claude-plugins-official true"
    fi
else
    log_info "You can enable it later: claude config set enabledPlugins.superpowers@claude-plugins-official true"
fi

echo ""
log_info "For all recommended plugins: ./scripts/claude-agents-cli.sh plugins"

# ============================================================================
# SUMMARY
# ============================================================================

echo ""
echo -e "${BOLD}${MAGENTA}╔══════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}${MAGENTA}║   Setup Complete!                            ║${NC}"
echo -e "${BOLD}${MAGENTA}╚══════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BOLD}What was installed:${NC}"
echo "  - 15 specialized agents"
echo "  - 8 slash commands (/commit, /pr, /review, /test, /explain, /refactor, /debug, /doc)"
echo "  - 5 workflow skills (deploy-pipeline, observability, IaC, security, scaffold)"
echo "  - Cognitive scripts (context router, pool coordinator, memory manager)"
echo ""
echo -e "${BOLD}Next steps:${NC}"
echo "  1. Open a terminal and run: claude"
echo "  2. The agents activate automatically based on your conversation"
echo "  3. Use /commit, /pr, /review, etc. for common workflows"
echo "  4. Check status anytime: ./scripts/claude-agents-cli.sh status"
echo ""
