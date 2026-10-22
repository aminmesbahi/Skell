package skell

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withGUIGOOS swaps the OS sentinel used by guiBinaryCandidates so a single
// host can exercise the lookup for every platform.
func withGUIGOOS(t *testing.T, goos string) {
	t.Helper()
	prev, prevExtra := guiGOOS, guiExtraLocations
	guiGOOS = goos
	// What happens to be installed on the machine running the tests must not
	// change the result.
	guiExtraLocations = func() []string { return nil }
	t.Cleanup(func() { guiGOOS, guiExtraLocations = prev, prevExtra })
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

func TestStartGUIProcessImpl_MissingBinaryReturnsError(t *testing.T) {
	withGUIGOOS(t, "linux")
	err := startGUIProcessImpl(filepath.Join(t.TempDir(), "missing-gui"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "launch desktop GUI")
}

func TestFindGUIBinary_FindsTheAppWhereTheInstallerPutsIt(t *testing.T) {
	withGUIGOOS(t, "windows")
	installed := filepath.Join(t.TempDir(), "Programs", "skell", "skell-gui.exe")
	require.NoError(t, os.MkdirAll(filepath.Dir(installed), 0o755))
	require.NoError(t, os.WriteFile(installed, []byte(""), 0o600))
	guiExtraLocations = func() []string { return []string{installed} }

	// skell.exe lives somewhere else entirely, with no app next to it.
	path, err := findGUIBinary(filepath.Join(t.TempDir(), "skell.exe"))
	require.NoError(t, err)
	assert.Equal(t, installed, path)
}

func TestDefaultGUILocations(t *testing.T) {
	withGUIGOOS(t, "windows")
	t.Setenv("LOCALAPPDATA", `C:\Users\me\AppData\Local`)
	t.Setenv("ProgramFiles", `C:\Program Files`)
	got := defaultGUILocations()
	assert.Contains(t, got, filepath.Join(`C:\Users\me\AppData\Local`, "Programs", "skell", "skell-gui.exe"))
	assert.Contains(t, got, filepath.Join(`C:\Program Files`, "Amin Mesbahi", "Skell", "Skell.exe"))
}

func TestGUIMissingHelpIsSpecificToTheSystem(t *testing.T) {
	withGUIGOOS(t, "windows")
	win := guiMissingHelp()
	assert.Contains(t, win, "not installed")
	assert.Contains(t, win, windowsInstallCommand)
	assert.Contains(t, win, "_installer.exe")
	assert.Contains(t, win, releasesURL)

	withGUIGOOS(t, "darwin")
	assert.Contains(t, guiMissingHelp(), ".dmg")

	withGUIGOOS(t, "linux")
	linux := guiMissingHelp()
	assert.Contains(t, linux, "linux_amd64.tar.gz")
	assert.Contains(t, linux, "GTK 4")
}

func TestGUICmd_WhenMissingExplainsWhatToDoInsteadOfListingFiles(t *testing.T) {
	withGUIGOOS(t, "windows")
	dir := t.TempDir()
	prev := currentExecutablePath
	currentExecutablePath = func() (string, error) { return filepath.Join(dir, "skell.exe"), nil }
	t.Cleanup(func() { currentExecutablePath = prev })

	out, err := executeCmd(t, "gui")
	require.Error(t, err)
	assert.Equal(t, "the desktop app is not installed", err.Error())
	assert.Contains(t, out, "not installed on this computer")
	assert.Contains(t, out, "install.ps1")
	assert.NotContains(t, out, "Skell-windows-amd64.exe", "no list of candidate file names")
}

// missingGUI sets up a machine without the desktop app, with the download and
// the launch replaced so nothing touches the network or starts a program.
func missingGUI(t *testing.T) (downloads *int, started *[]string) {
	t.Helper()
	withGUIGOOS(t, "windows")
	dir := t.TempDir()
	prevExe := currentExecutablePath
	currentExecutablePath = func() (string, error) { return filepath.Join(dir, "skell.exe"), nil }
	prevSupported, prevDownload, prevStart := desktopAppSupported, downloadDesktopApp, startGUIProcess
	n := 0
	var launched []string
	desktopAppSupported = func() bool { return true }
	downloadDesktopApp = func(*cobra.Command) (string, error) {
		n++
		return filepath.Join(dir, "skell-gui.exe"), nil
	}
	startGUIProcess = func(p string) error { launched = append(launched, p); return nil }
	t.Cleanup(func() {
		currentExecutablePath, desktopAppSupported, downloadDesktopApp, startGUIProcess = prevExe, prevSupported, prevDownload, prevStart
	})
	return &n, &launched
}

func TestGUICmd_WithoutATerminalExplainsAndNeverDownloadsUnasked(t *testing.T) {
	downloads, started := missingGUI(t)
	out, err := executeCmd(t, "gui")
	require.EqualError(t, err, "the desktop app is not installed")
	assert.Contains(t, out, "skell gui --install", "tells how to get it automatically")
	assert.Zero(t, *downloads)
	assert.Empty(t, *started)
}

func TestGUICmd_InstallFlagDownloadsThenOpensTheApp(t *testing.T) {
	downloads, started := missingGUI(t)
	out, err := executeCmd(t, "gui", "--install")
	require.NoError(t, err, out)
	assert.Equal(t, 1, *downloads)
	require.Len(t, *started, 1)
	assert.Equal(t, "skell-gui.exe", filepath.Base((*started)[0]))
	assert.Contains(t, out, "launched skell-gui.exe")
}

func TestGUICmd_FailedDownloadFallsBackToInstructions(t *testing.T) {
	_, started := missingGUI(t)
	downloadDesktopApp = func(*cobra.Command) (string, error) { return "", errors.New("no network") }
	out, err := executeCmd(t, "gui", "--install")
	require.EqualError(t, err, "the desktop app is not installed")
	assert.Contains(t, out, "Could not install the desktop app: no network")
	assert.Contains(t, out, windowsInstallCommand, "and says how to do it by hand")
	assert.Empty(t, *started)
}

func TestGUICmd_UnsupportedSystemJustExplains(t *testing.T) {
	downloads, _ := missingGUI(t)
	desktopAppSupported = func() bool { return false }
	out, err := executeCmd(t, "gui", "--install")
	require.EqualError(t, err, "the desktop app is not installed")
	assert.Contains(t, out, "not installed on this computer")
	assert.Zero(t, *downloads)
}

func TestMegabytes(t *testing.T) {
	assert.Equal(t, "size unknown", megabytes(0))
	assert.Equal(t, "1.5 MB", megabytes(3<<19))
}
