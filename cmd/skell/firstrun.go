package skell

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// installPageURL is where a new user finds the one-line installer for the
// command-line tool and the desktop app.
const installPageURL = "https://github.com/aminmesbahi/skell#install"

// firstRun is what a person sees when they double-click skell in a file
// manager instead of running it from a terminal. A bare "this is a command
// line tool" error is a dead end, so it opens the desktop app when one is next
// to the program and otherwise explains the options and keeps the window open.
type firstRun struct {
	in        *bufio.Reader
	out       io.Writer
	execPath  string
	findGUI   func(execPath string) (string, error)
	startGUI  func(path string) error
	openURL   func(url string) error
	openShell func(dir string) error
}

// run opens the desktop app when it is installed, otherwise shows the menu.
func (f *firstRun) run() {
	if gui, err := f.findGUI(f.execPath); err == nil {
		if err := f.startGUI(gui); err == nil {
			_, _ = fmt.Fprintf(f.out, "\n  Opening the Skell desktop app…\n")
			return
		}
	}
	f.menu()
}

func (f *firstRun) menu() {
	dir := filepath.Dir(f.execPath)
	_, _ = fmt.Fprint(f.out, `
  Skell — a package manager for Agent Skills

  Agent Skills are small folders that teach AI coding agents (Claude Code,
  Codex, GitHub Copilot, Cursor…) how to do a job well. Skell finds them,
  installs them, keeps them up to date and checks that they are safe.

  You opened Skell by double-clicking it. It is a command-line tool, so it is
  meant to be used from a terminal. Pick what you would like to do:

    1   Open a command prompt here, with skell ready to use
    2   Get the desktop app and install skell for every terminal
        (opens the install page in your browser)

  Press 1 or 2, or just press Enter to close this window.

`)
	for {
		_, _ = fmt.Fprint(f.out, "  > ")
		line, err := f.in.ReadString('\n')
		choice := strings.TrimSpace(line)
		switch choice {
		case "1":
			if oerr := f.openShell(dir); oerr != nil {
				_, _ = fmt.Fprintf(f.out, "  Could not open a command prompt: %v\n", oerr)
				continue
			}
			_, _ = fmt.Fprintln(f.out, "  Opened a command prompt. You can close this window.")
			return
		case "2":
			if oerr := f.openURL(installPageURL); oerr != nil {
				_, _ = fmt.Fprintf(f.out, "  Could not open the browser. Go to %s\n", installPageURL)
				continue
			}
			_, _ = fmt.Fprintf(f.out, "  Opened %s\n  The first command shown there installs skell and the desktop app.\n", installPageURL)
			return
		case "":
			return
		default:
			_, _ = fmt.Fprintln(f.out, "  Please press 1, 2 or Enter.")
		}
		if err != nil {
			return // input closed
		}
	}
}

// runFirstRunIfDoubleClicked handles a launch from a file manager and reports
// whether it did, in which case the caller should exit instead of running a
// command.
func runFirstRunIfDoubleClicked() bool {
	if len(os.Args) != 1 || !startedFromFileManager() {
		return false
	}
	execPath, err := currentExecutablePath()
	if err != nil {
		return false
	}
	f := &firstRun{
		in:        bufio.NewReader(os.Stdin),
		out:       os.Stdout,
		execPath:  execPath,
		findGUI:   findGUIBinary,
		startGUI:  startGUIProcess,
		openURL:   openBrowser,
		openShell: openTerminalHere,
	}
	f.run()
	return true
}
