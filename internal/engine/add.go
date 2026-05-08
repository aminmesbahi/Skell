package engine

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/registry"
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
			// URL points to a skills-root subdirectory rather than a single skill:
			// only fall back when the registry actually reports the skill missing.
			if errors.Is(err, registry.ErrSkillNotFound) && e.isSubPathDir(parsed.Alias, parsed.SubPath) {
				res.SkillName = ""
				res.Registered = true
				return res, nil
			}
			return AddResult{}, fmt.Errorf("add from URL: %w", err)
		}
		if !dryRun {
			if err := e.overrideInstalledSourceRepo(repoRoot, parsed.SkillName, strings.TrimRight(rawURL, "/")); err != nil {
				return AddResult{}, err
			}
		}
		res.Registered = true
		res.Installed = !dryRun
		return res, nil
	}

	m, err := manifest.Resolve(repoRoot)
	if err != nil {
		return AddResult{}, fmt.Errorf("no manifest found in %s — run 'skell init' first: %w", repoRoot, err)
	}

	if existing, exists := m.Registries[parsed.Alias]; exists {
		if existing != parsed.GitURL {
			return res, fmt.Errorf("registry alias %q already exists with a different URL (use a different --registry alias or edit skell.toml)", parsed.Alias)
		}
		return res, nil // already registered, same URL — nothing to do
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

// isSubPathDir reports whether subPath exists as a directory inside the cached
// clone for the given registry alias. Used to detect when a parsed "skill URL"
// actually points to a skills-root subdirectory rather than a single skill.
func (e *Engine) isSubPathDir(alias, subPath string) bool {
	if subPath == "" || e.cacheRoot == "" {
		return false
	}
	info, err := os.Stat(filepath.Join(e.cacheRoot, alias, filepath.FromSlash(subPath)))
	return err == nil && info.IsDir()
}

func (e *Engine) overrideInstalledSourceRepo(repoRoot, skillName, sourceRepo string) error {
	_, t, err := manifest.ResolveWithTarget(repoRoot)
	if err != nil {
		return fmt.Errorf("resolve manifest while storing source repo: %w", err)
	}

	lockPath := lockfile.PathFor(repoRoot, *t)
	lf, err := lockfile.Read(lockPath)
	if err != nil {
		return fmt.Errorf("read lock file while storing source repo: %w", err)
	}

	locked := lf.FindSkill(skillName)
	if locked == nil {
		return fmt.Errorf("skill %q not found in lock file while storing source repo", skillName)
	}
	locked.SourceRepo = sourceRepo

	if err := lockfile.Write(lockPath, lf); err != nil {
		return fmt.Errorf("write lock file while storing source repo: %w", err)
	}
	return nil
}
