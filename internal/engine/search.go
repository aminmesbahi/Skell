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
// from the global manifest are dropped.
func (e *Engine) SearchMerged(repoRoot, query, tag, lifecycle, owner string) ([]model.RegistrySkill, error) {
	globalRoot, _ := manifest.GlobalRootDir()
	isGlobal := filepath.Clean(repoRoot) == filepath.Clean(globalRoot)
	localM, localErr := manifest.Resolve(repoRoot)

	if isGlobal {
		if localErr != nil {
			return nil, localErr
		}
		return e.searchManifestWithSource(localM, "global", query, tag, lifecycle, owner)
	}

	merged, seen := e.optionalManifestSkills(localM, localErr, "local")
	merged = appendUniqueSkills(merged, seen, e.optionalGlobalSkills())
	return filterRegistrySkills(merged, query, tag, lifecycle, owner), nil
}

func (e *Engine) searchManifestWithSource(m *manifest.Manifest, source, query, tag, lifecycle, owner string) ([]model.RegistrySkill, error) {
	skills, err := e.listRegistryWithSource(m, source)
	if err != nil {
		return nil, err
	}
	return filterRegistrySkills(skills, query, tag, lifecycle, owner), nil
}

func (e *Engine) listRegistryWithSource(m *manifest.Manifest, source string) ([]model.RegistrySkill, error) {
	skills, err := e.ListRegistry(m)
	if err != nil {
		return nil, err
	}
	for i := range skills {
		skills[i].RegistrySource = source
	}
	return skills, nil
}

func (e *Engine) optionalManifestSkills(m *manifest.Manifest, manifestErr error, source string) ([]model.RegistrySkill, map[string]bool) {
	seen := make(map[string]bool)
	if manifestErr != nil {
		return nil, seen
	}
	skills, err := e.listRegistryWithSource(m, source)
	if err != nil {
		return nil, seen
	}
	for _, skill := range skills {
		seen[registrySkillKey(skill)] = true
	}
	return skills, seen
}

func (e *Engine) optionalGlobalSkills() []model.RegistrySkill {
	globalPath, err := manifest.GlobalPath()
	if err != nil {
		return nil
	}
	globalM, err := manifest.Read(globalPath)
	if err != nil {
		return nil
	}
	skills, err := e.listRegistryWithSource(globalM, "global")
	if err != nil {
		return nil
	}
	return skills
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
