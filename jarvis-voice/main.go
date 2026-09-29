package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unicode"
)

const (
	appName    = "jarvis-voice"
	appVersion = "0.2.0"

	// Hard cap on history size for the api brain. Voice turns are typically
	// short, so 20 messages = ~10 user/assistant pairs is plenty.
	historyMax = 20

	// Memory recall settings for the api brain: few hits, decent similarity
	// floor so weak matches don't drag in irrelevant history.
	recallK             = 4
	recallMinSimilarity = 0.35

	// How long a spoken permission question waits for a yes/no.
	permissionTimeout = 45 * time.Second

	// Upper bound for the exit wrap-up (session summary to Obsidian/memory).
	wrapUpTimeout = 2 * time.Minute
)

// wrapUpPrompt is sent to the claude brain when the user quits, so the
// session log gets written even if they never said goodbye.
const wrapUpPrompt = `[jarvis-voice] The user just closed Jarvis (this message is automatic, they are no longer listening). Wrap up now without asking anything: following the memory rules in your CLAUDE.md, write this session's summary to Obsidian (what we talked about, decisions, pending items) and update your auto-memory only if something durable came up. If a summary for this session already exists, update it instead of creating another. Any decision, new fact about the user, or pending item is worth saving; skip only if the session was pure small talk. Finish with one short sentence saying what you saved.`

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

	// First Ctrl+C ends the conversation (and triggers the wrap-up), a second
	// one quits immediately.
	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigs
		stop()
		<-sigs
		fmt.Println("\nforced exit")
		os.Exit(130)
	}()

	fmt.Printf("%s v%s\n", appName, appVersion)
	fmt.Printf("  brain   : %s\n", cfg.Brain)
	fmt.Printf("  mode    : %s\n", cfg.Mode)
	if cfg.Model != "" {
		fmt.Printf("  model   : %s\n", cfg.Model)
	}
	if cfg.Mode != "text" {
		fmt.Printf("  voice   : %s\n", cfg.Voice)
		dev, err := ResolveInputDevice()
		if err != nil {
			log.Fatalf("input device: %v", err)
		}
		fmt.Printf("  mic     : %s\n", dev)
	}

	brain, err := buildBrain(cfg)
	if err != nil {
		log.Fatalf("brain: %v", err)
	}
	defer brain.Close()

	var listener *Listener
	if cfg.Mode == "live" {
		listener = NewListener(cfg)
	}
	var speaker *Speaker
	if cfg.Mode != "text" {
		var mute func(bool)
		if listener != nil {
			mute = listener.SetMuted
		}
		speaker = NewSpeaker(cfg, mute)
	}

	utterances := startInput(ctx, cfg, listener, stop)
	s := &session{ctx: ctx, brain: brain, speaker: speaker, ptt: cfg.Mode == "ptt"}
	s.run(utterances)
	if cfg.Brain == "claude" && cfg.WrapUp {
		s.wrapUp()
	}
	fmt.Println("bye")
}

func buildBrain(cfg *Config) (Brain, error) {
	if cfg.Brain == "claude" {
		b, err := newClaudeBrain(cfg)
		if err != nil {
			return nil, err
		}
		fmt.Printf("  workdir : %s\n", cfg.Workspace)
		if b.MemoryOn() {
			fmt.Println("  memory  : on (jarvis-memory MCP)")
		} else {
			fmt.Println("  memory  : off (set JARVIS_VOICE_MEMORY=on to enable)")
		}
		return b, nil
	}

	// api brain: memory is optional and degrades gracefully.
	var mem *MemoryClient
	if cfg.MemoryEnabled {
		mctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		m, err := connectMemory(mctx, cfg.MemoryBin)
		cancel()
		if err != nil {
			log.Printf("memory disabled: %v", err)
			fmt.Println("  memory  : unavailable (continuing without it)")
		} else {
			mem = m
			fmt.Println("  memory  : on (jarvis-memory)")
		}
	} else {
		fmt.Println("  memory  : off (set JARVIS_VOICE_MEMORY=on to enable)")
	}
	return newAPIBrain(cfg, mem), nil
}

// startInput wires the configured capture mode to a channel of utterances.
func startInput(ctx context.Context, cfg *Config, listener *Listener, stop func()) <-chan string {
	fmt.Println()
	switch cfg.Mode {
	case "live":
		fmt.Println("Listening — just talk. Enter pauses/resumes the mic, Ctrl+C quits.")
		go func() {
			if err := listener.Run(ctx); err != nil && ctx.Err() == nil {
				log.Printf("listener: %v", err)
				stop()
			}
		}()
		go func() {
			sc := bufio.NewScanner(os.Stdin)
			for sc.Scan() {
				if listener.TogglePause() {
					fmt.Println("  ⏸  mic paused (Enter to resume)")
				} else {
					fmt.Println("  ▶  listening")
				}
			}
		}()
		return listener.Utterances()

	case "ptt":
		out := make(chan string)
		fmt.Printf("Press Enter to record %ds. Ctrl+C quits.\n", cfg.RecordSec)
		go func() {
			sc := bufio.NewScanner(os.Stdin)
			for sc.Scan() {
				text, err := recordTurn(ctx, cfg)
				if err != nil {
					log.Printf("record: %v", err)
					continue
				}
				if text == "" {
					fmt.Println("  (no speech detected)")
					continue
				}
				out <- text
			}
			stop()
		}()
		return out

	default: // text
		out := make(chan string)
		fmt.Println("Type a message and press Enter. Ctrl+D quits.")
		go func() {
			sc := bufio.NewScanner(os.Stdin)
			for sc.Scan() {
				if t := strings.TrimSpace(sc.Text()); t != "" {
					out <- t
				}
			}
			stop()
		}()
		return out
	}
}

// recordTurn is the push-to-talk path: fixed window → batch STT.
func recordTurn(ctx context.Context, cfg *Config) (string, error) {
	wav := filepath.Join(cfg.WorkDir, fmt.Sprintf("mic-%s.wav", time.Now().Format("20060102-150405")))
	fmt.Printf("• recording %ds...\n", cfg.RecordSec)
	if err := recordWAV(ctx, wav, time.Duration(cfg.RecordSec)*time.Second); err != nil {
		return "", err
	}
	text, err := deepgramTranscribe(ctx, cfg.DeepgramKey, wav)
	return strings.TrimSpace(text), err
}

// session is the turn-taking state machine between the user's utterances and
// the brain's events.
type session struct {
	ctx     context.Context
	brain   Brain
	speaker *Speaker // nil in text mode
	ptt     bool

	busy     bool
	turns    int // user turns sent this run
	queued   []string
	perm     *BrainEvent // pending permission question
	permTime *time.Timer
}

func (s *session) run(utterances <-chan string) {
	s.promptIdle()
	for {
		var permC <-chan time.Time
		if s.permTime != nil {
			permC = s.permTime.C
		}

		select {
		case <-s.ctx.Done():
			fmt.Println()
			return

		case u := <-utterances:
			fmt.Printf("  vos   : %s\n", u)
			s.onUtterance(u)

		case ev := <-s.brain.Events():
			s.onEvent(ev)

		case <-permC:
			s.answerPermission(false, "No confirmation was received in time.")
			s.say("No escuché respuesta, así que no lo hago.")
		}
	}
}

func (s *session) onUtterance(u string) {
	switch {
	case s.perm != nil:
		switch classifyYesNo(u) {
		case answerYes:
			s.answerPermission(true, "")
		case answerNo:
			s.answerPermission(false, "The user said no: "+u)
		default:
			s.say("No te entendí. ¿Sí o no?")
		}

	case s.busy:
		if isStopCommand(u) {
			if err := s.brain.Interrupt(); err != nil {
				log.Printf("interrupt: %v", err)
			}
			s.queued = nil
			fmt.Println("  ⏹  interrupted")
			return
		}
		s.queued = append(s.queued, u)
		fmt.Println("  (queued until the current task finishes)")

	default:
		s.send(u)
	}
}

func (s *session) onEvent(ev BrainEvent) {
	switch ev.Kind {
	case EventText:
		fmt.Printf("  jarvis: %s\n", ev.Text)
		s.say(ev.Text)

	case EventTool:
		fmt.Printf("  ⚙  %s\n", ev.Text)

	case EventPermission:
		s.perm = &ev
		s.permTime = time.NewTimer(permissionTimeout)
		fmt.Printf("  ❓ permission: %s\n", ev.Prompt)
		s.say(fmt.Sprintf("Necesito tu ok para %s. ¿Lo hago?", ev.Prompt))

	case EventError:
		fmt.Printf("  ✖  %s\n", ev.Text)
		s.say("Tuve un problema con eso. Te dejé el detalle en la consola.")

	case EventDone:
		s.busy = false
		if s.perm != nil { // turn ended without an answer (e.g. interrupted)
			s.perm = nil
			s.stopPermTimer()
		}
		if len(s.queued) > 0 {
			next := strings.Join(s.queued, ". ")
			s.queued = nil
			s.send(next)
			return
		}
		s.promptIdle()
	}
}

func (s *session) send(text string) {
	if err := s.brain.Send(text); err != nil {
		fmt.Printf("  ✖  %v\n", err)
		return
	}
	s.busy = true
	s.turns++
	fmt.Println("• thinking...")
}

func (s *session) answerPermission(allow bool, reason string) {
	if s.perm == nil {
		return
	}
	id := s.perm.RequestID
	s.perm = nil
	s.stopPermTimer()
	if allow {
		fmt.Println("  ✔  allowed")
	} else {
		fmt.Println("  ✖  denied")
	}
	if err := s.brain.Respond(id, allow, reason); err != nil {
		log.Printf("permission response: %v", err)
	}
}

func (s *session) stopPermTimer() {
	if s.permTime != nil {
		s.permTime.Stop()
		s.permTime = nil
	}
}

func (s *session) say(text string) {
	s.sayCtx(s.ctx, text)
}

func (s *session) sayCtx(ctx context.Context, text string) {
	if s.speaker == nil {
		return
	}
	if err := s.speaker.Say(ctx, text); err != nil && ctx.Err() == nil {
		log.Printf("speak: %v", err)
	}
}

// wrapUp asks the brain to log the session before exiting. The user is gone
// by now, so permissions can't be asked out loud: only writes to Obsidian and
// to Claude Code's auto-memory are approved, everything else is denied.
func (s *session) wrapUp() {
	if s.turns == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), wrapUpTimeout)
	defer cancel()

	// Settle whatever was in flight when the user quit.
	if s.perm != nil {
		s.answerPermission(false, "The user closed Jarvis before answering.")
	}
	if s.busy {
		if err := s.brain.Interrupt(); err != nil {
			log.Printf("interrupt: %v", err)
		}
		if !s.drainUntilDone(ctx, false) {
			return
		}
	}

	fmt.Println("• saving session summary... (Ctrl+C again to skip)")
	s.sayCtx(ctx, "Dale, guardo el resumen de la sesión.")
	if err := s.brain.Send(wrapUpPrompt); err != nil {
		log.Printf("wrap-up: %v", err)
		return
	}
	s.drainUntilDone(ctx, true)
}

// drainUntilDone consumes brain events until the turn ends. With verbose it
// prints progress; permission requests go through wrapUpAllowed.
func (s *session) drainUntilDone(ctx context.Context, verbose bool) bool {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("  ✖  wrap-up timed out")
			return false
		case ev := <-s.brain.Events():
			switch ev.Kind {
			case EventText:
				if verbose {
					fmt.Printf("  jarvis: %s\n", ev.Text)
				}
			case EventTool:
				if verbose {
					fmt.Printf("  ⚙  %s\n", ev.Text)
				}
			case EventPermission:
				allow := wrapUpAllowed(ev.ToolName, ev.Input)
				reason := ""
				if !allow {
					reason = "Only Obsidian and auto-memory writes are allowed during the exit wrap-up."
				}
				if verbose {
					fmt.Printf("  %s %s\n", map[bool]string{true: "✔", false: "✖"}[allow], ev.Prompt)
				}
				if err := s.brain.Respond(ev.RequestID, allow, reason); err != nil {
					log.Printf("permission response: %v", err)
				}
			case EventError:
				// An interrupted turn always ends in an error result; only
				// report errors from the wrap-up turn itself.
				if verbose {
					fmt.Printf("  ✖  %s\n", ev.Text)
				}
			case EventDone:
				return true
			}
		}
	}
}

// wrapUpAllowed is the unattended allowlist for the exit wrap-up.
func wrapUpAllowed(tool string, input json.RawMessage) bool {
	switch tool {
	case "mcp__obsidian-vault__write_note", "mcp__obsidian-vault__patch_note",
		"mcp__obsidian-vault__update_frontmatter", "mcp__obsidian-vault__manage_tags":
		return true
	case "Write", "Edit":
		var in struct {
			FilePath string `json:"file_path"`
		}
		_ = json.Unmarshal(input, &in)
		p := filepath.ToSlash(filepath.Clean(in.FilePath))
		return strings.Contains(p, "/.claude/projects/") && strings.Contains(p, "/memory/") &&
			!strings.Contains(p, "..")
	}
	return false
}

func (s *session) promptIdle() {
	if s.ptt {
		fmt.Print("\n[Enter to talk] ")
	}
}

// --- spoken yes/no and stop detection ---

type answer int

const (
	answerUnknown answer = iota
	answerYes
	answerNo
)

var (
	yesWords = set("si", "dale", "ok", "okay", "okey", "hacelo", "adelante", "confirmo",
		"confirmado", "obvio", "claro", "bueno", "listo", "afirmativo", "yes", "sure", "va", "sale", "mandale")
	noWords = set("no", "nop", "nope", "cancela", "cancelalo", "negativo", "deja", "dejalo",
		"frena", "para", "stop", "nah", "tampoco")
	stopWords = set("para", "stop", "cancela", "cancelalo", "frena", "basta", "detente", "callate")
)

// classifyYesNo is deliberately conservative: any "no" word wins, so a
// garbled or hedged answer never approves an action.
func classifyYesNo(u string) answer {
	words := tokens(u)
	for _, w := range words {
		if noWords[w] {
			return answerNo
		}
	}
	for _, w := range words {
		if yesWords[w] {
			return answerYes
		}
	}
	return answerUnknown
}

// isStopCommand matches short commands like "pará" or "cancelá". Longer
// sentences are real requests ("para mañana agendame...") and get queued.
func isStopCommand(u string) bool {
	words := tokens(u)
	return len(words) > 0 && len(words) <= 3 && stopWords[words[0]]
}

func tokens(s string) []string {
	s = stripAccents(strings.ToLower(s))
	return strings.FieldsFunc(s, func(r rune) bool { return !unicode.IsLetter(r) })
}

var accentReplacer = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u")

func stripAccents(s string) string { return accentReplacer.Replace(s) }

func set(words ...string) map[string]bool {
	m := make(map[string]bool, len(words))
	for _, w := range words {
		m[w] = true
	}
	return m
}
