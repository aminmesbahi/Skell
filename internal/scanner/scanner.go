// Package scanner discovers git repositories and reads their installed skills.
package scanner

import "github.com/aminmesbahi/skell/internal/model"

// ScanResult holds all findings for a single repository.
type ScanResult struct {
	RepoRoot       string
	InstalledSkills []model.InstalledSkill
	HasManifest    bool
	HasLockFile    bool
}

// ScanRepo scans a single repository root and returns its installed skill state.
func ScanRepo(repoRoot string) (*ScanResult, error) {
	// TODO: implement
	panic("not implemented")
}

// ScanAll walks a root directory, finds all git repositories under it,
// and returns a ScanResult for each.
func ScanAll(rootPath string) ([]ScanResult, error) {
	// TODO: implement
	panic("not implemented")
}

// IsGitRepo returns true if the given path is the root of a git repository.
func IsGitRepo(path string) bool {
	// TODO: implement
	panic("not implemented")
}

// SkillsDir returns the path to the .claude/skills/ directory for a given repo root.
func SkillsDir(repoRoot string) string {
	// TODO: implement
	panic("not implemented")
}
