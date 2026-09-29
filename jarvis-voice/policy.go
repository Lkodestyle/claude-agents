package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Confirmation policy. Every tool call outside Claude Code's allowlist comes
// back to us as a permission request; instead of asking out loud each time,
// Jarvis approves routine work silently and only asks for actions that are
// hard to undo or act on the user's behalf in the outside world.
//
// JARVIS_VOICE_CONFIRM selects the policy:
//   sensitive (default) — ask only for sensitiveAction() matches
//   all                 — ask for every request (the old behaviour)
//   none                — never ask (not recommended)

// safeRoots are directories where file writes never need confirmation: the
// Jarvis workspace and the extra dirs it was given. Set once at startup.
var safeRoots []string

// needsConfirmation decides whether a permission request is spoken to the
// user or approved silently.
func needsConfirmation(policy, tool string, input json.RawMessage) bool {
	switch policy {
	case "all":
		return true
	case "none":
		return false
	}
	return sensitiveAction(tool, input)
}

// Tool-name fragments that always need a spoken OK: sending things on the
// user's behalf, deleting, moving money.
var sensitiveToolWords = []string{
	"send", "reply", "forward", "delete", "trash", "remove", "spam",
	"comprar", "vender", "operar", "rescatar", "suscribir", "transfer",
}

// Bash commands that destroy data, publish work, escalate privileges or
// change infrastructure.
var sensitiveBash = regexp.MustCompile(`(?i)(^|[\s;&|(])(` + strings.Join([]string{
	`rm\s`, `rmdir\s`, `shred\s`, `dd\s`, `mkfs`, `truncate\s`,
	`sudo\s`, `su\s`, `doas\s`, `chmod\s+-R`, `chown\s+-R`,
	`git\s+push`, `git\s+reset\s+--hard`, `git\s+clean`, `git\s+branch\s+-D`, `git\s+rebase`,
	`gh\s+(pr\s+merge|release|repo\s+delete)`,
	`terraform\s+(apply|destroy|import|state\s+rm)`, `terraspace\s+(up|down|all\s+up)`, `terragrunt\s+(apply|destroy|run-all)`,
	`kubectl\s+(delete|apply|drain|scale|rollout)`, `helm\s+(install|upgrade|uninstall|delete)`,
	`aws\s+\S+\s+(delete|terminate|remove|put|create|update|stop)`, `az\s+\S+.*\s(delete|create|update|stop)`,
	`docker\s+(rm|rmi|system\s+prune|volume\s+rm|push)`, `docker\s+compose\s+down\s+-v`,
	`shutdown`, `reboot`, `poweroff`, `systemctl\s+(stop|disable|mask)`,
	`curl\s[^|]*\|\s*(sh|bash)`, `wget\s[^|]*\|\s*(sh|bash)`,
	`npm\s+publish`, `pip\s+install\s+--user`, `pacman\s+-(S|R)`, `paru\s+-(S|R)`, `yay\s+-(S|R)`,
	`crontab\s+-r`,
}, "|") + `)`)

func sensitiveAction(tool string, input json.RawMessage) bool {
	var in map[string]any
	_ = json.Unmarshal(input, &in)
	str := func(k string) string { s, _ := in[k].(string); return s }

	switch tool {
	case "Bash":
		return sensitiveBash.MatchString(str("command"))
	case "Write", "Edit", "MultiEdit", "NotebookEdit":
		p := str("file_path")
		if p == "" {
			p = str("notebook_path")
		}
		return sensitivePath(p)
	}

	name := strings.ToLower(tool)
	if strings.HasPrefix(name, "mcp__") {
		if i := strings.LastIndex(name, "__"); i >= 0 {
			name = name[i+2:] // judge the tool, not the server name
		}
	}
	for _, w := range sensitiveToolWords {
		if strings.Contains(name, w) {
			return true
		}
	}
	return false
}

// sensitivePath flags writes outside the user's home, or to config and
// credential locations inside it. Notes, projects and memory are fair game.
func sensitivePath(p string) bool {
	if p == "" {
		return true
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return true
	}
	p = filepath.Clean(p)
	base := filepath.Base(p)
	if base == ".env" || strings.HasPrefix(base, ".env.") {
		return true
	}
	for _, root := range safeRoots {
		if r, err := filepath.Rel(root, p); err == nil && !strings.HasPrefix(r, "..") {
			return false
		}
	}
	rel, err := filepath.Rel(home, p)
	if err != nil || rel == "." || strings.HasPrefix(rel, "..") {
		return true
	}
	rel = filepath.ToSlash(rel)
	if strings.HasPrefix(rel, ".claude/projects/") && strings.Contains(rel, "/memory/") {
		return false // Claude Code auto-memory
	}
	if strings.HasPrefix(rel, ".jarvis/") {
		return false
	}
	// Any other dotfile or dot-directory: shell rc, ssh, gnupg, aws, .config,
	// .claude settings...
	return strings.HasPrefix(rel, ".")
}
