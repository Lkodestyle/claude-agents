#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Pool Loader - Multi-Instance Coordination

Loads recent pool entries at session start to inform Claude
about work done by other instances.

Based on claude-cognitive Pool Coordinator patterns.

Usage: Called by Claude Code hooks (SessionStart)
"""

import json
import os
import sys
import io
from pathlib import Path
from datetime import datetime, timedelta
from typing import List, Dict, Optional

# Fix Windows encoding issues - ensure UTF-8 output
if sys.platform == "win32":
    try:
        sys.stdout.reconfigure(encoding='utf-8', errors='replace')
    except (AttributeError, io.UnsupportedOperation):
        pass
    try:
        sys.stderr.reconfigure(encoding='utf-8', errors='replace')
    except (AttributeError, io.UnsupportedOperation):
        pass


def get_project_root() -> Path:
    """Find project root by looking for .claude directory."""
    cwd = Path.cwd()

    if (cwd / ".claude").exists():
        return cwd

    for parent in cwd.parents:
        if (parent / ".claude").exists():
            return parent

    return cwd


def get_claude_dir() -> Path:
    """Get the .claude directory path."""
    env_root = os.environ.get("CONTEXT_DOCS_ROOT")
    if env_root:
        return Path(env_root)

    project_root = get_project_root()
    project_claude = project_root / ".claude"
    if project_claude.exists():
        return project_claude

    home_claude = Path.home() / ".claude"
    if home_claude.exists():
        return home_claude

    return project_claude


def get_instance_id() -> str:
    """Get the current instance identifier."""
    return os.environ.get("CLAUDE_INSTANCE", "default")


def load_pool_entries(pool_dir: Path, max_age_hours: int = 24) -> List[Dict]:
    """Load pool entries from the last N hours."""
    pool_file = pool_dir / "instance_state.jsonl"

    if not pool_file.exists():
        return []

    entries = []
    cutoff_time = datetime.now() - timedelta(hours=max_age_hours)

    try:
        with open(pool_file, "r", encoding="utf-8") as f:
            for line in f:
                line = line.strip()
                if not line:
                    continue
                try:
                    entry = json.loads(line)
                    timestamp = entry.get("timestamp")
                    if timestamp:
                        entry_time = datetime.fromisoformat(timestamp)
                        if entry_time >= cutoff_time:
                            entries.append(entry)
                except (json.JSONDecodeError, ValueError):
                    continue
    except Exception:
        return []

    return entries


def filter_relevant_entries(entries: List[Dict], current_instance: str) -> List[Dict]:
    """Filter entries relevant to current instance."""
    relevant = []

    for entry in entries:
        # Include all entries from other instances
        source = entry.get("source_instance", "")
        if source != current_instance:
            relevant.append(entry)
        # Also include own entries marked as "signaling" to others
        elif entry.get("action") == "signaling":
            relevant.append(entry)

    # Sort by timestamp (newest first)
    relevant.sort(key=lambda x: x.get("timestamp", ""), reverse=True)

    return relevant[:10]  # Limit to 10 most recent


def format_pool_summary(entries: List[Dict], current_instance: str) -> str:
    """Format pool entries into a readable summary."""
    if not entries:
        return ""

    output_parts = []
    output_parts.append("## 🔄 Recent Instance Activity\n\n")
    output_parts.append(f"*Current instance: {current_instance}*\n\n")

    for entry in entries:
        source = entry.get("source_instance", "?")
        action = entry.get("action", "unknown")
        topic = entry.get("topic", "")
        summary = entry.get("summary", "")
        affects = entry.get("affects", "")
        timestamp = entry.get("timestamp", "")

        # Format timestamp
        time_str = ""
        if timestamp:
            try:
                dt = datetime.fromisoformat(timestamp)
                time_str = dt.strftime("%H:%M")
            except ValueError:
                time_str = timestamp[:5]

        # Choose emoji based on action
        emoji = {
            "completed": "✅",
            "blocked": "🚫",
            "signaling": "📢",
            "in_progress": "🔄"
        }.get(action, "📝")

        output_parts.append(f"### {emoji} [{source}] {action.upper()}: {topic}\n")
        if time_str:
            output_parts.append(f"*{time_str}*\n\n")
        if summary:
            output_parts.append(f"{summary}\n\n")
        if affects:
            output_parts.append(f"**Affects:** {affects}\n\n")
        output_parts.append("---\n\n")

    output_parts.append("*Pool Coordinator: Use this context to avoid duplicating work*\n")

    return "".join(output_parts)


def main():
    """Main pool loader logic."""
    claude_dir = get_claude_dir()
    pool_dir = claude_dir / "pool"

    if not pool_dir.exists():
        print("", end="")
        return

    current_instance = get_instance_id()

    # Load recent entries
    entries = load_pool_entries(pool_dir)

    # Filter for relevance
    relevant = filter_relevant_entries(entries, current_instance)

    # Format output
    output = format_pool_summary(relevant, current_instance)

    print(output)


if __name__ == "__main__":
    main()
