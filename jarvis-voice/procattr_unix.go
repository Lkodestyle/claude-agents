//go:build !windows

package main

import (
	"os/exec"
	"syscall"
)

// detachFromTerminalSignals puts the child in its own process group so the
// terminal's Ctrl+C reaches only jarvis-voice, which then decides when the
// child exits (after the session wrap-up).
func detachFromTerminalSignals(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
