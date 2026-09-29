# jarvis-voice

Hands-free voice front-end for the optional Jarvis mode:
mic → Deepgram live STT → **Claude Code (headless)** → Edge TTS → speaker.

The brain is a long-lived `claude -p` session driven over the stream-json
control protocol (the same one the Agent SDKs use). Jarvis gets everything your
Claude Code already has: global MCPs (Calendar, Gmail, Notion, Obsidian...),
subagents, skills, hooks, web search and files. It runs on your Claude login,
not on an API key.

## Stack

| Layer | Component | Cost |
|-------|-----------|------|
| Audio I/O | ffmpeg / ffplay subprocess | $0 (system tool) |
| STT | Deepgram live WebSocket (`nova-3`), gated by a local voice detector | only speech is streamed, silence is never billed |
| Brain | Claude Code CLI (`claude -p`, stream-json) | your Claude plan usage |
| TTS | Microsoft Edge Read-Aloud (WSS) | $0, no API key |

Pure-Go, no CGO. Single binary cross-compiles for Windows / Linux / macOS.

## Prerequisites

- Go 1.25+
- ffmpeg with ffplay in `PATH`
- Claude Code installed and logged in (`claude` in `PATH`)
- `DEEPGRAM_API_KEY` in the repo-root `.env` (gitignored)

## Build & run

```bash
cd jarvis-voice
go build -o ../bin/jarvis-voice .
../bin/jarvis-voice
```

Just talk. Jarvis answers when you pause (~1.2s). While it speaks the mic is
muted so it never hears itself. **Enter** pauses/resumes the mic, **Ctrl+C** quits.

## How turns work

- **Actions need your spoken OK.** Read-only tools (files, web, calendar/mail
  reads, memory) run freely. Anything else — writing files, Bash, sending mail,
  creating events — makes Jarvis ask "Necesito tu ok para ... ¿Lo hago?".
  Answer "sí / dale" or "no". Any "no" in the answer wins; silence for 45s = no.
- **Stop a running task** by saying a short "pará" / "cancelá" / "stop".
- Anything else you say while it's working is queued for the next turn.

## Workspace & persona

The Claude Code session runs in `~/.jarvis/workspace`. On first run a
`CLAUDE.md` persona is created there — edit it freely, it's never overwritten.
`~/proyectos` (if it exists) is added as a readable directory.

## Configuration

| Var | Default | What it does |
|-----|---------|--------------|
| `JARVIS_VOICE_MODE` | `live` | `live` (hands-free), `ptt` (Enter + fixed window), `text` (type, no audio — for debugging) |
| `JARVIS_VOICE_BRAIN` | `claude` | `claude` (Claude Code, full tools) or `api` (tool-less Messages API chat, needs `ANTHROPIC_API_KEY`) |
| `JARVIS_VOICE_MODEL` | Claude Code default | Model id / alias passed to the brain |
| `JARVIS_VOICE` | `es-AR-TomasNeural` | Edge TTS voice |
| `JARVIS_VOICE_LANG` | `multi` | Deepgram language (`es`, `en`, `multi`...) |
| `JARVIS_VOICE_VAD_RMS` | `0` | Voice gate level. `0` = auto-calibrate at startup, `-1` = stream everything |
| `JARVIS_VOICE_WORKSPACE` | `~/.jarvis/workspace` | Claude Code working dir |
| `JARVIS_VOICE_ADD_DIRS` | `~/proyectos` | Extra readable dirs, comma-separated |
| `JARVIS_VOICE_ALLOWED_TOOLS` | read-only set | Tools that never need confirmation, comma-separated (replaces the default) |
| `JARVIS_VOICE_PERMISSION_MODE` | `manual` | Claude Code permission mode for the session |
| `JARVIS_VOICE_CONTINUE` | off | `on` = resume the last Jarvis conversation |
| `JARVIS_VOICE_MEMORY` | off | `on` = give the brain the `jarvis-memory` MCP |
| `JARVIS_VOICE_RECORD_SEC` | `8` | Recording window in `ptt` mode |
| `JARVIS_VOICE_INPUT` / `JARVIS_VOICE_INPUT_NAME` | auto | Mic override (full ffmpeg spec / name substring) |
| `JARVIS_VOICE_WORKDIR` | `~/.jarvis/voice` | Temp audio + `claude-stderr.log` |

`ANTHROPIC_API_KEY` is stripped from the Claude Code child process on purpose:
if inherited, Claude Code would bill the API and disable the claude.ai connectors.

### Recommended voices

- `es-AR-TomasNeural` — Argentine Spanish, male, neutral (default)
- `es-AR-ElenaNeural` — Argentine Spanish, female
- `es-MX-JorgeNeural` — Mexican Spanish, broadly understood across LatAm
- `en-US-GuyNeural` — closest to "Iron Man's Jarvis" feel, English
- `en-GB-RyanNeural` — British English, more on-the-nose Jarvis cosplay

## Roadmap

- **Done** — fixed-window loop (A), semantic memory (B), Claude Code brain with
  spoken permissions + hands-free listening (C).
- **Next** — barge-in (interrupt Jarvis while it talks, needs echo cancellation
  or headphones), local spoken reminders (systemd timers), wake word, streaming
  TTS for lower latency.
