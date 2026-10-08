// Package hasher computes and compares SHA-256 content hashes for skill directories.
package hasher

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// HashDir computes a deterministic SHA-256 hash over all files in a directory tree.
// Paths are hashed with forward slashes and sorted, so the same tree produces
// the same hash on every operating system.
func HashDir(dirPath string) (string, error) {
	return hashDir(dirPath, false)
}

// hashDir is HashDir with an optional legacy mode that hashes OS-native path
// separators (what Skell versions before cross-platform hashing produced).
func hashDir(dirPath string, native bool) (string, error) {
	if _, err := os.Stat(dirPath); err != nil {
		return "", err
	}
	var files []string
	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			rel, err := filepath.Rel(dirPath, path)
			if err != nil {
				return err
			}
			if !native {
				rel = filepath.ToSlash(rel)
			}
			files = append(files, rel)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(files)
	h := sha256.New()
	for _, rel := range files {
		if _, err := fmt.Fprintf(h, "%s\x00", rel); err != nil {
			return "", err
		}
		if err := hashFile(h, filepath.Join(dirPath, filepath.FromSlash(rel))); err != nil {
			return "", err
		}
		_, _ = h.Write([]byte{0})
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}

// hashFile streams a file into w rather than loading it fully into memory.
func hashFile(w io.Writer, path string) error {
	f, err := os.Open(path) //nolint:gosec
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = io.Copy(w, f)
	return err
}

// Verify returns true if the SHA-256 hash of the directory matches the expected hash.
// On platforms whose path separator is not "/", hashes written by older
// versions (which used native separators) are still accepted.
func Verify(dirPath, expectedHash string) (bool, error) {
	actual, err := HashDir(dirPath)
	if err != nil {
		return false, err
	}
	if actual == expectedHash {
		return true, nil
	}
	if os.PathSeparator != '/' {
		legacy, err := hashDir(dirPath, true)
		if err != nil {
			return false, err
		}
		return legacy == expectedHash, nil
	}
	return false, nil
}
