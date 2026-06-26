//go:build windows

package main

import (
	"debug/pe"
	"path/filepath"
)

const (
	peSubsystemWindowsGUI = 2 // IMAGE_SUBSYSTEM_WINDOWS_GUI
)

// isWindowsGUIBinary returns true when the PE binary at path has the Windows
// GUI subsystem flag (IMAGE_SUBSYSTEM_WINDOWS_GUI = 2). This is the definitive
// way to distinguish the Skell GUI executable from the skell CLI executable
// without running the binary:
//
//   - Skell.exe (Wails GUI)  → Subsystem = 2 (Windows GUI)
//   - skell.exe  (CLI)       → Subsystem = 3 (Windows console)
//
// Using the PE header is reliable even when os.Executable() returns an
// unexpected path and all path-comparison guards in isSelfExecutable fail.
func isWindowsGUIBinary(path string) bool {
	f, err := pe.Open(filepath.Clean(path))
	if err != nil {
		return false
	}
	defer f.Close()

	switch oh := f.OptionalHeader.(type) {
	case *pe.OptionalHeader32:
		return oh.Subsystem == peSubsystemWindowsGUI
	case *pe.OptionalHeader64:
		return oh.Subsystem == peSubsystemWindowsGUI
	}
	return false
}
