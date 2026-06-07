package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	appName    = "jarvis-voice"
	appVersion = "0.1.0"

	// Hard cap on history size to keep prompts bounded. Voice turns are
	// typically short, so 20 messages = ~10 user/assistant pairs is plenty.
	historyMax = 20
)

func main() {
	log.SetOutput(os.Stderr)
	loadDotEnv()

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if err := os.MkdirAll(cfg.WorkDir, 0o755); err != nil {
		log.Fatalf("workdir: %v", err)
	}

	fmt.Printf("%s v%s\n", appName, appVersion)
	fmt.Printf("  workdir : %s\n", cfg.WorkDir)
	fmt.Printf("  model   : %s\n", cfg.Model)
	fmt.Printf("  voice   : %s\n", cfg.Voice)
	fmt.Printf("  record  : %ds (fixed window)\n", cfg.RecordSec)

	// Resolve the input device once up front so the user sees which mic we
	// picked before they hit Enter for the first time. Failure here is fatal —
	// no point continuing if we can't capture audio.
	if dev, err := ResolveInputDevice(); err != nil {
		log.Fatalf("input device: %v", err)
	} else {
		fmt.Printf("  mic     : %s\n", dev)
	}

	fmt.Println()
	fmt.Println("Press Enter to record one turn. Ctrl+C to quit.")

	scanner := bufio.NewScanner(os.Stdin)
	history := []Message{}

	for {
		fmt.Print("\n[Enter to talk] ")
		if !scanner.Scan() {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		if err := turn(ctx, cfg, &history); err != nil {
			log.Printf("turn error: %v", err)
		}
		cancel()
	}
}

// turn runs one round of the conversation: record → STT → LLM → TTS → play.
// It mutates history so the next turn carries context.
func turn(ctx context.Context, cfg *Config, history *[]Message) error {
	ts := time.Now().Format("20060102-150405")
	micWav := filepath.Join(cfg.WorkDir, fmt.Sprintf("mic-%s.wav", ts))
	replyMP3 := filepath.Join(cfg.WorkDir, fmt.Sprintf("reply-%s.mp3", ts))

	fmt.Printf("• recording %ds...\n", cfg.RecordSec)
	if err := recordWAV(ctx, micWav, time.Duration(cfg.RecordSec)*time.Second); err != nil {
		return fmt.Errorf("record: %w", err)
	}

	fmt.Println("• transcribing...")
	transcript, err := deepgramTranscribe(ctx, cfg.DeepgramKey, micWav)
	if err != nil {
		return fmt.Errorf("stt: %w", err)
	}
	transcript = strings.TrimSpace(transcript)
	if transcript == "" {
		fmt.Println("  (no speech detected, skipping)")
		return nil
	}
	fmt.Printf("  you   : %s\n", transcript)

	*history = append(*history, Message{Role: "user", Content: transcript})

	fmt.Println("• thinking...")
	reply, err := anthropicChat(ctx, cfg, *history)
	if err != nil {
		return fmt.Errorf("llm: %w", err)
	}
	fmt.Printf("  jarvis: %s\n", reply)

	*history = append(*history, Message{Role: "assistant", Content: reply})
	*history = trimHistory(*history, historyMax)

	fmt.Println("• synthesizing voice...")
	if err := edgeTTS(ctx, cfg.Voice, reply, replyMP3); err != nil {
		return fmt.Errorf("tts: %w", err)
	}

	fmt.Println("• playing...")
	if err := playMP3(ctx, replyMP3); err != nil {
		return fmt.Errorf("play: %w", err)
	}
	return nil
}

// trimHistory keeps the latest N messages, preserving the user/assistant pair
// boundary so we don't ship Claude an orphaned assistant turn.
func trimHistory(h []Message, max int) []Message {
	if len(h) <= max {
		return h
	}
	start := len(h) - max
	// Make sure we don't start on an assistant turn (Anthropic requires the
	// first message to be from the user).
	if h[start].Role != "user" && start+1 < len(h) {
		start++
	}
	return h[start:]
}
