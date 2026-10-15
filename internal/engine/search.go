package engine

import (
	"path/filepath"
	"strings"

	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/model"
)

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
// from the global manifest are dropped. Sources that cannot be read are
// skipped; SearchMergedWithWarnings reports them.
func (e *Engine) SearchMerged(repoRoot, query, tag, lifecycle, owner string) ([]model.RegistrySkill, error) {
	results, _, err := e.SearchMergedWithWarnings(repoRoot, query, tag, lifecycle, owner)
	return results, err
}

// SearchMergedWithWarnings is SearchMerged plus the list of sources that could
// not be read (unreachable, mistyped or blocked-by-network), so callers can
// tell the user why a source contributed nothing instead of hiding it.
func (e *Engine) SearchMergedWithWarnings(repoRoot, query, tag, lifecycle, owner string) ([]model.RegistrySkill, []SourceError, error) {
	globalRoot, _ := manifest.GlobalRootDir()
	isGlobal := filepath.Clean(repoRoot) == filepath.Clean(globalRoot)
	localM, localErr := manifest.Resolve(repoRoot)

	if isGlobal {
		if localErr != nil {
			return nil, nil, localErr
		}
		skills, warns := e.listSourced(localM, "global")
		if len(skills) == 0 && len(warns) > 0 {
			return nil, warns, warns[0]
		}
		return filterRegistrySkills(skills, query, tag, lifecycle, owner), warns, nil
	}

	var warns []SourceError
	merged := []model.RegistrySkill(nil)
	seen := make(map[string]bool)
	if localErr == nil {
		var w []SourceError
		merged, w = e.listSourced(localM, "local")
		warns = append(warns, w...)
		for _, skill := range merged {
			seen[registrySkillKey(skill)] = true
		}
	}
	if globalPath, err := manifest.GlobalPath(); err == nil {
		if globalM, err := manifest.Read(globalPath); err == nil {
			skills, w := e.listSourced(globalM, "global")
			warns = append(warns, w...)
			merged = appendUniqueSkills(merged, seen, skills)
		}
	}
	return filterRegistrySkills(merged, query, tag, lifecycle, owner), warns, nil
}

// listSourced lists the readable skills of m's sources stamped with source
// ("local" or "global"), plus the sources that failed.
func (e *Engine) listSourced(m *manifest.Manifest, source string) ([]model.RegistrySkill, []SourceError) {
	skills, failed := e.ListRegistryPartial(m)
	for i := range skills {
		skills[i].RegistrySource = source
	}
	return skills, failed
}

func appendUniqueSkills(dst []model.RegistrySkill, seen map[string]bool, skills []model.RegistrySkill) []model.RegistrySkill {
	for _, skill := range skills {
		key := registrySkillKey(skill)
		if seen[key] {
			continue
		}
		seen[key] = true
		dst = append(dst, skill)
	}
	return dst
}

func registrySkillKey(skill model.RegistrySkill) string {
	return skill.RegistryAlias + "/" + skill.Name
}

func filterRegistrySkills(skills []model.RegistrySkill, query, tag, lifecycle, owner string) []model.RegistrySkill {
	var results []model.RegistrySkill
	for _, skill := range skills {
		if matchesFilter(skill, query, tag, lifecycle, owner) {
			results = append(results, skill)
		}
	}
	return results
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
