#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Session health check - runs on SessionStart.

Prints a few short lines so the first prompt of a session already knows what
is (not) available, instead of discovering it mid-task:

    [health] Obsidian: not running (port 27123) -> the `obsidian` MCP will fail; open Obsidian
    [health] Vault: /mnt/e/brain-vault not found -> `obsidian-vault` MCP reads will fail
    [health] Hooks: context-router.py declared in global AND project settings (runs twice)

Silent when everything is fine, except for a one-line summary. Never fails
the session: every check is wrapped and the script exits 0.
"""

import io
import json
import os
import socket
import sys
from pathlib import Path
from typing import List

if sys.platform == "win32":
    try:
        sys.stdout.reconfigure(encoding="utf-8", errors="replace")
    except (AttributeError, io.UnsupportedOperation):
        pass

MANAGED_HOOKS = ("context-router.py", "pool-loader.py", "pool-extractor.py", "session-health.py")
OBSIDIAN_REST_PORT = 27123


def claude_home() -> Path:
    return Path(os.environ.get("CLAUDE_HOME") or Path.home() / ".claude")


def project_dir() -> Path:
    return Path(os.environ.get("CLAUDE_PROJECT_DIR") or Path.cwd())


def load_json(path: Path) -> dict:
    try:
        with open(path, "r", encoding="utf-8") as fh:
            return json.load(fh)
    except (OSError, json.JSONDecodeError):
        return {}


def port_open(port: int, host: str = "127.0.0.1", timeout: float = 0.3) -> bool:
    try:
        with socket.create_connection((host, port), timeout=timeout):
            return True
    except OSError:
        return False


def mcp_servers() -> dict:
    """All MCP servers Claude Code would see here: user scope + this project."""
    servers = {}
    global_cfg = load_json(Path.home() / ".claude.json")
    servers.update(global_cfg.get("mcpServers", {}))
    for key, proj in global_cfg.get("projects", {}).items():
        if Path(key) == project_dir():
            servers.update(proj.get("mcpServers", {}))
    servers.update(load_json(project_dir() / ".mcp.json").get("mcpServers", {}))
    return servers


def check_obsidian(servers: dict) -> List[str]:
    out = []
    uses_rest = any("27123" in str(s.get("url", "")) for s in servers.values())
    if uses_rest and not port_open(OBSIDIAN_REST_PORT):
        out.append(f"Obsidian: not running (port {OBSIDIAN_REST_PORT}) -> the `obsidian` MCP will fail; open Obsidian or ignore it")
    return out


def check_vault(servers: dict) -> List[str]:
    """Vault path checks: ~/.claude/vault.json first, else the mcpvault server arg."""
    out = []
    vault_cfg = load_json(claude_home() / "vault.json")
    path = vault_cfg.get("path")
    source = "vault.json"
    if not path:
        for name, srv in servers.items():
            args = srv.get("args") or []
            if any("mcpvault" in str(a) for a in args) and args:
                path, source = args[-1], f"MCP `{name}`"
                break
    if path and not Path(path).expanduser().exists():
        out.append(f"Vault: {path} not found ({source}) -> vault reads/writes will fail (drive not mounted?)")
    if not vault_cfg:
        out.append("Vault: no ~/.claude/vault.json -> session-resume/save will ask which vault to use on this machine")
    return out


def check_hooks() -> List[str]:
    out = []
    proj_settings = project_dir() / ".claude" / "settings.json"
    glob_settings = claude_home() / "settings.json"
    if proj_settings.resolve() == glob_settings.resolve() or not proj_settings.exists():
        return out
    proj_text = proj_settings.read_text(encoding="utf-8", errors="ignore")
    glob_text = glob_settings.read_text(encoding="utf-8", errors="ignore") if glob_settings.exists() else ""
    dup = [h for h in MANAGED_HOOKS if h in proj_text and h in glob_text]
    if dup:
        out.append(f"Hooks: {', '.join(dup)} declared in global AND project settings -> run twice per event")
    return out


def check_agents() -> List[str]:
    agents_dir = claude_home() / "agents"
    if not agents_dir.is_dir():
        return ["Agents: ~/.claude/agents missing -> run claude-agents-cli.sh install"]
    broken = [p.name for p in agents_dir.glob("*.md") if p.is_symlink() and not p.exists()]
    if broken:
        return [f"Agents: broken symlinks ({', '.join(broken[:3])}{'...' if len(broken) > 3 else ''}) -> run claude-agents-cli.sh sync"]
    return []


def main() -> None:
    if os.environ.get("JARVIS_VOICE") == "1":
        return
    problems: List[str] = []
    try:
        servers = mcp_servers()
    except Exception:
        servers = {}
    for check in (lambda: check_obsidian(servers), lambda: check_vault(servers), check_hooks, check_agents):
        try:
            problems.extend(check())
        except Exception:
            continue

    if problems:
        print("\n".join(f"[health] {p}" for p in problems))
    else:
        n_agents = len(list((claude_home() / "agents").glob("*.md")))
        print(f"[health] OK: {n_agents} agents, hooks once, Obsidian reachable, vault configured")


if __name__ == "__main__":
    try:
        main()
    except Exception:
        pass
