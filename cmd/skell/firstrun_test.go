package skell

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type firstRunHarness struct {
	f       *firstRun
	out     *bytes.Buffer
	started []string
	urls    []string
	shells  []string
}

func newFirstRun(input string, guiPath string) *firstRunHarness {
	h := &firstRunHarness{out: &bytes.Buffer{}}
	h.f = &firstRun{
		in:       bufio.NewReader(strings.NewReader(input)),
		out:      h.out,
		execPath: `C:\Downloads\skell\skell.exe`,
		findGUI: func(string) (string, error) {
			if guiPath == "" {
				return "", errors.New("not found")
			}
			return guiPath, nil
		},
		startGUI:  func(p string) error { h.started = append(h.started, p); return nil },
		openURL:   func(u string) error { h.urls = append(h.urls, u); return nil },
		openShell: func(d string) error { h.shells = append(h.shells, d); return nil },
	}
	return h
}

func TestFirstRunOpensTheDesktopAppWhenInstalled(t *testing.T) {
	h := newFirstRun("", `C:\Downloads\skell\skell-gui.exe`)
	h.f.run()
	assert.Equal(t, []string{`C:\Downloads\skell\skell-gui.exe`}, h.started)
	assert.Contains(t, h.out.String(), "Opening the Skell desktop app")
	assert.NotContains(t, h.out.String(), "Pick what you would like to do")
}

func TestFirstRunFallsBackToTheMenuWhenTheAppWontStart(t *testing.T) {
	h := newFirstRun("\n", `C:\x\skell-gui.exe`)
	h.f.startGUI = func(string) error { return errors.New("boom") }
	h.f.run()
	assert.Contains(t, h.out.String(), "Pick what you would like to do")
}

func TestFirstRunMenuOptions(t *testing.T) {
	cases := []struct {
		name, input string
		shells      int
		urls        int
	}{
		{"open a terminal", "1\n", 1, 0},
		{"open the install page", "2\n", 0, 1},
		{"enter closes", "\n", 0, 0},
		{"closed input ends", "", 0, 0},
		{"asks again after a typo", "x\n1\n", 1, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newFirstRun(c.input, "")
			h.f.run()
			assert.Len(t, h.shells, c.shells)
			assert.Len(t, h.urls, c.urls)
			assert.Empty(t, h.started)
			assert.Contains(t, h.out.String(), "Skell — a package manager for Agent Skills")
		})
	}
}

func TestFirstRunTerminalOpensInTheProgramFolder(t *testing.T) {
	h := newFirstRun("1\n", "")
	h.f.run()
	assert.Equal(t, []string{filepath.Dir(h.f.execPath)}, h.shells[len(h.shells)-1:][0:1], "the folder part is platform dependent")
}

func TestFirstRunReportsFailuresAndKeepsAsking(t *testing.T) {
	h := newFirstRun("1\n2\n\n", "")
	h.f.openShell = func(string) error { return errors.New("no cmd") }
	h.f.openURL = func(string) error { return errors.New("no browser") }
	h.f.run()
	out := h.out.String()
	assert.Contains(t, out, "Could not open a command prompt: no cmd")
	assert.Contains(t, out, "Could not open the browser. Go to "+installPageURL)
}

func TestDoubleClickIsIgnoredWhenArgumentsAreGiven(t *testing.T) {
	// Normal command-line use never reaches the first-run screen.
	assert.False(t, runFirstRunIfDoubleClicked(), "tests run with arguments and from a terminal")
}

func TestFindGUIBinaryNeverReturnsTheCLIItself(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "skell.exe")
	require.NoError(t, os.WriteFile(exe, []byte("cli"), 0o755))

	old := guiGOOS
	guiGOOS = "windows"
	defer func() { guiGOOS = old }()

	// On a case-insensitive file system "Skell.exe" is the CLI itself.
	_, err := findGUIBinary(exe)
	assert.Error(t, err)

	gui := filepath.Join(dir, "skell-gui.exe")
	require.NoError(t, os.WriteFile(gui, []byte("gui"), 0o755))
	got, err := findGUIBinary(exe)
	require.NoError(t, err)
	assert.Equal(t, gui, got)
}

func TestBareSkellExplainsItselfInAFolderThatIsNotSetUp(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	out, err := executeCmd(t)
	require.NoError(t, err)
	for _, want := range []string{
		"package manager for Agent Skills",
		"SKILL.md",
		"What you can do",
		"Start here",
		"skell init",
		"skell --help",
	} {
		assert.Contains(t, out, want)
	}
	assert.NotContains(t, out, "  \n", "no trailing spaces on blank lines")
}

func TestBareSkellInASetUpProjectSkipsTheIntro(t *testing.T) {
	repo := makeRepoWithSkill(t, "pdf", "1.0.0")
	t.Chdir(repo)
	out, err := executeCmd(t)
	require.NoError(t, err)
	assert.NotContains(t, out, "What you can do")
	assert.Contains(t, out, "skills 1")
}

func TestHelpLeadsWithWhatSkellIs(t *testing.T) {
	out, err := executeCmd(t, "--help")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(out, tagline), out)
	assert.Contains(t, out, "the way npm or pip do for code")
}
