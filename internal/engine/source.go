package engine

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aminmesbahi/skell/internal/audit"
	"github.com/aminmesbahi/skell/internal/hasher"
	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/aminmesbahi/skell/internal/registry"
	"github.com/aminmesbahi/skell/internal/target"
)

// sourceProvider is implemented by providers that can say where a skill lives
// in a registry and install an exact revision (the real registry adapter).
// Providers that don't implement it (e.g. test doubles) keep the simpler
// version-only behaviour.
type sourceProvider interface {
	Source(reg registry.Registry, name string) (*registry.SkillSource, error)
	CopySkillAt(reg registry.Registry, name, commit, relPath, destPath string) error
}

// skillSource returns where a skill lives in its registry, or nil when the
// provider can't tell (unsupported provider, or any lookup error).
func (e *Engine) skillSource(reg registry.Registry, name string) *registry.SkillSource {
	sp, ok := e.provider.(sourceProvider)
	if !ok {
		return nil
	}
	src, err := sp.Source(reg, name)
	if err != nil {
		return nil
	}
	return src
}

// sourceChanged reports whether the registry's current content for a skill
// differs from what was installed (as recorded by the lock's content hash).
// known is false when that can't be determined, in which case callers fall
// back to comparing version strings.
func (e *Engine) sourceChanged(reg registry.Registry, locked model.InstalledSkill) (changed, known bool) {
	if locked.ContentHash == "" || locked.Linked {
		return false, false
	}
	src := e.skillSource(reg, locked.Name)
	if src == nil {
		return false, false
	}
	same, err := hasher.Verify(src.Dir, locked.ContentHash)
	if err != nil {
		return false, false
	}
	return !same, true
}

// DiffResult is the outcome of comparing an installed skill to its registry.
type DiffResult struct {
	Name           string `json:"name"`
	Registry       string `json:"registry"`
	InstalledVer   string `json:"installed_version,omitempty"`
	LatestVer      string `json:"latest_version,omitempty"`
	InstalledAt    string `json:"installed_commit,omitempty"`
	LatestCommit   string `json:"latest_commit,omitempty"`
	Patch          string `json:"patch"`
	LocallyChanged bool   `json:"locally_modified"`
}

// Diff compares an installed skill with the current registry version and
// returns a unified diff (installed → latest). Run 'skell cache refresh' (or
// pass refresh in the CLI) first to compare against the newest upstream.
func (e *Engine) Diff(repoRoot, skillName, targetID string, color bool) (*DiffResult, error) {
	if err := ValidateSkillName(skillName); err != nil {
		return nil, err
	}
	m, t, err := resolveManifestFor(repoRoot, targetID)
	if err != nil {
		return nil, fmt.Errorf("no manifest found in %s — run 'skell init' first: %w", repoRoot, err)
	}
	lf, err := lockfile.Read(lockfile.PathFor(repoRoot, *t))
	if err != nil {
		return nil, fmt.Errorf("lock file not found — run 'skell sync' first: %w", err)
	}
	locked := lf.FindSkill(skillName)
	if locked == nil {
		return nil, fmt.Errorf("skill %q is not installed", skillName)
	}
	if locked.Linked {
		return nil, fmt.Errorf("skill %q is linked to a local folder; its edits are already live", skillName)
	}
	alias, url, ok := e.resolveRegistryForLocked(m, *locked)
	if !ok {
		return nil, fmt.Errorf("registry %q for skill %q is not configured", alias, skillName)
	}
	if err := e.pol.CheckRegistry(url); err != nil {
		return nil, err
	}
	reg := registry.Registry{Alias: alias, URL: url}
	src := e.skillSource(reg, skillName)
	if src == nil {
		return nil, fmt.Errorf("could not locate skill %q in registry %q", skillName, alias)
	}

	skillDir := filepath.Join(t.SkillsDir(repoRoot), skillName)
	patch, err := registry.DiffDirs(skillDir, src.Dir, color)
	if err != nil {
		return nil, err
	}
	res := &DiffResult{
		Name:         skillName,
		Registry:     alias,
		InstalledVer: locked.Version,
		InstalledAt:  locked.Commit,
		LatestCommit: src.Commit,
		Patch:        patch,
	}
	if rs, err := e.provider.GetSkill(reg, skillName); err == nil {
		res.LatestVer = rs.Metadata.Version
	}
	if locked.ContentHash != "" {
		if same, err := hasher.Verify(skillDir, locked.ContentHash); err == nil {
			res.LocallyChanged = !same
		}
	}
	return res, nil
}

// installLocked reinstalls a skill exactly as recorded in the lock file
// (same registry commit) and verifies the result against the locked content
// hash. It returns handled=false when the lock entry has no commit or the
// provider can't pin revisions, so the caller can fall back to a normal
// install of the latest version.
func (e *Engine) installLocked(repoRoot string, t target.Target, m *manifest.Manifest, locked model.InstalledSkill) (handled bool, err error) {
	sp, ok := e.provider.(sourceProvider)
	if !ok || locked.Commit == "" {
		return false, nil
	}
	alias, url, found := e.resolveRegistryForLocked(m, locked)
	if !found {
		return false, nil
	}
	if err := e.pol.CheckRegistry(url); err != nil {
		return true, err
	}
	skillDir := filepath.Join(t.SkillsDir(repoRoot), locked.Name)
	reg := registry.Registry{Alias: alias, URL: url}
	if err := sp.CopySkillAt(reg, locked.Name, locked.Commit, locked.SourcePath, skillDir); err != nil {
		return true, fmt.Errorf("failed to install %q at locked commit: %w", locked.Name, err)
	}
	if locked.ContentHash != "" {
		same, err := hasher.Verify(skillDir, locked.ContentHash)
		if err != nil || !same {
			_ = os.RemoveAll(skillDir)
			return true, fmt.Errorf("skill %q does not match the content hash in skell.lock (registry history may have been rewritten); run 'skell upgrade %s' to accept the current version", locked.Name, locked.Name)
		}
	}
	if err := e.mirrorSkill(repoRoot, m, t, locked.Name); err != nil {
		return true, err
	}
	_ = e.logger.Log(audit.ActionInstall, locked.Name, locked.Version, alias, repoRoot)
	return true, nil
}
