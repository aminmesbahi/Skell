package skell

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindGUIBinary_PrefersBundledName(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "skell-gui.exe"), []byte(""), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Skell-windows-amd64.exe"), []byte(""), 0600))

	path, err := findGUIBinary(filepath.Join(dir, "skell.exe"))
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "skell-gui.exe"), path)
}

func TestFindGUIBinary_ReturnsErrorWhenMissing(t *testing.T) {
	_, err := findGUIBinary(filepath.Join(t.TempDir(), "skell.exe"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "desktop GUI not found")
}

func TestGUICmd_LaunchesSiblingExecutable(t *testing.T) {
	dir := t.TempDir()
	guiPath := filepath.Join(dir, "skell-gui.exe")
	require.NoError(t, os.WriteFile(guiPath, []byte(""), 0600))

	origExecPath := currentExecutablePath
	origStart := startGUIProcess
	t.Cleanup(func() {
		currentExecutablePath = origExecPath
		startGUIProcess = origStart
	})

	currentExecutablePath = func() (string, error) {
		return filepath.Join(dir, "skell.exe"), nil
	}

	launched := ""
	startGUIProcess = func(path string) error {
		launched = path
		return nil
	}

	out, err := executeCmd(t, "gui")
	require.NoError(t, err)
	assert.Equal(t, guiPath, launched)
	assert.Contains(t, out, "launched skell-gui.exe")
}

func TestGUICmd_PropagatesLaunchError(t *testing.T) {
	dir := t.TempDir()
	guiPath := filepath.Join(dir, "skell-gui.exe")
	require.NoError(t, os.WriteFile(guiPath, []byte(""), 0600))

	origExecPath := currentExecutablePath
	origStart := startGUIProcess
	t.Cleanup(func() {
		currentExecutablePath = origExecPath
		startGUIProcess = origStart
	})

	currentExecutablePath = func() (string, error) {
		return filepath.Join(dir, "skell.exe"), nil
	}
	startGUIProcess = func(string) error {
		return errors.New("boom")
	}

	_, err := executeCmd(t, "gui")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}
