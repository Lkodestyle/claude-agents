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

On **Ctrl+C** Jarvis first writes the session summary (Obsidian + auto-memory,
following the persona's memory rules) and then exits; press Ctrl+C again to
skip it. The user is gone at that point, so only Obsidian and auto-memory
writes are auto-approved — anything else is denied. Disable with
`JARVIS_VOICE_WRAPUP=off`.

## How turns work

- **Only sensitive actions need your spoken OK.** Routine work — reading,
  searching, writing notes and project files, creating calendar events, drafts,
  everyday commands — runs silently. Jarvis asks "¿Confirmás ...?" only for
  sending mail, deleting anything, `git push`/`reset --hard`, `rm`, `sudo`,
  infra changes (terraform apply, kubectl, aws/az mutations...), package
  installs, and writes to dotfiles, `.env` or outside your home (the workspace
  and `~/proyectos` are always fine). Tune with `JARVIS_VOICE_CONFIRM`.
- **Answering.** A short beep marks the moment the mic reopens. Your first
  words decide: "sí / dale..." approves (anything you add after is passed on
  as an instruction), "no..." denies, and a longer reply without either
  ("mandalo a Juan en vez de Pedro") is taken as a correction. No answer: it
  asks once more after 20s, then gives up.
- **Pauses.** A turn ends after 1.5s of silence; if you trail off on "y...",
  "o sea...", "que...", Jarvis waits a bit longer for the rest.
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
| `JARVIS_VOICE_CONFIRM` | `sensitive` | `sensitive` (ask only for risky actions), `all` (ask for everything), `none` |
| `JARVIS_VOICE_PAUSE_MS` | `1500` | Silence that ends your turn |
| `JARVIS_VOICE_EARCON` | on | `off` = no beep when the mic reopens after a question |
| `JARVIS_VOICE_CONTINUE` | off | `on` = resume the last Jarvis conversation |
| `JARVIS_VOICE_WRAPUP` | on | `off` = skip the session summary on exit |
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
