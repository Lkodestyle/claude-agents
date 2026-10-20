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
GLOBAL_COMMANDS="$CLAUDE_HOME/commands"
GLOBAL_SKILLS="$CLAUDE_HOME/skills"
VERSION="3.0.0"

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

    ${GREEN}obsidian${NC}    Setup Obsidian Vault MCP server
                Configures @bitbonsai/mcpvault for your vault

    ${GREEN}plugins${NC}     Show recommended plugins and skills

    ${GREEN}uninstall${NC}   Remove installed agents, scripts, and skills

    ${GREEN}help${NC}        Show this help message

${BOLD}EXAMPLES:${NC}
    ./scripts/claude-agents-cli.sh install              # Install with symlinks
    ./scripts/claude-agents-cli.sh install --copy       # Copy files
    ./scripts/claude-agents-cli.sh install --local      # Install to project
    ./scripts/claude-agents-cli.sh sync                 # Sync after git pull
    ./scripts/claude-agents-cli.sh status               # Check status
    ./scripts/claude-agents-cli.sh test                 # Test scripts
    ./scripts/claude-agents-cli.sh obsidian             # Setup Obsidian vault
    ./scripts/claude-agents-cli.sh plugins              # Show recommended plugins

${BOLD}ENVIRONMENT VARIABLES:${NC}
    CLAUDE_HOME         Override ~/.claude location
    CLAUDE_INSTANCE     Instance ID for pool coordination (A, B, C, etc.)
    MCP_PROXY_TOKEN     Token for mcp-proxy authentication
    NOTION_TOKEN        Notion integration token
    GITHUB_TOKEN        GitHub personal access token
    OBSIDIAN_VAULT_PATH Path to your Obsidian vault

${BOLD}MORE INFO:${NC}
    Repository: https://github.com/your-org/claude-agents
    Docs:       See CLAUDE.md in the repository

EOF
}

# ============================================================================
# UTILITIES
# ============================================================================

# Detected Python invocation (set by detect_python)
PYTHON_BIN=""

detect_python() {
    # Prefer python3, then python, then the Windows launcher `py -3`.
    if command -v python3 &> /dev/null; then
        PYTHON_BIN="python3"
    elif command -v python &> /dev/null; then
        PYTHON_BIN="python"
    elif command -v py &> /dev/null; then
        PYTHON_BIN="py -3"
    else
        log_error "Python not found. Please install Python 3.8+ (python3, python, or py)"
        return 1
    fi
    log_info "Python found: $($PYTHON_BIN --version 2>&1) (using '$PYTHON_BIN')"
    return 0
}

# Backwards-compatible alias used across the script.
check_python() { detect_python; }

# Merge the managed hooks (scripts/hooks.json) and the repo permissions into
# the target settings.json WITHOUT touching anything else the user has there
# (model, theme, autoMode, extra hooks, extra permissions...).
#
# Hook commands are rewritten to point at the install dir, using the detected
# Python interpreter. Only hooks that reference one of OUR scripts are replaced;
# hooks from other tools (e.g. the Jarvis persona hook) are preserved as-is.
install_settings() {
    local install_dir="$1"
    # Token used to locate the hook scripts inside each hook command. It is kept
    # LITERAL in the generated JSON (not expanded at install time) so the value
    # resolves at hook runtime in whatever environment Claude Code runs the hook:
    #   - global: '$HOME/.claude/scripts' -> sh expands $HOME (works in WSL AND
    #     native-Windows Git Bash, where $HOME is the Windows user profile and
    #     MSYS converts the arg to a Windows path for python3.exe).
    #   - local:  '${CLAUDE_PROJECT_DIR}/.claude/scripts' -> expanded by Claude
    #     Code itself to the native project path on each OS.
    # This avoids baking an absolute OS-specific path (e.g. /home/<user>/...),
    # which breaks when the same ~/.claude is shared (symlinked) between WSL and
    # native Windows.
    local hook_scripts="${2:-\$HOME/.claude/scripts}"
    local hooks_src="$REPO_DIR/scripts/hooks.json"
    local perms_src="$REPO_DIR/.claude/settings.json"
    local dst="$install_dir/settings.json"

    if [[ ! -f "$hooks_src" ]]; then
        log_warn "scripts/hooks.json not found, skipping settings"
        return 0
    fi

    local result
    result=$(PYTHON_BIN="$PYTHON_BIN" HOOK_SCRIPTS="$hook_scripts" \
        HOOKS_SRC="$hooks_src" PERMS_SRC="$perms_src" DST="$dst" \
        $PYTHON_BIN - <<'PYEOF'
import json, os, re, shutil, sys
from datetime import datetime

py = os.environ["PYTHON_BIN"]
hook_scripts = os.environ["HOOK_SCRIPTS"]
hooks_src = os.environ["HOOKS_SRC"]
perms_src = os.environ["PERMS_SRC"]
dst = os.environ["DST"]

script_re = re.compile(r"([\w.-]+\.py)")


def load(path):
    if not os.path.isfile(path):
        return {}
    with open(path, "r", encoding="utf-8") as fh:
        return json.load(fh)


managed = load(hooks_src).get("hooks", {})
managed_scripts = set()
for groups in managed.values():
    for group in groups:
        for hook in group.get("hooks", []):
            m = script_re.search(hook.get("command", ""))
            if m:
                managed_scripts.add(m.group(1))
sl_cmd = (load(hooks_src).get("statusLine") or {}).get("command", "")
if script_re.search(sl_cmd):
    managed_scripts.add(script_re.search(sl_cmd).group(1))


def is_managed(hook):
    m = script_re.search(hook.get("command", ""))
    return bool(m) and m.group(1) in managed_scripts


def rewrite(hook):
    m = script_re.search(hook.get("command", ""))
    hook = dict(hook)
    # Forward slashes + literal token: portable across WSL and Windows.
    hook["command"] = f'{py} "{hook_scripts}/{m.group(1)}"'
    return hook


existing = load(dst)
original = json.dumps(existing, sort_keys=True)
merged = dict(existing)

# --- hooks: drop every hook that references a managed script (wherever it
# lives), then add the managed ones under their event. Everything else stays.
hooks = {}
for event, groups in merged.get("hooks", {}).items():
    kept = []
    for group in groups:
        rest = [h for h in group.get("hooks", []) if not is_managed(h)]
        if rest:
            kept.append({**group, "hooks": rest})
    if kept:
        hooks[event] = kept
for event, groups in managed.items():
    for group in groups:
        new_group = dict(group)
        new_group["hooks"] = [rewrite(h) for h in group.get("hooks", [])]
        hooks.setdefault(event, []).append(new_group)
merged["hooks"] = hooks

# --- statusLine: set ours only if the user has none, or if theirs is ours
# (so re-installs refresh the path). A custom user status line is left alone.
managed_sl = load(hooks_src).get("statusLine")
if managed_sl:
    current = merged.get("statusLine") or {}
    if not current or is_managed(current):
        merged["statusLine"] = rewrite(managed_sl) if script_re.search(managed_sl.get("command", "")) else managed_sl

# --- permissions: union, user's entries first.
perms_repo = load(perms_src).get("permissions", {})
perms = dict(merged.get("permissions", {}))
for key in ("allow", "deny"):
    current = list(perms.get(key, []))
    for entry in perms_repo.get(key, []):
        if entry not in current:
            current.append(entry)
    if current:
        perms[key] = current
if perms:
    merged["permissions"] = perms

if json.dumps(merged, sort_keys=True) == original:
    print("unchanged")
    sys.exit(0)

if os.path.isfile(dst):
    backup = f"{dst}.backup.{datetime.now():%Y%m%d_%H%M%S}"
    shutil.copy2(dst, backup)
    print(f"backup {backup}")

os.makedirs(os.path.dirname(dst), exist_ok=True)
with open(dst, "w", encoding="utf-8") as fh:
    json.dump(merged, fh, indent=2)
    fh.write("\n")
print("written")
PYEOF
    )
    case "$result" in
        unchanged) log_info "Settings already up to date: $dst" ;;
        *backup*)  log_warn "Backed up previous settings: ${result#backup }"
                   log_info "Settings merged (hooks -> $hook_scripts): $dst" ;;
        *)         log_info "Settings written (hooks -> $hook_scripts): $dst" ;;
    esac
}

# Strip only OUR hooks from a settings.json (used by uninstall).
remove_managed_hooks() {
    local dst="$1"
    [[ -f "$dst" ]] || return 0
    HOOKS_SRC="$REPO_DIR/scripts/hooks.json" DST="$dst" $PYTHON_BIN - <<'PYEOF'
import json, os, re

script_re = re.compile(r"([\w.-]+\.py)")
with open(os.environ["HOOKS_SRC"], encoding="utf-8") as fh:
    managed = {m.group(1) for groups in json.load(fh)["hooks"].values()
               for g in groups for h in g["hooks"]
               for m in [script_re.search(h["command"])] if m}
with open(os.environ["DST"], encoding="utf-8") as fh:
    data = json.load(fh)
sl = (data.get("statusLine") or {}).get("command", "")
if "statusline.py" in sl:
    data.pop("statusLine", None)
hooks = {}
for event, groups in data.get("hooks", {}).items():
    kept = []
    for g in groups:
        rest = [h for h in g.get("hooks", [])
                if not (script_re.search(h.get("command", "")) and
                        script_re.search(h["command"]).group(1) in managed)]
        if rest:
            kept.append({**g, "hooks": rest})
    if kept:
        hooks[event] = kept
if hooks:
    data["hooks"] = hooks
else:
    data.pop("hooks", None)
with open(os.environ["DST"], "w", encoding="utf-8") as fh:
    json.dump(data, fh, indent=2)
    fh.write("\n")
PYEOF
    log_info "Managed hooks removed from $dst"
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

# Where user files displaced by an install go (one folder per run).
BACKUP_ROOT="$CLAUDE_HOME/backups/claude-agents-$(date +%Y%m%d_%H%M%S)"

# Install every entry of $src_dir into $dst_dir, ONE ENTRY AT A TIME, so files
# the user keeps in the same folder (other skills, other commands, hooks from
# other tools) survive. Entries we own are replaced; anything else is left
# alone. A user file that would be overwritten is moved to $BACKUP_ROOT first,
# but only if its content differs from what we are about to install.
install_entries() {
    local src_dir="$1" dst_dir="$2" mode="$3" kind="$4"
    [[ -d "$src_dir" ]] || return 0

    # Older installs symlinked the whole folder; turn it back into a real dir.
    if [[ -L "$dst_dir" ]]; then
        log_step "Replacing folder symlink with per-entry install: $dst_dir"
        rm "$dst_dir"
    fi
    mkdir -p "$dst_dir"

    local entry name target
    for entry in "$src_dir"/*; do
        [[ -e "$entry" ]] || continue
        name="$(basename "$entry")"
        [[ "$name" == "__pycache__" ]] && continue
        target="$dst_dir/$name"

        if [[ -L "$target" ]]; then
            rm "$target"
        elif [[ -e "$target" ]]; then
            if ! diff -rq "$entry" "$target" > /dev/null 2>&1; then
                mkdir -p "$BACKUP_ROOT/$kind"
                mv "$target" "$BACKUP_ROOT/$kind/$name"
                log_warn "Backed up modified $kind/$name -> $BACKUP_ROOT/$kind/"
            else
                rm -rf "$target"
            fi
        fi

        if [[ "$mode" == "symlink" ]]; then
            ln -s "$entry" "$target"
        else
            cp -r "$entry" "$target"
        fi
    done

    # Prune symlinks that point into this repo but whose source is gone
    # (e.g. an agent that was renamed or removed upstream).
    for target in "$dst_dir"/*; do
        if [[ -L "$target" && ! -e "$target" ]] && \
           [[ "$(readlink "$target")" == "$REPO_DIR/"* ]]; then
            rm "$target"
            log_step "Pruned stale $kind/$(basename "$target")"
        fi
    done

    log_info "$kind: $(ls -1 "$src_dir" | grep -vc __pycache__) entries installed (${mode}) into $dst_dir"
}

# Remove from $dst_dir only the entries that come from $src_dir.
remove_entries() {
    local src_dir="$1" dst_dir="$2" kind="$3"
    [[ -d "$dst_dir" ]] || return 0
    if [[ -L "$dst_dir" ]]; then
        rm "$dst_dir"
        log_info "Removed $kind folder symlink"
        return 0
    fi
    local entry name
    for entry in "$src_dir"/*; do
        name="$(basename "$entry")"
        [[ -e "$dst_dir/$name" || -L "$dst_dir/$name" ]] && rm -rf "$dst_dir/$name"
    done
    rmdir "$dst_dir" 2>/dev/null && log_info "Removed $kind" || log_info "Removed our $kind (folder kept: it still has your own files)"
}

# Router v2 kept its state next to keywords.json (global AND inside every
# project's .claude/). v3 keeps it under ~/.claude/router/, so the old files
# are dead weight: remove them from the install dir and the current project.
cleanup_legacy_state() {
    local dir f
    # Slash commands were removed from the repo (never used): drop our old links.
    for f in "$CLAUDE_HOME"/commands/*; do
        if [[ -L "$f" && ! -e "$f" && "$(readlink "$f")" == "$REPO_DIR/"* ]]; then
            rm -f "$f"
            log_step "Removed stale command link: $f"
        fi
    done
    for dir in "$CLAUDE_HOME" "$(pwd)/.claude"; do
        for f in attn_state.json attention_history.jsonl; do
            if [[ -f "$dir/$f" ]]; then
                rm -f "$dir/$f"
                log_step "Removed legacy router state: $dir/$f"
            fi
        done
    done
}

# Install (or refresh) agents, scripts, commands and skills into $install_dir.
install_tree() {
    local install_dir="$1" mode="$2"
    install_entries "$REPO_DIR/.claude/agents"   "$install_dir/agents"   "$mode" "agents"
    install_entries "$REPO_DIR/.claude/scripts"  "$install_dir/scripts"  "$mode" "scripts"
    install_entries "$REPO_DIR/.claude/skills"   "$install_dir/skills"   "$mode" "skills"
}

# "symlinked" / "copied" / "not installed", based on the first entry we own.
install_mode_of() {
    local src_dir="$1" dst_dir="$2"
    [[ -L "$dst_dir" ]] && { echo "symlinked (folder)"; return; }
    [[ -d "$dst_dir" ]] || { echo "not installed"; return; }
    local entry name
    for entry in "$src_dir"/*; do
        name="$(basename "$entry")"
        [[ "$name" == "__pycache__" ]] && continue
        if [[ -L "$dst_dir/$name" ]]; then echo "symlinked"; return; fi
        if [[ -e "$dst_dir/$name" ]]; then echo "copied"; return; fi
    done
    echo "not installed"
}

# ============================================================================
# MODS (hooks-module plugins shipped in mods/, installed through the repo's
# own marketplace so Claude Code reads them FROM THE REPO: a git pull + the
# /reload-plugins command is enough to pick up changes)
# ============================================================================

MODS="session-band guardrails"
MARKETPLACE="claude-agents"

install_mods() {
    if ! command -v claude &> /dev/null; then
        log_warn "claude CLI not found: mods not installed (run 'claude plugin marketplace add $REPO_DIR' later)"
        return 0
    fi
    log_step "Installing mods (plugins) from $REPO_DIR..."
    if ! claude plugin marketplace list 2>/dev/null | grep -q "❯ $MARKETPLACE\b"; then
        claude plugin marketplace add "$REPO_DIR" > /dev/null 2>&1 && log_info "Marketplace '$MARKETPLACE' registered (user scope)" \
            || { log_warn "Could not register the marketplace"; return 0; }
    fi
    local m
    for m in $MODS; do
        if claude plugin list 2>/dev/null | grep -q "❯ $m@$MARKETPLACE"; then
            log_info "mod $m: already installed"
        else
            claude plugin install "$m@$MARKETPLACE" -s user > /dev/null 2>&1 && log_info "mod $m: installed (user scope)" \
                || log_warn "mod $m: install failed (try: claude plugin install $m@$MARKETPLACE)"
        fi
    done
}

remove_mods() {
    command -v claude &> /dev/null || return 0
    local m
    for m in $MODS; do
        claude plugin uninstall "$m@$MARKETPLACE" -s user -y > /dev/null 2>&1 && log_info "mod $m: uninstalled"
    done
    claude plugin marketplace remove "$MARKETPLACE" > /dev/null 2>&1 && log_info "Marketplace '$MARKETPLACE' removed"
}

mods_status() {
    echo ""
    echo -e "${BOLD}Mods:${NC}"
    command -v claude &> /dev/null || { log_warn "claude CLI not found"; return 0; }
    local listing m
    listing=$(claude plugin list 2>/dev/null)
    for m in $MODS; do
        if echo "$listing" | grep -A5 "❯ $m@$MARKETPLACE" | grep -q "Status: .*enabled"; then
            log_info "$m: enabled"
        elif echo "$listing" | grep -q "❯ $m@$MARKETPLACE"; then
            log_warn "$m: installed but disabled"
        else
            log_warn "$m: not installed (run install)"
        fi
    done
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

    mkdir -p "$install_dir"
    if [[ "$mode" == "symlink" ]]; then
        log_info "Mode: symlink (auto-updates with git pull)"
    else
        log_info "Mode: copy (run 'sync' after git pull)"
    fi

    install_tree "$install_dir" "$mode"
    cleanup_legacy_state
    [[ "$target" == "global" ]] && install_mods

    # Per-machine vault config: detect this machine's Obsidian vault and write
    # ~/.claude/vault.json (skips silently when it already exists; every vault
    # is different per machine and may not be in git, so nothing is assumed).
    if [[ "$target" == "global" && ! -f "$CLAUDE_HOME/vault.json" ]]; then
        log_step "Detecting this machine's Obsidian vault..."
        $PYTHON_BIN "$REPO_DIR/.claude/scripts/vault-setup.py" ||             log_warn "No vault detected; run: $PYTHON_BIN $REPO_DIR/.claude/scripts/vault-setup.py --path /ruta/al/vault"
    fi

    # Install config files
    log_step "Installing configuration files..."

    if [[ ! -f "$install_dir/keywords.json" ]]; then
        cp "$REPO_DIR/.claude/keywords.json" "$install_dir/keywords.json"
        log_info "Keywords config installed"
    else
        log_warn "keywords.json already exists, skipping"
    fi

    if [[ "$target" == "local" ]]; then
        install_settings "$install_dir" '${CLAUDE_PROJECT_DIR}/.claude/scripts'
    else
        install_settings "$install_dir" '$HOME/.claude/scripts'
    fi

    # MCP servers are per project: Claude Code reads <project>/.mcp.json, so
    # there is nothing to install globally. Copy the repo's .mcp.json into a
    # project yourself if you want the same servers there.

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

    # Re-run the per-entry install in whatever mode is currently in use, so a
    # copy install picks up new/renamed/removed files and a symlink install
    # gets any new entries linked.
    local mode="copy"
    [[ "$(install_mode_of "$REPO_DIR/.claude/agents" "$GLOBAL_AGENTS")" == symlinked* ]] && mode="symlink"
    log_info "Detected install mode: $mode"
    install_tree "$CLAUDE_HOME" "$mode"
    cleanup_legacy_state
    install_mods

    # Always sync keywords.json (may have been updated)
    log_step "Syncing keywords.json..."
    cp "$REPO_DIR/.claude/keywords.json" "$CLAUDE_HOME/keywords.json"
    log_info "Keywords synced"

    # Regenerate settings.json so hook paths stay correct after updates
    log_step "Regenerating settings.json hook paths..."
    detect_python
    install_settings "$CLAUDE_HOME" '$HOME/.claude/scripts'

    log_header "Sync Complete!"
}

# ============================================================================
# STATUS
# ============================================================================

cmd_status() {
    log_header "Claude Agents Status"
    check_python

    # Check installation
    echo -e "${BOLD}Installation:${NC}"
    if [[ -d "$CLAUDE_HOME" ]]; then
        log_info "Claude home: $CLAUDE_HOME"
    else
        log_error "Claude home not found: $CLAUDE_HOME"
    fi

    local kind state
    for kind in agents scripts skills; do
        state="$(install_mode_of "$REPO_DIR/.claude/$kind" "$CLAUDE_HOME/$kind")"
        case "$state" in
            "not installed") log_warn "$kind: $state" ;;
            *) log_info "$kind: $state" ;;
        esac
    done

    # Hooks: ours should appear exactly once, and only in the global settings.
    echo ""
    echo -e "${BOLD}Hooks:${NC}"
    local project_settings="$(pwd)/.claude/settings.json"
    for script in context-router.py session-health.py pool-loader.py pool-extractor.py; do
        local n_global=0 n_project=0
        [[ -f "$CLAUDE_HOME/settings.json" ]] && n_global=$(grep -c "$script" "$CLAUDE_HOME/settings.json" || true)
        [[ -f "$project_settings" && "$project_settings" != "$CLAUDE_HOME/settings.json" ]] && n_project=$(grep -c "$script" "$project_settings" || true)
        if [[ $n_global -eq 1 && $n_project -eq 0 ]]; then
            log_info "$script: global"
        elif [[ $n_global -eq 0 && $n_project -eq 0 ]]; then
            log_warn "$script: not configured (run install)"
        else
            log_error "$script: global=$n_global project=$n_project -> would run $((n_global + n_project)) times per event"
        fi
    done

    mods_status

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


    # List skills
    echo ""
    if [[ -d "$GLOBAL_SKILLS" ]]; then
        local skill_count=$(ls -1d "$GLOBAL_SKILLS"/*/SKILL.md 2>/dev/null | wc -l)
        echo -e "${BOLD}Installed Skills ($skill_count):${NC}"
        for skill_dir in "$GLOBAL_SKILLS"/*/; do
            if [[ -f "${skill_dir}SKILL.md" ]]; then
                local name=$(basename "$skill_dir")
                printf "  ${MAGENTA}•${NC} %s\n" "$name"
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

    if [[ -d "$CLAUDE_HOME/router" ]]; then
        local projects=$(ls -1 "$CLAUDE_HOME/router"/*.json 2>/dev/null | wc -l)
        log_info "Router state: $projects project(s) tracked in $CLAUDE_HOME/router/"
    else
        log_warn "Router state: not initialized (runs on first prompt)"
    fi

    if [[ -n "${CLAUDE_INSTANCE:-}" ]]; then
        local pool_count=$(cat "$CLAUDE_HOME/pool/instance_state.jsonl" "$(pwd)/.claude/pool/instance_state.jsonl" 2>/dev/null | wc -l)
        log_info "Pool coordinator: ON (instance $CLAUDE_INSTANCE, $pool_count entries)"
    else
        log_info "Pool coordinator: off (opt-in: export CLAUDE_INSTANCE=A)"
    fi

    # MCP servers
    echo ""
    echo -e "${BOLD}MCP Configuration:${NC}"

    local mcp_file="$(pwd)/.mcp.json"
    if [[ -f "$mcp_file" ]]; then
        local servers=$($PYTHON_BIN -c "import json; print(', '.join(json.load(open('$mcp_file')).get('mcpServers', {})))" 2>/dev/null || echo "?")
        log_info "Project .mcp.json: $servers"
    else
        log_info "No .mcp.json in $(pwd) (MCP servers are configured per project)"
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
    if echo '{"prompt":"test terraform plan"}' | $PYTHON_BIN "$GLOBAL_SCRIPTS/context-router.py" > /dev/null 2>&1; then
        log_info "Context router: OK"
    else
        log_error "Context router: FAILED"
        ((errors++))
    fi

    # Test pool loader
    log_step "Testing pool loader..."
    if echo '{}' | CLAUDE_INSTANCE=test $PYTHON_BIN "$GLOBAL_SCRIPTS/pool-loader.py" > /dev/null 2>&1; then
        log_info "Pool loader: OK"
    else
        log_error "Pool loader: FAILED"
        ((errors++))
    fi

    # Test pool query
    log_step "Testing pool query..."
    if $PYTHON_BIN "$GLOBAL_SCRIPTS/pool-query.py" --count > /dev/null 2>&1; then
        log_info "Pool query: OK"
    else
        log_error "Pool query: FAILED"
        ((errors++))
    fi

    # Test keyword activation
    log_step "Testing keyword activation..."
    local output=$(echo '{"prompt":"help me with kubernetes deployment"}' | $PYTHON_BIN "$GLOBAL_SCRIPTS/context-router.py" 2>/dev/null)
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
# OBSIDIAN
# ============================================================================

cmd_obsidian() {
    if [[ -f "$REPO_DIR/scripts/setup-obsidian.sh" ]]; then
        bash "$REPO_DIR/scripts/setup-obsidian.sh" "$@"
    else
        log_error "setup-obsidian.sh not found in $REPO_DIR/scripts/"
        exit 1
    fi
}

# ============================================================================
# PLUGINS
# ============================================================================

cmd_plugins() {
    log_header "Recommended Plugins & Skills"

    echo -e "${BOLD}Essential Plugins:${NC}"
    echo ""
    printf "  ${GREEN}%-20s${NC} %s\n" "superpowers" "14 workflow skills: brainstorming, TDD, debugging, plans, code review, parallel agents"
    printf "  ${GREEN}%-20s${NC} %s\n" "skill-creator" "Create and test custom skills with eval framework"
    printf "  ${GREEN}%-20s${NC} %s\n" "context7" "Up-to-date library/framework documentation via MCP"
    echo ""

    echo -e "${BOLD}Optional Plugins (by use case):${NC}"
    echo ""
    printf "  ${CYAN}%-20s${NC} %s\n" "github" "GitHub repos, PRs, issues management"
    printf "  ${CYAN}%-20s${NC} %s\n" "gitlab" "GitLab repos, MRs, pipelines"
    printf "  ${CYAN}%-20s${NC} %s\n" "playwright" "E2E browser testing automation"
    printf "  ${CYAN}%-20s${NC} %s\n" "supabase" "Supabase backend-as-a-service"
    printf "  ${CYAN}%-20s${NC} %s\n" "firebase" "Firebase apps and services"
    printf "  ${CYAN}%-20s${NC} %s\n" "slack" "Slack messaging integration"
    printf "  ${CYAN}%-20s${NC} %s\n" "linear" "Linear issue tracking"
    printf "  ${CYAN}%-20s${NC} %s\n" "stripe" "Stripe payments integration"
    echo ""

    echo -e "${BOLD}Included Workflow Skills (installed with this repo):${NC}"
    echo ""
    printf "  ${MAGENTA}%-25s${NC} %s\n" "deploy-pipeline" "CI/CD + Docker + Cloud provider deployment"
    printf "  ${MAGENTA}%-25s${NC} %s\n" "observability-setup" "Monitoring + Logging + Alerting stack"
    printf "  ${MAGENTA}%-25s${NC} %s\n" "infrastructure-as-code" "Terraform + Networking + Cloud IaC"
    printf "  ${MAGENTA}%-25s${NC} %s\n" "security-hardening" "Security audit + CVE fix + Compliance"
    printf "  ${MAGENTA}%-25s${NC} %s\n" "full-stack-scaffold" "Project scaffolding with API + DB + Frontend"
    echo ""

    echo -e "${BOLD}How to install plugins:${NC}"
    echo ""
    echo "  1. Enable the superpowers plugin:"
    echo "     claude config set enabledPlugins.superpowers@claude-plugins-official true"
    echo ""
    echo "  2. External plugins (github, slack, etc.) are installed via Claude Code:"
    echo "     claude /plugins"
    echo ""
    echo "  3. Workflow skills from this repo are installed automatically with:"
    echo "     ./scripts/claude-agents-cli.sh install"
    echo ""

    echo -e "${BOLD}Integration with superpowers:${NC}"
    echo ""
    echo "  The workflow skills in this repo complement superpowers:"
    echo "  - Use superpowers:brainstorming BEFORE running a workflow skill"
    echo "  - Use superpowers:writing-plans to create implementation plans"
    echo "  - Use superpowers:verification-before-completion to validate results"
    echo ""
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

    check_python
    remove_entries "$REPO_DIR/.claude/agents"   "$GLOBAL_AGENTS"   "agents"
    remove_entries "$REPO_DIR/.claude/scripts"  "$GLOBAL_SCRIPTS"  "scripts"
    remove_entries "$REPO_DIR/.claude/skills"   "$GLOBAL_SKILLS"   "skills"
    remove_managed_hooks "$CLAUDE_HOME/settings.json"
    remove_mods

    # Keep config files by default
    log_warn "Preserved: keywords.json, permissions and any non-managed settings"
    log_info "Delete manually if needed: rm $CLAUDE_HOME/keywords.json"

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
        obsidian) cmd_obsidian "$@" ;;
        plugins)  cmd_plugins "$@" ;;
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
