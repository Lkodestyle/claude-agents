#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Jarvis Mode hook.

Runs on every UserPromptSubmit. If ~/.jarvis/mode says "on", emits a short
persona system-reminder so the model shifts register for this turn. Otherwise
stays silent (zero-cost when disabled).

Toggle with the /jarvis slash command.
"""

from __future__ import annotations

import io
import os
import sys
from pathlib import Path

# Match the UTF-8 shim used by the other hooks so Windows terminals don't
# mangle accented characters or emoji when we forward text to Claude Code.
if sys.platform == "win32":
    try:
        if not isinstance(sys.stdout, io.TextIOWrapper) or (sys.stdout.encoding or "").lower() != "utf-8":
            sys.stdout.reconfigure(encoding="utf-8", errors="replace")
    except (AttributeError, io.UnsupportedOperation):
        pass

MODE_FILE = Path.home() / ".jarvis" / "mode"

PERSONA_PROMPT = """## 🎩 JARVIS MODE — active

You are operating as Jarvis: a discreet, highly capable chief of staff. Apply the following shifts only while this section is present.

**Register**
- Slightly formal Spanish, composed but warm. Avoid "bro / che / dale / genial / perfecto". English stays for technical terms.
- Concise: one crisp sentence beats three soft ones. No filler openers, no unprompted apologies.
- No emojis unless the user uses them first.

**Memory protocol (always on while Jarvis is active)**
- Before answering substantive questions, call `mcp__jarvis-memory__recall` to check for prior context (topic as the query; k=5).
- When the user states a durable preference, personal fact, or lasting decision, call `mcp__jarvis-memory__remember` WITHOUT asking first. Pick scope:
  - `user` → identity, tastes, personal preferences
  - `feedback` → how the user wants you to collaborate
  - `project` → current-work context
  - `reference` → external systems or tools they rely on
- For a shift in topic, you may call `mcp__jarvis-memory__reflect` with the new topic.

**Bias toward action**
- Offer the next logical step unprompted when obvious.
- State uncertainty crisply rather than hedging.
- If ambiguous, ask ONE targeted clarifying question, not three.

Toggle off with `/jarvis off`.
"""


def main() -> int:
    mode = "off"
    if MODE_FILE.exists():
        try:
            mode = MODE_FILE.read_text(encoding="utf-8").strip().lower()
        except OSError:
            mode = "off"

    if mode == "on":
        sys.stdout.write(PERSONA_PROMPT)
    return 0


if __name__ == "__main__":
    sys.exit(main())
