package skell

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// releasesURL is the page where the desktop app is published.
const releasesURL = "https://github.com/aminmesbahi/skell/releases/latest"

// windowsInstallCommand installs skell and the desktop app together.
const windowsInstallCommand = "irm https://raw.githubusercontent.com/aminmesbahi/skell/main/install.ps1 | iex"

// errGUINotFound means no desktop app was found; the caller shows guiMissingHelp.
var errGUINotFound = errors.New("desktop GUI not found")

// guiExtraLocations returns the places the desktop app is normally installed
// (by the PowerShell installer, the Windows setup program, or by hand), so
// 'skell gui' finds it even when skell itself lives somewhere else. It is a
// variable so tests do not depend on what is installed on the machine.
var guiExtraLocations = defaultGUILocations

func defaultGUILocations() []string {
	home, _ := os.UserHomeDir()
	switch guiGOOS {
	case "windows":
		var out []string
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			out = append(out,
				filepath.Join(local, "Programs", "skell", "skell-gui.exe"), // install.ps1
				filepath.Join(local, "Programs", "Skell", "Skell.exe"),     // setup program, per user
			)
		}
		if pf := os.Getenv("ProgramFiles"); pf != "" {
			out = append(out, filepath.Join(pf, "Amin Mesbahi", "Skell", "Skell.exe")) // setup program, all users
		}
		return out
	case "darwin":
		if home != "" {
			return []string{
				filepath.Join(home, "Applications", "Skell", "Skell.app"), // downloaded by skell gui
				filepath.Join(home, "Applications", "Skell.app"),
			}
		}
	default:
		if home != "" {
			return []string{filepath.Join(home, ".local", "bin", "skell-gui"), filepath.Join(home, ".local", "bin", "Skell")}
		}
		return []string{"/usr/local/bin/skell-gui", "/opt/Skell/Skell"}
	}
	return nil
}

// guiMissingHelp tells someone who asked for the desktop app, and does not
// have it, exactly what to do next on their system.
func guiMissingHelp() string {
	var b strings.Builder
	b.WriteString("\n  The Skell desktop app is not installed on this computer.\n\n")
	switch guiGOOS {
	case "windows":
		fmt.Fprintf(&b, "  Easiest — one command that installs the app and skell together:\n\n      %s\n\n", windowsInstallCommand)
		fmt.Fprintf(&b, "  Or download the installer from the releases page:\n\n      skell_gui_<version>_windows_%s_installer.exe\n", runtime.GOARCH)
	case "darwin":
		b.WriteString("  Download the app from the releases page, open the .dmg and drag Skell to Applications:\n\n      skell_gui_<version>_darwin_arm64.dmg\n")
	default:
		b.WriteString("  Download the app from the releases page, unpack it and put the Skell file next to skell\n  (it needs GTK 4 and WebKitGTK 6 installed):\n\n      skell_gui_<version>_linux_amd64.tar.gz\n")
	}
	fmt.Fprintf(&b, "\n  Releases: %s\n\n", releasesURL)
	return b.String()
}
