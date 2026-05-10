//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

// hideConsoleWindow makes a child process invisible on Windows. The Wails GUI
// build uses the Windows GUI subsystem (no parent console), so each subprocess
// spawned via os/exec would otherwise allocate its own console window — which
// flashes up briefly on every page load. CREATE_NO_WINDOW (0x08000000)
// suppresses that allocation.
func hideConsoleWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
}
