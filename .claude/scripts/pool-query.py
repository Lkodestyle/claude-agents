#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Pool Query - CLI tool to query pool state

Query and display pool entries for debugging and monitoring.

Usage:
    python pool-query.py                    # Show recent entries
    python pool-query.py --since 1h         # Entries from last hour
    python pool-query.py --instance A       # Entries from instance A
    python pool-query.py --action completed # Only completed entries
"""

import argparse
import json
import os
import sys
import io
from pathlib import Path
from datetime import datetime, timedelta
from typing import List, Dict

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


def get_claude_dir() -> Path:
    """Get the .claude directory path."""
    env_root = os.environ.get("CONTEXT_DOCS_ROOT")
    if env_root:
        return Path(env_root)

    cwd = Path.cwd()
    if (cwd / ".claude").exists():
        return cwd / ".claude"

    for parent in cwd.parents:
        if (parent / ".claude").exists():
            return parent / ".claude"

    return Path.home() / ".claude"


def parse_duration(duration_str: str) -> timedelta:
    """Parse duration string like '1h', '30m', '2d'."""
    if not duration_str:
        return timedelta(hours=24)

    unit = duration_str[-1].lower()
    try:
        value = int(duration_str[:-1])
    except ValueError:
        return timedelta(hours=24)

    if unit == 'h':
        return timedelta(hours=value)
    elif unit == 'm':
        return timedelta(minutes=value)
    elif unit == 'd':
        return timedelta(days=value)
    else:
        return timedelta(hours=24)


def load_pool_entries(pool_dir: Path) -> List[Dict]:
    """Load all pool entries."""
    pool_file = pool_dir / "instance_state.jsonl"

    if not pool_file.exists():
        return []

    entries = []
    try:
        with open(pool_file, "r", encoding="utf-8") as f:
            for line in f:
                line = line.strip()
                if line:
                    try:
                        entries.append(json.loads(line))
                    except json.JSONDecodeError:
                        continue
    except Exception:
        return []

    return entries


def filter_entries(
    entries: List[Dict],
    since: timedelta = None,
    instance: str = None,
    action: str = None
) -> List[Dict]:
    """Filter entries based on criteria."""
    filtered = entries

    if since:
        cutoff = datetime.now() - since
        filtered = [
            e for e in filtered
            if datetime.fromisoformat(e.get("timestamp", "")) >= cutoff
        ]

    if instance:
        filtered = [
            e for e in filtered
            if e.get("source_instance", "").lower() == instance.lower()
        ]

    if action:
        filtered = [
            e for e in filtered
            if e.get("action", "").lower() == action.lower()
        ]

    return filtered


def format_entry(entry: Dict) -> str:
    """Format a single entry for display."""
    source = entry.get("source_instance", "?")
    action = entry.get("action", "unknown")
    topic = entry.get("topic", "")
    summary = entry.get("summary", "")
    timestamp = entry.get("timestamp", "")[:19]

    emoji = {
        "completed": "✅",
        "blocked": "🚫",
        "signaling": "📢",
        "in_progress": "🔄"
    }.get(action, "📝")

    lines = [
        f"{emoji} [{source}] {action.upper()}: {topic}",
        f"   Time: {timestamp}",
    ]

    if summary:
        lines.append(f"   Summary: {summary}")

    if entry.get("affects"):
        lines.append(f"   Affects: {entry['affects']}")

    return "\n".join(lines)


def main():
    parser = argparse.ArgumentParser(description="Query pool state")
    parser.add_argument("--since", "-s", help="Time window (e.g., 1h, 30m, 2d)")
    parser.add_argument("--instance", "-i", help="Filter by instance ID")
    parser.add_argument("--action", "-a", help="Filter by action type")
    parser.add_argument("--json", "-j", action="store_true", help="Output as JSON")
    parser.add_argument("--count", "-c", action="store_true", help="Show count only")

    args = parser.parse_args()

    claude_dir = get_claude_dir()
    pool_dir = claude_dir / "pool"

    entries = load_pool_entries(pool_dir)

    # Apply filters
    since = parse_duration(args.since) if args.since else timedelta(hours=24)
    filtered = filter_entries(entries, since, args.instance, args.action)

    # Sort by timestamp descending
    filtered.sort(key=lambda x: x.get("timestamp", ""), reverse=True)

    if args.count:
        print(f"Entries: {len(filtered)}")
        return

    if args.json:
        print(json.dumps(filtered, indent=2))
        return

    if not filtered:
        print("No pool entries found")
        return

    print(f"=== Pool Entries ({len(filtered)}) ===\n")
    for entry in filtered[:20]:  # Limit to 20
        print(format_entry(entry))
        print()


if __name__ == "__main__":
    main()
