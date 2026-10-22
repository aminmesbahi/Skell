package main

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain keeps every test independent of whatever CLI a local build has put
// in embedded/, and of the real cache folder.
func TestMain(m *testing.M) {
	embeddedCLISource = fstest.MapFS{"embedded/README.md": {Data: []byte("placeholder")}}
	tmp, err := os.MkdirTemp("", "skell-gui-cli-")
	if err != nil {
		panic(err)
	}
	cliInstallRoot = func() (string, error) { return tmp, nil }
	code := m.Run()
	_ = os.RemoveAll(tmp)
	os.Exit(code)
}

func cliFileName() string {
	if runtime.GOOS == "windows" {
		return "skell.exe"
	}
	return "skell"
}

// withEmbeddedCLI installs a fake embedded CLI and a private unpack folder for one test.
func withEmbeddedCLI(t *testing.T, content string) string {
	t.Helper()
	prevSrc, prevRoot := embeddedCLISource, cliInstallRoot
	root := t.TempDir()
	embeddedCLISource = fstest.MapFS{"embedded/" + cliFileName(): {Data: []byte(content)}, "embedded/README.md": {Data: []byte("x")}}
	cliInstallRoot = func() (string, error) { return root, nil }
	embeddedCLIOnce = sync.Once{}
	t.Cleanup(func() {
		embeddedCLISource, cliInstallRoot = prevSrc, prevRoot
		embeddedCLIOnce = sync.Once{}
	})
	return root
}

func TestPlaceholderOnlyMeansNoEmbeddedCLI(t *testing.T) {
	_, err := unpackEmbeddedCLI()
	assert.ErrorIs(t, err, errNoEmbeddedCLI)
}

func TestUnpackEmbeddedCLIWritesItOnceUnderAHashFolder(t *testing.T) {
	root := withEmbeddedCLI(t, "cli v1")

	first, err := unpackEmbeddedCLI()
	require.NoError(t, err)
	assert.Equal(t, cliFileName(), filepath.Base(first))
	assert.Equal(t, root, filepath.Dir(filepath.Dir(first)))
	b, _ := os.ReadFile(first)
	assert.Equal(t, "cli v1", string(b))
	if runtime.GOOS != "windows" {
		info, _ := os.Stat(first)
		assert.NotZero(t, info.Mode()&0o100, "must be executable")
	}

	// A second run reuses it.
	info1, _ := os.Stat(first)
	second, err := unpackEmbeddedCLI()
	require.NoError(t, err)
	assert.Equal(t, first, second)
	info2, _ := os.Stat(second)
	assert.Equal(t, info1.ModTime(), info2.ModTime(), "not rewritten")

	// A different build goes into its own folder, so a running copy is never overwritten.
	embeddedCLISource = fstest.MapFS{"embedded/" + cliFileName(): {Data: []byte("cli v2 is longer")}}
	third, err := unpackEmbeddedCLI()
	require.NoError(t, err)
	assert.NotEqual(t, first, third)
	assert.FileExists(t, first)
	entries, _ := os.ReadDir(root)
	assert.Len(t, entries, 2)
}

func TestResolveFallsBackToTheEmbeddedCLI(t *testing.T) {
	withEmbeddedCLI(t, "embedded cli")
	empty := t.TempDir()
	prevExec, prevLook, prevDirs := currentExecutable, lookPath, extraToolSearchDirs
	currentExecutable = func() (string, error) { return filepath.Join(empty, "skell-gui.exe"), nil }
	lookPath = func(string) (string, error) { return "", os.ErrNotExist }
	extraToolSearchDirs = func() []string { return nil }
	t.Cleanup(func() { currentExecutable, lookPath, extraToolSearchDirs = prevExec, prevLook, prevDirs })
	t.Setenv("SKELL_BIN", "")

	got, err := skellBin()
	require.NoError(t, err)
	b, _ := os.ReadFile(got)
	assert.Equal(t, "embedded cli", string(b))
}

func TestResolvePrefersACLINextToTheApp(t *testing.T) {
	withEmbeddedCLI(t, "embedded cli")
	dir := t.TempDir()
	sibling := filepath.Join(dir, cliFileName())
	require.NoError(t, os.WriteFile(sibling, []byte("sibling cli"), 0o755))
	prevExec, prevLook, prevDirs := currentExecutable, lookPath, extraToolSearchDirs
	currentExecutable = func() (string, error) { return filepath.Join(dir, "skell-gui.exe"), nil }
	lookPath = func(string) (string, error) { return "", os.ErrNotExist }
	extraToolSearchDirs = func() []string { return nil }
	t.Cleanup(func() { currentExecutable, lookPath, extraToolSearchDirs = prevExec, prevLook, prevDirs })
	t.Setenv("SKELL_BIN", "")

	got, err := skellBin()
	require.NoError(t, err)
	assert.Equal(t, sibling, got)
}

func TestResolveStillReportsMissingWhenNothingIsEmbedded(t *testing.T) {
	embeddedCLIOnce = sync.Once{}
	t.Cleanup(func() { embeddedCLIOnce = sync.Once{} })
	empty := t.TempDir()
	prevExec, prevLook, prevDirs := currentExecutable, lookPath, extraToolSearchDirs
	currentExecutable = func() (string, error) { return filepath.Join(empty, "skell-gui.exe"), nil }
	lookPath = func(string) (string, error) { return "", os.ErrNotExist }
	extraToolSearchDirs = func() []string { return nil }
	t.Cleanup(func() { currentExecutable, lookPath, extraToolSearchDirs = prevExec, prevLook, prevDirs })
	t.Setenv("SKELL_BIN", "")

	_, err := skellBin()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "binary not found")
}
