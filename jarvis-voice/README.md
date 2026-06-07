# jarvis-voice

Push-to-talk Go CLI that closes the voice loop for the optional Jarvis
mode: capture mic → Deepgram STT → Claude → Edge TTS → speaker.

This is **Milestone A** — the audio loop only. Memory integration with
`jarvis-memory` is Milestone B.

## Stack

| Layer | Component | Cost |
|-------|-----------|------|
| Audio I/O | ffmpeg / ffplay subprocess | $0 (system tool) |
| STT | Deepgram REST API (`nova-3` model) | $200 free credit on signup, ~$0.0043/min after |
| LLM | Anthropic Messages API (Sonnet by default) | pay-per-use |
| TTS | Microsoft Edge Read-Aloud (WSS) | $0 forever, no API key |

Pure-Go, no CGO. Single binary cross-compiles for Windows / Linux / macOS.

## Prerequisites

- Go 1.25+
- ffmpeg with ffplay in `PATH` (`winget install ffmpeg` on Windows)
- Three keys in the repo-root `.env` (gitignored):
  ```
  DEEPGRAM_API_KEY=...
  ANTHROPIC_API_KEY=sk-ant-api...
  # VOYAGE_API_KEY only needed for jarvis-memory, not voice
  ```

## Build & run

```bash
cd jarvis-voice
go mod tidy
go build -o ../bin/jarvis-voice .
../bin/jarvis-voice
```

You'll see:

```
jarvis-voice v0.1.0
  workdir : ~/.jarvis/voice
  model   : claude-sonnet-4-5-20250929
  voice   : es-AR-TomasNeural
  record  : 8s (fixed window)

Press Enter to record one turn. Ctrl+C to quit.

[Enter to talk]
```

Press Enter, speak for 8 seconds, then listen.

## Configuration

All env vars are optional except the two API keys.

| Var | Default | What it does |
|-----|---------|--------------|
| `JARVIS_VOICE_MODEL` | `claude-sonnet-4-5-20250929` | Anthropic model id |
| `JARVIS_VOICE` | `es-AR-TomasNeural` | Edge TTS voice. See [voice list](https://learn.microsoft.com/en-us/azure/ai-services/speech-service/language-support?tabs=tts) |
| `JARVIS_VOICE_RECORD_SEC` | `8` | Fixed recording window (seconds) |
| `JARVIS_VOICE_WORKDIR` | `~/.jarvis/voice` | Where temp `.wav` and `.mp3` files go |

### Recommended voices

- `es-AR-TomasNeural` — Argentine Spanish, male, neutral (default)
- `es-AR-ElenaNeural` — Argentine Spanish, female
- `es-MX-JorgeNeural` — Mexican Spanish, broadly understood across LatAm
- `en-US-GuyNeural` — closest to "Iron Man's Jarvis" feel, English
- `en-GB-RyanNeural` — British English, more on-the-nose Jarvis cosplay

## Roadmap

- **Milestone A (this)** — fixed-window recording, no memory.
- **Milestone B** — connect to `jarvis-memory` via the MCP client SDK so
  spoken interactions share the same long-term memory as text sessions.
- **Later** — silence detection (auto-stop), wake word ("Hey Jarvis"),
  streaming TTS for lower perceived latency.
