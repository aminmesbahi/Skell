// Package hasher computes and compares SHA-256 content hashes for skill directories.
package hasher

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// HashDir computes a deterministic SHA-256 hash over all files in a directory tree.
// Files are sorted by path before hashing to ensure reproducibility.
func HashDir(dirPath string) (string, error) {
	if _, err := os.Stat(dirPath); err != nil {
		return "", err
	}
	var files []string
	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(files)
	h := sha256.New()
	for _, f := range files {
		rel, err := filepath.Rel(dirPath, f)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(h, "%s\x00", rel)
		data, err := os.ReadFile(f)
		if err != nil {
			return "", err
		}
		h.Write(data)
		h.Write([]byte{0})
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}

// Verify returns true if the SHA-256 hash of the directory matches the expected hash.
func Verify(dirPath, expectedHash string) (bool, error) {
	actual, err := HashDir(dirPath)
	if err != nil {
		return false, err
	}
	return actual == expectedHash, nil
}
