package main

import (
	"encoding/json"
	"testing"
)

func TestClassifyYesNo(t *testing.T) {
	cases := map[string]answer{
		"Sí, dale.":          answerYes,
		"Dale":               answerYes,
		"ok hacelo":          answerYes,
		"No.":                answerNo,
		"No, cancelá.":       answerNo,
		"sí... no, mejor no": answerNo, // any "no" wins
		"eh, ¿qué dijiste?":  answerUnknown,
		"":                   answerUnknown,
	}
	for in, want := range cases {
		if got := classifyYesNo(in); got != want {
			t.Errorf("classifyYesNo(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestIsStopCommand(t *testing.T) {
	cases := map[string]bool{
		"Pará.":                            true,
		"Cancelá eso":                      true,
		"stop":                             true,
		"para mañana agendame una reunión": false,
		"decime el clima":                  false,
	}
	for in, want := range cases {
		if got := isStopCommand(in); got != want {
			t.Errorf("isStopCommand(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestDescribeAction(t *testing.T) {
	mk := func(tool string, input map[string]any) permissionRequest {
		raw, _ := json.Marshal(input)
		return permissionRequest{Subtype: "can_use_tool", ToolName: tool, Input: raw}
	}
	cases := []struct {
		req  permissionRequest
		want string
	}{
		{mk("Write", map[string]any{"file_path": "/tmp/x/hola.txt"}), "escribir el archivo hola.txt"},
		{mk("Bash", map[string]any{"command": "ls", "description": "List files"}), "correr un comando para list files"},
		{mk("mcp__claude_ai_Gmail__send_message", map[string]any{"to": "a@b.com", "subject": "Hola"}),
			"usar send message en Gmail: a@b.com, Hola"},
	}
	for _, c := range cases {
		if got := describeAction(c.req); got != c.want {
			t.Errorf("describeAction(%s) = %q, want %q", c.req.ToolName, got, c.want)
		}
	}
}

func TestSpeakable(t *testing.T) {
	got := speakable("**Listo**, mirá https://example.com/x `code`")
	want := "Listo , mirá el link code"
	if got != want {
		t.Errorf("speakable = %q, want %q", got, want)
	}
}

func TestWrapUpAllowed(t *testing.T) {
	path := func(p string) json.RawMessage {
		raw, _ := json.Marshal(map[string]string{"file_path": p})
		return raw
	}
	cases := []struct {
		tool  string
		input json.RawMessage
		want  bool
	}{
		{"mcp__obsidian-vault__write_note", nil, true},
		{"mcp__obsidian-vault__delete_note", nil, false},
		{"Write", path("/home/u/.claude/projects/-home-u--jarvis-workspace/memory/x.md"), true},
		{"Write", path("/home/u/.claude/projects/p/memory/../../../../.bashrc"), false},
		{"Write", path("/home/u/.bashrc"), false},
		{"Bash", nil, false},
		{"mcp__claude_ai_Gmail__send_message", nil, false},
	}
	for _, c := range cases {
		if got := wrapUpAllowed(c.tool, c.input); got != c.want {
			t.Errorf("wrapUpAllowed(%s, %s) = %v, want %v", c.tool, c.input, got, c.want)
		}
	}
}
