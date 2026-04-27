// Package scanner discovers git repositories and reads their installed skills.
package scanner

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aminmesbahi/skell/internal/model"
)

// ScanResult holds all findings for a single repository.
type ScanResult struct {
	RepoRoot        string
	InstalledSkills []model.InstalledSkill
	HasManifest     bool
	HasLockFile     bool
}

// IsGitRepo returns true if the given path is the root of a git repository.
// Accepts both classic clones (.git is a directory) and worktrees / submodules
// (.git is a file pointing at the gitdir).
func IsGitRepo(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil
}

// HasSkellManifest returns true if the given path contains a .claude/skell.toml file.
func HasSkellManifest(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".claude", "skell.toml"))
	return err == nil
}

// SkillsDir returns the path to the .claude/skills/ directory for a given repo root.
func SkillsDir(repoRoot string) string {
	return filepath.Join(repoRoot, ".claude", "skills")
}

// ScanRepo scans a single repository root and returns its installed skill state.
func ScanRepo(repoRoot string) (*ScanResult, error) {
	result := &ScanResult{RepoRoot: repoRoot}

	claudeDir := filepath.Join(repoRoot, ".claude")
	_, err := os.Stat(filepath.Join(claudeDir, "skell.toml"))
	result.HasManifest = err == nil

	_, err = os.Stat(filepath.Join(claudeDir, "skell.lock"))
	result.HasLockFile = err == nil

	skillsPath := SkillsDir(repoRoot)
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
				InstalledPath: filepath.Join(".claude", "skills", e.Name()),
			})
		}
	}
	return result, nil
}

const maxScanDepth = 6

var scanSkipDirs = map[string]bool{
	".git":         true,
	".claude":      true,
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
