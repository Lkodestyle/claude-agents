#!/bin/bash
# claude-agents-cli.sh - CLI completo para gestionar Claude DevOps Agents
#
# Uso:
#   ./scripts/claude-agents-cli.sh install [--symlink|--copy]
#   ./scripts/claude-agents-cli.sh sync
#   ./scripts/claude-agents-cli.sh status
#   ./scripts/claude-agents-cli.sh test
#   ./scripts/claude-agents-cli.sh uninstall
#   ./scripts/claude-agents-cli.sh help

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# Logging functions
log_info() { echo -e "${GREEN}[✓]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[!]${NC} $1"; }
log_error() { echo -e "${RED}[✗]${NC} $1"; }
log_step() { echo -e "${BLUE}[→]${NC} $1"; }
log_header() { echo -e "\n${BOLD}${CYAN}$1${NC}\n"; }

# Get paths
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"
CLAUDE_HOME="${CLAUDE_HOME:-$HOME/.claude}"
GLOBAL_AGENTS="$CLAUDE_HOME/agents"
GLOBAL_SCRIPTS="$CLAUDE_HOME/scripts"
VERSION="2.0.0"

# ============================================================================
# HELP
# ============================================================================

show_help() {
    cat << EOF
${BOLD}Claude DevOps Agents CLI v${VERSION}${NC}

${BOLD}USAGE:${NC}
    claude-agents-cli.sh <command> [options]

${BOLD}COMMANDS:${NC}
    ${GREEN}install${NC}     Install agents and cognitive scripts
                --symlink   Use symlinks (default, auto-updates with git pull)
                --copy      Copy files (independent of repo)
                --global    Install to ~/.claude (default)
                --local     Install to current project .claude/

    ${GREEN}sync${NC}        Sync configuration from repo to installed location
                Useful after git pull to update scripts

    ${GREEN}status${NC}      Show installation status and agent list

    ${GREEN}test${NC}        Test the context router and pool coordinator
                Validates Python scripts work correctly

    ${GREEN}proxy${NC}       Start mcp-proxy server (requires Go installation)
                --background  Run in background

    ${GREEN}uninstall${NC}   Remove installed agents and scripts

    ${GREEN}help${NC}        Show this help message

${BOLD}EXAMPLES:${NC}
    ./scripts/claude-agents-cli.sh install              # Install with symlinks
    ./scripts/claude-agents-cli.sh install --copy       # Copy files
    ./scripts/claude-agents-cli.sh install --local      # Install to project
    ./scripts/claude-agents-cli.sh sync                 # Sync after git pull
    ./scripts/claude-agents-cli.sh status               # Check status
    ./scripts/claude-agents-cli.sh test                 # Test scripts

${BOLD}ENVIRONMENT VARIABLES:${NC}
    CLAUDE_HOME         Override ~/.claude location
    CLAUDE_INSTANCE     Instance ID for pool coordination (A, B, C, etc.)
    MCP_PROXY_TOKEN     Token for mcp-proxy authentication
    NOTION_TOKEN        Notion integration token
    GITHUB_TOKEN        GitHub personal access token

${BOLD}MORE INFO:${NC}
    Repository: https://github.com/your-org/claude-agents
    Docs:       See CLAUDE.md in the repository

EOF
}

# ============================================================================
# UTILITIES
# ============================================================================

check_python() {
    if command -v python3 &> /dev/null; then
        log_info "Python3 found: $(python3 --version)"
        return 0
    else
        log_error "Python3 not found. Please install Python 3.8+"
        return 1
    fi
}

check_repo() {
    if [[ ! -d "$REPO_DIR/.claude/agents" ]]; then
        log_error "Invalid repository structure. Missing .claude/agents"
        exit 1
    fi
}

backup_if_exists() {
    local target="$1"
    if [[ -e "$target" && ! -L "$target" ]]; then
        local backup="${target}.backup.$(date +%Y%m%d_%H%M%S)"
        log_warn "Backing up existing: $target -> $backup"
        mv "$target" "$backup"
    elif [[ -L "$target" ]]; then
        log_step "Removing existing symlink: $target"
        rm "$target"
    fi
}

# ============================================================================
# INSTALL
# ============================================================================

cmd_install() {
    local mode="symlink"
    local target="global"

    while [[ $# -gt 0 ]]; do
        case $1 in
            --symlink) mode="symlink"; shift ;;
            --copy) mode="copy"; shift ;;
            --global) target="global"; shift ;;
            --local) target="local"; shift ;;
            *) log_error "Unknown option: $1"; exit 1 ;;
        esac
    done

    log_header "Installing Claude DevOps Agents"
    check_repo
    check_python

    # Determine target directory
    local install_dir
    if [[ "$target" == "local" ]]; then
        install_dir="$(pwd)/.claude"
        log_info "Installing to project: $install_dir"
    else
        install_dir="$CLAUDE_HOME"
        log_info "Installing globally: $install_dir"
    fi

    # Create directories
    mkdir -p "$install_dir"/{agents,scripts,pool}

    # Install agents
    log_step "Installing agents..."
    local agents_target="$install_dir/agents"
    backup_if_exists "$agents_target"

    if [[ "$mode" == "symlink" ]]; then
        ln -s "$REPO_DIR/.claude/agents" "$agents_target"
        log_info "Agents symlinked (auto-updates with git pull)"
    else
        cp -r "$REPO_DIR/.claude/agents" "$agents_target"
        log_info "Agents copied"
    fi

    # Install scripts
    log_step "Installing cognitive scripts..."
    local scripts_target="$install_dir/scripts"
    backup_if_exists "$scripts_target"

    if [[ "$mode" == "symlink" ]]; then
        ln -s "$REPO_DIR/.claude/scripts" "$scripts_target"
        log_info "Scripts symlinked"
    else
        cp -r "$REPO_DIR/.claude/scripts" "$scripts_target"
        log_info "Scripts copied"
    fi

    # Install config files
    log_step "Installing configuration files..."

    if [[ ! -f "$install_dir/keywords.json" ]]; then
        cp "$REPO_DIR/.claude/keywords.json" "$install_dir/keywords.json"
        log_info "Keywords config installed"
    else
        log_warn "keywords.json already exists, skipping"
    fi

    if [[ ! -f "$install_dir/settings.json" ]]; then
        cp "$REPO_DIR/.claude/settings.json" "$install_dir/settings.json"
        log_info "Settings config installed"
    else
        log_warn "settings.json already exists, skipping"
    fi

    # Install MCP config if global
    if [[ "$target" == "global" && ! -f "$HOME/.mcp.json" ]]; then
        cp "$REPO_DIR/.mcp.json" "$HOME/.mcp.json"
        log_info "MCP config installed to ~/.mcp.json"
    fi

    # Summary
    log_header "Installation Complete!"
    cmd_status_brief "$install_dir"
}

# ============================================================================
# SYNC
# ============================================================================

cmd_sync() {
    log_header "Syncing Claude Agents Configuration"
    check_repo

    # Check if using symlinks (no sync needed)
    if [[ -L "$GLOBAL_AGENTS" ]]; then
        log_info "Agents using symlink - auto-synced via git"
    else
        log_step "Syncing agents..."
        rsync -av --delete "$REPO_DIR/.claude/agents/" "$GLOBAL_AGENTS/"
        log_info "Agents synced"
    fi

    if [[ -L "$GLOBAL_SCRIPTS" ]]; then
        log_info "Scripts using symlink - auto-synced via git"
    else
        log_step "Syncing scripts..."
        rsync -av --delete "$REPO_DIR/.claude/scripts/" "$GLOBAL_SCRIPTS/"
        log_info "Scripts synced"
    fi

    # Always sync keywords.json (may have been updated)
    log_step "Syncing keywords.json..."
    cp "$REPO_DIR/.claude/keywords.json" "$CLAUDE_HOME/keywords.json"
    log_info "Keywords synced"

    log_header "Sync Complete!"
}

# ============================================================================
# STATUS
# ============================================================================

cmd_status() {
    log_header "Claude Agents Status"

    # Check installation
    echo -e "${BOLD}Installation:${NC}"
    if [[ -d "$CLAUDE_HOME" ]]; then
        log_info "Claude home: $CLAUDE_HOME"
    else
        log_error "Claude home not found: $CLAUDE_HOME"
    fi

    if [[ -L "$GLOBAL_AGENTS" ]]; then
        log_info "Agents: symlinked -> $(readlink "$GLOBAL_AGENTS")"
    elif [[ -d "$GLOBAL_AGENTS" ]]; then
        log_info "Agents: copied"
    else
        log_error "Agents: not installed"
    fi

    if [[ -L "$GLOBAL_SCRIPTS" ]]; then
        log_info "Scripts: symlinked -> $(readlink "$GLOBAL_SCRIPTS")"
    elif [[ -d "$GLOBAL_SCRIPTS" ]]; then
        log_info "Scripts: copied"
    else
        log_warn "Scripts: not installed"
    fi

    # List agents
    echo ""
    echo -e "${BOLD}Installed Agents ($(ls -1 "$GLOBAL_AGENTS"/*.md 2>/dev/null | wc -l)):${NC}"

    if [[ -d "$GLOBAL_AGENTS" ]]; then
        for agent in "$GLOBAL_AGENTS"/*.md; do
            if [[ -f "$agent" ]]; then
                local name=$(basename "$agent" .md)
                printf "  ${GREEN}•${NC} %-15s\n" "$name"
            fi
        done
    fi

    # Check cognitive features
    echo ""
    echo -e "${BOLD}Cognitive Features:${NC}"

    if [[ -f "$CLAUDE_HOME/keywords.json" ]]; then
        log_info "Keywords config: present"
    else
        log_warn "Keywords config: missing"
    fi

    if [[ -f "$CLAUDE_HOME/attn_state.json" ]]; then
        local turn_count=$(python3 -c "import json; print(json.load(open('$CLAUDE_HOME/attn_state.json')).get('turn_count', 0))" 2>/dev/null || echo "?")
        log_info "Attention state: $turn_count turns tracked"
    else
        log_warn "Attention state: not initialized"
    fi

    if [[ -f "$CLAUDE_HOME/pool/instance_state.jsonl" ]]; then
        local pool_count=$(wc -l < "$CLAUDE_HOME/pool/instance_state.jsonl" 2>/dev/null || echo "0")
        log_info "Pool entries: $pool_count"
    else
        log_warn "Pool: not initialized"
    fi

    # MCP servers
    echo ""
    echo -e "${BOLD}MCP Configuration:${NC}"

    if [[ -f "$HOME/.mcp.json" ]]; then
        log_info "MCP config: ~/.mcp.json"
        local servers=$(python3 -c "import json; print(len(json.load(open('$HOME/.mcp.json')).get('mcpServers', {})))" 2>/dev/null || echo "?")
        log_info "MCP servers configured: $servers"
    else
        log_warn "MCP config: not found"
    fi

    # Environment
    echo ""
    echo -e "${BOLD}Environment:${NC}"
    log_info "CLAUDE_INSTANCE: ${CLAUDE_INSTANCE:-not set}"
    [[ -n "${NOTION_TOKEN:-}" ]] && log_info "NOTION_TOKEN: set" || log_warn "NOTION_TOKEN: not set"
    [[ -n "${GITHUB_TOKEN:-}" ]] && log_info "GITHUB_TOKEN: set" || log_warn "GITHUB_TOKEN: not set"
}

cmd_status_brief() {
    local dir="${1:-$CLAUDE_HOME}"

    echo ""
    echo -e "${BOLD}Agents installed:${NC}"
    for agent in "$dir/agents"/*.md; do
        if [[ -f "$agent" ]]; then
            local name=$(basename "$agent" .md)
            printf "  ${GREEN}•${NC} %s\n" "$name"
        fi
    done

    echo ""
    log_info "Run 'claude-agents-cli.sh status' for full details"
    log_info "Use '/agents' in Claude Code to manage agents"
}

# ============================================================================
# TEST
# ============================================================================

cmd_test() {
    log_header "Testing Claude Agents Configuration"
    check_python

    local errors=0

    # Test context router
    log_step "Testing context router..."
    if echo '{"prompt":"test terraform plan"}' | python3 "$GLOBAL_SCRIPTS/context-router.py" > /dev/null 2>&1; then
        log_info "Context router: OK"
    else
        log_error "Context router: FAILED"
        ((errors++))
    fi

    # Test pool loader
    log_step "Testing pool loader..."
    if python3 "$GLOBAL_SCRIPTS/pool-loader.py" > /dev/null 2>&1; then
        log_info "Pool loader: OK"
    else
        log_error "Pool loader: FAILED"
        ((errors++))
    fi

    # Test pool query
    log_step "Testing pool query..."
    if python3 "$GLOBAL_SCRIPTS/pool-query.py" --count > /dev/null 2>&1; then
        log_info "Pool query: OK"
    else
        log_error "Pool query: FAILED"
        ((errors++))
    fi

    # Test keyword activation
    log_step "Testing keyword activation..."
    local output=$(echo '{"prompt":"help me with kubernetes deployment"}' | python3 "$GLOBAL_SCRIPTS/context-router.py" 2>/dev/null)
    if echo "$output" | grep -q "kubernetes" 2>/dev/null; then
        log_info "Keyword activation: OK (kubernetes detected)"
    else
        log_warn "Keyword activation: No match (may need keywords.json)"
    fi

    # Summary
    echo ""
    if [[ $errors -eq 0 ]]; then
        log_header "All Tests Passed!"
    else
        log_error "Tests completed with $errors errors"
        exit 1
    fi
}

# ============================================================================
# PROXY
# ============================================================================

cmd_proxy() {
    local background=false

    while [[ $# -gt 0 ]]; do
        case $1 in
            --background) background=true; shift ;;
            *) log_error "Unknown option: $1"; exit 1 ;;
        esac
    done

    log_header "Starting MCP Proxy"

    if ! command -v mcp-proxy &> /dev/null; then
        log_error "mcp-proxy not found"
        log_info "Install with: go install github.com/tbxark/mcp-proxy@latest"
        exit 1
    fi

    local config="$REPO_DIR/mcp-proxy-config.json"
    if [[ ! -f "$config" ]]; then
        log_error "Config not found: $config"
        exit 1
    fi

    log_info "Using config: $config"

    if [[ "$background" == true ]]; then
        mcp-proxy -config "$config" &
        log_info "MCP proxy started in background (PID: $!)"
    else
        log_info "Starting MCP proxy (Ctrl+C to stop)..."
        mcp-proxy -config "$config"
    fi
}

# ============================================================================
# UNINSTALL
# ============================================================================

cmd_uninstall() {
    log_header "Uninstalling Claude Agents"

    if [[ -e "$GLOBAL_AGENTS" ]]; then
        rm -rf "$GLOBAL_AGENTS"
        log_info "Removed agents"
    fi

    if [[ -e "$GLOBAL_SCRIPTS" ]]; then
        rm -rf "$GLOBAL_SCRIPTS"
        log_info "Removed scripts"
    fi

    # Keep config files by default
    log_warn "Config files preserved (keywords.json, settings.json)"
    log_info "Delete manually if needed: rm $CLAUDE_HOME/*.json"

    log_header "Uninstall Complete"
}

# ============================================================================
# MAIN
# ============================================================================

main() {
    echo -e "${BOLD}${MAGENTA}"
    echo "╔══════════════════════════════════════════╗"
    echo "║   Claude DevOps Agents CLI v${VERSION}        ║"
    echo "╚══════════════════════════════════════════╝"
    echo -e "${NC}"

    local command="${1:-help}"
    shift || true

    case "$command" in
        install)  cmd_install "$@" ;;
        sync)     cmd_sync "$@" ;;
        status)   cmd_status "$@" ;;
        test)     cmd_test "$@" ;;
        proxy)    cmd_proxy "$@" ;;
        uninstall) cmd_uninstall "$@" ;;
        help|--help|-h) show_help ;;
        *)
            log_error "Unknown command: $command"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

main "$@"
