package main

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// The desktop app ships the skell command-line tool it was built with, so it
// works on a computer that has never had skell installed, and the app and the
// CLI it drives are always the same version. The build puts the CLI for the
// target system in embedded/ (see gui/Taskfile.yml); a build without it — such
// as `go vet` or the unit tests — simply has no embedded CLI.
//
//go:embed all:embedded
var embeddedFiles embed.FS

// embeddedCLISource is where the embedded CLI is read from; tests replace it.
var embeddedCLISource fs.FS = embeddedFiles

// cliInstallRoot is the folder the embedded CLI is unpacked under. Tests replace it.
var cliInstallRoot = func() (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cache, "Skell", "cli"), nil
}

var (
	embeddedCLIOnce sync.Once
	embeddedCLIPath string
	embeddedCLIErr  error
)

// errNoEmbeddedCLI means this build of the app carries no CLI.
var errNoEmbeddedCLI = errors.New("this build of the app does not include the skell command-line tool")

// embeddedCLIData returns the embedded CLI for this system, if there is one.
func embeddedCLIData() ([]byte, error) {
	for _, name := range []string{"embedded/skell.exe", "embedded/skell"} {
		if runtime.GOOS == "windows" && name == "embedded/skell" {
			continue
		}
		if runtime.GOOS != "windows" && name == "embedded/skell.exe" {
			continue
		}
		data, err := fs.ReadFile(embeddedCLISource, name)
		if err == nil && len(data) > 0 {
			return data, nil
		}
	}
	return nil, errNoEmbeddedCLI
}

// ensureEmbeddedCLI unpacks the embedded CLI (once per version, into a folder
// named after the hash of its content) and returns its path. A running copy is
// never overwritten: a different build gets a different folder.
func ensureEmbeddedCLI() (string, error) {
	embeddedCLIOnce.Do(func() {
		embeddedCLIPath, embeddedCLIErr = unpackEmbeddedCLI()
	})
	return embeddedCLIPath, embeddedCLIErr
}

func unpackEmbeddedCLI() (string, error) {
	data, err := embeddedCLIData()
	if err != nil {
		return "", err
	}
	root, err := cliInstallRoot()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	dir := filepath.Join(root, hex.EncodeToString(sum[:])[:12])
	name := "skell"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	dest := filepath.Join(dir, name)
	if info, err := os.Stat(dest); err == nil && info.Size() == int64(len(data)) {
		return dest, nil // unpacked by an earlier run
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("cannot prepare %s: %w", dir, err)
	}
	// Write beside the destination and rename, so a crash never leaves a
	// half-written program that the size check above would accept.
	tmp, err := os.CreateTemp(dir, ".skell-*")
	if err != nil {
		return "", err
	}
	tmpName := tmp.Name()
	_, werr := tmp.Write(data)
	cerr := tmp.Close()
	if werr != nil || cerr != nil {
		_ = os.Remove(tmpName)
		return "", errors.Join(werr, cerr)
	}
	if err := os.Chmod(tmpName, 0o755); err != nil {
		_ = os.Remove(tmpName)
		return "", err
	}
	if err := os.Rename(tmpName, dest); err != nil {
		_ = os.Remove(tmpName)
		if info, serr := os.Stat(dest); serr == nil && info.Size() == int64(len(data)) {
			return dest, nil // another instance won the race
		}
		return "", err
	}
	return dest, nil
}
