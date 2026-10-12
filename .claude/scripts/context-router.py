#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Context Router v3 - Agent suggestions from prompt keywords

Runs on every UserPromptSubmit. Scores each agent from the keywords in the
prompt (plus co-activation and a per-turn decay so recent topics linger) and
emits ONE short line naming the agents worth delegating to, e.g.:

    [context-router] Suggested agents: terraform, aws · related: networking · skill: infrastructure-as-code

It never injects the agent files themselves: Claude Code already gives each
subagent its own prompt, so repeating it in the main context only burns tokens.
Set "output": "full" in keywords.json to get the old HOT/WARM content injection.

State lives under ~/.claude/router/<project>.json (never inside the project).
"""

import io
import json
import os
import re
import sys
import unicodedata
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Tuple

# Windows terminals: make sure we emit UTF-8 regardless of the console codepage.
if sys.platform == "win32":
    for stream in (sys.stdout, sys.stderr):
        try:
            if not isinstance(stream, io.TextIOWrapper) or (stream.encoding or "").lower() != "utf-8":
                stream.reconfigure(encoding="utf-8", errors="replace")
        except (AttributeError, io.UnsupportedOperation):
            pass

DEFAULTS = {
    "thresholds": {"hot": 0.8, "warm": 0.25, "max_hot": 4, "max_warm": 4, "max_chars": 25000},
    "co_activation_boost": 0.35,
    "default_decay": 0.75,
    "output": "compact",
    "history": False,
}


# --------------------------------------------------------------------------
# Locations
# --------------------------------------------------------------------------

def project_root() -> Path:
    """Project dir: what Claude Code says, else the nearest .claude/ upwards."""
    env = os.environ.get("CLAUDE_PROJECT_DIR")
    if env:
        return Path(env)
    cwd = Path.cwd()
    for candidate in (cwd, *cwd.parents):
        if (candidate / ".claude").exists():
            return candidate
    return cwd


def global_claude_dir() -> Path:
    env = os.environ.get("CONTEXT_DOCS_ROOT")
    return Path(env) if env else Path.home() / ".claude"


def find_config() -> Tuple[Path, Path]:
    """Return (keywords.json path, dir holding agents/).

    A project may ship its own keywords.json to override the global one; the
    agents folder is only needed for "full" output mode.
    """
    for base in (project_root() / ".claude", global_claude_dir()):
        cfg = base / "keywords.json"
        if cfg.exists():
            return cfg, base
    return global_claude_dir() / "keywords.json", global_claude_dir()


def state_path() -> Path:
    """One state file per project, kept OUT of the project tree."""
    root = project_root()
    slug = re.sub(r"[^A-Za-z0-9]+", "-", str(root)).strip("-")[-80:] or "root"
    return global_claude_dir() / "router" / f"{slug}.json"


# --------------------------------------------------------------------------
# IO
# --------------------------------------------------------------------------

def load_json(path: Path, fallback):
    try:
        with open(path, "r", encoding="utf-8") as fh:
            return json.load(fh)
    except (OSError, json.JSONDecodeError):
        return fallback


def save_json(path: Path, data) -> None:
    try:
        path.parent.mkdir(parents=True, exist_ok=True)
        with open(path, "w", encoding="utf-8") as fh:
            json.dump(data, fh, indent=2)
    except OSError:
        pass  # state is a nice-to-have; never fail the hook over it


def read_prompt() -> str:
    if sys.stdin.isatty():
        return ""
    raw = sys.stdin.read()
    try:
        return json.loads(raw).get("prompt", "")
    except (json.JSONDecodeError, AttributeError):
        return raw


def normalize(text: str) -> str:
    """Lowercase and strip accents so 'aplicación' matches 'aplicacion'."""
    text = unicodedata.normalize("NFKD", text.lower())
    return "".join(ch for ch in text if not unicodedata.combining(ch))


def agent_name(file_path: str) -> str:
    return Path(file_path).stem


# --------------------------------------------------------------------------
# Scoring
# --------------------------------------------------------------------------

def match_keywords(prompt: str, config: dict) -> List[str]:
    prompt = normalize(prompt)
    matched = []
    for file_path, keywords in config.get("keywords", {}).items():
        for keyword in keywords:
            if re.search(r"\b" + re.escape(normalize(keyword)) + r"\b", prompt):
                matched.append(file_path)
                break
    return matched


def update_scores(scores: Dict[str, float], matched: List[str], config: dict) -> Dict[str, float]:
    decay_rates = config.get("decay_rates", {})
    default_decay = config.get("default_decay", DEFAULTS["default_decay"])
    boost = config.get("co_activation_boost", DEFAULTS["co_activation_boost"])

    for file_path in config.get("keywords", {}):
        scores.setdefault(file_path, 0.0)

    scores = {k: v * decay_rates.get(k, default_decay) for k, v in scores.items()}

    for file_path in matched:
        scores[file_path] = 1.0
    for file_path in matched:
        for related in config.get("co_activation", {}).get(file_path, []):
            scores[related] = min(1.0, scores.get(related, 0.0) + boost)

    warm = config.get("thresholds", {}).get("warm", DEFAULTS["thresholds"]["warm"])
    for file_path in config.get("pinned", []):
        scores[file_path] = max(scores.get(file_path, 0.0), warm + 0.1)

    return scores


def categorize(scores: Dict[str, float], config: dict) -> Tuple[List[str], List[str]]:
    th = {**DEFAULTS["thresholds"], **config.get("thresholds", {})}
    # Backwards compatibility with the v2 key names.
    max_hot = th.get("max_hot_files", th["max_hot"])
    max_warm = th.get("max_warm_files", th["max_warm"])

    hot, warm = [], []
    for file_path, score in sorted(scores.items(), key=lambda kv: kv[1], reverse=True):
        if score >= th["hot"] and len(hot) < max_hot:
            hot.append(file_path)
        elif th["warm"] <= score < th["hot"] and len(warm) < max_warm:
            warm.append(file_path)
    return hot, warm


def suggest_skills(hot: List[str], config: dict) -> List[str]:
    """A workflow skill is suggested when its anchor agent (first in the list)
    and at least one other of its agents are HOT."""
    hot_names = {agent_name(f) for f in hot}
    return [skill for skill, agents in config.get("skills", {}).items()
            if agents and agents[0] in hot_names and len(hot_names & set(agents)) >= 2]


# --------------------------------------------------------------------------
# Output
# --------------------------------------------------------------------------

def compact_output(hot: List[str], warm: List[str], skills: List[str]) -> str:
    if not hot and not warm:
        return ""
    parts = []
    if hot:
        parts.append("Suggested agents: " + ", ".join(agent_name(f) for f in hot))
    if warm:
        label = "related: " if hot else "Possibly relevant agents: "
        parts.append(label + ", ".join(agent_name(f) for f in warm))
    if skills:
        parts.append("skill: " + ", ".join(skills))
    return "[context-router] " + " · ".join(parts)


def read_agent(base: Path, file_path: str, full: bool) -> str:
    try:
        with open(base / file_path, "r", encoding="utf-8") as fh:
            return fh.read() if full else "".join(fh.readlines()[:25])
    except OSError:
        return ""


def full_output(base: Path, hot: List[str], warm: List[str], scores: Dict[str, float], config: dict) -> str:
    """Legacy v2 behaviour: inject agent files (HOT = full, WARM = header)."""
    max_chars = config.get("thresholds", {}).get("max_chars", DEFAULTS["thresholds"]["max_chars"])
    out = ["## 🧠 Active Agent Context\n\n"]
    total = len(out[0])
    for title, files, full in (("### 🔥 HOT (Full Context)\n\n", hot, True),
                               ("### 🌡️ WARM (Headers Only)\n\n", warm, False)):
        if not files:
            continue
        out.append(title)
        total += len(title)
        for file_path in files:
            section = f"#### {file_path} (score: {scores.get(file_path, 0):.2f})\n\n{read_agent(base, file_path, full)}\n\n---\n\n"
            if total + len(section) <= max_chars:
                out.append(section)
                total += len(section)
    out.append(f"\n*Context Router: {len(hot)} HOT, {len(warm)} WARM, {total} chars*\n")
    return "".join(out)


def log_history(state_file: Path, turn: int, prompt: str, hot: List[str], warm: List[str]) -> None:
    entry = {"turn": turn, "timestamp": datetime.now().isoformat(),
             "prompt_preview": prompt[:100], "hot": hot, "warm": warm}
    try:
        with open(state_file.with_suffix(".history.jsonl"), "a", encoding="utf-8") as fh:
            fh.write(json.dumps(entry) + "\n")
    except OSError:
        pass


# --------------------------------------------------------------------------
# Main
# --------------------------------------------------------------------------

def main() -> None:
    # Jarvis voice sessions must skip the router: no agent-context injection
    if os.environ.get("JARVIS_VOICE") == "1":
        return

    prompt = read_prompt()
    if not prompt.strip():
        return

    cfg_path, base = find_config()
    config = {**DEFAULTS, **load_json(cfg_path, {})}
    if not config.get("keywords"):
        return

    state_file = state_path()
    state = load_json(state_file, {"scores": {}, "turn_count": 0})

    matched = match_keywords(prompt, config)
    scores = update_scores(state.get("scores", {}), matched, config)
    hot, warm = categorize(scores, config)

    state.update({"scores": scores, "turn_count": state.get("turn_count", 0) + 1,
                  "last_update": datetime.now().isoformat()})
    save_json(state_file, state)
    if config.get("history"):
        log_history(state_file, state["turn_count"], prompt, hot, warm)

    if config.get("output") == "full":
        output = full_output(base, hot, warm, scores, config)
    else:
        output = compact_output(hot, warm, suggest_skills(hot, config))
    if output:
        print(output)


if __name__ == "__main__":
    main()
