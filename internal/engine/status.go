package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aminmesbahi/skell/internal/frontmatter"
	"github.com/aminmesbahi/skell/internal/hasher"
	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/aminmesbahi/skell/internal/registry"
	"github.com/aminmesbahi/skell/internal/target"
)

// Status returns the comparison between registry and local state for a repository.
// Skills that cannot be found in the registry are marked StatusUnknown.
//
// Deprecated: prefer StatusFor which allows specifying a target agent platform.
func (e *Engine) Status(repoRoot string) ([]model.StatusEntry, error) {
	return e.StatusFor(repoRoot, "")
}

// StatusFor returns the comparison between registry and local state for a
// repository and optional target. When targetID is empty, statuses from ALL
// detected targets are aggregated. When targetID is provided, only that
// target is checked.
func (e *Engine) StatusFor(repoRoot, targetID string) ([]model.StatusEntry, error) {
	if targetID != "" {
		return e.statusSingleTarget(repoRoot, targetID)
	}

	detected := detectManaged(repoRoot)
	if len(detected) == 0 {
		return e.statusSingleTarget(repoRoot, target.Default)
	}

	var all []model.StatusEntry
	var firstErr error
	for _, t := range detected {
		entries, err := e.statusSingleTarget(repoRoot, t.ID)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		all = append(all, entries...)
	}
	if len(all) == 0 && firstErr != nil {
		return nil, firstErr
	}
	return all, nil
}

// statusSingleTarget returns status entries for a single target.
func (e *Engine) statusSingleTarget(repoRoot, targetID string) ([]model.StatusEntry, error) {
	t, err := resolveTarget(repoRoot, targetID)
	if err != nil {
		return nil, err
	}
	m, _, err := manifest.ResolveWithTarget(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("no manifest found in %s: %w", repoRoot, err)
	}

	lf, err := lockfile.Read(lockfile.PathFor(repoRoot, t))
	if err != nil {
		return nil, fmt.Errorf("lock file not found — run 'skell sync' to create one: %w", err)
	}

	var entries []model.StatusEntry
	locked := make(map[string]bool, len(lf.Skills))
	for _, s := range lf.Skills {
		locked[s.Name] = true
		entries = append(entries, e.statusEntryForSkill(m, t, repoRoot, s))
	}

	for _, name := range unlockedSkillDirs(t, repoRoot, locked) {
		entries = append(entries, model.StatusEntry{Name: name, Status: model.StatusUnknown})
	}
	return entries, nil
}

// unlockedSkillDirs returns the names of skill directories under the target's
// skills/ folder that are not recorded in the provided locked set.
func unlockedSkillDirs(t target.Target, repoRoot string, locked map[string]bool) []string {
	dirEntries, err := os.ReadDir(t.SkillsDir(repoRoot))
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range dirEntries {
		if e.IsDir() && !locked[e.Name()] && !strings.HasPrefix(e.Name(), ".skell-") {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out
}

// statusEntryForSkill derives the status of a single installed skill by consulting
// the local hash, the manifest registry map, and the remote registry metadata.
func (e *Engine) statusEntryForSkill(m *manifest.Manifest, t target.Target, repoRoot string, locked model.InstalledSkill) model.StatusEntry {
	entry := model.StatusEntry{Name: locked.Name, Installed: locked.Version, Target: t.ID}

	if locked.Pinned {
		entry.Status = model.StatusPinned
		return entry
	}
	if locked.Linked {
		entry.Status = model.StatusLinked
		return entry
	}

	// Computed first but not returned early, so the registry lookup below can
	// still populate Latest and apply lifecycle precedence.
	var localStatus model.SkillStatus
	skillDir := filepath.Join(t.SkillsDir(repoRoot), locked.Name)
	if locked.ContentHash != "" {
		ok, hashErr := hasher.Verify(skillDir, locked.ContentHash)
		switch {
		case hashErr != nil:
			entry.Status = model.StatusUnknown
			return entry
		case !ok:
			localStatus = model.StatusLocallyModified
		}
	}

	alias := locked.Registry
	if alias == "" {
		alias = "default"
	}
	registryURL, ok := e.resolveRegistryURL(m, alias)
	if !ok {
		entry.Status = firstNonEmptyStatus(localStatus, model.StatusUnknown)
		return entry
	}

	if e.pol.CheckRegistry(registryURL) != nil {
		entry.Status = firstNonEmptyStatus(localStatus, model.StatusUnknown)
		return entry
	}

	reg := registry.Registry{Alias: alias, URL: registryURL}
	rs, err := e.provider.GetSkill(reg, locked.Name)
	if err != nil {
		entry.Status = firstNonEmptyStatus(localStatus, model.StatusUnknown)
		return entry
	}

	entry.Latest = rs.Metadata.Version
	versionStatus := resolveVersionStatus(locked.Version, rs)

	// Content comparison catches upstream edits that didn't bump the version
	// (common: most skills carry no version at all).
	if changed, known := e.sourceChanged(reg, locked); known {
		entry.Changed = changed
		switch versionStatus {
		case model.StatusUpToDate, model.StatusUnversioned, model.StatusMissingMetadata:
			if changed {
				versionStatus = model.StatusOutdated
			} else if versionStatus != model.StatusUpToDate {
				versionStatus = model.StatusUpToDate
			}
		}
	}

	// Report local modification, but let deprecated/archived take precedence.
	if localStatus != "" {
		switch versionStatus {
		case model.StatusDeprecated, model.StatusArchived:
			entry.Status = versionStatus
		default:
			entry.Status = localStatus
		}
		return entry
	}

	entry.Status = versionStatus
	return entry
}

// firstNonEmptyStatus returns a if it is set, otherwise b.
func firstNonEmptyStatus(a, b model.SkillStatus) model.SkillStatus {
	if a != "" {
		return a
	}
	return b
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
//
// Deprecated: prefer InfoFor which allows specifying a target agent platform.
func (e *Engine) Info(repoRoot, skillName, source string) (*model.InfoResult, error) {
	return e.InfoFor(repoRoot, skillName, source, "")
}

// InfoFor returns the full detail for a single named skill from local state
// for the given optional target. When targetID is empty the target is
// auto-detected.
func (e *Engine) InfoFor(repoRoot, skillName, source, targetID string) (*model.InfoResult, error) {
	if err := ValidateSkillName(skillName); err != nil {
		return nil, err
	}
	result := &model.InfoResult{}
	t, err := resolveTarget(repoRoot, targetID)
	if err != nil {
		return nil, err
	}

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
	regs := e.effectiveRegistries(m)
	for _, alias := range sortedAliases(regs) {
		if e.pol.CheckRegistry(regs[alias]) != nil {
			continue
		}
		reg := registry.Registry{Alias: alias, URL: regs[alias]}
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
