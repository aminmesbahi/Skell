//go:build windows

package main

import (
	"os"
	"syscall"
	"unsafe"
)

func sameFileRobust(left, right string) bool {
	if left == "" || right == "" {
		return false
	}
	// Try standard os.SameFile first (may work in many cases)
	if fi1, err := os.Stat(left); err == nil {
		if fi2, err := os.Stat(right); err == nil {
			if os.SameFile(fi1, fi2) {
				return true
			}
		}
	}
	// Fallback to explicit handle + GetFileInformationByHandle for reliable
	// VolumeSerialNumber + FileIndex on Windows, even when plain Stat doesn't
	// populate the full fileStat ID fields.
	return sameFileByHandle(left, right)
}

func sameFileByHandle(p1, p2 string) bool {
	h1, err := openFileForInfo(p1)
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(h1)

	h2, err := openFileForInfo(p2)
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(h2)

	var info1, info2 syscall.ByHandleFileInformation
	if err := syscall.GetFileInformationByHandle(h1, &info1); err != nil {
		return false
	}
	if err := syscall.GetFileInformationByHandle(h2, &info2); err != nil {
		return false
	}

	return info1.VolumeSerialNumber == info2.VolumeSerialNumber &&
		info1.FileIndexHigh == info2.FileIndexHigh &&
		info1.FileIndexLow == info2.FileIndexLow
}

func openFileForInfo(path string) (syscall.Handle, error) {
	// Ensure we have a UTF16 pointer. Support long paths by ensuring \\?\ prefix if needed
	// (CreateFile on Win10+ with manifest often handles, but be defensive).
	p16, err := syscall.UTF16PtrFromString(ensureLongPathPrefix(path))
	if err != nil {
		return 0, err
	}

	// Open with minimal access, share everything, for info only.
	h, err := syscall.CreateFile(
		p16,
		syscall.GENERIC_READ,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		return 0, err
	}
	return h, nil
}

// ensureLongPathPrefix adds the \\?\ prefix on Windows for paths that look like
// they might exceed MAX_PATH, or always for safety with absolute paths.
func ensureLongPathPrefix(p string) string {
	if len(p) < 260 {
		return p
	}
	if len(p) >= 4 && p[:4] == `\\?\` {
		return p
	}
	// Only for drive absolute paths
	if len(p) >= 2 && p[1] == ':' && (p[2] == '\\' || p[2] == '/') {
		return `\\?\` + p
	}
	return p
}
