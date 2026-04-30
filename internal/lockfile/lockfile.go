// Package lockfile handles reading and writing skell.lock files.
package lockfile

import (
	"encoding/json"
	"os"

	"github.com/aminmesbahi/skell/internal/model"
	"github.com/aminmesbahi/skell/internal/target"
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

// Path returns the lock file path for a repo using the legacy Claude layout.
// Prefer PathFor when target-awareness matters.
func Path(repoRoot string) string {
	return target.MustLookup(target.Default).LockPath(repoRoot)
}

// PathFor returns the lock file path for a given repository and target.
func PathFor(repoRoot string, t target.Target) string {
	return t.LockPath(repoRoot)
}

// Locate searches every known target directory for a skell.lock and returns
// the first one found. The fallback path (legacy Claude layout) is returned
// when none exists, so callers can still create a fresh lock file in a
// predictable location.
func Locate(repoRoot string) string {
	for _, t := range target.All() {
		p := t.LockPath(repoRoot)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return Path(repoRoot)
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
