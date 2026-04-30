// Package manifest handles reading and writing skell.toml files.
package manifest

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/aminmesbahi/skell/internal/target"
)

// SkillEntry is a single skill declaration inside skell.toml [skills].
type SkillEntry struct {
	Version  string `toml:"version"`
	Registry string `toml:"registry"`
	Pinned   bool   `toml:"pinned"`
}

// Manifest represents the full contents of a skell.toml file.
//
// Target identifies which AI-agent platform layout the manifest is bound to
// ("claude", "codex", "copilot", "cursor"). Empty means the legacy Claude
// layout, kept for backward compatibility with manifests written by Skell
// versions prior to multi-target support.
type Manifest struct {
	Target     string                `toml:"target,omitempty"`
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

// GlobalPath returns the path to the global manifest (~/.skell/.claude/skell.toml).
// The location is preserved for backward compatibility; the global manifest is
// not bound to any specific target.
func GlobalPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".skell", ".claude", "skell.toml"), nil
}

// GlobalRootDir returns the global Skell root directory (~/.skell).
func GlobalRootDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".skell"), nil
}

// EnsureGlobal creates the global manifest if it does not already exist.
func EnsureGlobal() error {
	path, err := GlobalPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return Write(path, &Manifest{
		Registries: map[string]string{},
		Skills:     map[string]SkillEntry{},
	})
}

// LocalPath returns the path to the local manifest inside a repository root
// for the legacy Claude layout. Prefer LocalPathFor when target-awareness
// matters.
func LocalPath(repoRoot string) string {
	return target.MustLookup(target.Default).ManifestPath(repoRoot)
}

// LocalPathFor returns the manifest path inside a repository for the given
// target.
func LocalPathFor(repoRoot string, t target.Target) string {
	return t.ManifestPath(repoRoot)
}

// Resolve returns the effective manifest for a given repository root.
// It probes every known target's directory and returns the first manifest
// found. The Manifest.Target field is populated from the on-disk value when
// present, otherwise inferred from which directory the file came from.
func Resolve(repoRoot string) (*Manifest, error) {
	m, _, err := ResolveWithTarget(repoRoot)
	return m, err
}

// ResolveWithTarget returns the effective manifest along with the target the
// manifest is bound to. Discovery order:
//  1. Manifest in any target directory under repoRoot (target.All() order).
//     The first one found wins.
//  2. Returns a not-found error referring to the legacy Claude path so error
//     messages remain stable for users who never adopted other targets.
func ResolveWithTarget(repoRoot string) (*Manifest, *target.Target, error) {
	for _, t := range target.All() {
		path := t.ManifestPath(repoRoot)
		if _, err := os.Stat(path); err != nil {
			continue
		}
		m, err := Read(path)
		if err != nil {
			return nil, nil, err
		}
		if m.Target == "" {
			m.Target = t.ID
		}
		effective := t
		if m.Target != t.ID {
			if resolved, lookupErr := target.Lookup(m.Target); lookupErr == nil {
				effective = resolved
			}
		}
		return m, &effective, nil
	}
	return nil, nil, fmt.Errorf("open %s: no such file or directory", LocalPath(repoRoot))
}
