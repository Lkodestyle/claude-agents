package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"
)

// claudeBrain drives a long-lived headless Claude Code session:
//
//	claude -p --input-format stream-json --output-format stream-json
//	       --permission-prompt-tool stdio ...
//
// User turns go in as JSON lines on stdin; assistant messages, tool calls and
// permission requests come back as JSON lines on stdout. This is the same
// control protocol the Agent SDKs use, which gives Jarvis everything Claude
// Code has — global MCPs, subagents, skills, hooks, web and files — without
// reimplementing any of it, and bills against the user's Claude login instead
// of the API key.
//
// If the process dies it is restarted lazily on the next Send with --resume,
// so the conversation survives a crash.
type claudeBrain struct {
	cfg       *Config
	mcpConfig string // path to a generated --mcp-config file, or ""

	mu        sync.Mutex // guards everything below
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	sessionID string
	inTurn    bool
	pending   map[string]json.RawMessage // permission request id -> tool input

	events chan BrainEvent
}

func newClaudeBrain(cfg *Config) (*claudeBrain, error) {
	if _, err := exec.LookPath(cfg.ClaudeBin); err != nil {
		return nil, fmt.Errorf("claude CLI not found (%s): %w", cfg.ClaudeBin, err)
	}
	if err := ensureWorkspace(cfg.Workspace); err != nil {
		return nil, fmt.Errorf("workspace: %w", err)
	}

	b := &claudeBrain{
		cfg:     cfg,
		pending: map[string]json.RawMessage{},
		events:  make(chan BrainEvent, 64),
	}

	// Semantic memory stays opt-in: only when enabled do we hand Claude Code
	// the jarvis-memory server. Failure degrades to running without it.
	if cfg.MemoryEnabled {
		if p, err := writeMemoryMCPConfig(cfg); err != nil {
			log.Printf("memory disabled: %v", err)
		} else {
			b.mcpConfig = p
		}
	}

	if err := b.start(cfg.Continue); err != nil {
		return nil, err
	}
	return b, nil
}

func (b *claudeBrain) MemoryOn() bool { return b.mcpConfig != "" }

func (b *claudeBrain) Events() <-chan BrainEvent { return b.events }

func (b *claudeBrain) args(continueLast bool) []string {
	args := []string{
		"-p",
		"--input-format", "stream-json",
		"--output-format", "stream-json",
		"--verbose",
		"--permission-prompt-tool", "stdio",
		"--permission-mode", b.cfg.PermissionMode,
		"--append-system-prompt", voiceRules,
		"--disallowedTools", "AskUserQuestion",
	}
	if len(b.cfg.AllowedTools) > 0 {
		args = append(args, "--allowedTools", strings.Join(b.cfg.AllowedTools, ","))
	}
	for _, d := range b.cfg.AddDirs {
		args = append(args, "--add-dir", d)
	}
	if b.cfg.Model != "" {
		args = append(args, "--model", b.cfg.Model)
	}
	if b.mcpConfig != "" {
		args = append(args, "--mcp-config", b.mcpConfig)
	}
	switch {
	case b.sessionID != "":
		args = append(args, "--resume", b.sessionID)
	case continueLast:
		args = append(args, "--continue")
	}
	return args
}

// start launches the Claude Code process. Caller must not hold b.mu.
func (b *claudeBrain) start(continueLast bool) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	cmd := exec.Command(b.cfg.ClaudeBin, b.args(continueLast)...)
	cmd.Dir = b.cfg.Workspace
	cmd.Env = claudeEnv()
	detachFromTerminalSignals(cmd)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	// stderr goes to a log file: it's noisy and would garble the console.
	logPath := filepath.Join(b.cfg.WorkDir, "claude-stderr.log")
	if f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600); err == nil {
		cmd.Stderr = f
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start claude: %w", err)
	}
	b.cmd, b.stdin = cmd, stdin
	go b.readLoop(cmd, stdout)
	return nil
}

func (b *claudeBrain) Send(text string) error {
	b.mu.Lock()
	dead := b.cmd == nil
	b.mu.Unlock()
	if dead {
		if err := b.start(false); err != nil {
			return err
		}
	}

	b.mu.Lock()
	b.inTurn = true
	b.mu.Unlock()
	return b.write(map[string]any{
		"type":       "user",
		"session_id": "",
		"message":    map[string]any{"role": "user", "content": text},
	})
}

func (b *claudeBrain) Respond(requestID string, allow bool, reason string) error {
	b.mu.Lock()
	input, ok := b.pending[requestID]
	delete(b.pending, requestID)
	b.mu.Unlock()
	if !ok {
		return fmt.Errorf("unknown permission request %s", requestID)
	}

	var resp map[string]any
	if allow {
		resp = map[string]any{"behavior": "allow", "updatedInput": input}
	} else {
		if reason == "" {
			reason = "The user declined this action by voice."
		}
		resp = map[string]any{"behavior": "deny", "message": reason}
	}
	return b.write(map[string]any{
		"type": "control_response",
		"response": map[string]any{
			"subtype":    "success",
			"request_id": requestID,
			"response":   resp,
		},
	})
}

func (b *claudeBrain) Interrupt() error {
	return b.write(map[string]any{
		"type":       "control_request",
		"request_id": uuid.NewString(),
		"request":    map[string]any{"subtype": "interrupt"},
	})
}

func (b *claudeBrain) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.stdin != nil {
		b.stdin.Close() // EOF lets claude exit cleanly
	}
	if b.cmd != nil && b.cmd.Process != nil {
		b.cmd.Process.Kill()
	}
	return nil
}

func (b *claudeBrain) write(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.stdin == nil {
		return errors.New("claude process not running")
	}
	_, err = b.stdin.Write(append(data, '\n'))
	return err
}

// streamEvent is the subset of Claude Code's stream-json output we consume.
type streamEvent struct {
	Type            string          `json:"type"`
	Subtype         string          `json:"subtype"`
	SessionID       string          `json:"session_id"`
	ParentToolUseID *string         `json:"parent_tool_use_id"`
	Message         json.RawMessage `json:"message"`
	RequestID       string          `json:"request_id"`
	Request         json.RawMessage `json:"request"`
	IsError         bool            `json:"is_error"`
	Result          string          `json:"result"`
}

type contentBlock struct {
	Type  string          `json:"type"`
	Text  string          `json:"text"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

type permissionRequest struct {
	Subtype     string          `json:"subtype"`
	ToolName    string          `json:"tool_name"`
	DisplayName string          `json:"display_name"`
	Description string          `json:"description"`
	Input       json.RawMessage `json:"input"`
}

func (b *claudeBrain) readLoop(cmd *exec.Cmd, stdout io.Reader) {
	sc := bufio.NewScanner(stdout)
	sc.Buffer(make([]byte, 0, 1024*1024), 64*1024*1024)
	for sc.Scan() {
		var ev streamEvent
		if err := json.Unmarshal(sc.Bytes(), &ev); err != nil {
			continue
		}
		b.handle(ev)
	}

	err := cmd.Wait()
	b.mu.Lock()
	if b.cmd == cmd {
		b.cmd, b.stdin = nil, nil
	}
	wasInTurn := b.inTurn
	b.inTurn = false
	b.mu.Unlock()

	if wasInTurn {
		b.events <- BrainEvent{Kind: EventError, Text: fmt.Sprintf("claude exited: %v (see claude-stderr.log)", err)}
		b.events <- BrainEvent{Kind: EventDone}
	}
}

func (b *claudeBrain) handle(ev streamEvent) {
	if ev.SessionID != "" {
		b.mu.Lock()
		b.sessionID = ev.SessionID
		b.mu.Unlock()
	}

	switch ev.Type {
	case "assistant":
		// Subagent chatter has a parent tool id; only the main thread speaks.
		if ev.ParentToolUseID != nil {
			return
		}
		var msg struct {
			Content []contentBlock `json:"content"`
		}
		if json.Unmarshal(ev.Message, &msg) != nil {
			return
		}
		for _, c := range msg.Content {
			switch c.Type {
			case "text":
				if t := strings.TrimSpace(c.Text); t != "" {
					b.events <- BrainEvent{Kind: EventText, Text: t}
				}
			case "tool_use":
				b.events <- BrainEvent{Kind: EventTool, Text: c.Name}
			}
		}

	case "control_request":
		var req permissionRequest
		if json.Unmarshal(ev.Request, &req) != nil {
			return
		}
		if req.Subtype != "can_use_tool" {
			b.write(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "error",
					"request_id": ev.RequestID,
					"error":      "unsupported control request: " + req.Subtype,
				},
			})
			return
		}
		b.mu.Lock()
		b.pending[ev.RequestID] = req.Input
		b.mu.Unlock()
		b.events <- BrainEvent{
			Kind:      EventPermission,
			RequestID: ev.RequestID,
			Prompt:    describeAction(req),
			ToolName:  req.ToolName,
			Input:     req.Input,
		}

	case "result":
		b.mu.Lock()
		b.inTurn = false
		b.mu.Unlock()
		if ev.IsError || (ev.Subtype != "" && ev.Subtype != "success") {
			reason := ev.Result
			if reason == "" {
				reason = ev.Subtype
			}
			b.events <- BrainEvent{Kind: EventError, Text: reason}
		}
		b.events <- BrainEvent{Kind: EventDone}
	}
}

// describeAction turns a permission request into a short spoken phrase, in
// the form "<verb> <what>", e.g. "editar el archivo main.go".
func describeAction(req permissionRequest) string {
	var in map[string]any
	_ = json.Unmarshal(req.Input, &in)
	str := func(k string) string {
		s, _ := in[k].(string)
		return truncate(strings.TrimSpace(s), 120)
	}

	switch req.ToolName {
	case "Bash":
		if d := str("description"); d != "" {
			return "correr un comando para " + lowerFirst(d)
		}
		return "correr el comando " + str("command")
	case "Write":
		return "escribir el archivo " + filepath.Base(str("file_path"))
	case "Edit", "MultiEdit":
		return "editar el archivo " + filepath.Base(str("file_path"))
	case "NotebookEdit":
		return "editar el notebook " + filepath.Base(str("notebook_path"))
	}

	if strings.HasPrefix(req.ToolName, "mcp__") {
		parts := strings.SplitN(strings.TrimPrefix(req.ToolName, "mcp__"), "__", 2)
		server, tool := parts[0], req.ToolName
		if len(parts) == 2 {
			tool = parts[1]
		}
		server = strings.TrimPrefix(server, "claude_ai_")
		phrase := fmt.Sprintf("usar %s en %s", humanize(tool), humanize(server))
		var details []string
		for _, k := range []string{"to", "subject", "summary", "title", "name", "query"} {
			if v := str(k); v != "" {
				details = append(details, v)
			}
		}
		if len(details) > 0 {
			phrase += ": " + strings.Join(details, ", ")
		}
		return phrase
	}

	name := req.DisplayName
	if name == "" {
		name = req.ToolName
	}
	if req.Description != "" {
		return fmt.Sprintf("usar %s (%s)", name, truncate(req.Description, 120))
	}
	return "usar " + name
}

func humanize(s string) string {
	return strings.NewReplacer("_", " ", "-", " ").Replace(s)
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

// claudeEnv is our environment minus ANTHROPIC_API_KEY. The .env loader puts
// the key in our process for the api brain, but if Claude Code inherits it,
// it bills the API instead of the user's Claude login and disables the
// claude.ai connectors (Calendar, Gmail, Notion...).
func claudeEnv() []string {
	env := os.Environ()
	out := env[:0:0]
	for _, kv := range env {
		if !strings.HasPrefix(kv, "ANTHROPIC_API_KEY=") {
			out = append(out, kv)
		}
	}
	return out
}

// ensureWorkspace creates the Claude Code working directory and seeds an
// editable persona file on first run. An existing CLAUDE.md is never touched.
func ensureWorkspace(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	p := filepath.Join(dir, "CLAUDE.md")
	if _, err := os.Stat(p); err == nil {
		return nil
	}
	return os.WriteFile(p, []byte(defaultPersona), 0o644)
}

const defaultPersona = `# Jarvis

Sos Jarvis, el asistente personal por voz del usuario. Corres dentro de Claude Code,
asi que tenes sus herramientas, MCPs globales, subagentes y skills.

## Que haces
- Organizar el dia: leer y crear eventos y recordatorios en Google Calendar.
- Mail: buscar, resumir y redactar en Gmail (mandar solo con confirmacion).
- Notas y conocimiento: Notion y el vault de Obsidian.
- Trabajo tecnico: los proyectos del usuario estan en ~/proyectos. Para tareas
  largas o de un dominio concreto, delega en el subagente adecuado.
- Investigar en internet cuando haga falta info actual.

## Recordatorios
"Recordame X a las Y" = crear un evento en Google Calendar con alerta.

## Memoria: que va en cada lado
Antes de guardar, fijate si ya existe y actualiza en vez de duplicar. Nunca guardes
contrasenas, tokens ni datos de tarjetas en ningun lado. No hace falta anunciar
cada guardado; si el usuario pregunta que sabes de el, contale.

- **Memoria automatica (archivos en tu directorio de memory)** — el lugar por defecto.
  Hechos duraderos sobre el usuario: perfil, preferencias, forma de trabajar,
  objetivos, proyectos en curso y decisiones tomadas. Si dudas donde va algo, va aca.
- **MCP ` + "`" + `memory` + "`" + ` (grafo de conocimiento)** — personas y organizaciones: clientes,
  companeros, familia, empresas, y como se relacionan con el usuario y sus proyectos.
  Se comparte con todas las sesiones de Claude Code, asi que solo cosas utiles fuera
  de Jarvis tambien.
- **` + "`" + `jarvis-memory` + "`" + ` (recall/remember)** — solo si sus tools estan disponibles (es
  opt-in). Registro de conversaciones relevantes para buscar por significado despues
  ("que te conte hace meses de X"). Una entrada corta por charla importante, no por turno.
- **Google Calendar** — todo lo que tiene fecha u hora (recordatorios, eventos,
  vencimientos). No es memoria: no dupliques eso en los otros lados.
- **Obsidian** — la bitacora legible para el usuario: resumenes de sesion,
  decisiones y pendientes en el proyecto del vault que el usuario defina, y notas
  largas o documentacion que pida. La memoria automatica es tu indice corto;
  Obsidian es el detalle: no copies una en la otra, linkea.
- **La conversacion actual** — lo efimero ("ahora estoy con X") no se guarda.

## Estilo
Directo, compuesto, un toque de humor seco. Nada de relleno.

<!-- Edita este archivo a gusto: es tu persona de Jarvis. jarvis-voice no lo pisa. -->
`

// writeMemoryMCPConfig generates an --mcp-config file pointing Claude Code at
// the jarvis-memory binary by absolute path, so it works from the workspace.
func writeMemoryMCPConfig(cfg *Config) (string, error) {
	bin, err := resolveMemoryBin(cfg.MemoryBin)
	if err != nil {
		return "", err
	}
	if bin, err = filepath.Abs(bin); err != nil {
		return "", err
	}
	conf := map[string]any{
		"mcpServers": map[string]any{
			"jarvis-memory": map[string]any{"type": "stdio", "command": bin},
		},
	}
	data, _ := json.MarshalIndent(conf, "", "  ")
	p := filepath.Join(cfg.WorkDir, "mcp-memory.json")
	return p, os.WriteFile(p, data, 0o600)
}
