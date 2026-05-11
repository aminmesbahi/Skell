//go:build !windows

package registry

import "os/exec"

func hideSubprocessWindow(*exec.Cmd) {}
