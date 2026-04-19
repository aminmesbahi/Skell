// Package manifest handles reading and writing skell.toml files.
package manifest

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// SkillEntry is a single skill declaration inside skell.toml [skills].
type SkillEntry struct {
	Version  string `toml:"version"`
	Registry string `toml:"registry"`
	Pinned   bool   `toml:"pinned"`
}

// Manifest represents the full contents of a skell.toml file.
type Manifest struct {
	Registries map[string]string     `toml:"registries"`
	Skills     map[string]SkillEntry `toml:"skills"`
}

// Read parses a skell.toml file at the given path.
func Read(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if _, err := toml.Decode(string(data), &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// Write serialises a Manifest to a skell.toml file at the given path.
func Write(path string, m *Manifest) error {
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(m); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0600)
}

// GlobalPath returns the path to the global manifest (~/.skell/skell.toml).
func GlobalPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".skell", "skell.toml"), nil
}

// GlobalRootDir returns the global Skell root directory (~/.skell).
// When the --global flag is set, this is used as the "repository root" so that
// engine methods naturally target ~/.skell/.claude/skills/ for installed skills
// and fall back to ~/.skell/skell.toml for the manifest.
func GlobalRootDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".skell"), nil
}

// EnsureGlobal creates the global manifest (~/.skell/skell.toml) with an empty
// skeleton if it does not already exist. Safe to call multiple times.
func EnsureGlobal() error {
	path, err := GlobalPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		return nil // already exists
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return Write(path, &Manifest{
		Registries: map[string]string{},
		Skills:     map[string]SkillEntry{},
	})
}

// LocalPath returns the path to the local manifest inside a repository root.
func LocalPath(repoRoot string) string {
	return filepath.Join(repoRoot, ".claude", "skell.toml")
}

// Resolve returns the effective manifest for a given repository root.
// It first looks for a local manifest (.claude/skell.toml inside repoRoot).
// The global manifest (~/.skell/skell.toml) is only consulted when repoRoot
// itself is the global directory — normal project repos do not inherit it.
func Resolve(repoRoot string) (*Manifest, error) {
	localPath := LocalPath(repoRoot)
	if _, err := os.Stat(localPath); err == nil {
		return Read(localPath)
	}

	// Fall back to the global manifest only when operating on the global dir.
	globalDir, err := GlobalRootDir()
	if err == nil && filepath.Clean(repoRoot) == filepath.Clean(globalDir) {
		globalPath, err := GlobalPath()
		if err != nil {
			return nil, err
		}
		return Read(globalPath)
	}

	return nil, fmt.Errorf("open %s: no such file or directory", localPath)
}
