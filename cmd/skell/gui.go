package skell

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var currentExecutablePath = os.Executable
var startGUIProcess = startGUIProcessImpl

// guiGOOS is overridable in tests so candidate-list logic can be exercised
// regardless of the host platform.
var guiGOOS = runtime.GOOS

func newGUICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gui",
		Short: "Launch the desktop GUI",
		Long: `Launches the Skell desktop GUI when a GUI executable is installed
next to the skell CLI binary.`,
		Example: `  skell gui`,
		RunE: func(cmd *cobra.Command, args []string) error {
			execPath, err := currentExecutablePath()
			if err != nil {
				return fmt.Errorf("locate current executable: %w", err)
			}

			guiPath, err := findGUIBinary(execPath)
			if err != nil {
				return err
			}

			if err := startGUIProcess(guiPath); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  launched %s\n", filepath.Base(guiPath))
			return nil
		},
	}
	return cmd
}

// guiBinaryCandidates returns the paths to probe for a GUI launcher, in the
// order they should be tried. The list is OS-specific because the Wails build
// produces a `.app` bundle on macOS, a `.exe` on Windows, and a raw ELF
// binary on Linux.
func guiBinaryCandidates(dir string) []string {
	switch guiGOOS {
	case "darwin":
		// Sibling locations win over /Applications so a local dev build
		// next to the CLI takes precedence over a previously-installed app.
		return []string{
			filepath.Join(dir, "Skell.app"),
			filepath.Join(dir, "skell-gui"),
			filepath.Join(dir, "Skell"),
			"/Applications/Skell.app",
		}
	case "windows":
		return []string{
			filepath.Join(dir, "skell-gui.exe"),
			filepath.Join(dir, "Skell.exe"),
			filepath.Join(dir, "Skell-windows-amd64.exe"),
		}
	default:
		return []string{
			filepath.Join(dir, "skell-gui"),
			filepath.Join(dir, "Skell"),
		}
	}
}

func findGUIBinary(execPath string) (string, error) {
	dir := filepath.Dir(execPath)
	for _, candidate := range guiBinaryCandidates(dir) {
		info, err := os.Stat(candidate)
		if err != nil {
			continue
		}
		// On macOS a `.app` bundle is a directory; everything else must be a
		// regular file.
		if guiGOOS == "darwin" && strings.HasSuffix(candidate, ".app") {
			if info.IsDir() {
				return candidate, nil
			}
			continue
		}
		if !info.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("desktop GUI not found: looked for %s", strings.Join(guiBinaryCandidates(dir), ", "))
}

func startGUIProcessImpl(guiPath string) error {
	var cmd *exec.Cmd
	if guiGOOS == "darwin" && strings.HasSuffix(guiPath, ".app") {
		// `open -n` launches a new instance of the app bundle without
		// blocking, mirroring the behaviour of double-clicking it in Finder.
		cmd = exec.CommandContext(context.Background(), "/usr/bin/open", "-n", guiPath) //nolint:gosec
	} else {
		cmd = exec.CommandContext(context.Background(), guiPath) //nolint:gosec
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("launch desktop GUI: %w", err)
	}
	return nil
}
