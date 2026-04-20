package engine

import (
	"fmt"

	"github.com/aminmesbahi/skell/internal/manifest"
)

// AddResult describes what AddFromURL did.
type AddResult struct {
	// Alias is the registry alias that was used or added.
	Alias string
	// SkillName is non-empty when a specific skill was installed.
	SkillName string
	// Registered is true when the registry alias was newly added to skell.toml.
	Registered bool
	// Installed is true when skill files were written to the repository.
	Installed bool
}

// AddFromURL installs a skill (or registers a registry) from a GitHub tree URL
// or plain git URL.
//
//   - If the URL points to a specific skill directory (≥2 path segments after the branch),
//     the skill is installed and the registry is auto-registered in skell.toml.
//   - If the URL points to a registry root (≤1 path segment after the branch),
//     the registry is registered in skell.toml so future search/install can use it.
//
// When dryRun is true no files are written.
func (e *Engine) AddFromURL(repoRoot, rawURL string, dryRun bool) (AddResult, error) {
	parsed, err := ParseSkillURL(rawURL)
	if err != nil {
		return AddResult{}, err
	}

	res := AddResult{
		Alias:     parsed.Alias,
		SkillName: parsed.SkillName,
	}

	if parsed.SkillName != "" {
		if err := e.Install(repoRoot, parsed.SkillName, parsed.Alias, parsed.GitURL, dryRun); err != nil {
			return AddResult{}, fmt.Errorf("add from URL: %w", err)
		}
		res.Registered = true
		res.Installed = !dryRun
		return res, nil
	}

	m, err := manifest.Resolve(repoRoot)
	if err != nil {
		return AddResult{}, fmt.Errorf("no manifest found in %s — run 'skell init' first: %w", repoRoot, err)
	}

	if _, exists := m.Registries[parsed.Alias]; exists {
		return res, nil
	}

	if !dryRun {
		if m.Registries == nil {
			m.Registries = make(map[string]string)
		}
		m.Registries[parsed.Alias] = parsed.GitURL
		if err := manifest.Write(manifest.LocalPath(repoRoot), m); err != nil {
			return AddResult{}, fmt.Errorf("failed to save registry to manifest: %w", err)
		}
	}
	res.Registered = !dryRun
	return res, nil
}
