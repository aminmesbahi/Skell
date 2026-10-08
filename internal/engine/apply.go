package engine

import (
	"fmt"
	"slices"
	"sort"

	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/target"
)

// ApplyReport describes what merging a baseline manifest changed.
type ApplyReport struct {
	AddedRegistries []string `json:"added_registries"`
	AddedSkills     []string `json:"added_skills"`
	AddedMirrors    []string `json:"added_mirrors"`
	// Kept lists entries the project already had and that were left as-is
	// (the project's own choices win over the baseline).
	Kept []string `json:"kept,omitempty"`
}

// ApplyBaseline merges a shared "baseline" manifest (e.g. a team's standard
// skill set) into the project's manifest: missing sources, skills and mirror
// targets are added; anything the project already declares is kept. A source
// alias that exists with a different URL is an error, so a baseline can never
// silently redirect a project's skills to another repository. Nothing is
// installed — run Sync afterwards.
func (e *Engine) ApplyBaseline(repoRoot string, base *manifest.Manifest, dryRun bool) (*ApplyReport, error) {
	m, t, err := manifest.ResolveWithTarget(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("no manifest found in %s — run 'skell init' first: %w", repoRoot, err)
	}
	rep := &ApplyReport{}
	if m.Registries == nil {
		m.Registries = map[string]string{}
	}
	if m.Skills == nil {
		m.Skills = map[string]manifest.SkillEntry{}
	}

	for _, alias := range sortedAliases(base.Registries) {
		url := base.Registries[alias]
		if err := validateAliasName(alias); err != nil {
			return nil, err
		}
		if cur, ok := m.Registries[alias]; ok {
			if cur != url {
				return nil, fmt.Errorf("source %q is %s here but %s in the baseline — rename one of them", alias, cur, url)
			}
			rep.Kept = append(rep.Kept, "source "+alias)
			continue
		}
		m.Registries[alias] = url
		rep.AddedRegistries = append(rep.AddedRegistries, alias)
	}

	names := make([]string, 0, len(base.Skills))
	for n := range base.Skills {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := ValidateSkillName(name); err != nil {
			return nil, err
		}
		if _, ok := m.Skills[name]; ok {
			rep.Kept = append(rep.Kept, "skill "+name)
			continue
		}
		m.Skills[name] = base.Skills[name]
		rep.AddedSkills = append(rep.AddedSkills, name)
	}

	for _, id := range base.Mirrors {
		mt, err := target.Lookup(id)
		if err != nil || mt.ID == t.ID || slices.Contains(m.Mirrors, mt.ID) {
			continue
		}
		m.Mirrors = append(m.Mirrors, mt.ID)
		rep.AddedMirrors = append(rep.AddedMirrors, mt.ID)
	}
	sort.Strings(m.Mirrors)

	if dryRun {
		return rep, nil
	}
	if err := manifest.Write(manifest.LocalPathFor(repoRoot, *t), m); err != nil {
		return nil, err
	}
	return rep, nil
}

// validateAliasName rejects aliases that can't be used as a cache folder name.
func validateAliasName(alias string) error {
	if err := ValidateSkillName(alias); err != nil {
		return fmt.Errorf("invalid source alias %q", alias)
	}
	return nil
}
