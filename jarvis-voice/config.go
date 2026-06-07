package main

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
)

// Config holds everything the voice loop needs at runtime. Fields are sourced
// from environment variables — the .env loader runs in main() before this is
// called, so a repo-root .env populates anything not already set in the shell.
type Config struct {
	DeepgramKey  string
	AnthropicKey string

	// Anthropic model id. Override with JARVIS_VOICE_MODEL.
	Model string

	// Edge TTS voice id. See https://learn.microsoft.com/en-us/azure/ai-services/speech-service/language-support
	// Some good defaults:
	//   es-AR-TomasNeural   (Argentine Spanish, male)
	//   es-AR-ElenaNeural   (Argentine Spanish, female)
	//   es-MX-JorgeNeural   (Mexican Spanish, male — clear and neutral)
	//   en-US-GuyNeural     (American English, male — closest to "Jarvis")
	Voice string

	// Recording duration in seconds. Fixed window for the MVP; we'll add
	// silence-trim later. Override with JARVIS_VOICE_RECORD_SEC.
	RecordSec int

	// Working directory for transient audio files (mic recordings + TTS output).
	WorkDir string

	// System prompt sent to Claude. Designed for spoken responses: short,
	// composed, no filler. The persona stays close to /jarvis mode in text.
	SystemPrompt string
}

const defaultSystemPrompt = `Sos Jarvis, el asistente personal del usuario. Tus respuestas se reproducen por audio, asi que aplica estas reglas SIEMPRE:

- Maximo 1 a 3 oraciones cortas. La voz no aguanta parrafos.
- Sin rellenos: nada de "claro!", "por supuesto", "perfecto". Empieza con la respuesta.
- Sin listas con bullets ni markdown. Es voz, no texto.
- Tono compuesto pero calido, en espanol casual neutro. Frases simples.
- Si necesitas mas info, hace UNA sola pregunta corta y especifica.
- Si no sabes algo, decilo en una oracion. No fabriques datos.
- Para numeros largos, leelos en grupos chicos para que se entienda al hablar.

Hablas en nombre del proyecto Jarvis del usuario. Si te pide algo del codigo o sistema, contesta a alto nivel y sugeri que abra Claude Code para detalles.`

func loadConfig() (*Config, error) {
	cfg := &Config{
		DeepgramKey:  os.Getenv("DEEPGRAM_API_KEY"),
		AnthropicKey: os.Getenv("ANTHROPIC_API_KEY"),
		Model:        getenv("JARVIS_VOICE_MODEL", "claude-sonnet-4-5-20250929"),
		Voice:        getenv("JARVIS_VOICE", "es-AR-TomasNeural"),
		RecordSec:    getenvInt("JARVIS_VOICE_RECORD_SEC", 8),
		WorkDir:      getenv("JARVIS_VOICE_WORKDIR", defaultWorkDir()),
		SystemPrompt: defaultSystemPrompt,
	}

	var missing []string
	if cfg.DeepgramKey == "" {
		missing = append(missing, "DEEPGRAM_API_KEY")
	}
	if cfg.AnthropicKey == "" {
		missing = append(missing, "ANTHROPIC_API_KEY")
	}
	if len(missing) > 0 {
		return nil, errors.New("missing required env vars: " + joinComma(missing))
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

func getenvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func joinComma(s []string) string {
	out := ""
	for i, v := range s {
		if i > 0 {
			out += ", "
		}
		out += v
	}
	return out
}
