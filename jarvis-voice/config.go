package main

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config holds everything the voice loop needs at runtime. Fields are sourced
// from environment variables — the .env loader runs in main() before this is
// called, so a repo-root .env populates anything not already set in the shell.
type Config struct {
	DeepgramKey  string
	AnthropicKey string

	// Brain selects who answers: "claude" (default) drives a headless Claude
	// Code session with all its tools, MCPs and subagents; "api" is the
	// original tool-less Messages API chat. Override with JARVIS_VOICE_BRAIN.
	Brain string

	// Mode selects how speech is captured: "live" (default) listens
	// continuously and ends a turn when you stop talking; "ptt" is the
	// original Enter + fixed window; "text" reads turns from stdin and prints
	// replies (handy for debugging the brain without a mic).
	Mode string

	// Model id. Empty with the claude brain = whatever Claude Code defaults
	// to. Override with JARVIS_VOICE_MODEL.
	Model string

	// Edge TTS voice id. See https://learn.microsoft.com/en-us/azure/ai-services/speech-service/language-support
	// Some good defaults:
	//   es-AR-TomasNeural   (Argentine Spanish, male)
	//   es-AR-ElenaNeural   (Argentine Spanish, female)
	//   es-MX-JorgeNeural   (Mexican Spanish, male — clear and neutral)
	//   en-US-GuyNeural     (American English, male — closest to "Jarvis")
	Voice string

	// Recording duration in seconds for ptt mode. Override with
	// JARVIS_VOICE_RECORD_SEC.
	RecordSec int

	// Working directory for transient audio files (mic recordings + TTS output).
	WorkDir string

	// Semantic memory via the jarvis-memory MCP server.
	// Opt-in — off unless JARVIS_VOICE_MEMORY is "on"/"true"/"1".
	MemoryEnabled bool

	// Explicit path to the jarvis-memory binary (JARVIS_MEMORY_BIN).
	// Empty = auto-discover next to our executable, then PATH.
	MemoryBin string

	// System prompt for the api brain. The claude brain uses voiceRules
	// (appended to Claude Code's own system prompt) instead.
	SystemPrompt string

	// --- claude brain ---

	// Claude Code executable (JARVIS_CLAUDE_BIN, default "claude" from PATH).
	ClaudeBin string

	// Directory the Claude Code session runs in. Holds the editable persona
	// in CLAUDE.md. Override with JARVIS_VOICE_WORKSPACE.
	Workspace string

	// Extra directories Claude may read without asking (JARVIS_VOICE_ADD_DIRS,
	// comma-separated). Defaults to ~/proyectos when it exists.
	AddDirs []string

	// Claude Code permission mode (JARVIS_VOICE_PERMISSION_MODE). "manual"
	// makes every write/exec/MCP action come back to us for a spoken
	// confirmation, regardless of the mode set in the user's global settings.
	PermissionMode string

	// Tools that never need confirmation (JARVIS_VOICE_ALLOWED_TOOLS,
	// comma-separated, replaces the default read-only list).
	AllowedTools []string

	// Which permission requests are asked out loud (JARVIS_VOICE_CONFIRM):
	// "sensitive" (default), "all" or "none". See policy.go.
	ConfirmPolicy string

	// Silence that ends a turn, in ms (JARVIS_VOICE_PAUSE_MS). Longer is
	// more forgiving with thinking pauses, shorter answers faster.
	PauseMs int

	// Beep when the mic reopens after a question (JARVIS_VOICE_EARCON,
	// on by default; "off" disables).
	Earcon bool

	// Continue the most recent Jarvis session instead of starting fresh
	// (JARVIS_VOICE_CONTINUE=on).
	Continue bool

	// Ask the brain to write the session summary on exit
	// (JARVIS_VOICE_WRAPUP, on by default; "off" disables).
	WrapUp bool

	// --- live listening ---

	// RMS level (int16 scale) that opens the local voice gate. 0 = calibrate
	// from ambient noise at startup. -1 = no gate, stream everything.
	// Override with JARVIS_VOICE_VAD_RMS.
	VADThreshold int
}

// voiceRules are appended to Claude Code's system prompt. They only cover how
// to talk; who Jarvis is lives in the workspace CLAUDE.md so the user can
// edit it without rebuilding.
const voiceRules = `You are running as a VOICE assistant. Everything you write as text is converted to speech and played to the user. Rules:
- Reply in the user's language (default: Spanish, casual rioplatense voseo). Friendly and respectful: no insults or crude slang (never "boludo", "che boludo", etc.).
- 1 to 3 short sentences per reply. No markdown, no bullet lists, no code blocks, no URLs, no file paths read out in full.
- Before a tool call that may take a while, say in a few words what you are about to do ("Dale, reviso tu calendario.").
- Do NOT ask for permission or confirmation in text before using tools: the system approves routine actions silently and asks the user out loud for sensitive ones (sending, deleting, pushing, infra changes). Just act. Still ask a clarifying question when the request itself is unclear.
- If a tool call is denied with the user's words, do what those words say instead; don't ask the same thing again.
- Write the Bash tool "description" in the user's language, as a short verb phrase ("borrar la carpeta build"): it is read aloud when confirmation is needed.
- After working, report only the outcome, not the steps.
- If you need information, ask ONE short, specific question.
- Never use the AskUserQuestion tool; ask in plain text instead.
- The user may be misheard by speech-to-text: if a request is ambiguous or sounds garbled, confirm what you understood before acting.
- Read long numbers in small groups so they are understandable when spoken.`

const defaultSystemPrompt = `Sos Jarvis, el asistente personal del usuario. Tus respuestas se reproducen por audio, asi que aplica estas reglas SIEMPRE:

- Maximo 1 a 3 oraciones cortas. La voz no aguanta parrafos.
- Sin rellenos: nada de "claro!", "por supuesto", "perfecto". Empieza con la respuesta.
- Sin listas con bullets ni markdown. Es voz, no texto.
- Tono compuesto pero calido, en espanol casual neutro. Frases simples.
- Si necesitas mas info, hace UNA sola pregunta corta y especifica.
- Si no sabes algo, decilo en una oracion. No fabriques datos.
- Para numeros largos, leelos en grupos chicos para que se entienda al hablar.

Hablas en nombre del proyecto Jarvis del usuario. Si te pide algo del codigo o sistema, contesta a alto nivel y sugeri que abra Claude Code para detalles.`

// defaultAllowedTools are read-only / low-risk tools that run without a
// spoken confirmation. Anything else (writes, Bash, sending mail, creating
// events...) is asked out loud first.
var defaultAllowedTools = []string{
	"Read", "Glob", "Grep", "WebSearch", "WebFetch", "Task", "Agent", "TodoWrite",
	"mcp__jarvis-memory__recall",
	"mcp__jarvis-memory__remember",
	"mcp__claude_ai_Google_Calendar__list_events",
	"mcp__claude_ai_Google_Calendar__search_events",
	"mcp__claude_ai_Google_Calendar__get_event",
	"mcp__claude_ai_Google_Calendar__list_calendars",
	"mcp__claude_ai_Google_Calendar__suggest_time",
	"mcp__claude_ai_Gmail__search_threads",
	"mcp__claude_ai_Gmail__get_thread",
	"mcp__claude_ai_Gmail__get_message",
	"mcp__claude_ai_Notion__notion-search",
	"mcp__claude_ai_Notion__notion-fetch",
	"mcp__obsidian-vault__read_note",
	"mcp__obsidian-vault__read_multiple_notes",
	"mcp__obsidian-vault__search_notes",
	"mcp__obsidian-vault__list_directory",
	"mcp__context7__resolve-library-id",
	"mcp__context7__query-docs",
}

func loadConfig() (*Config, error) {
	home, _ := os.UserHomeDir()
	cfg := &Config{
		DeepgramKey:    os.Getenv("DEEPGRAM_API_KEY"),
		AnthropicKey:   os.Getenv("ANTHROPIC_API_KEY"),
		Brain:          strings.ToLower(getenv("JARVIS_VOICE_BRAIN", "claude")),
		Mode:           strings.ToLower(getenv("JARVIS_VOICE_MODE", "live")),
		Model:          os.Getenv("JARVIS_VOICE_MODEL"),
		Voice:          getenv("JARVIS_VOICE", "es-AR-TomasNeural"),
		RecordSec:      getenvInt("JARVIS_VOICE_RECORD_SEC", 8),
		WorkDir:        getenv("JARVIS_VOICE_WORKDIR", defaultWorkDir()),
		SystemPrompt:   defaultSystemPrompt,
		MemoryEnabled:  getenvBool("JARVIS_VOICE_MEMORY"),
		MemoryBin:      os.Getenv("JARVIS_MEMORY_BIN"),
		ClaudeBin:      getenv("JARVIS_CLAUDE_BIN", "claude"),
		Workspace:      getenv("JARVIS_VOICE_WORKSPACE", filepath.Join(home, ".jarvis", "workspace")),
		PermissionMode: getenv("JARVIS_VOICE_PERMISSION_MODE", "manual"),
		AllowedTools:   splitList(os.Getenv("JARVIS_VOICE_ALLOWED_TOOLS")),
		VADThreshold:   getenvInt("JARVIS_VOICE_VAD_RMS", 0),
		Continue:       getenvBool("JARVIS_VOICE_CONTINUE"),
		ConfirmPolicy:  strings.ToLower(getenv("JARVIS_VOICE_CONFIRM", "sensitive")),
		PauseMs:        getenvInt("JARVIS_VOICE_PAUSE_MS", 1500),
		Earcon:         os.Getenv("JARVIS_VOICE_EARCON") != "off",
		WrapUp:         os.Getenv("JARVIS_VOICE_WRAPUP") != "off",
	}

	if len(cfg.AllowedTools) == 0 {
		cfg.AllowedTools = defaultAllowedTools
	}
	if dirs := splitList(os.Getenv("JARVIS_VOICE_ADD_DIRS")); len(dirs) > 0 {
		cfg.AddDirs = dirs
	} else if p := filepath.Join(home, "proyectos"); dirExists(p) {
		cfg.AddDirs = []string{p}
	}

	switch cfg.Brain {
	case "claude", "api":
	default:
		return nil, errors.New("JARVIS_VOICE_BRAIN must be 'claude' or 'api'")
	}
	switch cfg.ConfirmPolicy {
	case "sensitive", "all", "none":
	default:
		return nil, errors.New("JARVIS_VOICE_CONFIRM must be 'sensitive', 'all' or 'none'")
	}
	if cfg.PauseMs < 500 {
		cfg.PauseMs = 500
	}
	switch cfg.Mode {
	case "live", "ptt", "text":
	default:
		return nil, errors.New("JARVIS_VOICE_MODE must be 'live', 'ptt' or 'text'")
	}
	if cfg.Brain == "api" && cfg.Model == "" {
		cfg.Model = "claude-sonnet-5-5"
	}

	var missing []string
	if cfg.DeepgramKey == "" && cfg.Mode != "text" {
		missing = append(missing, "DEEPGRAM_API_KEY")
	}
	if cfg.AnthropicKey == "" && cfg.Brain == "api" {
		missing = append(missing, "ANTHROPIC_API_KEY")
	}
	if len(missing) > 0 {
		return nil, errors.New("missing required env vars: " + strings.Join(missing, ", "))
	}
	return cfg, nil
}

func defaultWorkDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".jarvis", "voice")
	}
	return filepath.Join(home, ".jarvis", "voice")
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvBool(key string) bool {
	switch os.Getenv(key) {
	case "on", "true", "1":
		return true
	}
	return false
}

func getenvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func splitList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func dirExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}
