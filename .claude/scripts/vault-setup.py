#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
vault-setup: generate this machine's ~/.claude/vault.json by DETECTION.

Every machine has its own Obsidian vault (different path, maybe no git, maybe
a different folder layout), so nothing vault-related lives in the repo. This
script looks at what actually exists on this machine and writes the config the
session skills, the health hook, the status line and the session-band mod read.

Detection, in order:
  1. Vaults registered in Obsidian's own config (obsidian.json) on Linux,
     Windows (incl. WSL view of %APPDATA%) and macOS.
  2. For each candidate: is it a git repo? does it have 02-projects/ (or any
     folder holding <name>/<name>.md project notes)? a CLAUDE.md rules note?
     a session template? a validator script?
  3. One candidate -> write the config. Several -> list them and take
     --path. None -> explain and take --path.

Usage:
  python3 vault-setup.py                 # detect and write (asks nothing)
  python3 vault-setup.py --path /x/y     # use this vault path
  python3 vault-setup.py --dry-run       # print what would be written
  python3 vault-setup.py --force         # overwrite an existing vault.json

Never guesses from another machine: only what exists here counts.
"""

import argparse
import json
import os
import subprocess
import sys
from pathlib import Path


def claude_home() -> Path:
    return Path(os.environ.get("CLAUDE_HOME") or Path.home() / ".claude")


def obsidian_configs() -> list[Path]:
    home = Path.home()
    candidates = [
        home / ".config" / "obsidian" / "obsidian.json",            # Linux
        home / "Library" / "Application Support" / "obsidian" / "obsidian.json",  # macOS
    ]
    appdata = os.environ.get("APPDATA")
    if appdata:
        candidates.append(Path(appdata) / "obsidian" / "obsidian.json")  # Windows
    # WSL view of the Windows config
    users = Path("/mnt/c/Users")
    if users.is_dir():
        try:
            for user in users.iterdir():
                candidates.append(user / "AppData" / "Roaming" / "obsidian" / "obsidian.json")
        except OSError:
            pass
    return [c for c in candidates if c.is_file()]


def registered_vaults() -> list[Path]:
    seen: list[Path] = []
    for cfg in obsidian_configs():
        try:
            data = json.loads(cfg.read_text(encoding="utf-8"))
        except (OSError, json.JSONDecodeError):
            continue
        for entry in (data.get("vaults") or {}).values():
            raw = entry.get("path")
            if not raw:
                continue
            path = Path(raw)
            # A Windows path read from WSL: translate C:\x -> /mnt/c/x
            if not path.exists() and len(raw) > 2 and raw[1] == ":":
                translated = Path("/mnt/" + raw[0].lower() + raw[2:].replace("\\", "/"))
                if translated.exists():
                    path = translated
            if path.exists() and path not in seen:
                seen.append(path)
    return seen


def is_git(path: Path) -> dict | None:
    try:
        run = subprocess.run(["git", "-C", str(path), "rev-parse", "--abbrev-ref", "HEAD"],
                             capture_output=True, text=True, timeout=5)
    except (OSError, subprocess.SubprocessError):
        return None
    if run.returncode != 0:
        return None
    branch = run.stdout.strip() or "master"
    has_remote = subprocess.run(["git", "-C", str(path), "remote"],
                                capture_output=True, text=True, timeout=5).stdout.strip() != ""
    return {"branch": branch, "push": has_remote}


def detect_projects_dir(path: Path) -> str | None:
    for name in ("02-projects", "projects", "proyectos", "Projects"):
        if (path / name).is_dir():
            return name
    return None


def detect_sessions_subdir(path: Path, projects_dir: str | None) -> str:
    if projects_dir:
        for proj in (path / projects_dir).iterdir():
            if proj.is_dir():
                for name in ("sesiones", "sessions"):
                    if (proj / name).is_dir():
                        return name
    return "sesiones"


def build_config(path: Path) -> dict:
    cfg: dict = {"name": path.name, "path": str(path), "mcp": "filesystem"}
    projects_dir = detect_projects_dir(path)
    if projects_dir:
        cfg["projects_dir"] = projects_dir
        cfg["sessions_subdir"] = detect_sessions_subdir(path, projects_dir)
    if (path / "CLAUDE.md").is_file():
        cfg["rules_note"] = "CLAUDE.md"
    for tpl in ("_templates/template-sesion.md", "_templates/template-session.md", "templates/sesion.md"):
        if (path / tpl).is_file():
            cfg["session_template"] = tpl
            break
    if (path / ".claude" / "validate-vault.py").is_file():
        cfg["validate"] = "python3 .claude/validate-vault.py"
    git = is_git(path)
    if git:
        cfg["git"] = git
    return cfg


def mcp_vault_registered() -> bool:
    """Whether an obsidian-vault (mcpvault) server exists in the user config."""
    try:
        data = json.loads((Path.home() / ".claude.json").read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        return False
    servers = dict(data.get("mcpServers") or {})
    for proj in (data.get("projects") or {}).values():
        servers.update(proj.get("mcpServers") or {})
    return any("mcpvault" in str(s.get("args", "")) for s in servers.values())


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--path", help="vault path (skips detection)")
    ap.add_argument("--dry-run", action="store_true", help="print the config, write nothing")
    ap.add_argument("--force", action="store_true", help="overwrite an existing vault.json")
    args = ap.parse_args()

    target = claude_home() / "vault.json"
    if target.exists() and not args.force and not args.dry_run:
        print(f"[vault-setup] {target} ya existe; usa --force para regenerarlo. Contenido actual:")
        print(target.read_text(encoding="utf-8"))
        return 0

    if args.path:
        path = Path(args.path).expanduser()
        if not path.is_dir():
            print(f"[vault-setup] {path} no existe o no es un directorio", file=sys.stderr)
            return 1
        candidates = [path]
    else:
        candidates = registered_vaults()

    if not candidates:
        print("[vault-setup] No encontre vaults registrados en Obsidian en esta maquina.")
        print("  Decime la ruta: python3 vault-setup.py --path /ruta/al/vault")
        return 1
    if len(candidates) > 1 and not args.path:
        print("[vault-setup] Hay varios vaults registrados; elegi uno con --path:")
        for c in candidates:
            print(f"  - {c}")
        return 1

    cfg = build_config(candidates[0])
    if mcp_vault_registered():
        cfg["mcp"] = "obsidian-vault"

    text = json.dumps(cfg, indent=2, ensure_ascii=False) + "\n"
    if args.dry_run:
        print(text, end="")
        return 0
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_text(text, encoding="utf-8")
    print(f"[vault-setup] escrito {target}:")
    print(text, end="")
    if cfg["mcp"] == "filesystem":
        print("[vault-setup] tip: para leer el vault via MCP de archivos, registra mcpvault:")
        print(f"  claude mcp add obsidian-vault -s user -- npx @bitbonsai/mcpvault@latest {cfg['path']}")
    if "git" not in cfg:
        print("[vault-setup] este vault NO esta en git: las skills no haran pull/push, solo lectura/escritura local.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
