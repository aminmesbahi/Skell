package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aminmesbahi/skell/internal/audit"
	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/aminmesbahi/skell/internal/target"
)

// Remove deletes a skill from the target repository and updates skell.toml and skell.lock.
// When dryRun is true no files are modified.
func (e *Engine) Remove(repoRoot, skillName string, dryRun bool) error {
	return e.RemoveFor(repoRoot, skillName, "", dryRun)
}

// RemoveFor is Remove for an explicit agent target (empty targetID
// auto-detects). Mirror copies of the skill are removed too.
func (e *Engine) RemoveFor(repoRoot, skillName, targetID string, dryRun bool) error {
	if err := ValidateSkillName(skillName); err != nil {
		return err
	}
	t, err := resolveTarget(repoRoot, targetID)
	if err != nil {
		return err
	}
	skillDir := filepath.Join(t.SkillsDir(repoRoot), skillName)
	// Lstat so a linked skill whose source folder is gone can still be removed.
	if _, err := os.Lstat(skillDir); os.IsNotExist(err) {
		return fmt.Errorf("skill %q is not installed in %s", skillName, repoRoot)
	}

	if dryRun {
		return nil
	}

	if err := os.RemoveAll(skillDir); err != nil {
		return fmt.Errorf("failed to remove skill %q: %w", skillName, err)
	}

	lockPath := lockfile.PathFor(repoRoot, t)
	if lf, err := lockfile.Read(lockPath); err == nil {
		lf.Remove(skillName)
		if err := lockfile.Write(lockPath, lf); err != nil {
			return fmt.Errorf("skill %q removed from disk but failed to update lock file: %w", skillName, err)
		}
	}

	if m, err := manifest.Read(manifest.LocalPathFor(repoRoot, t)); err == nil {
		delete(m.Skills, skillName)
		if err := manifest.Write(manifest.LocalPathFor(repoRoot, t), m); err != nil {
			return fmt.Errorf("skill %q removed from disk but failed to update manifest: %w", skillName, err)
		}
		e.unmirrorSkill(repoRoot, m, t, skillName)
	}

	_ = e.logger.Log(audit.ActionRemove, skillName, "", "", repoRoot)
	return nil
}

// SyncReport summarises the outcome of a Sync operation.
type SyncReport struct {
	Installed []string
	Removed   []string
	// Untracked lists skills present on disk that are neither in the manifest
	// nor in the lock file (hand-authored, Skell-unmanaged). They are reported
	// but left in place unless prune is requested.
	Untracked []string
	// Mirrored lists skills copied into mirror targets (see Manifest.Mirrors).
	Mirrored []string
}

// SyncDiffError is returned by Sync when checkOnly=true and the repo differs from the manifest.
type SyncDiffError struct {
	Missing []string // in manifest but not installed
	Extra   []string // managed by Skell (in lock) but no longer in the manifest
	// StaleMirrors lists "<target>/<skill>" mirror copies that are missing or
	// differ from the primary install.
	StaleMirrors []string
}

func (e *SyncDiffError) Error() string {
	var parts []string
	if len(e.Missing) > 0 {
		parts = append(parts, "missing: "+strings.Join(e.Missing, ", "))
	}
	if len(e.Extra) > 0 {
		parts = append(parts, "extra: "+strings.Join(e.Extra, ", "))
	}
	if len(e.StaleMirrors) > 0 {
		parts = append(parts, "stale mirrors: "+strings.Join(e.StaleMirrors, ", "))
	}
	return "repo differs from manifest — " + strings.Join(parts, "; ")
}

// Sync applies skell.toml to the repository: installs missing skills and removes
// skills Skell previously managed (present in the lock file) that are no longer
// in the manifest.
//
// Hand-authored skills that were never tracked by Skell (not in the lock file)
// are NOT removed by default — they are reported in SyncReport.Untracked. This
// honours the safety model (design §15 and Open Decisions #4/#5): Skell never
// silently deletes work it did not install. Pass prune=true to also remove them.
//
// checkOnly returns a non-nil *SyncDiffError (exit non-zero) if any managed
// drift exists. dryRun returns the report without writing any files.
func (e *Engine) Sync(repoRoot string, checkOnly, dryRun, prune bool) (*SyncReport, error) {
	m, t, err := manifest.ResolveWithTarget(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("no manifest found in %s — run 'skell init' first: %w", repoRoot, err)
	}

	// Scoped to t: e.List(repoRoot) aggregates installed skills across every
	// detected target, which would let a same-named skill installed for a
	// different target (e.g. .claude) mask a genuinely missing one here, and
	// — worse, with prune=true — mark a different target's skill removable.
	listed, err := e.ListFor(repoRoot, t.ID)
	if err != nil {
		return nil, err
	}
	// The lock file lists what Skell installed, but a skill whose folder was
	// deleted (or never checked out) is not actually present: count only
	// skills that exist on disk so sync restores the missing ones.
	var installed []model.InstalledSkill
	for _, s := range listed {
		if info, err := os.Stat(filepath.Join(t.SkillsDir(repoRoot), s.Name)); err == nil && info.IsDir() {
			installed = append(installed, s)
		}
	}

	locked := lockedSkillNames(repoRoot, *t)
	missing, removable, untracked := computeSyncDiff(m, installed, locked, prune)

	if checkOnly {
		var stale []string
		if lf, err := lockfile.Read(lockfile.PathFor(repoRoot, *t)); err == nil {
			stale = staleMirrors(repoRoot, m, *t, lf)
		}
		if len(missing) > 0 || len(removable) > 0 || len(stale) > 0 {
			return nil, &SyncDiffError{Missing: missing, Extra: removable, StaleMirrors: stale}
		}
		return &SyncReport{Untracked: untracked}, nil
	}

	if dryRun {
		return &SyncReport{Installed: missing, Removed: removable, Untracked: untracked}, nil
	}

	report, err := e.applySyncChanges(repoRoot, m, *t, missing, removable)
	if err != nil {
		return nil, err
	}
	report.Untracked = untracked
	mr, err := e.RefreshMirrors(repoRoot)
	if err != nil {
		return nil, err
	}
	if len(mr.Targets) > 0 {
		report.Mirrored = mr.Copied
	}
	return report, nil
}

// lockedSkillNames returns the set of skill names recorded in the lock file.
func lockedSkillNames(repoRoot string, t target.Target) map[string]bool {
	out := map[string]bool{}
	if lf, err := lockfile.Read(lockfile.PathFor(repoRoot, t)); err == nil {
		for _, s := range lf.Skills {
			out[s.Name] = true
		}
	}
	return out
}

// computeSyncDiff classifies skills into:
//   - missing:   declared in the manifest but absent on disk (to install)
//   - removable: on disk and managed by Skell (in lock) but no longer in the
//     manifest, or — when prune is true — any on-disk skill not in the manifest
//   - untracked: on disk but in neither the manifest nor the lock file, and not
//     being pruned (reported, never deleted)
func computeSyncDiff(m *manifest.Manifest, installed []model.InstalledSkill, locked map[string]bool, prune bool) (missing, removable, untracked []string) {
	installedSet := make(map[string]bool, len(installed))
	for _, s := range installed {
		installedSet[s.Name] = true
	}
	for name := range m.Skills {
		if !installedSet[name] {
			missing = append(missing, name)
		}
	}
	for _, s := range installed {
		if _, ok := m.Skills[s.Name]; ok || s.Linked {
			continue // declared, or a local 'skell link' (never pruned)
		}
		switch {
		case locked[s.Name] || prune:
			removable = append(removable, s.Name)
		default:
			untracked = append(untracked, s.Name)
		}
	}
	sort.Strings(missing)
	sort.Strings(removable)
	sort.Strings(untracked)
	return missing, removable, untracked
}

// applySyncChanges installs missing skills and removes extra skills, returning the report.
//
// A missing skill that is recorded in skell.lock with a registry commit is
// reinstalled at exactly that commit and checked against the locked content
// hash, so every machine (and CI) gets byte-identical skills. Skills without a
// lock entry get the registry's current version.
func (e *Engine) applySyncChanges(repoRoot string, m *manifest.Manifest, t target.Target, missing, extra []string) (*SyncReport, error) {
	report := &SyncReport{}
	lf, _ := lockfile.Read(lockfile.PathFor(repoRoot, t))

	for _, name := range missing {
		if lf != nil {
			if locked := lf.FindSkill(name); locked != nil {
				handled, err := e.installLocked(repoRoot, t, m, *locked)
				if err != nil {
					return nil, fmt.Errorf("failed to install %q during sync: %w", name, err)
				}
				if handled {
					report.Installed = append(report.Installed, name)
					continue
				}
			}
		}
		entry := m.Skills[name]
		alias := entry.Registry
		if alias == "" {
			alias = "default"
		}
		if err := e.InstallTo(repoRoot, name, alias, "", m.Target, false); err != nil {
			return nil, fmt.Errorf("failed to install %q during sync: %w", name, err)
		}
		report.Installed = append(report.Installed, name)
	}

	for _, name := range extra {
		if err := e.RemoveFor(repoRoot, name, m.Target, false); err != nil {
			return nil, fmt.Errorf("failed to remove %q during sync: %w", name, err)
		}
		report.Removed = append(report.Removed, name)
	}

	if len(report.Installed)+len(report.Removed) > 0 {
		_ = e.logger.Log(audit.ActionSync, "", "", "", repoRoot)
	}
	return report, nil
}
