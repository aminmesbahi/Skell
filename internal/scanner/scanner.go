// Package scanner discovers git repositories and reads their installed skills.
package scanner

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aminmesbahi/skell/internal/model"
	"github.com/aminmesbahi/skell/internal/target"
)

// ScanResult holds all findings for a single repository.
type ScanResult struct {
	RepoRoot        string
	InstalledSkills []model.InstalledSkill
	HasManifest     bool
	HasLockFile     bool
	// Target is the AI-agent platform layout detected for this repo, or empty
	// when no recognised layout was found.
	Target string
}

// IsGitRepo returns true if the given path is the root of a git repository.
// Accepts both classic clones (.git is a directory) and worktrees / submodules
// (.git is a file pointing at the gitdir).
func IsGitRepo(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil
}

// HasSkellManifest returns true when any known target has a skell.toml in path.
func HasSkellManifest(path string) bool {
	for _, t := range target.All() {
		if _, err := os.Stat(t.ManifestPath(path)); err == nil {
			return true
		}
	}
	return false
}

// SkillsDir returns the path to the skills directory for the legacy Claude
// layout. Prefer SkillsDirFor for target-aware code.
func SkillsDir(repoRoot string) string {
	return target.MustLookup(target.Default).SkillsDir(repoRoot)
}

// SkillsDirFor returns the skills directory for the given target.
func SkillsDirFor(repoRoot string, t target.Target) string {
	return t.SkillsDir(repoRoot)
}

// ScanRepo scans a single repository root and returns its installed skill state.
// The active target is auto-detected; when more than one target directory is
// present, the one with a manifest wins.
func ScanRepo(repoRoot string) (*ScanResult, error) {
	t, ok := target.DetectPrimary(repoRoot)
	if !ok {
		t = target.MustLookup(target.Default)
	}
	return ScanRepoFor(repoRoot, t)
}

// ScanRepoFor scans a single repository for a specific target.
func ScanRepoFor(repoRoot string, t target.Target) (*ScanResult, error) {
	result := &ScanResult{RepoRoot: repoRoot, Target: t.ID}

	_, err := os.Stat(t.ManifestPath(repoRoot))
	result.HasManifest = err == nil

	_, err = os.Stat(t.LockPath(repoRoot))
	result.HasLockFile = err == nil

	skillsPath := t.SkillsDir(repoRoot)
	entries, err := os.ReadDir(skillsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			result.InstalledSkills = append(result.InstalledSkills, model.InstalledSkill{
				Name:          e.Name(),
				InstalledPath: t.InstalledRelPath(e.Name()),
			})
		}
	}
	return result, nil
}

const maxScanDepth = 6

var scanSkipDirs = map[string]bool{
	".git":         true,
	".claude":      true,
	".codex":       true,
	".cursor":      true,
	".github":      true,
	"node_modules": true,
	"vendor":       true,
	".venv":        true,
	"venv":         true,
	"target":       true,
	"dist":         true,
	"build":        true,
	".next":        true,
	".cache":       true,
}

// ScanAll walks rootPath recursively (up to maxScanDepth) and returns a
// ScanResult for every git repository or Skell-managed directory it finds.
// Once a directory qualifies it is not descended into.
func ScanAll(rootPath string) ([]ScanResult, error) {
	info, err := os.Stat(rootPath)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("scanner: %q is not a directory", rootPath)
	}

	var results []ScanResult
	if err := scanDir(rootPath, 0, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func scanDir(path string, depth int, out *[]ScanResult) error {
	if depth > maxScanDepth {
		return nil
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if scanSkipDirs[name] {
			continue
		}
		child := filepath.Join(path, name)
		if IsGitRepo(child) || HasSkellManifest(child) {
			sr, err := ScanRepo(child)
			if err != nil {
				return err
			}
			*out = append(*out, *sr)
			continue
		}
		if err := scanDir(child, depth+1, out); err != nil {
			return err
		}
	}
	return nil
}
