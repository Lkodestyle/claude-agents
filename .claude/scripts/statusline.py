#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Status line for Claude Code: context, cache, cost and git at a glance.

Claude Code pipes a JSON document on stdin on every update (see
https://code.claude.com/docs/en/statusline) and shows whatever we print.

Line 1:  Opus · claude-agents · main* · PR #11
Line 2:  ctx ████░░░░░░ 38% (76k/200k) · cache 91% warm · $0.42 · 12m · 5h 23% · 7d 41%

Everything is optional: fields that are absent (older versions, no git, no
cache stats yet) are simply skipped, and any exception prints a minimal line
instead of breaking the UI.
"""

import json
import os
import subprocess
import sys

RESET = "\033[0m"
DIM = "\033[2m"
GREEN = "\033[32m"
YELLOW = "\033[33m"
RED = "\033[31m"
CYAN = "\033[36m"


def color_for(pct: float) -> str:
    if pct >= 85:
        return RED
    if pct >= 60:
        return YELLOW
    return GREEN


def bar(pct: float, width: int = 10) -> str:
    filled = max(0, min(width, round(pct / 100 * width)))
    return "█" * filled + "░" * (width - filled)


def fmt_tokens(n) -> str:
    try:
        n = int(n)
    except (TypeError, ValueError):
        return "?"
    return f"{n / 1000:.0f}k" if n >= 1000 else str(n)


def fmt_duration(ms) -> str:
    try:
        s = int(ms) // 1000
    except (TypeError, ValueError):
        return ""
    return f"{s // 3600}h{(s % 3600) // 60:02d}m" if s >= 3600 else f"{s // 60}m"


def git_branch(cwd: str) -> str:
    try:
        out = subprocess.run(
            ["git", "-C", cwd, "status", "--porcelain=v1", "--branch"],
            capture_output=True, text=True, timeout=1,
        ).stdout.splitlines()
    except (OSError, subprocess.SubprocessError):
        return ""
    if not out or not out[0].startswith("## "):
        return ""
    head = out[0][3:]
    branch = head.split("...")[0].split(" ")[0]
    dirty = "*" if len(out) > 1 else ""
    ahead = ""
    if "[ahead" in head:
        ahead = " ↑" + head.split("[ahead ")[1].split("]")[0].split(",")[0]
    return f"{branch}{dirty}{ahead}"


def main() -> None:
    data = json.load(sys.stdin)

    # ---- line 1: model · project · branch · PR
    parts = [f"{CYAN}{data.get('model', {}).get('display_name', '?')}{RESET}"]
    ws = data.get("workspace", {})
    cwd = ws.get("current_dir") or data.get("cwd") or os.getcwd()
    parts.append(os.path.basename(cwd.rstrip("/")) or cwd)
    if ws.get("git_worktree"):
        parts.append(f"wt:{ws['git_worktree']}")
    branch = git_branch(cwd)
    if branch:
        parts.append(branch)
    pr = data.get("pr", {})
    if pr.get("number"):
        parts.append(f"PR #{pr['number']}")
    if data.get("fast_mode"):
        parts.append(f"{YELLOW}fast{RESET}")
    effort = data.get("effort", {}).get("level")
    if effort and effort != "high":
        parts.append(f"effort:{effort}")
    print(f" {DIM}·{RESET} ".join(parts))

    # ---- line 2: context · cache · cost · time · rate limits
    cw = data.get("context_window", {})
    pct = cw.get("used_percentage")
    segs = []
    if pct is not None:
        used = cw.get("total_input_tokens")
        size = cw.get("context_window_size")
        detail = f" ({fmt_tokens(used)}/{fmt_tokens(size)})" if used and size else ""
        segs.append(f"ctx {color_for(pct)}{bar(pct)} {pct:.0f}%{RESET}{detail}")

    pc = data.get("prompt_cache")
    if pc:
        ratio = pc.get("hit_ratio")
        if ratio is not None:
            state = "warm" if pc.get("warm") else "cold"
            c = GREEN if ratio >= 0.8 else YELLOW if ratio >= 0.5 else RED
            segs.append(f"cache {c}{ratio * 100:.0f}%{RESET} {state}")

    cost = data.get("cost", {})
    if cost.get("total_cost_usd") is not None:
        segs.append(f"${cost['total_cost_usd']:.2f}")
    dur = fmt_duration(cost.get("total_duration_ms"))
    if dur:
        segs.append(dur)

    rl = data.get("rate_limits", {})
    for key, label in (("five_hour", "5h"), ("seven_day", "7d")):
        p = rl.get(key, {}).get("used_percentage")
        if p is not None:
            segs.append(f"{label} {color_for(p)}{p:.0f}%{RESET}")

    if segs:
        print(f" {DIM}·{RESET} ".join(segs))


if __name__ == "__main__":
    try:
        main()
    except Exception:  # never break the UI over a status line
        print("claude")
