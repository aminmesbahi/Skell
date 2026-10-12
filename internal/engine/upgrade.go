package engine

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aminmesbahi/skell/internal/audit"
	"github.com/aminmesbahi/skell/internal/hasher"
	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/aminmesbahi/skell/internal/registry"
	"github.com/aminmesbahi/skell/internal/target"
)

// Upgrade updates one or all skills in a repository to the latest registry version.
// When skillName is empty every upgradeable skill is processed.
// Pinned skills are skipped unless force is true.
// Locally-modified skills halt the upgrade unless force is true.
// When dryRun is true no files are written; the returned report lists what would change.
func (e *Engine) Upgrade(repoRoot, skillName string, force, dryRun bool) (*UpgradeReport, error) {
	return e.UpgradeFor(repoRoot, skillName, "", force, dryRun)
}

// UpgradeFor is Upgrade for an explicit agent target (empty targetID uses the
// manifest found in repoRoot).
func (e *Engine) UpgradeFor(repoRoot, skillName, targetID string, force, dryRun bool) (*UpgradeReport, error) {
	if skillName != "" {
		if err := ValidateSkillName(skillName); err != nil {
			return nil, err
		}
	}
	m, t, err := resolveManifestFor(repoRoot, targetID)
	if err != nil {
		return nil, fmt.Errorf("no manifest found in %s — run 'skell init' first: %w", repoRoot, err)
	}

	lf, err := lockfile.Read(lockfile.PathFor(repoRoot, *t))
	if err != nil {
		return nil, fmt.Errorf("lock file not found — run 'skell install' first: %w", err)
	}

	candidates, err := buildUpgradeCandidates(lf, skillName)
	if err != nil {
		return nil, err
	}

	report := &UpgradeReport{}

	for _, locked := range candidates {
		if err := e.upgradeOne(repoRoot, *t, m, locked, force, dryRun, report); err != nil {
			return nil, err
		}
	}

	if !dryRun && len(report.Upgraded) > 0 {
		manifestPath := manifest.LocalPathFor(repoRoot, *t)
		if err := os.MkdirAll(filepath.Dir(manifestPath), 0755); err != nil {
			return nil, err
		}
		if err := manifest.Write(manifestPath, m); err != nil {
			return nil, err
		}
	}

	return report, nil
}

// buildUpgradeCandidates returns the list of skills to consider for upgrade.
func buildUpgradeCandidates(lf *lockfile.LockFile, skillName string) ([]model.InstalledSkill, error) {
	if skillName == "" {
		return lf.Skills, nil
	}
	locked := lf.FindSkill(skillName)
	if locked == nil {
		return nil, fmt.Errorf("skill %q is not installed", skillName)
	}
	return []model.InstalledSkill{*locked}, nil
}

// upgradeOne processes a single skill candidate: skip, dry-run, or perform the real upgrade.
func (e *Engine) upgradeOne(repoRoot string, t target.Target, m *manifest.Manifest, locked model.InstalledSkill, force, dryRun bool, report *UpgradeReport) error {
	if locked.Pinned && !force {
		report.Skipped = append(report.Skipped, locked.Name+" (pinned)")
		return nil
	}
	if locked.Linked {
		report.Skipped = append(report.Skipped, locked.Name+" (linked to a local folder)")
		return nil
	}

	alias, registryURL, ok := e.resolveRegistryForLocked(m, locked)
	if !ok {
		report.Skipped = append(report.Skipped, locked.Name+" (unknown registry)")
		return nil
	}

	if err := e.pol.CheckRegistry(registryURL); err != nil {
		report.Skipped = append(report.Skipped, locked.Name+" (blocked by policy)")
		return nil
	}

	reg := registry.Registry{Alias: alias, URL: registryURL}
	rs, err := e.provider.GetSkill(reg, locked.Name)
	if err != nil {
		return fmt.Errorf("could not fetch skill %q from registry %q: %w", locked.Name, alias, err)
	}

	skillDir := filepath.Join(t.SkillsDir(repoRoot), locked.Name)

	// Prefer comparing actual content: it catches upstream edits that didn't
	// bump the version and avoids pointless reinstalls of unversioned skills.
	upstreamUnchanged := false
	if changed, known := e.sourceChanged(reg, locked); known {
		upstreamUnchanged = !changed
	} else if rs.Metadata.Version == locked.Version && rs.Metadata.Version != "" {
		upstreamUnchanged = true
	}
	if upstreamUnchanged {
		switch {
		case !locallyModified(skillDir, locked):
			report.Skipped = append(report.Skipped, locked.Name+" (already up-to-date)")
			return nil
		case !force:
			report.Skipped = append(report.Skipped, locked.Name+" (up to date upstream; local edits kept — use --force to discard them)")
			return nil
		}
		// --force on an unchanged source: restore the pristine copy below.
	}

	if err := checkLocallyModified(skillDir, locked, force); err != nil {
		return err
	}

	src := e.skillSource(reg, locked.Name)
	if risk := upgradeRisk(locked.Name, skillDir, rs, src, reg); risk != nil {
		report.NeedsReview = append(report.NeedsReview, *risk)
		if !dryRun && e.upgradeGate != nil && !e.upgradeGate(*risk) {
			report.Skipped = append(report.Skipped, locked.Name+" (needs review: "+risk.Summary()+")")
			return nil
		}
	}

	if dryRun {
		latestCommit := ""
		if src := e.skillSource(reg, locked.Name); src != nil {
			latestCommit = src.Commit
		}
		report.Upgraded = append(report.Upgraded, upgradeLabel(locked, rs, latestCommit))
		return nil
	}

	return e.performSkillUpgrade(repoRoot, t, m, locked, rs, reg, alias, registryURL, skillDir, report)
}

// resolveRegistryForLocked returns the effective registry alias and URL for a locked skill.
func (e *Engine) resolveRegistryForLocked(m *manifest.Manifest, locked model.InstalledSkill) (alias, url string, ok bool) {
	alias = locked.Registry
	if alias == "" {
		alias = "default"
	}
	url, ok = e.resolveRegistryURL(m, alias)
	return alias, url, ok
}

// checkLocallyModified returns an error if the skill has been modified locally and force is false.
func checkLocallyModified(skillDir string, locked model.InstalledSkill, force bool) error {
	if force {
		return nil
	}
	if locallyModified(skillDir, locked) {
		return fmt.Errorf(
			"skill %q has local modifications; use --force to overwrite or commit your changes first",
			locked.Name,
		)
	}
	return nil
}

// locallyModified reports whether the installed copy no longer matches the
// content hash recorded when it was installed.
func locallyModified(skillDir string, locked model.InstalledSkill) bool {
	if locked.ContentHash == "" {
		return false
	}
	ok, err := hasher.Verify(skillDir, locked.ContentHash)
	return err == nil && !ok
}

// performSkillUpgrade copies the new skill version, rehashes, updates lock + manifest, and logs.
func (e *Engine) performSkillUpgrade(repoRoot string, t target.Target, m *manifest.Manifest, locked model.InstalledSkill, rs *model.RegistrySkill, reg registry.Registry, alias, registryURL, skillDir string, report *UpgradeReport) error {
	src := e.skillSource(reg, locked.Name)
	if err := e.copySkill(reg, locked.Name, rs.Metadata.Version, skillDir); err != nil {
		return fmt.Errorf("failed to upgrade skill %q: %w", locked.Name, err)
	}

	hash, err := hasher.HashDir(skillDir)
	if err != nil {
		return fmt.Errorf("failed to hash upgraded skill: %w", err)
	}

	sourceRepo := registryURL
	if locked.SourceRepo != "" {
		sourceRepo = locked.SourceRepo // keep a more specific URL recorded by 'skell add'
	}
	if err := e.updateLockFile(repoRoot, t, locked.Name, alias, sourceRepo, rs, hash, src); err != nil {
		return err
	}

	if entry, exists := m.Skills[locked.Name]; exists {
		entry.Version = rs.Metadata.Version
		m.Skills[locked.Name] = entry
	}

	if err := e.mirrorAfterChange(repoRoot, t, locked.Name); err != nil {
		return err
	}
	_ = e.logger.Log(audit.ActionUpgrade, locked.Name, rs.Metadata.Version, alias, repoRoot)
	latestCommit := ""
	if src != nil {
		latestCommit = src.Commit
	}
	report.Upgraded = append(report.Upgraded, upgradeLabel(locked, rs, latestCommit))
	return nil
}

// upgradeLabel renders "<name> (<old> → <new>)" using versions where skills
// have them and short commits where they don't.
func upgradeLabel(locked model.InstalledSkill, rs *model.RegistrySkill, latestCommit string) string {
	return fmt.Sprintf("%s (%s → %s)", locked.Name,
		model.VersionLabel(locked.Version, locked.Commit), model.VersionLabel(rs.Metadata.Version, latestCommit))
}

// UpgradeReport summarises the outcome of an Upgrade operation.
type UpgradeReport struct {
	Upgraded []string // "<name> (<old> → <new>)"
	Skipped  []string // "<name> (<reason>)"
	// NeedsReview lists upgrades that add scripts or broad tool permissions
	// compared with the installed copy. Those the upgrade gate declined are
	// also in Skipped; in a dry run they are reported but not blocked.
	NeedsReview []UpgradeRisk
}
