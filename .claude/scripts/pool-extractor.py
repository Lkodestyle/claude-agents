#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Pool Extractor - Extract completion/blocker signals

Extracts pool blocks from Claude responses and saves them
for other instances to see.

Based on claude-cognitive Pool Coordinator patterns.

Usage: Called by Claude Code hooks (Stop)
"""

import json
import os
import re
import sys
import io
import uuid
from pathlib import Path
from datetime import datetime
from typing import Dict, Optional, List

# Fix Windows encoding issues (fallback for non-WSL usage)
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


def get_session_id() -> str:
    """Get or generate session ID."""
    return os.environ.get("CLAUDE_SESSION_ID", str(uuid.uuid4())[:8])


def extract_pool_blocks(text: str) -> List[Dict]:
    """Extract explicit pool blocks from text.

    Format:
    ```pool
    INSTANCE: A
    ACTION: completed
    TOPIC: Setup authentication
    SUMMARY: Implemented JWT with refresh
    AFFECTS: auth.py, session.py
    BLOCKS: Session management can proceed
    ```
    """
    pattern = r'```pool\s*(.*?)```'
    matches = re.findall(pattern, text, re.DOTALL | re.IGNORECASE)

    blocks = []
    for match in matches:
        block = parse_pool_block(match)
        if block:
            blocks.append(block)

    return blocks


def parse_pool_block(block_text: str) -> Optional[Dict]:
    """Parse a single pool block into structured data."""
    lines = block_text.strip().split('\n')

    data = {}
    for line in lines:
        if ':' in line:
            key, value = line.split(':', 1)
            key = key.strip().lower()
            value = value.strip()

            # Map keys
            key_map = {
                'instance': 'source_instance',
                'action': 'action',
                'topic': 'topic',
                'summary': 'summary',
                'affects': 'affects',
                'blocks': 'blocks'
            }

            if key in key_map:
                data[key_map[key]] = value

    # Validate required fields
    if 'action' in data and 'topic' in data:
        return data

    return None


def detect_implicit_signals(text: str) -> List[Dict]:
    """Detect implicit completion/blocker signals in text.

    Looks for patterns like:
    - "Successfully completed X"
    - "Finished implementing Y"
    - "Blocked by Z"
    - "Cannot proceed until W"
    """
    signals = []

    # Completion patterns
    completion_patterns = [
        r'(?:successfully|finished|completed|done)\s+(?:implementing|creating|setting up|configuring|deploying)\s+(.{10,80})',
        r'(?:the|this)\s+(.{10,50})\s+(?:is now|has been)\s+(?:complete|ready|deployed|configured)',
        r'i\'ve\s+(?:finished|completed|done)\s+(.{10,80})',
    ]

    for pattern in completion_patterns:
        matches = re.findall(pattern, text, re.IGNORECASE)
        for match in matches:
            signals.append({
                'action': 'completed',
                'topic': match.strip()[:80],
                'summary': f'Detected completion: {match.strip()[:100]}'
            })

    # Blocker patterns
    blocker_patterns = [
        r'(?:blocked|waiting|cannot proceed)\s+(?:by|on|until)\s+(.{10,80})',
        r'(?:needs|requires|depends on)\s+(.{10,80})\s+(?:first|before)',
    ]

    for pattern in blocker_patterns:
        matches = re.findall(pattern, text, re.IGNORECASE)
        for match in matches:
            signals.append({
                'action': 'blocked',
                'topic': match.strip()[:80],
                'summary': f'Detected blocker: {match.strip()[:100]}'
            })

    return signals[:3]  # Limit to 3 implicit signals


def save_pool_entry(pool_dir: Path, entry: Dict) -> None:
    """Save a pool entry to the pool file."""
    pool_file = pool_dir / "instance_state.jsonl"

    # Ensure pool directory exists
    pool_dir.mkdir(parents=True, exist_ok=True)

    # Add metadata
    entry["id"] = str(uuid.uuid4())
    entry["timestamp"] = datetime.now().isoformat()
    entry["source_instance"] = entry.get("source_instance", get_instance_id())
    entry["session_id"] = get_session_id()

    try:
        with open(pool_file, "a", encoding="utf-8") as f:
            f.write(json.dumps(entry) + "\n")
    except Exception as e:
        sys.stderr.write(f"Failed to save pool entry: {e}\n")


def cleanup_old_entries(pool_dir: Path, max_entries: int = 100) -> None:
    """Keep only the most recent entries."""
    pool_file = pool_dir / "instance_state.jsonl"

    if not pool_file.exists():
        return

    try:
        with open(pool_file, "r", encoding="utf-8") as f:
            lines = f.readlines()

        if len(lines) > max_entries:
            # Keep only the last max_entries
            with open(pool_file, "w", encoding="utf-8") as f:
                f.writelines(lines[-max_entries:])
    except Exception:
        pass


def read_transcript_from_stdin() -> str:
    """Read conversation transcript from stdin."""
    if not sys.stdin.isatty():
        try:
            return sys.stdin.read()
        except Exception:
            return ""
    return ""


def main():
    """Main pool extractor logic."""
    claude_dir = get_claude_dir()
    pool_dir = claude_dir / "pool"

    # Read transcript
    transcript = read_transcript_from_stdin()

    if not transcript:
        return

    # Extract explicit pool blocks
    explicit_blocks = extract_pool_blocks(transcript)

    # If no explicit blocks, try implicit detection
    if not explicit_blocks:
        implicit_signals = detect_implicit_signals(transcript)
        for signal in implicit_signals:
            save_pool_entry(pool_dir, signal)
    else:
        for block in explicit_blocks:
            save_pool_entry(pool_dir, block)

    # Cleanup old entries
    cleanup_old_entries(pool_dir)


if __name__ == "__main__":
    main()
