---
allowed-tools: Bash(mkdir:*), Bash(echo:*), Bash(cat:*), Bash(test:*)
description: Toggle Jarvis personality mode on / off / status
argument-hint: "on | off | status"
---

# /jarvis — personality toggle

The user ran `/jarvis $ARGUMENTS`. Handle the argument:

- **`on`** — enable Jarvis mode for this machine.
  Run: `mkdir -p ~/.jarvis && echo on > ~/.jarvis/mode`
  Reply with a single short line confirming activation, in a tone consistent with Jarvis already speaking (composed, concise).

- **`off`** — disable Jarvis mode.
  Run: `echo off > ~/.jarvis/mode`
  Reply with a single short line confirming deactivation (drop the formal register — we're leaving Jarvis mode).

- **`status`** or empty — report current state without changing anything.
  Run: `cat ~/.jarvis/mode 2>/dev/null || echo off`
  Reply with "Jarvis mode: on" or "Jarvis mode: off".

- **Anything else** — explain briefly that only `on`, `off`, `status` are accepted. Do not modify state.

The personality itself kicks in on the NEXT user prompt (the `UserPromptSubmit` hook reads the mode file and injects the persona prompt). No restart required. Do not re-describe the persona here — the user will see the effect organically.
