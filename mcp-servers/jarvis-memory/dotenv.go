package main

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// loadDotEnv tries to populate env vars from a .env file near the running
// binary. It stops at the first file it finds. Existing env vars are NEVER
// overwritten, so anything the MCP client (Claude Code) passes in wins.
//
// Search order:
//   1. $PWD/.env                       — cwd when the binary was invoked
//   2. <exeDir>/.env                   — binary's own directory
//   3. <exeDir>/../.env                — one level up (typical bin/ layout)
//   4. <exeDir>/../../.env             — two levels up (./bin/jarvis-memory inside a repo)
//   5. $HOME/.jarvis/.env              — user-global config
//
// Soft-fails silently: if no file is found we just log a debug line and move on.
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

	for _, path := range candidates {
		abs, _ := filepath.Abs(path)
		if applyDotEnv(abs) {
			log.Printf("loaded env from %s", abs)
			return
		}
	}
	log.Printf("no .env found (searched %d locations)", len(candidates))
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

		// Strip a single layer of matched surrounding quotes.
		if len(val) >= 2 && (val[0] == '"' || val[0] == '\'') && val[len(val)-1] == val[0] {
			val = val[1 : len(val)-1]
		}

		// Only set if not already defined by the parent process. This keeps
		// Claude Code's env substitution authoritative when it sets values,
		// but still lets .env fill in gaps.
		if existing := os.Getenv(key); existing == "" {
			_ = os.Setenv(key, val)
		}
	}
	return true
}
