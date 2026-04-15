// Package lockfile handles reading and writing skell.lock files.
package lockfile

import "github.com/aminmesbahi/skell/internal/model"

// LockFile represents the full contents of a skell.lock file.
type LockFile struct {
	SkellVersion string                `json:"skell_version"`
	LockedAt     string                `json:"locked_at"`
	Skills       []model.InstalledSkill `json:"skills"`
}

// Read parses a skell.lock file at the given path.
func Read(path string) (*LockFile, error) {
	// TODO: implement
	panic("not implemented")
}

// Write serialises a LockFile to a skell.lock file at the given path.
func Write(path string, lf *LockFile) error {
	// TODO: implement
	panic("not implemented")
}

// Path returns the expected lock file path for a given repository root.
func Path(repoRoot string) string {
	// TODO: implement
	panic("not implemented")
}

// FindSkill returns the InstalledSkill entry for the given name, or nil if not present.
func (lf *LockFile) FindSkill(name string) *model.InstalledSkill {
	// TODO: implement
	panic("not implemented")
}

// Upsert adds or replaces the lock entry for a skill.
func (lf *LockFile) Upsert(skill model.InstalledSkill) {
	// TODO: implement
	panic("not implemented")
}

// Remove deletes the lock entry for a skill by name.
func (lf *LockFile) Remove(name string) {
	// TODO: implement
	panic("not implemented")
}
