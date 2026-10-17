package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/aminmesbahi/skell/internal/audit"
	"github.com/aminmesbahi/skell/internal/frontmatter"
	"github.com/aminmesbahi/skell/internal/hasher"
	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/aminmesbahi/skell/internal/registry"
	"github.com/aminmesbahi/skell/internal/scanner"
	"github.com/aminmesbahi/skell/internal/target"
)

// Install copies a skill from the registry into the target repository.
// When dryRun is true no files are written.
//
// Deprecated: prefer InstallTo which allows specifying a target agent platform.
func (e *Engine) Install(repoRoot, skillName, registryAlias, registryURL string, dryRun bool) error {
	return e.InstallTo(repoRoot, skillName, registryAlias, registryURL, "", dryRun)
}

// InstallTo copies a skill from the registry into the target repository for
// the specified agent platform. When targetID is empty the target is resolved
// from the existing manifest (same as Install). When targetID is provided, the
// skill is installed into that agent's layout; if no manifest exists for the
// target, one is auto-created (equivalent to running `skell init --target <id>`
// first). When dryRun is true no files are written.
func (e *Engine) InstallTo(repoRoot, skillName, registryAlias, registryURL, targetID string, dryRun bool) (retErr error) {
	if err := ValidateSkillName(skillName); err != nil {
		return err
	}

	var m *manifest.Manifest
	var t *target.Target

	if targetID != "" {
		// Explicit target: resolve it and ensure a manifest exists.
		resolved, err := target.Lookup(targetID)
		if err != nil {
			return fmt.Errorf("invalid --target: %w", err)
		}
		t = &resolved
		manifestPath := t.ManifestPath(repoRoot)
		if _, statErr := os.Stat(manifestPath); statErr == nil {
			m, err = manifest.Read(manifestPath)
			if err != nil {
				return err
			}
		} else {
			// An agent that the project already serves as a mirror of another
			// agent has no manifest of its own: install through that primary
			// (which copies the skill into the mirror) instead of creating a
			// second, competing manifest.
			if primary, ok := mirroringPrimary(repoRoot, *t); ok {
				if lf, lerr := lockfile.Read(lockfile.PathFor(repoRoot, primary)); lerr == nil && lf.FindSkill(skillName) != nil {
					if info, serr := os.Stat(filepath.Join(primary.SkillsDir(repoRoot), skillName)); serr == nil && info.IsDir() {
						return nil // already installed there and mirrored
					}
				}
				return e.InstallTo(repoRoot, skillName, registryAlias, registryURL, primary.ID, dryRun)
			}
			// Auto-init the target. It starts with the sources the project's
			// other agent manifests already use, so "install for Copilot" in a
			// project set up for Claude finds the same skills. A dry run only
			// describes this in memory, and a failed install removes what it
			// created instead of leaving an empty manifest behind.
			inherited := inheritedSources(repoRoot, *t)
			if dryRun {
				m = &manifest.Manifest{Target: t.ID, Registries: inherited, Skills: map[string]manifest.SkillEntry{}}
			} else {
				if err := e.InitFor(repoRoot, *t); err != nil {
					return fmt.Errorf("auto-init for target %s failed: %w", t.ID, err)
				}
				defer func() {
					if retErr != nil {
						removeAutoInit(repoRoot, *t)
					}
				}()
				m, err = manifest.Read(manifestPath)
				if err != nil {
					return err
				}
				if len(inherited) > 0 {
					m.Registries = inherited
					if err := manifest.Write(manifestPath, m); err != nil {
						return err
					}
				}
			}
		}
	} else {
		var resolveErr error
		m, t, resolveErr = manifest.ResolveWithTarget(repoRoot)
		if resolveErr != nil {
			return fmt.Errorf("no manifest found in %s — run 'skell init' first: %w", repoRoot, resolveErr)
		}
	}

	// An unknown --registry is reported before anything else, so it is not
	// masked by an "already installed" error for the same skill.
	if registryAlias != "" && registryURL == "" {
		if _, ok := e.resolveRegistryURL(m, registryAlias); !ok {
			return fmt.Errorf("registry %q not configured in manifest — add it to skell.toml or supply --registry-url <url>", registryAlias)
		}
	}

	skillDir := filepath.Join(t.SkillsDir(repoRoot), skillName)
	if err := ensureSkillNotInstalled(repoRoot, *t, skillDir, skillName); err != nil {
		return err
	}

	registryAlias, existingURL, registryNeedsAdding, err := e.resolveInstallRegistry(m, skillName, registryAlias, registryURL)
	if err != nil {
		return err
	}

	if err := e.pol.CheckRegistry(existingURL); err != nil {
		return err
	}

	reg := registry.Registry{Alias: registryAlias, URL: existingURL}
	rs, err := e.fetchInstallSkill(reg, skillName, registryAlias)
	if err != nil {
		return err
	}

	// Register the registry only after the skill was found, so a failed
	// install never leaves a stale registry entry behind.
	if err := autoRegisterInstallRegistry(repoRoot, *t, m, registryAlias, registryURL, registryNeedsAdding, dryRun); err != nil {
		return err
	}

	if dryRun {
		return nil
	}
	return e.installFetchedSkill(repoRoot, *t, m, reg, skillName, skillDir, existingURL, rs)
}

// mirroringPrimary finds the agent whose manifest lists t as a mirror target.
func mirroringPrimary(repoRoot string, t target.Target) (target.Target, bool) {
	for _, other := range target.All() {
		if other.ID == t.ID {
			continue
		}
		om, err := manifest.Read(other.ManifestPath(repoRoot))
		if err != nil {
			continue
		}
		if slices.Contains(om.Mirrors, t.ID) {
			return other, true
		}
	}
	return target.Target{}, false
}

// inheritedSources returns the skill sources declared by the project's
// manifests for agents other than except, so a newly enabled agent starts with
// the same sources.
func inheritedSources(repoRoot string, except target.Target) map[string]string {
	out := map[string]string{}
	for _, other := range target.All() {
		if other.ID == except.ID {
			continue
		}
		om, err := manifest.Read(other.ManifestPath(repoRoot))
		if err != nil {
			continue
		}
		for alias, url := range om.Registries {
			if _, taken := out[alias]; !taken {
				out[alias] = url
			}
		}
	}
	return out
}

// removeAutoInit deletes the manifest created by an install that then failed,
// and its folder when nothing else is in it.
func removeAutoInit(repoRoot string, t target.Target) {
	if entries, err := os.ReadDir(t.SkillsDir(repoRoot)); err == nil && len(entries) > 0 {
		return // the install got far enough to put files here; keep the manifest
	}
	_ = os.Remove(t.ManifestPath(repoRoot))
	_ = os.Remove(t.SkillsDir(repoRoot))          // only succeeds when empty
	_ = os.Remove(filepath.Join(repoRoot, t.Dir)) // likewise
}

func ensureSkillNotInstalled(repoRoot string, t target.Target, skillDir, skillName string) error {
	lf, err := lockfile.Read(lockfile.PathFor(repoRoot, t))
	if err != nil {
		return nil
	}
	locked := lf.FindSkill(skillName)
	if locked == nil {
		return nil
	}
	info, statErr := os.Stat(skillDir)
	if statErr == nil && info.IsDir() {
		return fmt.Errorf("skill %q is already installed; use 'skell upgrade %s' or 'skell remove %s' first", skillName, skillName, skillName)
	}
	return nil
}

// resolveInstallRegistry decides which registry to install skillName from.
// With no alias and no URL it uses a registry named "default" if configured,
// otherwise it searches every configured source and uses the one that has the
// skill (an error lists the candidates when several do).
func (e *Engine) resolveInstallRegistry(m *manifest.Manifest, skillName, registryAlias, registryURL string) (alias, existingURL string, needsAdding bool, err error) {
	if registryAlias == "" && registryURL == "" {
		if url, ok := e.resolveRegistryURL(m, "default"); ok {
			return "default", url, false, nil
		}
		alias, existingURL, err = e.findSkillRegistry(m, skillName)
		return alias, existingURL, false, err
	}
	alias = registryAlias
	if alias == "" {
		alias = "default"
	}
	existingURL, ok := e.resolveRegistryURL(m, alias)
	if ok {
		return alias, existingURL, false, nil
	}
	if registryURL == "" {
		return "", "", false, fmt.Errorf("registry %q not configured in manifest — add it to skell.toml or supply --registry-url <url>", alias)
	}
	return alias, registryURL, true, nil
}

// findSkillRegistry returns the single configured registry that provides name.
func (e *Engine) findSkillRegistry(m *manifest.Manifest, name string) (string, string, error) {
	regs := e.effectiveRegistries(m)
	if len(regs) == 0 {
		return "", "", fmt.Errorf("no skill sources configured — add one with 'skell add <source>' (browse ideas with 'skell catalog')")
	}
	var hits []string
	for _, a := range sortedAliases(regs) {
		if e.pol.CheckRegistry(regs[a]) != nil {
			continue
		}
		if _, err := e.provider.GetSkill(registry.Registry{Alias: a, URL: regs[a]}, name); err == nil {
			hits = append(hits, a)
		}
	}
	switch len(hits) {
	case 0:
		return "", "", fmt.Errorf("%w: %q in any configured source (%s) — try 'skell search %s'",
			registry.ErrSkillNotFound, name, strings.Join(sortedAliases(regs), ", "), name)
	case 1:
		return hits[0], regs[hits[0]], nil
	default:
		return "", "", fmt.Errorf("skill %q exists in several sources (%s) — choose one with --registry <alias>",
			name, strings.Join(hits, ", "))
	}
}

func autoRegisterInstallRegistry(repoRoot string, t target.Target, m *manifest.Manifest, registryAlias, registryURL string, registryNeedsAdding, dryRun bool) error {
	if !registryNeedsAdding || dryRun {
		return nil
	}
	if m.Registries == nil {
		m.Registries = make(map[string]string)
	}
	m.Registries[registryAlias] = registryURL
	if err := manifest.Write(manifest.LocalPathFor(repoRoot, t), m); err != nil {
		return fmt.Errorf("failed to add registry %q to manifest: %w", registryAlias, err)
	}
	return nil
}

func (e *Engine) fetchInstallSkill(reg registry.Registry, skillName, registryAlias string) (*model.RegistrySkill, error) {
	rs, err := e.provider.GetSkill(reg, skillName)
	if err != nil {
		return nil, fmt.Errorf("could not fetch skill %q from registry %q: %w", skillName, registryAlias, err)
	}
	return rs, nil
}

func (e *Engine) installFetchedSkill(repoRoot string, t target.Target, m *manifest.Manifest, reg registry.Registry, skillName, skillDir, sourceURL string, rs *model.RegistrySkill) error {
	if err := os.MkdirAll(t.SkillsDir(repoRoot), 0755); err != nil {
		return fmt.Errorf("failed to create skills directory: %w", err)
	}
	src := e.skillSource(reg, skillName)
	if err := e.copySkill(reg, skillName, rs.Metadata.Version, skillDir); err != nil {
		return fmt.Errorf("failed to install skill %q: %w", skillName, err)
	}
	hash, err := hasher.HashDir(skillDir)
	if err != nil {
		return fmt.Errorf("failed to hash installed skill: %w", err)
	}
	if err := e.updateLockFile(repoRoot, t, skillName, reg.Alias, sourceURL, rs, hash, src); err != nil {
		return err
	}
	if err := e.updateManifest(repoRoot, t, m, skillName, reg.Alias, rs.Metadata.Version); err != nil {
		return err
	}
	if err := e.mirrorAfterChange(repoRoot, t, skillName); err != nil {
		return err
	}
	_ = e.logger.Log(audit.ActionInstall, skillName, rs.Metadata.Version, reg.Alias, repoRoot)
	return nil
}

// updateLockFile adds or replaces the lock entry for the installed skill.
// src (optional) records the registry commit and path so the install can be
// reproduced exactly; an existing pin is preserved.
func (e *Engine) updateLockFile(repoRoot string, t target.Target, skillName, registryAlias, registryURL string, rs *model.RegistrySkill, hash string, src *registry.SkillSource) error {
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
	entry := model.InstalledSkill{
		Name:          skillName,
		Version:       rs.Metadata.Version,
		Registry:      registryAlias,
		SourceRepo:    registryURL,
		InstalledPath: t.InstalledRelPath(skillName),
		InstalledAt:   now,
		ContentHash:   hash,
	}
	if src != nil {
		entry.Commit = src.Commit
		entry.SourcePath = src.RelPath
	}
	if prev := lf.FindSkill(skillName); prev != nil {
		entry.Pinned = prev.Pinned
	}
	lf.Upsert(entry)

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
