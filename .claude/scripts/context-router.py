#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Context Router v2.0 - Intelligent Agent Activation

Based on claude-cognitive patterns, this script provides:
- Keyword-based activation of agent documentation
- Co-activation of related agents
- Decay mechanism for unused agents
- Tiered injection (HOT/WARM/COLD)

Usage: Called by Claude Code hooks (UserPromptSubmit)
"""

import json
import os
import sys
import re
import io
from pathlib import Path
from datetime import datetime
from typing import Dict, List, Tuple, Optional

# Fix Windows encoding issues - ensure UTF-8 output
if sys.platform == "win32":
    try:
        # Only wrap if not already wrapped
        if not isinstance(sys.stdout, io.TextIOWrapper) or sys.stdout.encoding.lower() != 'utf-8':
            sys.stdout.reconfigure(encoding='utf-8', errors='replace')
    except (AttributeError, io.UnsupportedOperation):
        # Fallback for older Python or special environments
        pass
    try:
        if not isinstance(sys.stderr, io.TextIOWrapper) or sys.stderr.encoding.lower() != 'utf-8':
            sys.stderr.reconfigure(encoding='utf-8', errors='replace')
    except (AttributeError, io.UnsupportedOperation):
        pass


def get_project_root() -> Path:
    """Find project root by looking for .claude directory."""
    cwd = Path.cwd()

    # Check current directory
    if (cwd / ".claude").exists():
        return cwd

    # Walk up the directory tree
    for parent in cwd.parents:
        if (parent / ".claude").exists():
            return parent

    return cwd


def get_claude_dir() -> Path:
    """Get the .claude directory path."""
    # Check environment variable first
    env_root = os.environ.get("CONTEXT_DOCS_ROOT")
    if env_root:
        return Path(env_root)

    # Check project-local .claude
    project_root = get_project_root()
    project_claude = project_root / ".claude"
    if project_claude.exists():
        return project_claude

    # Fallback to global ~/.claude
    home_claude = Path.home() / ".claude"
    if home_claude.exists():
        return home_claude

    return project_claude


def load_config(claude_dir: Path) -> dict:
    """Load keywords.json configuration."""
    config_path = claude_dir / "keywords.json"
    if config_path.exists():
        with open(config_path, "r", encoding="utf-8") as f:
            return json.load(f)
    return {
        "keywords": {},
        "co_activation": {},
        "pinned": [],
        "decay_rates": {},
        "thresholds": {
            "hot": 0.8,
            "warm": 0.25,
            "max_hot_files": 4,
            "max_warm_files": 8,
            "max_chars": 25000
        }
    }


def load_attention_state(claude_dir: Path) -> dict:
    """Load current attention state."""
    state_path = claude_dir / "attn_state.json"
    if state_path.exists():
        with open(state_path, "r", encoding="utf-8") as f:
            return json.load(f)
    return {
        "scores": {},
        "turn_count": 0,
        "last_update": None
    }


def save_attention_state(claude_dir: Path, state: dict) -> None:
    """Save attention state."""
    state_path = claude_dir / "attn_state.json"
    state["last_update"] = datetime.now().isoformat()
    with open(state_path, "w", encoding="utf-8") as f:
        json.dump(state, f, indent=2)


def get_prompt_from_stdin() -> str:
    """Read prompt from stdin (provided by Claude Code hook)."""
    if not sys.stdin.isatty():
        try:
            input_data = sys.stdin.read()
            data = json.loads(input_data)
            return data.get("prompt", "")
        except (json.JSONDecodeError, KeyError):
            return input_data
    return ""


def find_keyword_matches(prompt: str, config: dict) -> List[str]:
    """Find agent files that match keywords in the prompt."""
    prompt_lower = prompt.lower()
    matched_files = []

    for file_path, keywords in config.get("keywords", {}).items():
        for keyword in keywords:
            # Use word boundaries for more precise matching
            pattern = r'\b' + re.escape(keyword.lower()) + r'\b'
            if re.search(pattern, prompt_lower):
                if file_path not in matched_files:
                    matched_files.append(file_path)
                break

    return matched_files


def apply_decay(scores: dict, config: dict) -> dict:
    """Apply decay to all attention scores."""
    decay_rates = config.get("decay_rates", {})
    default_decay = 0.75

    new_scores = {}
    for file_path, score in scores.items():
        decay_rate = decay_rates.get(file_path, default_decay)
        new_scores[file_path] = score * decay_rate

    return new_scores


def apply_activation(scores: dict, matched_files: List[str]) -> dict:
    """Apply activation to matched files (set to 1.0)."""
    for file_path in matched_files:
        scores[file_path] = 1.0
    return scores


def apply_co_activation(scores: dict, matched_files: List[str], config: dict) -> dict:
    """Apply co-activation boost to related files."""
    co_activation = config.get("co_activation", {})
    boost = 0.35

    for matched_file in matched_files:
        related_files = co_activation.get(matched_file, [])
        for related_file in related_files:
            current_score = scores.get(related_file, 0)
            scores[related_file] = min(1.0, current_score + boost)

    return scores


def apply_pinned(scores: dict, config: dict) -> dict:
    """Ensure pinned files stay above WARM threshold."""
    pinned = config.get("pinned", [])
    thresholds = config.get("thresholds", {})
    warm_threshold = thresholds.get("warm", 0.25)

    for file_path in pinned:
        current_score = scores.get(file_path, 0)
        scores[file_path] = max(current_score, warm_threshold + 0.1)

    return scores


def categorize_files(scores: dict, config: dict) -> Tuple[List[str], List[str], List[str]]:
    """Categorize files into HOT, WARM, COLD based on scores."""
    thresholds = config.get("thresholds", {})
    hot_threshold = thresholds.get("hot", 0.8)
    warm_threshold = thresholds.get("warm", 0.25)
    max_hot = thresholds.get("max_hot_files", 4)
    max_warm = thresholds.get("max_warm_files", 8)

    # Sort by score descending
    sorted_files = sorted(scores.items(), key=lambda x: x[1], reverse=True)

    hot_files = []
    warm_files = []
    cold_files = []

    for file_path, score in sorted_files:
        if score >= hot_threshold and len(hot_files) < max_hot:
            hot_files.append(file_path)
        elif score >= warm_threshold and len(warm_files) < max_warm:
            warm_files.append(file_path)
        else:
            cold_files.append(file_path)

    return hot_files, warm_files, cold_files


def read_file_content(claude_dir: Path, file_path: str, full: bool = True) -> str:
    """Read file content, optionally truncated to first 25 lines."""
    full_path = claude_dir / file_path
    if not full_path.exists():
        return ""

    try:
        with open(full_path, "r", encoding="utf-8") as f:
            if full:
                return f.read()
            else:
                lines = f.readlines()[:25]
                return "".join(lines)
    except Exception:
        return ""


def build_context_injection(
    claude_dir: Path,
    hot_files: List[str],
    warm_files: List[str],
    scores: dict,
    config: dict
) -> str:
    """Build the context string to inject."""
    thresholds = config.get("thresholds", {})
    max_chars = thresholds.get("max_chars", 25000)

    output_parts = []
    total_chars = 0

    # Header
    header = "## 🧠 Active Agent Context\n\n"
    output_parts.append(header)
    total_chars += len(header)

    # HOT files (full content)
    if hot_files:
        hot_header = "### 🔥 HOT (Full Context)\n\n"
        output_parts.append(hot_header)
        total_chars += len(hot_header)

        for file_path in hot_files:
            content = read_file_content(claude_dir, file_path, full=True)
            score = scores.get(file_path, 0)

            file_section = f"#### {file_path} (score: {score:.2f})\n\n{content}\n\n---\n\n"

            if total_chars + len(file_section) <= max_chars:
                output_parts.append(file_section)
                total_chars += len(file_section)

    # WARM files (headers only - first 25 lines)
    if warm_files:
        warm_header = "### 🌡️ WARM (Headers Only)\n\n"
        output_parts.append(warm_header)
        total_chars += len(warm_header)

        for file_path in warm_files:
            content = read_file_content(claude_dir, file_path, full=False)
            score = scores.get(file_path, 0)

            file_section = f"#### {file_path} (score: {score:.2f})\n\n{content}\n\n---\n\n"

            if total_chars + len(file_section) <= max_chars:
                output_parts.append(file_section)
                total_chars += len(file_section)

    # Footer with stats
    footer = f"\n*Context Router: {len(hot_files)} HOT, {len(warm_files)} WARM, {total_chars} chars*\n"
    output_parts.append(footer)

    return "".join(output_parts)


def log_attention_history(claude_dir: Path, state: dict, hot: List[str], warm: List[str], prompt: str) -> None:
    """Append to attention history for analytics."""
    history_path = claude_dir / "attention_history.jsonl"

    entry = {
        "turn": state.get("turn_count", 0),
        "timestamp": datetime.now().isoformat(),
        "prompt_preview": prompt[:100] if prompt else "",
        "hot": hot,
        "warm": warm,
        "scores_snapshot": {k: round(v, 2) for k, v in list(state.get("scores", {}).items())[:10]}
    }

    try:
        with open(history_path, "a", encoding="utf-8") as f:
            f.write(json.dumps(entry) + "\n")
    except Exception:
        pass  # Don't fail if history logging fails


def main():
    """Main context routing logic."""
    # Get directories
    claude_dir = get_claude_dir()

    if not claude_dir.exists():
        print("", end="")  # No output if no .claude directory
        return

    # Load configuration and state
    config = load_config(claude_dir)
    state = load_attention_state(claude_dir)

    # Get prompt from stdin
    prompt = get_prompt_from_stdin()

    if not prompt:
        print("", end="")
        return

    # Initialize scores if empty
    scores = state.get("scores", {})

    # Initialize all agent files with 0 if not present
    for file_path in config.get("keywords", {}).keys():
        if file_path not in scores:
            scores[file_path] = 0.0

    # Step 1: Apply decay to all scores
    scores = apply_decay(scores, config)

    # Step 2: Find keyword matches
    matched_files = find_keyword_matches(prompt, config)

    # Step 3: Apply activation to matched files
    scores = apply_activation(scores, matched_files)

    # Step 4: Apply co-activation
    scores = apply_co_activation(scores, matched_files, config)

    # Step 5: Apply pinned files
    scores = apply_pinned(scores, config)

    # Step 6: Categorize files
    hot_files, warm_files, cold_files = categorize_files(scores, config)

    # Step 7: Build context injection
    context = build_context_injection(claude_dir, hot_files, warm_files, scores, config)

    # Step 8: Update state
    state["scores"] = scores
    state["turn_count"] = state.get("turn_count", 0) + 1
    save_attention_state(claude_dir, state)

    # Step 9: Log history
    log_attention_history(claude_dir, state, hot_files, warm_files, prompt)

    # Output context (Claude Code will inject this)
    print(context)


if __name__ == "__main__":
    main()
