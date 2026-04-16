// Package scanner discovers git repositories and reads their installed skills.
package scanner

import (
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
func IsGitRepo(path string) bool {
	info, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil && info.IsDir()
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

// ScanAll walks a root directory, finds all git repositories under it,
// and returns a ScanResult for each.
func ScanAll(rootPath string) ([]ScanResult, error) {
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}
	var results []ScanResult
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(rootPath, e.Name())
		if IsGitRepo(path) {
			sr, err := ScanRepo(path)
			if err != nil {
				return nil, err
			}
			results = append(results, *sr)
		}
	}
	return results, nil
}
