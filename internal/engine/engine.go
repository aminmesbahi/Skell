// Package engine is the comparison and action layer.
// It coordinates registry, scanner, manifest, lockfile, and hasher to implement
// every skell command.
package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aminmesbahi/skell/internal/audit"
	"github.com/aminmesbahi/skell/internal/frontmatter"
	"github.com/aminmesbahi/skell/internal/hasher"
	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/aminmesbahi/skell/internal/policy"
	"github.com/aminmesbahi/skell/internal/registry"
	"github.com/aminmesbahi/skell/internal/scanner"
	"github.com/aminmesbahi/skell/internal/target"
	"github.com/aminmesbahi/skell/internal/version"
)

// skellVersion returns the build's CLI version for embedding into lock files.
func skellVersion() string {
	return version.Version
}

// RegistryProvider abstracts registry operations, enabling testability.
type RegistryProvider interface {
	GetSkill(reg registry.Registry, name string) (*model.RegistrySkill, error)
	CopySkillTo(reg registry.Registry, name, version, destPath string) error
	ListSkills(reg registry.Registry) ([]model.RegistrySkill, error)
}

// Engine wires together all internal subsystems.
type Engine struct {
	provider  RegistryProvider
	cacheRoot string
	logger    *audit.Logger
	pol       *policy.Config
}

// New creates a ready-to-use Engine backed by the real registry adapter.
func New(cacheRoot string) *Engine {
	e := &Engine{
		provider:  registry.NewAdapter(cacheRoot),
		cacheRoot: cacheRoot,
		logger:    defaultAuditLogger(),
		pol:       loadPolicy(),
	}
	return e
}

// newWithProvider creates an Engine with an injected provider (used in tests).
func newWithProvider(p RegistryProvider) *Engine {
	return &Engine{provider: p, logger: defaultAuditLogger(), pol: loadPolicy()}
}

// defaultAuditLogger returns a Logger writing to ~/.skell/audit.log.
func defaultAuditLogger() *audit.Logger {
	home, err := os.UserHomeDir()
	if err != nil {
		return audit.NewLogger(filepath.Join(os.TempDir(), ".skell", "audit.log"))
	}
	return audit.NewLogger(filepath.Join(home, ".skell", "audit.log"))
}

// loadPolicy reads ~/.skell/config.toml.
func loadPolicy() *policy.Config {
	home, err := os.UserHomeDir()
	if err != nil {
		return &policy.Config{}
	}
	cfg, err := policy.Read(filepath.Join(home, ".skell", "config.toml"))
	if err != nil {
		return &policy.Config{}
	}
	return cfg
}

// ResolveTarget returns the active target for a repository. Resolution order:
//  1. explicit target recorded in skell.toml
//  2. directory of the manifest that was discovered (e.g. .codex/)
//  3. an existing skills/ directory belonging to a known target
//  4. the default target (claude) so brand-new repos behave as before
func ResolveTarget(repoRoot string) target.Target {
	if _, t, err := manifest.ResolveWithTarget(repoRoot); err == nil && t != nil {
		return *t
	}
	if t, ok := target.DetectPrimary(repoRoot); ok {
		return t
	}
	return target.MustLookup(target.Default)
}

// resolveTargetForExisting is like ResolveTarget but only returns a value when
// the repository already has a manifest or skills directory. Useful for
// commands that must not silently create files in the default location for
// repos that have not yet been initialised.
func resolveTargetForExisting(repoRoot string) (target.Target, bool) {
	if _, t, err := manifest.ResolveWithTarget(repoRoot); err == nil && t != nil {
		return *t, true
	}
	return target.DetectPrimary(repoRoot)
}

// List returns all installed skills for the given repository root.
// It reads the lock file when available; falls back to scanning the skills directory.
func (e *Engine) List(repoRoot string) ([]model.InstalledSkill, error) {
	t := ResolveTarget(repoRoot)
	lf, err := lockfile.Read(lockfile.PathFor(repoRoot, t))
	if err == nil {
		return lf.Skills, nil
	}

	// No lock file — synthesise entries from the skills directory.
	sr, err := scanner.ScanRepoFor(repoRoot, t)
	if err != nil {
		return nil, fmt.Errorf("failed to scan repository: %w", err)
	}
	return sr.InstalledSkills, nil
}

// ListRegistry returns all skills available in all registries configured in the manifest.
func (e *Engine) ListRegistry(m *manifest.Manifest) ([]model.RegistrySkill, error) {
	var all []model.RegistrySkill
	for alias, url := range m.Registries {
		reg := registry.Registry{Alias: alias, URL: url}
		skills, err := e.provider.ListSkills(reg)
		if err != nil {
			return nil, fmt.Errorf("failed to list skills from registry %q: %w", alias, err)
		}
		for i := range skills {
			skills[i].RegistryAlias = alias
			skills[i].RegistryURL = url
		}
		all = append(all, skills...)
	}
	return all, nil
}

// Status returns the comparison between registry and local state for a repository.
// Skills that cannot be found in the registry are marked StatusUnknown.
func (e *Engine) Status(repoRoot string) ([]model.StatusEntry, error) {
	m, t, err := manifest.ResolveWithTarget(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("no manifest found in %s: %w", repoRoot, err)
	}

	lf, err := lockfile.Read(lockfile.PathFor(repoRoot, *t))
	if err != nil {
		return nil, fmt.Errorf("lock file not found — run 'skell sync' to create one: %w", err)
	}

	var entries []model.StatusEntry
	for _, locked := range lf.Skills {
		entries = append(entries, e.statusEntryForSkill(m, *t, repoRoot, locked))
	}
	return entries, nil
}

// statusEntryForSkill derives the status of a single installed skill by consulting
// the local hash, the manifest registry map, and the remote registry metadata.
func (e *Engine) statusEntryForSkill(m *manifest.Manifest, t target.Target, repoRoot string, locked model.InstalledSkill) model.StatusEntry {
	entry := model.StatusEntry{Name: locked.Name, Installed: locked.Version}

	if locked.Pinned {
		entry.Status = model.StatusPinned
		return entry
	}

	skillDir := filepath.Join(t.SkillsDir(repoRoot), locked.Name)
	if locked.ContentHash != "" {
		ok, hashErr := hasher.Verify(skillDir, locked.ContentHash)
		if hashErr != nil {
			entry.Status = model.StatusUnknown
			return entry
		}
		if !ok {
			entry.Status = model.StatusLocallyModified
			return entry
		}
	}

	alias := locked.Registry
	if alias == "" {
		alias = "default"
	}
	registryURL, ok := m.Registries[alias]
	if !ok {
		entry.Status = model.StatusUnknown
		return entry
	}

	rs, err := e.provider.GetSkill(registry.Registry{Alias: alias, URL: registryURL}, locked.Name)
	if err != nil {
		entry.Status = model.StatusUnknown
		return entry
	}

	entry.Latest = rs.Metadata.Version
	entry.Status = resolveVersionStatus(locked.Version, rs)
	return entry
}

// resolveVersionStatus maps registry lifecycle and version data to a SkillStatus.
func resolveVersionStatus(installedVersion string, rs *model.RegistrySkill) model.SkillStatus {
	switch rs.Metadata.Lifecycle {
	case model.LifecycleDeprecated:
		return model.StatusDeprecated
	case model.LifecycleArchived:
		return model.StatusArchived
	}
	if installedVersion == "" && rs.Metadata.Version == "" {
		// Both unversioned: treat as unversioned (caller decides whether to
		// reinstall on upgrade).
		return model.StatusUnversioned
	}
	if installedVersion == "" {
		return model.StatusMissingMetadata
	}
	if rs.Metadata.Version == "" {
		return model.StatusUnversioned
	}
	if rs.Metadata.Version != installedVersion {
		return model.StatusOutdated
	}
	return model.StatusUpToDate
}

// Info returns the full detail for a single named skill from local state.
// Pass source="registry" to fetch from the remote registry instead (requires a configured registry).
func (e *Engine) Info(repoRoot, skillName, source string) (*model.InfoResult, error) {
	result := &model.InfoResult{}
	t := ResolveTarget(repoRoot)

	if source != "registry" {
		// Local frontmatter
		skillDir := filepath.Join(t.SkillsDir(repoRoot), skillName)
		if rs, err := frontmatter.ParseDir(skillDir); err == nil {
			result.Skill = rs
		}

		// Lock file entry
		if lf, err := lockfile.Read(lockfile.PathFor(repoRoot, t)); err == nil {
			result.Lock = lf.FindSkill(skillName)
		}

		if result.Skill != nil || result.Lock != nil {
			result.Status = model.StatusUpToDate
			if result.Skill != nil && result.Lock != nil {
				if ok, err := hasher.Verify(skillDir, result.Lock.ContentHash); err == nil && !ok {
					result.Status = model.StatusLocallyModified
				}
			}
			return result, nil
		}

		if source == "local" {
			return nil, fmt.Errorf("skill %q not found in %s", skillName, repoRoot)
		}
	}

	// Registry lookup.
	m, err := manifest.Resolve(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("skill %q not found in %s", skillName, repoRoot)
	}
	for alias, url := range m.Registries {
		reg := registry.Registry{Alias: alias, URL: url}
		rs, err := e.provider.GetSkill(reg, skillName)
		if err != nil {
			continue
		}
		result.Skill = rs
		result.Status = model.StatusUnknown
		return result, nil
	}

	return nil, fmt.Errorf("skill %q not found in %s or any configured registry", skillName, repoRoot)
}

// Install copies a skill from the registry into the target repository.
// When dryRun is true no files are written.
func (e *Engine) Install(repoRoot, skillName, registryAlias, registryURL string, dryRun bool) error {
	m, t, err := manifest.ResolveWithTarget(repoRoot)
	if err != nil {
		return fmt.Errorf("no manifest found in %s — run 'skell init' first: %w", repoRoot, err)
	}

	skillDir := filepath.Join(t.SkillsDir(repoRoot), skillName)
	if lf, err := lockfile.Read(lockfile.PathFor(repoRoot, *t)); err == nil {
		if locked := lf.FindSkill(skillName); locked != nil {
			if info, statErr := os.Stat(skillDir); statErr == nil && info.IsDir() {
				return fmt.Errorf("skill %q is already installed; use 'skell upgrade %s' or 'skell remove %s' first", skillName, skillName, skillName)
			}
		}
	}

	if registryAlias == "" {
		registryAlias = "default"
	}

	existingURL, ok := m.Registries[registryAlias]
	registryNeedsAdding := false
	if !ok {
		if registryURL == "" {
			return fmt.Errorf("registry %q not configured in manifest — add it to skell.toml or supply --registry-url <url>", registryAlias)
		}
		existingURL = registryURL
		registryNeedsAdding = true
	}

	if err := e.pol.CheckRegistry(existingURL); err != nil {
		return err
	}

	// Auto-register only on a real install; a preview must not edit skell.toml.
	if registryNeedsAdding && !dryRun {
		if m.Registries == nil {
			m.Registries = make(map[string]string)
		}
		m.Registries[registryAlias] = registryURL
		if err := manifest.Write(manifest.LocalPathFor(repoRoot, *t), m); err != nil {
			return fmt.Errorf("failed to add registry %q to manifest: %w", registryAlias, err)
		}
	}

	reg := registry.Registry{Alias: registryAlias, URL: existingURL}

	rs, err := e.provider.GetSkill(reg, skillName)
	if err != nil {
		return fmt.Errorf("could not fetch skill %q from registry %q: %w", skillName, registryAlias, err)
	}

	if dryRun {
		return nil
	}

	skillsDir := t.SkillsDir(repoRoot)
	destPath := skillDir

	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return fmt.Errorf("failed to create skills directory: %w", err)
	}

	if err := e.provider.CopySkillTo(reg, skillName, rs.Metadata.Version, destPath); err != nil {
		return fmt.Errorf("failed to copy skill %q: %w", skillName, err)
	}

	hash, err := hasher.HashDir(destPath)
	if err != nil {
		return fmt.Errorf("failed to hash installed skill: %w", err)
	}

	if err := e.updateLockFile(repoRoot, *t, skillName, registryAlias, existingURL, rs, hash); err != nil {
		return err
	}

	if err := e.updateManifest(repoRoot, *t, m, skillName, registryAlias, rs.Metadata.Version); err != nil {
		return err
	}

	_ = e.logger.Log(audit.ActionInstall, skillName, rs.Metadata.Version, registryAlias, repoRoot)
	return nil
}

// updateLockFile adds or replaces the lock entry for the installed skill.
func (e *Engine) updateLockFile(repoRoot string, t target.Target, skillName, registryAlias, registryURL string, rs *model.RegistrySkill, hash string) error {
	lockPath := lockfile.PathFor(repoRoot, t)

	var lf *lockfile.LockFile
	if _, err := os.Stat(lockPath); err == nil {
		lf, err = lockfile.Read(lockPath)
		if err != nil {
			return fmt.Errorf("failed to read lock file: %w", err)
		}
	} else {
		lf = &lockfile.LockFile{SkellVersion: skellVersion(), Skills: []model.InstalledSkill{}}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	lf.LockedAt = now
	lf.Upsert(model.InstalledSkill{
		Name:          skillName,
		Version:       rs.Metadata.Version,
		Registry:      registryAlias,
		SourceRepo:    registryURL,
		InstalledPath: t.InstalledRelPath(skillName),
		InstalledAt:   now,
		ContentHash:   hash,
	})

	if err := os.MkdirAll(filepath.Dir(lockPath), 0755); err != nil {
		return fmt.Errorf("failed to create %s directory: %w", t.Dir, err)
	}

	return lockfile.Write(lockPath, lf)
}

// updateManifest adds the skill to skell.toml if not already present.
func (e *Engine) updateManifest(repoRoot string, t target.Target, m *manifest.Manifest, skillName, registryAlias, version string) error {
	if m.Skills == nil {
		m.Skills = make(map[string]manifest.SkillEntry)
	}
	if m.Target == "" {
		m.Target = t.ID
	}
	m.Skills[skillName] = manifest.SkillEntry{
		Version:  version,
		Registry: registryAlias,
	}
	manifestPath := manifest.LocalPathFor(repoRoot, t)
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0755); err != nil {
		return err
	}
	return manifest.Write(manifestPath, m)
}

// Init creates a skell.toml from the skills currently installed in a repository.
// The target argument selects the on-disk layout (claude, codex, copilot, cursor).
// Pass an empty target to auto-detect from existing folders, falling back to the
// default (Claude) for fresh repos.
func (e *Engine) Init(repoRoot string) error {
	return e.InitFor(repoRoot, target.Target{})
}

// InitFor is like Init but lets the caller pin the layout to a specific target.
// When t is the zero value, the active target is auto-detected.
func (e *Engine) InitFor(repoRoot string, t target.Target) error {
	if t.ID == "" {
		if detected, ok := target.DetectPrimary(repoRoot); ok {
			t = detected
		} else {
			t = target.MustLookup(target.Default)
		}
	}

	manifestPath := manifest.LocalPathFor(repoRoot, t)
	if _, err := os.Stat(manifestPath); err == nil {
		return fmt.Errorf("skell.toml already exists at %s; delete it first or edit it manually", manifestPath)
	}

	scanResult, err := scanner.ScanRepoFor(repoRoot, t)
	if err != nil {
		return fmt.Errorf("failed to scan repository: %w", err)
	}

	skills := make(map[string]manifest.SkillEntry)
	for _, s := range scanResult.InstalledSkills {
		skillDir := filepath.Join(t.SkillsDir(repoRoot), s.Name)
		entry := manifest.SkillEntry{}
		if rs, err := frontmatter.ParseDir(skillDir); err == nil {
			entry.Version = rs.Metadata.Version
		}
		skills[s.Name] = entry
	}

	m := &manifest.Manifest{
		Target:     t.ID,
		Registries: map[string]string{},
		Skills:     skills,
	}

	if err := os.MkdirAll(filepath.Dir(manifestPath), 0755); err != nil {
		return fmt.Errorf("failed to create %s directory: %w", t.Dir, err)
	}
	return manifest.Write(manifestPath, m)
}

// Upgrade updates one or all skills in a repository to the latest registry version.
// When skillName is empty every upgradeable skill is processed.
// Pinned skills are skipped unless force is true.
// Locally-modified skills halt the upgrade unless force is true.
// When dryRun is true no files are written; the returned report lists what would change.
func (e *Engine) Upgrade(repoRoot, skillName string, force, dryRun bool) (*UpgradeReport, error) {
	m, t, err := manifest.ResolveWithTarget(repoRoot)
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

	alias, registryURL, ok := resolveRegistryForLocked(m, locked)
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

	if rs.Metadata.Version == locked.Version && rs.Metadata.Version != "" {
		report.Skipped = append(report.Skipped, locked.Name+" (already up-to-date)")
		return nil
	}

	skillDir := filepath.Join(t.SkillsDir(repoRoot), locked.Name)
	if err := checkLocallyModified(skillDir, locked, force); err != nil {
		return err
	}

	if dryRun {
		report.Upgraded = append(report.Upgraded, fmt.Sprintf("%s (%s → %s)", locked.Name, locked.Version, rs.Metadata.Version))
		return nil
	}

	return e.performSkillUpgrade(repoRoot, t, m, locked, rs, reg, alias, registryURL, skillDir, report)
}

// resolveRegistryForLocked returns the effective registry alias and URL for a locked skill.
func resolveRegistryForLocked(m *manifest.Manifest, locked model.InstalledSkill) (alias, url string, ok bool) {
	alias = locked.Registry
	if alias == "" {
		alias = "default"
	}
	url, ok = m.Registries[alias]
	return alias, url, ok
}

// checkLocallyModified returns an error if the skill has been modified locally and force is false.
func checkLocallyModified(skillDir string, locked model.InstalledSkill, force bool) error {
	if locked.ContentHash == "" || force {
		return nil
	}
	ok, hashErr := hasher.Verify(skillDir, locked.ContentHash)
	if hashErr == nil && !ok {
		return fmt.Errorf(
			"skill %q has local modifications; use --force to overwrite or commit your changes first",
			locked.Name,
		)
	}
	return nil
}

// performSkillUpgrade copies the new skill version, rehashes, updates lock + manifest, and logs.
func (e *Engine) performSkillUpgrade(repoRoot string, t target.Target, m *manifest.Manifest, locked model.InstalledSkill, rs *model.RegistrySkill, reg registry.Registry, alias, registryURL, skillDir string, report *UpgradeReport) error {
	if err := e.provider.CopySkillTo(reg, locked.Name, rs.Metadata.Version, skillDir); err != nil {
		return fmt.Errorf("failed to copy skill %q: %w", locked.Name, err)
	}

	hash, err := hasher.HashDir(skillDir)
	if err != nil {
		return fmt.Errorf("failed to hash upgraded skill: %w", err)
	}

	if err := e.updateLockFile(repoRoot, t, locked.Name, alias, registryURL, rs, hash); err != nil {
		return err
	}

	if entry, exists := m.Skills[locked.Name]; exists {
		entry.Version = rs.Metadata.Version
		m.Skills[locked.Name] = entry
	}

	_ = e.logger.Log(audit.ActionUpgrade, locked.Name, rs.Metadata.Version, alias, repoRoot)
	report.Upgraded = append(report.Upgraded, fmt.Sprintf("%s (%s → %s)", locked.Name, locked.Version, rs.Metadata.Version))
	return nil
}

// UpgradeReport summarises the outcome of an Upgrade operation.
type UpgradeReport struct {
	Upgraded []string // "<name> (<old> → <new>)"
	Skipped  []string // "<name> (<reason>)"
}

// Remove deletes a skill from the target repository and updates skell.toml and skell.lock.
// When dryRun is true no files are modified.
func (e *Engine) Remove(repoRoot, skillName string, dryRun bool) error {
	t := ResolveTarget(repoRoot)
	skillDir := filepath.Join(t.SkillsDir(repoRoot), skillName)
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		return fmt.Errorf("skill %q is not installed in %s", skillName, repoRoot)
	}

	if dryRun {
		return nil
	}

	if err := os.RemoveAll(skillDir); err != nil {
		return fmt.Errorf("failed to remove skill %q: %w", skillName, err)
	}

	lockPath := lockfile.PathFor(repoRoot, t)
	lf, err := lockfile.Read(lockPath)
	if err == nil {
		lf.Remove(skillName)
		_ = lockfile.Write(lockPath, lf)
	}

	m, err := manifest.Resolve(repoRoot)
	if err == nil {
		delete(m.Skills, skillName)
		_ = manifest.Write(manifest.LocalPathFor(repoRoot, t), m)
	}

	_ = e.logger.Log(audit.ActionRemove, skillName, "", "", repoRoot)
	return nil
}

// SyncReport summarises the outcome of a Sync operation.
type SyncReport struct {
	Installed []string
	Removed   []string
}

// SyncDiffError is returned by Sync when checkOnly=true and the repo differs from the manifest.
type SyncDiffError struct {
	Missing []string // in manifest but not installed
	Extra   []string // installed but not in manifest
}

func (e *SyncDiffError) Error() string {
	var parts []string
	if len(e.Missing) > 0 {
		parts = append(parts, "missing: "+strings.Join(e.Missing, ", "))
	}
	if len(e.Extra) > 0 {
		parts = append(parts, "extra: "+strings.Join(e.Extra, ", "))
	}
	return "repo differs from manifest — " + strings.Join(parts, "; ")
}

// Sync applies skell.toml to the repository: installs missing skills, removes unlisted ones.
// checkOnly returns a non-nil *SyncDiffError (exit non-zero) if any differences exist.
// dryRun returns the report without writing any files.
func (e *Engine) Sync(repoRoot string, checkOnly, dryRun bool) (*SyncReport, error) {
	m, err := manifest.Resolve(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("no manifest found in %s — run 'skell init' first: %w", repoRoot, err)
	}

	installed, err := e.List(repoRoot)
	if err != nil {
		return nil, err
	}

	missing, extra := computeSyncDiff(m, installed)

	if checkOnly {
		if len(missing) > 0 || len(extra) > 0 {
			return nil, &SyncDiffError{Missing: missing, Extra: extra}
		}
		return &SyncReport{}, nil
	}

	if dryRun {
		return &SyncReport{Installed: missing, Removed: extra}, nil
	}

	return e.applySyncChanges(repoRoot, m, missing, extra)
}

// computeSyncDiff returns which skills are missing from disk and which are extra (not in manifest).
func computeSyncDiff(m *manifest.Manifest, installed []model.InstalledSkill) (missing, extra []string) {
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
		if _, ok := m.Skills[s.Name]; !ok {
			extra = append(extra, s.Name)
		}
	}
	return missing, extra
}

// applySyncChanges installs missing skills and removes extra skills, returning the report.
func (e *Engine) applySyncChanges(repoRoot string, m *manifest.Manifest, missing, extra []string) (*SyncReport, error) {
	report := &SyncReport{}

	for _, name := range missing {
		entry := m.Skills[name]
		alias := entry.Registry
		if alias == "" {
			alias = "default"
		}
		if err := e.Install(repoRoot, name, alias, "", false); err != nil {
			return nil, fmt.Errorf("failed to install %q during sync: %w", name, err)
		}
		report.Installed = append(report.Installed, name)
	}

	for _, name := range extra {
		if err := e.Remove(repoRoot, name, false); err != nil {
			return nil, fmt.Errorf("failed to remove %q during sync: %w", name, err)
		}
		report.Removed = append(report.Removed, name)
	}

	if len(report.Installed)+len(report.Removed) > 0 {
		_ = e.logger.Log(audit.ActionSync, "", "", "", repoRoot)
	}
	return report, nil
}

// Search queries configured registries for skills matching query, tag, lifecycle, and owner.
// All filters are optional; an empty filter matches everything.
func (e *Engine) Search(m *manifest.Manifest, query, tag, lifecycle, owner string) ([]model.RegistrySkill, error) {
	all, err := e.ListRegistry(m)
	if err != nil {
		return nil, err
	}

	var results []model.RegistrySkill
	for _, s := range all {
		if matchesFilter(s, query, tag, lifecycle, owner) {
			results = append(results, s)
		}
	}
	return results, nil
}

// SearchMerged queries skills from the local manifest of repoRoot AND from the
// global manifest (~/.skell), stamping each result with RegistrySource="local"
// or RegistrySource="global". When repoRoot IS the global root, all results are
// stamped "global". Local results take priority: duplicate (alias+name) pairs
// from the global manifest are dropped.
func (e *Engine) SearchMerged(repoRoot, query, tag, lifecycle, owner string) ([]model.RegistrySkill, error) {
	globalRoot, _ := manifest.GlobalRootDir()
	isGlobal := filepath.Clean(repoRoot) == filepath.Clean(globalRoot)

	// Resolve local manifest (or global if repoRoot == globalRoot).
	localM, localErr := manifest.Resolve(repoRoot)

	if isGlobal {
		if localErr != nil {
			return nil, localErr
		}
		skills, err := e.ListRegistry(localM)
		if err != nil {
			return nil, err
		}
		for i := range skills {
			skills[i].RegistrySource = "global"
		}
		var results []model.RegistrySkill
		for _, s := range skills {
			if matchesFilter(s, query, tag, lifecycle, owner) {
				results = append(results, s)
			}
		}
		return results, nil
	}

	// Local skills.
	var merged []model.RegistrySkill
	seen := make(map[string]bool)

	if localErr == nil {
		localSkills, err := e.ListRegistry(localM)
		if err == nil {
			for i := range localSkills {
				localSkills[i].RegistrySource = "local"
				key := localSkills[i].RegistryAlias + "/" + localSkills[i].Name
				seen[key] = true
			}
			merged = append(merged, localSkills...)
		}
	}

	// Merge global skills — skip duplicates already in local.
	globalPath, err := manifest.GlobalPath()
	if err == nil {
		globalM, err := manifest.Read(globalPath)
		if err == nil {
			globalSkills, err := e.ListRegistry(globalM)
			if err == nil {
				for i := range globalSkills {
					globalSkills[i].RegistrySource = "global"
					key := globalSkills[i].RegistryAlias + "/" + globalSkills[i].Name
					if !seen[key] {
						merged = append(merged, globalSkills[i])
					}
				}
			}
		}
	}

	var results []model.RegistrySkill
	for _, s := range merged {
		if matchesFilter(s, query, tag, lifecycle, owner) {
			results = append(results, s)
		}
	}
	return results, nil
}

// matchesFilter returns true when the skill satisfies all non-empty filter criteria.
func matchesFilter(s model.RegistrySkill, query, tag, lifecycle, owner string) bool {
	if query != "" {
		q := strings.ToLower(query)
		if !strings.Contains(strings.ToLower(s.Name), q) &&
			!strings.Contains(strings.ToLower(s.Description), q) &&
			!strings.Contains(strings.ToLower(s.Metadata.Tags), q) {
			return false
		}
	}
	if tag != "" && !strings.Contains(strings.ToLower(s.Metadata.Tags), strings.ToLower(tag)) {
		return false
	}
	if lifecycle != "" && string(s.Metadata.Lifecycle) != lifecycle {
		return false
	}
	if owner != "" && !strings.EqualFold(s.Metadata.Owner, owner) {
		return false
	}
	return true
}

// Pin marks an installed skill as pinned in skell.toml and skell.lock.
// If version is non-empty it pins to that specific version; otherwise the
// currently installed version is used. Pinning a skill that has no version
// (in either the lock or the override) is rejected because there is nothing
// stable to pin to (see design §8.3).
func (e *Engine) Pin(repoRoot, skillName, version string) error {
	m, t, err := manifest.ResolveWithTarget(repoRoot)
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
	m, t, err := manifest.ResolveWithTarget(repoRoot)
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

// CacheStatus returns a human-readable summary of the local registry cache.
func (e *Engine) CacheStatus() (string, error) {
	a := registry.NewAdapter(e.cacheRoot)
	return a.CacheStatus()
}

// CacheClear removes all locally cached registry data.
func (e *Engine) CacheClear() error {
	a := registry.NewAdapter(e.cacheRoot)
	return a.CacheClear()
}

// CacheRefresh fetches the latest from all registries configured in the manifest.
func (e *Engine) CacheRefresh(m *manifest.Manifest) error {
	a := registry.NewAdapter(e.cacheRoot)
	for alias, url := range m.Registries {
		reg := registry.Registry{Alias: alias, URL: url}
		if err := a.CacheRefresh(reg); err != nil {
			return fmt.Errorf("failed to refresh registry %q: %w", alias, err)
		}
	}
	return nil
}

// Doctor runs all diagnostic checks on a repository.
func (e *Engine) Doctor(repoRoot string) ([]DiagnosticIssue, error) {
	var issues []DiagnosticIssue

	// 1. Manifest
	m, t, err := manifest.ResolveWithTarget(repoRoot)
	if err != nil {
		issues = append(issues, DiagnosticIssue{
			Severity: SeverityError,
			Code:     "no-manifest",
			Message:  "no manifest (skell.toml) found",
			Hint:     "run 'skell init' to create one",
		})
		return issues, nil
	}

	// 2. Lock file
	lockPath := lockfile.PathFor(repoRoot, *t)
	lf, err := lockfile.Read(lockPath)
	if err != nil {
		issues = append(issues, DiagnosticIssue{
			Severity: SeverityWarning,
			Code:     "no-lockfile",
			Message:  "no lock file found (skell.lock)",
			Hint:     "run 'skell sync' to create one",
		})
		lf = &lockfile.LockFile{}
	}

	// 3. Per-skill checks
	skillsDir := t.SkillsDir(repoRoot)
	for _, locked := range lf.Skills {
		skillDir := filepath.Join(skillsDir, locked.Name)

		// Directory present?
		if _, err := os.Stat(skillDir); os.IsNotExist(err) {
			issues = append(issues, DiagnosticIssue{
				Severity: SeverityError,
				Code:     "missing-dir",
				Message:  fmt.Sprintf("skill %q is in lock file but directory is missing", locked.Name),
				Hint:     fmt.Sprintf("run 'skell install %s' to reinstall", locked.Name),
			})
			continue
		}

		// SKILL.md parseable?
		if _, err := frontmatter.ParseDir(skillDir); err != nil {
			issues = append(issues, DiagnosticIssue{
				Severity: SeverityWarning,
				Code:     "malformed-frontmatter",
				Message:  fmt.Sprintf("skill %q: SKILL.md is missing or malformed", locked.Name),
				Hint:     "verify the skill directory is intact",
			})
		}

		// Content hash matches?
		if locked.ContentHash != "" {
			ok, err := hasher.Verify(skillDir, locked.ContentHash)
			if err != nil {
				issues = append(issues, DiagnosticIssue{
					Severity: SeverityWarning,
					Code:     "hash-error",
					Message:  fmt.Sprintf("skill %q: could not verify content hash: %v", locked.Name, err),
				})
			} else if !ok {
				issues = append(issues, DiagnosticIssue{
					Severity: SeverityWarning,
					Code:     "locally-modified",
					Message:  fmt.Sprintf("skill %q: content hash mismatch (locally modified)", locked.Name),
					Hint:     "run 'skell upgrade' or 'skell install' to restore",
				})
			}
		}
	}

	// 4. Installed skills not in manifest
	if _, err := os.Stat(skillsDir); err == nil {
		entries, _ := os.ReadDir(skillsDir)
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			if _, ok := m.Skills[entry.Name()]; !ok {
				issues = append(issues, DiagnosticIssue{
					Severity: SeverityWarning,
					Code:     "untracked-skill",
					Message:  fmt.Sprintf("skill %q is installed but not in manifest", entry.Name()),
					Hint:     "run 'skell sync' to reconcile",
				})
			}
		}
	}

	return issues, nil
}
