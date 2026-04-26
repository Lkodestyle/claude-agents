package main

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// loadDotEnv mirrors the loader in mcp-servers/jarvis-memory: search the cwd,
// the binary's directory, and a couple of parents for a .env file. Existing
// env vars are NEVER overwritten — anything the parent shell sets wins.
//
// This is what lets the user keep all their keys (VOYAGE, DEEPGRAM, ANTHROPIC)
// in a single .env at the repo root and have both the memory MCP server AND
// the voice CLI pick them up automatically.
func loadDotEnv() {
	candidates := []string{".env"}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, ".env"),
			filepath.Join(exeDir, "..", ".env"),
			filepath.Join(exeDir, "..", "..", ".env"),
		)
	}

	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, ".jarvis", ".env"))
	}

	for _, p := range candidates {
		abs, _ := filepath.Abs(p)
		if applyDotEnv(abs) {
			log.Printf("loaded env from %s", abs)
			return
		}
	}
}

func applyDotEnv(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		eq := strings.IndexByte(line, '=')
		if eq <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		val := strings.TrimSpace(line[eq+1:])
		if len(val) >= 2 && (val[0] == '"' || val[0] == '\'') && val[len(val)-1] == val[0] {
			val = val[1 : len(val)-1]
		}
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
	return true
}
