package main

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"time"
)

// Speaker turns text into audio (Edge TTS → ffplay). It mutes the listener
// for the duration of playback plus a short tail, so Jarvis never hears
// itself through the speakers.
type Speaker struct {
	cfg   *Config
	mute  func(bool) // nil when there is no live listener
	count atomic.Int64
}

// echoTail keeps the mic muted briefly after playback: speaker/driver buffers
// still flush audio after ffplay exits.
const echoTail = 350 * time.Millisecond

func NewSpeaker(cfg *Config, mute func(bool)) *Speaker {
	return &Speaker{cfg: cfg, mute: mute}
}

func (s *Speaker) Say(ctx context.Context, text string) error {
	text = speakable(text)
	if text == "" {
		return nil
	}
	n := s.count.Add(1)
	path := filepath.Join(s.cfg.WorkDir, fmt.Sprintf("reply-%s-%d.mp3", time.Now().Format("150405"), n))
	if err := edgeTTS(ctx, s.cfg.Voice, text, path); err != nil {
		return fmt.Errorf("tts: %w", err)
	}

	if s.mute != nil {
		s.mute(true)
		defer func() {
			time.Sleep(echoTail)
			s.mute(false)
		}()
	}
	if err := playMP3(ctx, path); err != nil {
		return fmt.Errorf("play: %w", err)
	}
	return nil
}

var (
	reURL      = regexp.MustCompile(`https?://\S+`)
	reMarkdown = regexp.MustCompile("[*_`#>|]+")
	reSpaces   = regexp.MustCompile(`\s+`)
)

// speakable strips what sounds bad when read aloud: markdown symbols and
// raw URLs (the model is told not to produce them, but it sometimes does).
func speakable(s string) string {
	s = reURL.ReplaceAllString(s, "el link")
	s = reMarkdown.ReplaceAllString(s, " ")
	return strings.TrimSpace(reSpaces.ReplaceAllString(s, " "))
}
