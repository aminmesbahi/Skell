package engine

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/registry"
)

// AddResult describes what AddFromURL did.
type AddResult struct {
	// Alias is the registry alias that was used or added.
	Alias string
	// URL is the registry URL (or local path) that was used.
	URL string
	// SkillName is non-empty when a specific skill was installed.
	SkillName string
	// Registered is true when the registry alias was newly added to skell.toml.
	Registered bool
	// Installed is true when skill files were written to the repository.
	Installed bool
}

// AddOptions tunes AddFromURLWith.
type AddOptions struct {
	// Alias forces the registry alias instead of deriving it from the URL.
	Alias string
	// DryRun previews without writing files.
	DryRun bool
}

var githubShorthandRx = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]*/[A-Za-z0-9._-]+(/[^\s]+)?$`)

// ExpandAddArg turns the short forms accepted by 'skell add' into a full URL
// or path. lookup resolves catalog ids (may be nil). Forms:
//
//   - an existing directory (relative or absolute) → its absolute path
//   - a catalog id such as "anthropic"             → the catalog source URL
//   - owner/repo                                   → https://github.com/owner/repo
//   - owner/repo/path/to/skill                     → that skill inside the repo
//
// Anything else (full URLs, scp-style git URLs) is returned unchanged. The
// boolean reports whether a catalog id matched, so callers can use the id as
// the registry alias.
func ExpandAddArg(raw string, lookup func(id string) (string, bool)) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.Contains(raw, "://") || isLocalPathForAdd(raw) || strings.HasPrefix(raw, "git@") {
		return raw, false
	}
	if info, err := os.Stat(raw); err == nil && info.IsDir() {
		if abs, err := filepath.Abs(raw); err == nil {
			return abs, false
		}
	}
	if !strings.Contains(raw, "/") && lookup != nil {
		if u, ok := lookup(raw); ok {
			return u, true
		}
	}
	if githubShorthandRx.MatchString(raw) {
		parts := strings.SplitN(raw, "/", 3)
		repo := "https://github.com/" + parts[0] + "/" + strings.TrimSuffix(parts[1], ".git")
		if len(parts) == 3 && strings.Trim(parts[2], "/") != "" {
			// The branch segment is only a placeholder: the registry is
			// always cloned at its default branch and skills are found by name.
			return repo + "/tree/HEAD/" + strings.Trim(parts[2], "/"), false
		}
		return repo, false
	}
	return raw, false
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
	return e.AddFromURLWith(repoRoot, rawURL, AddOptions{DryRun: dryRun})
}

// AddFromURLWith is AddFromURL with options (explicit alias).
func (e *Engine) AddFromURLWith(repoRoot, rawURL string, opts AddOptions) (AddResult, error) {
	parsed, err := ParseSkillURL(rawURL)
	if err != nil {
		return AddResult{}, err
	}

	m, err := manifest.Resolve(repoRoot)
	if err != nil {
		return AddResult{}, fmt.Errorf("no manifest found in %s — run 'skell init' first: %w", repoRoot, err)
	}

	alias, err := e.chooseAlias(m, parsed, opts.Alias)
	if err != nil {
		return AddResult{}, err
	}
	parsed.Alias = alias

	res := AddResult{
		Alias:     parsed.Alias,
		URL:       parsed.GitURL,
		SkillName: parsed.SkillName,
	}

	if parsed.SkillName != "" {
		if err := e.Install(repoRoot, parsed.SkillName, parsed.Alias, parsed.GitURL, opts.DryRun); err != nil {
			// URL points to a skills-root subdirectory rather than a single skill:
			// only fall back when the registry actually reports the skill missing.
			if errors.Is(err, registry.ErrSkillNotFound) && e.isSubPathDir(parsed.Alias, parsed.SubPath) {
				res.SkillName = ""
				return e.registerAlias(repoRoot, m, parsed, res, opts.DryRun)
			}
			return AddResult{}, fmt.Errorf("add from URL: %w", err)
		}
		if !opts.DryRun {
			if err := e.overrideInstalledSourceRepo(repoRoot, parsed.SkillName, strings.TrimRight(rawURL, "/")); err != nil {
				return AddResult{}, err
			}
		}
		res.Registered = true
		res.Installed = !opts.DryRun
		return res, nil
	}

	return e.registerAlias(repoRoot, m, parsed, res, opts.DryRun)
}

func (e *Engine) registerAlias(repoRoot string, m *manifest.Manifest, parsed ParsedSkillURL, res AddResult, dryRun bool) (AddResult, error) {
	if existing, ok := e.effectiveRegistries(m)[parsed.Alias]; ok && existing == parsed.GitURL {
		return res, nil // already registered (here or globally) — nothing to do
	}
	if !dryRun {
		_, t, err := manifest.ResolveWithTarget(repoRoot)
		if err != nil {
			return AddResult{}, err
		}
		if m.Registries == nil {
			m.Registries = make(map[string]string)
		}
		m.Registries[parsed.Alias] = parsed.GitURL
		if err := manifest.Write(manifest.LocalPathFor(repoRoot, *t), m); err != nil {
			return AddResult{}, fmt.Errorf("failed to save registry to manifest: %w", err)
		}
	}
	res.Registered = !dryRun
	return res, nil
}

// chooseAlias picks the registry alias for an added URL. An explicit alias
// wins (and must not clash). A derived alias that is already taken by a
// different URL — e.g. two repos both named "skills" — is disambiguated with
// the repository owner ("anthropics-skills").
func (e *Engine) chooseAlias(m *manifest.Manifest, parsed ParsedSkillURL, explicit string) (string, error) {
	regs := e.effectiveRegistries(m)
	free := func(a string) bool {
		existing, ok := regs[a]
		return !ok || existing == parsed.GitURL
	}
	if explicit != "" {
		explicit = strings.ToLower(strings.TrimSpace(explicit))
		if !free(explicit) {
			return "", fmt.Errorf("registry alias %q already exists with a different URL (%s)", explicit, regs[explicit])
		}
		return explicit, nil
	}
	if free(parsed.Alias) {
		return parsed.Alias, nil
	}
	if owner := repoOwner(parsed.GitURL); owner != "" {
		alt := strings.ToLower(owner + "-" + parsed.Alias)
		if free(alt) {
			return alt, nil
		}
	}
	return "", fmt.Errorf("registry alias %q already exists with a different URL (%s); pass --alias to choose another", parsed.Alias, regs[parsed.Alias])
}

// repoOwner extracts "owner" from https://host/owner/repo.
func repoOwner(gitURL string) string {
	i := strings.Index(gitURL, "://")
	if i < 0 {
		return ""
	}
	parts := strings.Split(strings.Trim(gitURL[i+3:], "/"), "/")
	if len(parts) < 3 {
		return ""
	}
	return path.Clean(parts[1])
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
