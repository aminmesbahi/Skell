// Package engine is the comparison and action layer.
// It coordinates registry, scanner, manifest, lockfile, and hasher to implement
// every skell command.
package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aminmesbahi/skell/internal/frontmatter"
	"github.com/aminmesbahi/skell/internal/hasher"
	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/aminmesbahi/skell/internal/registry"
	"github.com/aminmesbahi/skell/internal/scanner"
)

// skellVersion is embedded into generated lock files.
const skellVersion = "0.1.0"

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
}

// New creates a ready-to-use Engine backed by the real registry adapter.
func New(cacheRoot string) *Engine {
	return &Engine{provider: registry.NewAdapter(cacheRoot), cacheRoot: cacheRoot}
}

// newWithProvider creates an Engine with an injected provider (used in tests).
func newWithProvider(p RegistryProvider) *Engine {
	return &Engine{provider: p}
}

// List returns all installed skills for the given repository root.
// It reads the lock file when available; falls back to scanning the skills directory.
func (e *Engine) List(repoRoot string) ([]model.InstalledSkill, error) {
	lf, err := lockfile.Read(lockfile.Path(repoRoot))
	if err == nil {
		return lf.Skills, nil
	}

	// No lock file — synthesise entries from the skills directory.
	sr, err := scanner.ScanRepo(repoRoot)
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
		all = append(all, skills...)
	}
	return all, nil
}

// Status returns the comparison between registry and local state for a repository.
func (e *Engine) Status(repoRoot string) ([]model.StatusEntry, error) {
	// TODO: implement
	panic("not implemented")
}

// Info returns the full detail for a single named skill from local state.
// Pass source="registry" to fetch from the remote registry instead (requires a configured registry).
func (e *Engine) Info(repoRoot, skillName, source string) (*model.InfoResult, error) {
	result := &model.InfoResult{}

	// Local frontmatter
	skillDir := filepath.Join(repoRoot, ".claude", "skills", skillName)
	rs, err := frontmatter.ParseDir(skillDir)
	if err == nil {
		result.Skill = rs
	}

	// Lock file entry
	lf, err := lockfile.Read(lockfile.Path(repoRoot))
	if err == nil {
		result.Lock = lf.FindSkill(skillName)
	}

	if result.Skill == nil && result.Lock == nil {
		return nil, fmt.Errorf("skill %q not found in %s", skillName, repoRoot)
	}

	result.Status = model.StatusUpToDate
	if result.Skill != nil && result.Lock != nil {
		ok, err := hasher.Verify(skillDir, result.Lock.ContentHash)
		if err == nil && !ok {
			result.Status = model.StatusLocallyModified
		}
	}

	return result, nil
}

// Install copies a skill from the registry into the target repository.
// When dryRun is true the files are not written; the resolved skill metadata is still fetched.
func (e *Engine) Install(repoRoot, skillName, registryAlias string, dryRun bool) error {
	m, err := manifest.Resolve(repoRoot)
	if err != nil {
		return fmt.Errorf("no manifest found in %s — run 'skell init' first: %w", repoRoot, err)
	}

	if registryAlias == "" {
		registryAlias = "default"
	}
	registryURL, ok := m.Registries[registryAlias]
	if !ok {
		return fmt.Errorf("registry %q not configured in manifest", registryAlias)
	}

	reg := registry.Registry{Alias: registryAlias, URL: registryURL}

	rs, err := e.provider.GetSkill(reg, skillName)
	if err != nil {
		return fmt.Errorf("could not fetch skill %q from registry %q: %w", skillName, registryAlias, err)
	}

	if dryRun {
		return nil
	}

	skillsDir := scanner.SkillsDir(repoRoot)
	destPath := filepath.Join(skillsDir, skillName)

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

	if err := e.updateLockFile(repoRoot, skillName, registryAlias, registryURL, rs, hash); err != nil {
		return err
	}

	return e.updateManifest(repoRoot, m, skillName, registryAlias, rs.Metadata.Version)
}

// updateLockFile adds or replaces the lock entry for the installed skill.
func (e *Engine) updateLockFile(repoRoot, skillName, registryAlias, registryURL string, rs *model.RegistrySkill, hash string) error {
	lockPath := lockfile.Path(repoRoot)

	var lf *lockfile.LockFile
	if _, err := os.Stat(lockPath); err == nil {
		lf, err = lockfile.Read(lockPath)
		if err != nil {
			return fmt.Errorf("failed to read lock file: %w", err)
		}
	} else {
		lf = &lockfile.LockFile{SkellVersion: skellVersion, Skills: []model.InstalledSkill{}}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	lf.LockedAt = now
	lf.Upsert(model.InstalledSkill{
		Name:          skillName,
		Version:       rs.Metadata.Version,
		Registry:      registryAlias,
		SourceRepo:    registryURL,
		InstalledPath: filepath.Join(".claude", "skills", skillName),
		InstalledAt:   now,
		ContentHash:   hash,
	})

	claudeDir := filepath.Dir(lockPath)
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return fmt.Errorf("failed to create .claude directory: %w", err)
	}

	return lockfile.Write(lockPath, lf)
}

// updateManifest adds the skill to skell.toml if not already present.
func (e *Engine) updateManifest(repoRoot string, m *manifest.Manifest, skillName, registryAlias, version string) error {
	if m.Skills == nil {
		m.Skills = make(map[string]manifest.SkillEntry)
	}
	m.Skills[skillName] = manifest.SkillEntry{
		Version:  version,
		Registry: registryAlias,
	}
	manifestPath := manifest.LocalPath(repoRoot)
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0755); err != nil {
		return err
	}
	return manifest.Write(manifestPath, m)
}

// Init creates a skell.toml from the skills currently installed in a repository.
func (e *Engine) Init(repoRoot string) error {
	manifestPath := manifest.LocalPath(repoRoot)
	if _, err := os.Stat(manifestPath); err == nil {
		return fmt.Errorf("skell.toml already exists at %s; delete it first or edit it manually", manifestPath)
	}

	scanResult, err := scanner.ScanRepo(repoRoot)
	if err != nil {
		return fmt.Errorf("failed to scan repository: %w", err)
	}

	skills := make(map[string]manifest.SkillEntry)
	for _, s := range scanResult.InstalledSkills {
		skillDir := filepath.Join(repoRoot, ".claude", "skills", s.Name)
		entry := manifest.SkillEntry{Registry: "default"}
		if rs, err := frontmatter.ParseDir(skillDir); err == nil {
			entry.Version = rs.Metadata.Version
		}
		skills[s.Name] = entry
	}

	m := &manifest.Manifest{
		Registries: map[string]string{},
		Skills:     skills,
	}

	if err := os.MkdirAll(filepath.Dir(manifestPath), 0755); err != nil {
		return fmt.Errorf("failed to create .claude directory: %w", err)
	}
	return manifest.Write(manifestPath, m)
}

// Upgrade updates one or all skills in a repository to the latest registry version.
func (e *Engine) Upgrade(repoRoot, skillName string, force, dryRun bool) error {
	// TODO: implement
	panic("not implemented")
}

// Remove deletes a skill from the target repository.
func (e *Engine) Remove(repoRoot, skillName string, dryRun bool) error {
	// TODO: implement
	panic("not implemented")
}

// Pin marks an installed skill as pinned in skell.toml and skell.lock.
// If version is non-empty it pins to that specific version; otherwise the
// currently installed version is used.
func (e *Engine) Pin(repoRoot, skillName, version string) error {
	m, err := manifest.Resolve(repoRoot)
	if err != nil {
		return fmt.Errorf("no manifest found in %s — run 'skell init' first: %w", repoRoot, err)
	}

	entry, ok := m.Skills[skillName]
	if !ok {
		return fmt.Errorf("skill %q not found in manifest", skillName)
	}

	lf, err := lockfile.Read(lockfile.Path(repoRoot))
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
		return fmt.Errorf("skill %q has no installed version; specify --version explicitly", skillName)
	}

	// Update manifest entry.
	entry.Pinned = true
	entry.Version = pinVersion
	m.Skills[skillName] = entry

	// Update lock file entry.
	locked.Pinned = true
	locked.Version = pinVersion
	lf.Upsert(*locked)

	if err := manifest.Write(manifest.LocalPath(repoRoot), m); err != nil {
		return fmt.Errorf("failed to update manifest: %w", err)
	}
	return lockfile.Write(lockfile.Path(repoRoot), lf)
}

// Unpin removes the pinned flag from a skill in skell.toml and skell.lock.
func (e *Engine) Unpin(repoRoot, skillName string) error {
	m, err := manifest.Resolve(repoRoot)
	if err != nil {
		return fmt.Errorf("no manifest found in %s — run 'skell init' first: %w", repoRoot, err)
	}

	entry, ok := m.Skills[skillName]
	if !ok {
		return fmt.Errorf("skill %q not found in manifest", skillName)
	}

	lf, err := lockfile.Read(lockfile.Path(repoRoot))
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

	if err := manifest.Write(manifest.LocalPath(repoRoot), m); err != nil {
		return fmt.Errorf("failed to update manifest: %w", err)
	}
	return lockfile.Write(lockfile.Path(repoRoot), lf)
}

// Sync applies skell.toml to the repository: installs missing, removes unlisted.
func (e *Engine) Sync(repoRoot string, checkOnly, dryRun bool) error {
	// TODO: implement
	panic("not implemented")
}

// Search queries configured registries for skills matching the criteria.
func (e *Engine) Search(m *manifest.Manifest, query, tag, lifecycle, owner string) ([]model.RegistrySkill, error) {
	// TODO: implement
	panic("not implemented")
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
	// TODO: implement
	panic("not implemented")
}
