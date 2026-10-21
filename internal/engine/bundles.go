package engine

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/target"
)

// BundleInfo is a bundle as shown to users.
type BundleInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Skills      []string `json:"skills"`
	// Installed counts how many of the bundle's skills are already installed.
	Installed int `json:"installed"`
}

func sortedBundleNames(b map[string]manifest.Bundle) []string {
	names := make([]string, 0, len(b))
	for n := range b {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// ListBundles returns the bundles declared in a project's manifest.
func (e *Engine) ListBundles(repoRoot string) ([]BundleInfo, error) {
	m, t, err := manifest.ResolveWithTarget(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("no manifest found in %s — run 'skell init' first: %w", repoRoot, err)
	}
	installed := map[string]bool{}
	if skills, err := e.ListFor(repoRoot, t.ID); err == nil {
		for _, s := range skills {
			installed[s.Name] = true
		}
	}
	out := []BundleInfo{}
	for _, name := range sortedBundleNames(m.Bundles) {
		b := m.Bundles[name]
		info := BundleInfo{Name: name, Description: b.Description, Skills: append([]string{}, b.Skills...)}
		for _, s := range b.Skills {
			if installed[s] {
				info.Installed++
			}
		}
		out = append(out, info)
	}
	return out, nil
}

// validateBundle checks a bundle name and its members.
var bundleNameRx = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

func validateBundle(name string, skills []string) error {
	if !bundleNameRx.MatchString(name) {
		return fmt.Errorf("invalid bundle name %q: use lowercase letters, digits and hyphens", name)
	}
	if len(skills) == 0 {
		return fmt.Errorf("bundle %q needs at least one skill", name)
	}
	for _, s := range skills {
		if err := ValidateSkillName(s); err != nil {
			return fmt.Errorf("bundle %q: %w", name, err)
		}
	}
	return nil
}

// SaveBundle creates or replaces a bundle in the project's manifest.
func (e *Engine) SaveBundle(repoRoot, name, description string, skills []string) error {
	skills = dedupeStrings(skills)
	if err := validateBundle(name, skills); err != nil {
		return err
	}
	m, t, err := manifest.ResolveWithTarget(repoRoot)
	if err != nil {
		return fmt.Errorf("no manifest found in %s — run 'skell init' first: %w", repoRoot, err)
	}
	if m.Bundles == nil {
		m.Bundles = map[string]manifest.Bundle{}
	}
	m.Bundles[name] = manifest.Bundle{Description: strings.TrimSpace(description), Skills: skills}
	return manifest.Write(manifest.LocalPathFor(repoRoot, *t), m)
}

// DeleteBundle removes a bundle (its skills stay installed).
func (e *Engine) DeleteBundle(repoRoot, name string) error {
	m, t, err := manifest.ResolveWithTarget(repoRoot)
	if err != nil {
		return fmt.Errorf("no manifest found in %s — run 'skell init' first: %w", repoRoot, err)
	}
	if _, ok := m.Bundles[name]; !ok {
		return fmt.Errorf("no bundle named %q", name)
	}
	delete(m.Bundles, name)
	return manifest.Write(manifest.LocalPathFor(repoRoot, *t), m)
}

// ExpandBundles replaces "@bundle" arguments with the skills the bundle
// contains, keeping every other argument and removing duplicates. Members of
// a bundle that are already installed are left out, so installing a bundle
// twice (or on a project that has part of it) just fills the gaps.
func (e *Engine) ExpandBundles(repoRoot string, args []string) ([]string, error) {
	var out []string
	var m *manifest.Manifest
	var installed map[string]bool
	for _, a := range args {
		if !strings.HasPrefix(a, "@") {
			out = append(out, a)
			continue
		}
		if m == nil {
			var t *target.Target
			var err error
			if m, t, err = manifest.ResolveWithTarget(repoRoot); err != nil {
				return nil, fmt.Errorf("no manifest found in %s — run 'skell init' first: %w", repoRoot, err)
			}
			installed = map[string]bool{}
			if have, err := e.ListFor(repoRoot, t.ID); err == nil {
				for _, s := range have {
					installed[s.Name] = true
				}
			}
		}
		b, ok := m.Bundles[strings.TrimPrefix(a, "@")]
		if !ok {
			return nil, fmt.Errorf("no bundle named %q — see 'skell bundle list'", strings.TrimPrefix(a, "@"))
		}
		for _, s := range b.Skills {
			if !installed[s] {
				out = append(out, s)
			}
		}
	}
	return dedupeStrings(out), nil
}

func dedupeStrings(in []string) []string {
	var out []string
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s != "" && !slices.Contains(out, s) {
			out = append(out, s)
		}
	}
	return out
}
