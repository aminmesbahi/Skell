//go:build windows

package skell

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"github.com/inconshreveable/mousetrap"
	"github.com/spf13/cobra"
)

func init() {
	// Cobra would otherwise print "This is a command line tool. You need to
	// open cmd.exe…" and quit when started from Explorer; firstrun.go gives
	// that person something useful instead.
	cobra.MousetrapHelpText = ""
}

// startedFromFileManager reports whether this process was started by a
// double-click: either Explorer is its parent, or it has a console of its own
// with no shell attached (Explorer's short-lived helper processes, shortcuts
// and "Open with" can leave the parent gone, so the parent check alone misses
// some double-clicks). A program run from a terminal shares the console with
// its shell, so it never matches.
func startedFromFileManager() bool {
	if !isInteractive() {
		return false
	}
	return mousetrap.StartedByExplorer() || ownsItsConsole()
}

var procGetConsoleProcessList = syscall.NewLazyDLL("kernel32.dll").NewProc("GetConsoleProcessList")

// ownsItsConsole reports whether this is the only process attached to its console.
func ownsItsConsole() bool {
	var pids [2]uint32
	n, _, _ := procGetConsoleProcessList.Call(uintptr(unsafe.Pointer(&pids[0])), uintptr(len(pids)))
	return n == 1
}

// createNewConsole gives the child its own window.
const createNewConsole = 0x00000010

// openTerminalHere opens a command prompt in dir with dir on PATH, so "skell"
// works immediately even though it is not installed anywhere.
func openTerminalHere(dir string) error {
	cmd := exec.CommandContext(context.Background(), "cmd.exe") //nolint:gosec
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "PATH="+dir+";"+os.Getenv("PATH"))
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: createNewConsole,
		// cmd strips the outer quotes when the text holds & characters.
		CmdLine: `cmd.exe /k "title Skell & echo. & echo   skell is ready. Type  skell  to see what to do next,  or  skell --help  for every command. & echo."`,
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start cmd.exe: %w", err)
	}
	return nil
}

// openBrowser opens url in the default browser.
func openBrowser(url string) error {
	cmd := exec.CommandContext(context.Background(), "rundll32", "url.dll,FileProtocolHandler", url) //nolint:gosec
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("open browser: %w", err)
	}
	return nil
}
