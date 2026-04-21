#!/usr/bin/env python3
"""Enable jarvis-memory (or any MCP from .mcp.json) in the user's Claude Code config.

Claude Code silently ignores project-level MCP servers that aren't in each
project's `enabledMcpjsonServers` list. This is by design (opt-in security),
but means the MCP never launches until we flip that switch.

This script:
  1. Backs up ~/.claude.json to ~/.claude.json.bak-<timestamp>
  2. Finds all project entries whose path matches the target repo
  3. Adds the requested server name to each project's enabledMcpjsonServers
  4. Writes back, preserving indentation (2 spaces)

Safe to run multiple times (idempotent).
"""
from __future__ import annotations

import argparse
import json
import os
import shutil
import sys
import time
from pathlib import Path


def find_config() -> Path:
    """Return the path to ~/.claude.json, regardless of platform."""
    home = Path(os.path.expanduser("~"))
    cfg = home / ".claude.json"
    if not cfg.exists():
        print(f"error: {cfg} not found", file=sys.stderr)
        sys.exit(1)
    return cfg


def normalize(path: str) -> str:
    """Normalize a path for comparison (case-insensitive, forward slashes)."""
    return path.replace("\\", "/").rstrip("/").lower()


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--server",
        default="jarvis-memory",
        help="MCP server name (from .mcp.json) to enable (default: jarvis-memory)",
    )
    parser.add_argument(
        "--repo",
        default=str(Path(__file__).resolve().parent.parent),
        help="Repo path to match project entries against (default: this repo)",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Show what would change without writing",
    )
    args = parser.parse_args()

    cfg_path = find_config()
    target = normalize(args.repo)

    with cfg_path.open("r", encoding="utf-8") as f:
        data = json.load(f)

    projects = data.get("projects", {})
    matched: list[str] = []
    already_ok: list[str] = []

    for proj_path, proj in projects.items():
        if normalize(proj_path) != target:
            continue

        enabled = proj.setdefault("enabledMcpjsonServers", [])
        if args.server in enabled:
            already_ok.append(proj_path)
            continue
        enabled.append(args.server)
        matched.append(proj_path)

    if not matched and not already_ok:
        print(f"error: no project entries match {args.repo!r}", file=sys.stderr)
        print("       candidates in config:", file=sys.stderr)
        for p in projects.keys():
            print(f"         - {p}", file=sys.stderr)
        return 2

    for p in already_ok:
        print(f"ok (already enabled)  {p}")
    for p in matched:
        print(f"will enable           {p}")

    if args.dry_run:
        print("\n(dry-run, no changes written)")
        return 0

    if not matched:
        print("\nnothing to do — all matching projects already have it enabled")
        return 0

    # Back up before writing.
    backup = cfg_path.with_name(f"{cfg_path.name}.bak-{int(time.time())}")
    shutil.copy2(cfg_path, backup)
    print(f"\nbackup -> {backup}")

    with cfg_path.open("w", encoding="utf-8") as f:
        json.dump(data, f, indent=2, ensure_ascii=False)
        f.write("\n")

    print(f"wrote  -> {cfg_path}")
    print(f"\n{args.server} enabled in {len(matched)} project entry/entries.")
    print("Restart Claude Code (/exit, then `claude`) to pick up the change.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
