//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

// detachFromTerminalSignals starts the child in a new process group, which
// makes Windows stop delivering the console's Ctrl+C to it.
func detachFromTerminalSignals(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
}
