package skell

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var currentExecutablePath = os.Executable
var startGUIProcess = startGUIProcessImpl

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

func findGUIBinary(execPath string) (string, error) {
	dir := filepath.Dir(execPath)
	for _, name := range []string{"skell-gui.exe", "Skell.exe", "Skell-windows-amd64.exe", "skell-gui", "Skell"} {
		candidate := filepath.Join(dir, name)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("desktop GUI not found next to skell; expected skell-gui.exe or Skell-windows-amd64.exe in %s", dir)
}

func startGUIProcessImpl(guiPath string) error {
	cmd := exec.CommandContext(context.Background(), guiPath) //nolint:gosec
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("launch desktop GUI: %w", err)
	}
	return nil
}
