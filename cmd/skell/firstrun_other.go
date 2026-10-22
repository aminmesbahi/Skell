//go:build !windows

package skell

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
)

// File managers on macOS and Linux do not run command-line programs by
// double-click in a way that needs special handling, so nothing is detected.
func startedFromFileManager() bool { return false }

func openTerminalHere(string) error { return errors.New("not supported on this platform") }

// openBrowser opens url in the default browser.
func openBrowser(url string) error {
	opener := "xdg-open"
	if runtime.GOOS == "darwin" {
		opener = "open"
	}
	if err := exec.CommandContext(context.Background(), opener, url).Start(); err != nil { //nolint:gosec
		return fmt.Errorf("open browser: %w", err)
	}
	return nil
}
