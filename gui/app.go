package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// SkellResult mirrors the JSON output contract used by the frontend.
type SkellResult struct {
	Stdout  string `json:"stdout"`
	Stderr  string `json:"stderr"`
	Success bool   `json:"success"`
}

// FileEntry represents a filesystem entry (file or directory).
type FileEntry struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Path  string `json:"path"`
}

// App is the Wails application struct. All exported methods are bound to the frontend.
type App struct {
	ctx context.Context
}

// NewApp creates a new App instance.
func NewApp() *App {
	return &App{}
}

// startup is called by Wails when the application starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// skellBin returns the path to the skell binary, searching PATH.
func skellBin() (string, error) {
	return exec.LookPath("skell")
}

// RunSkell executes the skell CLI with the provided arguments and returns stdout/stderr.
func (a *App) RunSkell(args []string) SkellResult {
	bin, err := skellBin()
	if err != nil {
		return SkellResult{
			Stderr:  "skell binary not found in PATH. Install skell first: https://github.com/aminmesbahi/Skell",
			Success: false,
		}
	}

	cmd := exec.Command(bin, args...) //nolint:gosec // args come from trusted frontend
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	return SkellResult{
		Stdout:  stdout.String(),
		Stderr:  stderr.String(),
		Success: err == nil,
	}
}

// ReadFileContent reads and returns the contents of a file.
func (a *App) ReadFileContent(path string) (string, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ListDirectory returns the immediate children of a directory.
func (a *App) ListDirectory(path string) ([]FileEntry, error) {
	entries, err := os.ReadDir(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	result := make([]FileEntry, 0, len(entries))
	for _, e := range entries {
		result = append(result, FileEntry{
			Name:  e.Name(),
			IsDir: e.IsDir(),
			Path:  filepath.Join(path, e.Name()),
		})
	}
	return result, nil
}

// SkellVersion returns the output of `skell version`.
func (a *App) SkellVersion() string {
	r := a.RunSkell([]string{"version"})
	return strings.TrimSpace(r.Stdout)
}

// SelectDirectory opens a native directory picker dialog and returns the selected path.
func (a *App) SelectDirectory() string {
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Repository",
	})
	if err != nil {
		return ""
	}
	return path
}

// AuditLogPath returns the platform-correct path to ~/.skell/audit.log.
func (a *App) AuditLogPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".skell", "audit.log")
}

// GlobalRootDir returns the global Skell root directory (~/.skell) and ensures
// the global manifest exists so that `skell search --repo <path>` can resolve it.
func (a *App) GlobalRootDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	root := filepath.Join(home, ".skell")
	manifestDir := filepath.Join(root, ".claude")
	manifestPath := filepath.Join(manifestDir, "skell.toml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		if mkErr := os.MkdirAll(manifestDir, 0700); mkErr == nil {
			_ = os.WriteFile(manifestPath, []byte("[registries]\n[skills]\n"), 0600)
		}
	}
	return root
}
