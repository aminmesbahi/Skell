package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ValidateDirectory checks a local or cached skill before installation.
func (a *App) ValidateDirectory(path string) ([]SkillValidationResult, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("skill path must be a directory")
	}
	result := a.RunSkell([]string{"validate", "--path", path, "--json"})
	if strings.TrimSpace(result.Stdout) == "" {
		return nil, fmt.Errorf("validation unavailable: %s", result.Stderr)
	}
	return parseValidationOutput(result.Stdout)
}

// OpenSkillFolder opens an existing directory with the platform's file manager.
func (a *App) OpenSkillFolder(path string) error {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory")
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer.exe", absolute)
	case "darwin":
		cmd = exec.Command("open", absolute)
	default:
		cmd = exec.Command("xdg-open", absolute)
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}
