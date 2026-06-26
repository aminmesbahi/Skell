//go:build !windows

package main

// isWindowsGUIBinary always returns false on non-Windows platforms where the
// PE format is not used. On Linux/macOS the CLI binary is an ELF or Mach-O
// file; a GUI executable would be a different format entirely and cannot
// masquerade as a CLI via case-insensitive filename collision.
func isWindowsGUIBinary(_ string) bool { return false }
