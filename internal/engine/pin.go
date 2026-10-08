package engine

import (
	"fmt"

	"github.com/aminmesbahi/skell/internal/audit"
	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/manifest"
)

// Pin marks an installed skill as pinned in skell.toml and skell.lock.
// If version is non-empty it pins to that specific version; otherwise the
// currently installed version is used. Pinning a skill that has no version
// (in either the lock or the override) is rejected because there is nothing
// stable to pin to (see design §8.3).
func (e *Engine) Pin(repoRoot, skillName, version string) error {
	return e.PinFor(repoRoot, skillName, version, "")
}

// PinFor is Pin for an explicit agent target (empty targetID auto-detects).
func (e *Engine) PinFor(repoRoot, skillName, version, targetID string) error {
	if err := ValidateSkillName(skillName); err != nil {
		return err
	}
	m, t, err := resolveManifestFor(repoRoot, targetID)
	if err != nil {
		return fmt.Errorf("no manifest found in %s — run 'skell init' first: %w", repoRoot, err)
	}

	entry, ok := m.Skills[skillName]
	if !ok {
		return fmt.Errorf("skill %q not found in manifest", skillName)
	}

	lockPath := lockfile.PathFor(repoRoot, *t)
	lf, err := lockfile.Read(lockPath)
	if err != nil {
		return fmt.Errorf("lock file not found — run 'skell install %s' first: %w", skillName, err)
	}
	locked := lf.FindSkill(skillName)
	if locked == nil {
		return fmt.Errorf("skill %q not found in lock file — run 'skell install %s' first", skillName, skillName)
	}

	pinVersion := version
	if pinVersion == "" {
		pinVersion = locked.Version
	}
	if pinVersion == "" {
		return fmt.Errorf("cannot pin %q: skill has no version metadata; supply --version to pin to a specific revision", skillName)
	}

	// Update manifest entry.
	entry.Pinned = true
	entry.Version = pinVersion
	m.Skills[skillName] = entry

	// Update lock file entry.
	locked.Pinned = true
	locked.Version = pinVersion
	lf.Upsert(*locked)

	if err := manifest.Write(manifest.LocalPathFor(repoRoot, *t), m); err != nil {
		return fmt.Errorf("failed to update manifest: %w", err)
	}
	if err := lockfile.Write(lockPath, lf); err != nil {
		return err
	}
	_ = e.logger.Log(audit.ActionPin, skillName, pinVersion, "", repoRoot)
	return nil
}

// Unpin removes the pinned flag from a skill in skell.toml and skell.lock.
func (e *Engine) Unpin(repoRoot, skillName string) error {
	return e.UnpinFor(repoRoot, skillName, "")
}

// UnpinFor is Unpin for an explicit agent target (empty targetID auto-detects).
func (e *Engine) UnpinFor(repoRoot, skillName, targetID string) error {
	if err := ValidateSkillName(skillName); err != nil {
		return err
	}
	m, t, err := resolveManifestFor(repoRoot, targetID)
	if err != nil {
		return fmt.Errorf("no manifest found in %s — run 'skell init' first: %w", repoRoot, err)
	}

	entry, ok := m.Skills[skillName]
	if !ok {
		return fmt.Errorf("skill %q not found in manifest", skillName)
	}

	lockPath := lockfile.PathFor(repoRoot, *t)
	lf, err := lockfile.Read(lockPath)
	if err != nil {
		return fmt.Errorf("lock file not found: %w", err)
	}
	locked := lf.FindSkill(skillName)
	if locked == nil {
		return fmt.Errorf("skill %q not found in lock file", skillName)
	}

	entry.Pinned = false
	m.Skills[skillName] = entry

	locked.Pinned = false
	lf.Upsert(*locked)

	if err := manifest.Write(manifest.LocalPathFor(repoRoot, *t), m); err != nil {
		return fmt.Errorf("failed to update manifest: %w", err)
	}
	if err := lockfile.Write(lockPath, lf); err != nil {
		return err
	}
	_ = e.logger.Log(audit.ActionUnpin, skillName, "", "", repoRoot)
	return nil
}
