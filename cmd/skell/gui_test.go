package skell

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withGUIGOOS swaps the OS sentinel used by guiBinaryCandidates so a single
// host can exercise the lookup for every platform.
func withGUIGOOS(t *testing.T, goos string) {
	t.Helper()
	prev := guiGOOS
	guiGOOS = goos
	t.Cleanup(func() { guiGOOS = prev })
}

func TestFindGUIBinary_Windows_PrefersBundledName(t *testing.T) {
	withGUIGOOS(t, "windows")
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "skell-gui.exe"), []byte(""), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Skell-windows-amd64.exe"), []byte(""), 0600))

	path, err := findGUIBinary(filepath.Join(dir, "skell.exe"))
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "skell-gui.exe"), path)
}

func TestFindGUIBinary_Darwin_PrefersSiblingAppBundle(t *testing.T) {
	withGUIGOOS(t, "darwin")
	dir := t.TempDir()
	bundle := filepath.Join(dir, "Skell.app")
	require.NoError(t, os.MkdirAll(filepath.Join(bundle, "Contents", "MacOS"), 0o755))

	path, err := findGUIBinary(filepath.Join(dir, "skell"))
	require.NoError(t, err)
	assert.Equal(t, bundle, path)
}

func TestFindGUIBinary_Darwin_FallsBackToRawBinary(t *testing.T) {
	withGUIGOOS(t, "darwin")
	dir := t.TempDir()
	gui := filepath.Join(dir, "skell-gui")
	require.NoError(t, os.WriteFile(gui, []byte(""), 0o755))

	path, err := findGUIBinary(filepath.Join(dir, "skell"))
	require.NoError(t, err)
	assert.Equal(t, gui, path)
}

func TestFindGUIBinary_Linux_FindsSiblingBinary(t *testing.T) {
	withGUIGOOS(t, "linux")
	dir := t.TempDir()
	gui := filepath.Join(dir, "skell-gui")
	require.NoError(t, os.WriteFile(gui, []byte(""), 0o755))

	path, err := findGUIBinary(filepath.Join(dir, "skell"))
	require.NoError(t, err)
	assert.Equal(t, gui, path)
}

func TestFindGUIBinary_ReturnsErrorWhenMissing(t *testing.T) {
	// Pin to linux so the lookup doesn't probe /Applications/Skell.app on a
	// macOS host (which may exist when the developer is testing the GUI).
	withGUIGOOS(t, "linux")
	_, err := findGUIBinary(filepath.Join(t.TempDir(), "skell"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "desktop GUI not found")
}

func TestGUICmd_LaunchesSiblingExecutable(t *testing.T) {
	withGUIGOOS(t, "windows")
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
	withGUIGOOS(t, "windows")
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
