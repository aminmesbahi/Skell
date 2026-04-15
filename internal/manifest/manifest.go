// Package manifest handles reading and writing skell.toml files.
package manifest

import "github.com/aminmesbahi/skell/internal/model"

// SkillEntry is a single skill declaration inside skell.toml [skills].
type SkillEntry struct {
	Version  string `toml:"version"`
	Registry string `toml:"registry"`
	Pinned   bool   `toml:"pinned"`
}

// Manifest represents the full contents of a skell.toml file.
type Manifest struct {
	Registries map[string]string        `toml:"registries"`
	Skills     map[string]SkillEntry    `toml:"skills"`
}

// Read parses a skell.toml file at the given path.
func Read(path string) (*Manifest, error) {
	// TODO: implement
	panic("not implemented")
}

// Write serialises a Manifest to a skell.toml file at the given path.
func Write(path string, m *Manifest) error {
	// TODO: implement
	panic("not implemented")
}

// GlobalPath returns the path to the global manifest (~/.skell/skell.toml).
func GlobalPath() (string, error) {
	// TODO: implement
	panic("not implemented")
}

// LocalPath returns the path to the local manifest inside a repository root.
func LocalPath(repoRoot string) string {
	// TODO: implement
	panic("not implemented")
}

// Resolve returns the effective manifest for a given repository root,
// preferring the local manifest over the global one.
func Resolve(repoRoot string) (*Manifest, error) {
	// TODO: implement
	_ = model.SkillMetadata{} // ensure model import is used once wired
	panic("not implemented")
}
