// Package lockfile handles reading and writing skell.lock files.
package lockfile

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/aminmesbahi/skell/internal/model"
)

// LockFile represents the full contents of a skell.lock file.
type LockFile struct {
	SkellVersion string                 `json:"skell_version"`
	LockedAt     string                 `json:"locked_at"`
	Skills       []model.InstalledSkill `json:"skills"`
}

// Read parses a skell.lock file at the given path.
func Read(path string) (*LockFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var lf LockFile
	if err := json.Unmarshal(data, &lf); err != nil {
		return nil, err
	}
	return &lf, nil
}

// Write serialises a LockFile to a skell.lock file at the given path.
func Write(path string, lf *LockFile) error {
	data, err := json.MarshalIndent(lf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// Path returns the expected lock file path for a given repository root.
func Path(repoRoot string) string {
	return filepath.Join(repoRoot, ".claude", "skell.lock")
}

// FindSkill returns the InstalledSkill entry for the given name, or nil if not present.
func (lf *LockFile) FindSkill(name string) *model.InstalledSkill {
	for i := range lf.Skills {
		if lf.Skills[i].Name == name {
			return &lf.Skills[i]
		}
	}
	return nil
}

// Upsert adds or replaces the lock entry for a skill.
func (lf *LockFile) Upsert(skill model.InstalledSkill) {
	for i := range lf.Skills {
		if lf.Skills[i].Name == skill.Name {
			lf.Skills[i] = skill
			return
		}
	}
	lf.Skills = append(lf.Skills, skill)
}

// Remove deletes the lock entry for a skill by name.
func (lf *LockFile) Remove(name string) {
	filtered := lf.Skills[:0]
	for _, s := range lf.Skills {
		if s.Name != name {
			filtered = append(filtered, s)
		}
	}
	lf.Skills = filtered
}
