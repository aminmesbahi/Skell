//go:build windows

package registry

import (
	"os/exec"
	"syscall"
)

func hideSubprocessWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
}
