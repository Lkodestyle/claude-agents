#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Pool Extractor - Extract completion/blocker signals

Runs on the Claude Code `Stop` hook. Per the hooks contract, stdin carries a
JSON payload (session_id, transcript_path, stop_hook_active, ...) — NOT the
conversation text. We read the transcript from `transcript_path`, take the
last assistant message, and persist any explicit ```pool blocks so other
instances can see them.

Only explicit blocks are extracted: deterministic hooks beat regex guessing.
The pool block protocol Claude emits (documented in CLAUDE.md):

    ```pool
    INSTANCE: A
    ACTION: completed | blocked | in_progress | signaling
    TOPIC: Setup authentication
    SUMMARY: Implemented JWT with refresh
    AFFECTS: auth.py, session.py
    BLOCKS: Session management can proceed
    ```

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
    """Find project root: CLAUDE_PROJECT_DIR (set by the hook runner) wins."""
    env_root = os.environ.get("CLAUDE_PROJECT_DIR")
    if env_root and (Path(env_root) / ".claude").exists():
        return Path(env_root)

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


def read_hook_payload() -> Dict:
    """Read the JSON payload Claude Code pipes to Stop hooks."""
    if sys.stdin.isatty():
        return {}
    try:
        raw = sys.stdin.read()
        return json.loads(raw) if raw.strip() else {}
    except (json.JSONDecodeError, OSError):
        return {}


def message_text(message: Dict) -> str:
    """Flatten a transcript message's content into plain text.

    Content is either a string or a list of typed blocks; only text blocks
    can carry a pool block, so everything else is skipped.
    """
    content = message.get("content", "")
    if isinstance(content, str):
        return content
    parts = []
    if isinstance(content, list):
        for block in content:
            if isinstance(block, dict) and block.get("type") == "text":
                parts.append(block.get("text", ""))
    return "\n".join(parts)


def last_assistant_text(transcript_path: str) -> str:
    """Return the text of the last assistant message in the transcript JSONL."""
    if not transcript_path:
        return ""
    path = Path(transcript_path)
    if not path.exists():
        return ""

    last = ""
    try:
        with open(path, "r", encoding="utf-8") as f:
            for line in f:
                line = line.strip()
                if not line:
                    continue
                try:
                    entry = json.loads(line)
                except json.JSONDecodeError:
                    continue
                message = entry.get("message") or {}
                if entry.get("type") == "assistant" or message.get("role") == "assistant":
                    text = message_text(message)
                    if text:
                        last = text
    except OSError:
        return ""

    return last


def extract_pool_blocks(text: str) -> List[Dict]:
    """Extract explicit pool blocks from text."""
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


def save_pool_entry(pool_dir: Path, entry: Dict, session_id: str) -> None:
    """Save a pool entry to the pool file."""
    pool_file = pool_dir / "instance_state.jsonl"

    # Ensure pool directory exists
    pool_dir.mkdir(parents=True, exist_ok=True)

    # Add metadata
    entry["id"] = str(uuid.uuid4())
    entry["timestamp"] = datetime.now().isoformat()
    entry["source_instance"] = entry.get("source_instance", get_instance_id())
    entry["session_id"] = session_id

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


def main():
    """Main pool extractor logic."""
    payload = read_hook_payload()

    # Loop guard: if this Stop event was itself triggered by a stop hook,
    # bail out (per Anthropic hooks docs).
    if payload.get("stop_hook_active"):
        return

    text = last_assistant_text(payload.get("transcript_path", ""))
    if not text:
        return

    blocks = extract_pool_blocks(text)
    if not blocks:
        return

    claude_dir = get_claude_dir()
    pool_dir = claude_dir / "pool"
    session_id = payload.get("session_id", "") or str(uuid.uuid4())[:8]

    for block in blocks:
        save_pool_entry(pool_dir, block, session_id)

    cleanup_old_entries(pool_dir)


if __name__ == "__main__":
    main()
