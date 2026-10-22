package skell

import (
	"context"
	"errors"
	"fmt"
	"io"

	"path/filepath"
	"runtime"
	"time"

	"github.com/aminmesbahi/skell/internal/guiinstall"
	"github.com/aminmesbahi/skell/internal/selfupdate"
	"github.com/spf13/cobra"
)

// errGUINotInstalled is returned when the person declined, or could not be
// asked, to download the desktop app.
var errGUINotInstalled = errors.New("the desktop app is not installed")

// downloadDesktopApp fetches and installs the latest desktop app and returns
// the program to launch. It is a variable so tests do not touch the network.
var downloadDesktopApp = downloadDesktopAppImpl

// desktopAppSupported reports whether a desktop app is published for this
// system; a variable so tests do not depend on the machine they run on.
var desktopAppSupported = guiinstall.Here

// offerDesktopAppInstall deals with 'skell gui' finding no app. In a terminal
// it asks before downloading; with --install it does not ask; with nobody to
// ask it explains and says how to download it.
func offerDesktopAppInstall(cmd *cobra.Command, assumeYes bool) (string, error) {
	out, errw := cmd.OutOrStdout(), cmd.ErrOrStderr()

	if !desktopAppSupported() {
		_, _ = fmt.Fprint(errw, guiMissingHelp())
		return "", errGUINotInstalled
	}
	if !assumeYes {
		if !isInteractive() {
			_, _ = fmt.Fprint(errw, guiMissingHelp())
			_, _ = fmt.Fprintln(errw, "  To download and install it automatically, run:  skell gui --install")
			_, _ = fmt.Fprintln(errw)
			return "", errGUINotInstalled
		}
		_, _ = fmt.Fprintln(out, "\n  The Skell desktop app is not installed on this computer.")
		_, _ = fmt.Fprintln(out, "  It can be downloaded from GitHub now: the latest release, checked against its")
		_, _ = fmt.Fprintln(out, "  published SHA-256, installed for your user only (no administrator rights).")
		_, _ = fmt.Fprintln(out)
		if !confirm(cmd, "Download and open it now?", true) {
			_, _ = fmt.Fprint(errw, guiMissingHelp())
			return "", errGUINotInstalled
		}
	}

	path, err := downloadDesktopApp(cmd)
	if err != nil {
		_, _ = fmt.Fprintf(errw, "\n  Could not install the desktop app: %v\n", err)
		_, _ = fmt.Fprint(errw, guiMissingHelp())
		return "", errGUINotInstalled
	}
	return path, nil
}

func downloadDesktopAppImpl(cmd *cobra.Command) (string, error) {
	out := cmd.OutOrStdout()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	_, _ = fmt.Fprintln(out, "  Looking up the latest release…")
	rel, err := selfupdate.New("aminmesbahi", "skell").LatestRelease()
	if err != nil {
		return "", err
	}
	choice, err := guiinstall.Choose(rel, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return "", err
	}
	dir, err := guiinstall.DefaultDir(runtime.GOOS)
	if err != nil {
		return "", err
	}
	_, _ = fmt.Fprintf(out, "  Downloading %s (%s)\n", choice.Asset.Name, megabytes(choice.Asset.Size))

	in := guiinstall.NewInstaller()
	in.Progress = progressPrinter(out)
	path, err := in.Install(ctx, rel, runtime.GOOS, runtime.GOARCH, dir)
	_, _ = fmt.Fprintln(out)
	if err != nil {
		return "", err
	}
	_, _ = fmt.Fprintf(out, "  %s  checked and installed to %s\n", markOK, dir)
	if exe, eerr := currentExecutablePath(); eerr == nil && filepath.Dir(exe) != dir {
		_, _ = fmt.Fprintln(out, "     (skell itself is not added to your PATH — the installer from the README does that)")
	}
	return path, nil
}

func megabytes(n int64) string {
	if n <= 0 {
		return "size unknown"
	}
	return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
}

// progressPrinter shows "downloaded x of y MB" on one line in a terminal and
// stays quiet otherwise (logs do not want hundreds of updates).
func progressPrinter(w io.Writer) func(done, total int64) {
	if !isInteractive() {
		return nil
	}
	lastPct := -1
	return func(done, total int64) {
		if total <= 0 {
			return
		}
		pct := int(done * 100 / total)
		if pct == lastPct {
			return
		}
		lastPct = pct
		_, _ = fmt.Fprintf(w, "\r  %.1f / %.1f MB  (%d%%)", float64(done)/(1<<20), float64(total)/(1<<20), pct)
	}
}
