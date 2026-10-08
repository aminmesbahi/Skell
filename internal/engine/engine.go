package engine

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"sort"

	"github.com/aminmesbahi/skell/internal/audit"
	"github.com/aminmesbahi/skell/internal/config"
	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/aminmesbahi/skell/internal/policy"
	"github.com/aminmesbahi/skell/internal/registry"
	"github.com/aminmesbahi/skell/internal/scanner"
	"github.com/aminmesbahi/skell/internal/target"
	"github.com/aminmesbahi/skell/internal/version"
)

func skellVersion() string {
	return version.Version
}

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
	// homeRoot is the Skell home directory (e.g. ~/.skell) used to locate the
	// global config.toml [sources]. Empty disables global sources, which keeps
	// tests hermetic (no dependence on the developer's real ~/.skell).
	homeRoot string
	logger   *audit.Logger
	pol      *policy.Config
	// requireValidation gates install/upgrade on spec validation. Initialised
	// from policy; the CLI can override it per-invocation (e.g. --no-validate).
	requireValidation bool
}

// New creates a ready-to-use Engine backed by the real registry adapter.
// cacheRoot is expected to be <home>/cache, so the Skell home root is its parent.
// The audit log and policy are read from that same home root, keeping all of
// Skell's state (cache, sources, policy, audit) under one directory and honoring
// SKELL_HOME via the caller's cacheRoot.
func New(cacheRoot string) *Engine {
	home := filepath.Dir(cacheRoot)
	pol := loadPolicy(home)
	e := &Engine{
		provider:          registry.NewAdapter(cacheRoot),
		cacheRoot:         cacheRoot,
		homeRoot:          home,
		logger:            audit.NewLogger(filepath.Join(home, "audit.log")),
		pol:               pol,
		requireValidation: pol.RequireValidation,
	}
	return e
}

// SetRequireValidation overrides whether install/upgrade validate skills before
// writing them. Used by the CLI (--no-validate / --validate) to override the
// policy default for a single invocation.
func (e *Engine) SetRequireValidation(v bool) { e.requireValidation = v }

// newWithProvider creates an Engine with an injected provider (used in tests).
// homeRoot is intentionally left empty so global sources are not consulted, and
// the audit logger is a no-op so tests never write to the real audit log.
func newWithProvider(p RegistryProvider) *Engine {
	return &Engine{provider: p, logger: audit.NewLogger(""), pol: &policy.Config{}}
}

// loadPolicy reads <home>/config.toml. A missing file means no policy; a file
// that exists but cannot be parsed fails closed (all registries refused) so a
// typo cannot silently disable enterprise controls.
func loadPolicy(home string) *policy.Config {
	cfg, err := policy.Read(filepath.Join(home, "config.toml"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &policy.Config{}
		}
		return policy.Invalid(err)
	}
	return cfg
}

// ResolveTarget returns the active target for a repository. Resolution order:
//  1. explicit target recorded in skell.toml
//  2. directory of the manifest that was discovered (e.g. .codex/)
//  3. an existing skills/ directory belonging to a known target
//  4. the default target (claude) so brand-new repos behave as before
func ResolveTarget(repoRoot string) target.Target {
	if _, t, err := manifest.ResolveWithTarget(repoRoot); err == nil && t != nil {
		return *t
	}
	if t, ok := target.DetectPrimary(repoRoot); ok {
		return t
	}
	return target.MustLookup(target.Default)
}

// resolveTarget returns the target for a repository, optionally overridden by
// targetID. When targetID is empty, resolution falls back to the standard
// auto-detection logic.
func resolveTarget(repoRoot, targetID string) (target.Target, error) {
	if targetID != "" {
		return target.Lookup(targetID)
	}
	return ResolveTarget(repoRoot), nil
}

// List returns all installed skills for the given repository root.
// It reads the lock file when available; falls back to scanning the skills directory.
//
// Deprecated: prefer ListFor which allows specifying a target agent platform.
func (e *Engine) List(repoRoot string) ([]model.InstalledSkill, error) {
	return e.ListFor(repoRoot, "")
}

// ListFor returns all installed skills for the given repository root and
// optional target. When targetID is empty, skills from ALL detected targets
// are aggregated. When targetID is provided, only skills for that target
// are returned.
func (e *Engine) ListFor(repoRoot, targetID string) ([]model.InstalledSkill, error) {
	if targetID != "" {
		return e.listSingleTarget(repoRoot, targetID)
	}

	// Aggregate across all detected targets.
	detected := detectManaged(repoRoot)
	if len(detected) == 0 {
		// Nothing detected — fall back to the default target.
		return e.listSingleTarget(repoRoot, target.Default)
	}

	var all []model.InstalledSkill
	var firstErr error
	for _, t := range detected {
		skills, err := e.listSingleTarget(repoRoot, t.ID)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		all = append(all, skills...)
	}
	if len(all) == 0 && firstErr != nil {
		return nil, firstErr
	}
	return all, nil
}

// listSingleTarget returns installed skills for a single target.
func (e *Engine) listSingleTarget(repoRoot, targetID string) ([]model.InstalledSkill, error) {
	t, err := resolveTarget(repoRoot, targetID)
	if err != nil {
		return nil, err
	}
	lf, err := lockfile.Read(lockfile.PathFor(repoRoot, t))
	if err == nil {
		return lf.Skills, nil
	}

	// No lock file — synthesise entries from the skills directory.
	sr, err := scanner.ScanRepoFor(repoRoot, t)
	if err != nil {
		return nil, fmt.Errorf("failed to scan repository: %w", err)
	}
	return sr.InstalledSkills, nil
}

// effectiveRegistries merges global sources (~/.skell/config.toml [sources])
// with the project manifest; project definitions win on alias conflict. All
// alias→URL resolution flows through here so a globally-configured source
// behaves the same as one declared in skell.toml.
func (e *Engine) effectiveRegistries(m *manifest.Manifest) map[string]string {
	out := make(map[string]string)
	if global, err := config.SourcesFrom(e.homeRoot); err == nil {
		maps.Copy(out, global)
	}
	if m != nil {
		maps.Copy(out, m.Registries)
	}
	return out
}

// resolveRegistryURL returns the URL for a single alias across the effective
// (global + project) registry set.
func (e *Engine) resolveRegistryURL(m *manifest.Manifest, alias string) (string, bool) {
	url, ok := e.effectiveRegistries(m)[alias]
	return url, ok
}

// sortedAliases returns the registry aliases in a stable (lexicographic) order
// so that iteration, output, and collision precedence are deterministic rather
// than dependent on Go's randomized map ordering.
func sortedAliases(regs map[string]string) []string {
	aliases := make([]string, 0, len(regs))
	for alias := range regs {
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)
	return aliases
}

// ListRegistry returns all skills available in all registries configured in the manifest
// plus any globally configured sources (from Settings).
func (e *Engine) ListRegistry(m *manifest.Manifest) ([]model.RegistrySkill, error) {
	regs := e.effectiveRegistries(m)

	var all []model.RegistrySkill
	for _, alias := range sortedAliases(regs) {
		url := regs[alias]
		if e.pol.CheckRegistry(url) != nil {
			continue // blocked by policy: never clone or list it
		}
		reg := registry.Registry{Alias: alias, URL: url}
		skills, err := e.provider.ListSkills(reg)
		if err != nil {
			return nil, fmt.Errorf("failed to list registry %q: %w", alias, err)
		}
		for i := range skills {
			skills[i].RegistryAlias = alias
			skills[i].RegistryURL = url
		}
		all = append(all, skills...)
	}
	return all, nil
}

func resolveManifestFor(repoRoot, targetID string) (*manifest.Manifest, *target.Target, error) {
	if targetID == "" {
		return manifest.ResolveWithTarget(repoRoot)
	}
	t, err := resolveTarget(repoRoot, targetID)
	if err != nil {
		return nil, nil, err
	}
	m, err := manifest.Read(manifest.LocalPathFor(repoRoot, t))
	return m, &t, err
}
