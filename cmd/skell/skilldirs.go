package skell

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/spf13/cobra"
)

// skillRef points at one skill folder on disk.
type skillRef struct {
	Repo string `json:"repo,omitempty"`
	Name string `json:"name"`
	Dir  string `json:"path"`
}

// installedSkillRefs resolves installed skills in a repo to folders. With
// names it keeps only those; an unknown name is an error.
func installedSkillRefs(eng *engine.Engine, repo, targetID string, names []string) ([]skillRef, error) {
	skills, err := eng.ListFor(repo, targetID)
	if err != nil {
		return nil, err
	}
	var refs []skillRef
	seen := map[string]bool{}
	for _, s := range skills {
		dir := filepath.Join(repo, filepath.FromSlash(s.InstalledPath))
		if len(names) > 0 && !contains(names, s.Name) {
			continue
		}
		if seen[dir] {
			continue
		}
		seen[dir] = true
		refs = append(refs, skillRef{Repo: repo, Name: s.Name, Dir: dir})
	}
	for _, n := range names {
		found := false
		for _, r := range refs {
			if r.Name == n {
				found = true
			}
		}
		if !found {
			return nil, fmt.Errorf("skill %q is not installed in %s", n, repo)
		}
	}
	sort.Slice(refs, func(i, j int) bool { return refs[i].Name < refs[j].Name })
	return refs, nil
}

// pathSkillRefs finds skill folders at or below path: the folder itself when
// it holds a SKILL.md, otherwise every subfolder that does.
func pathSkillRefs(path string) ([]skillRef, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a folder", path)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(filepath.Join(abs, "SKILL.md")); err == nil {
		return []skillRef{{Name: filepath.Base(abs), Dir: abs}}, nil
	}
	var refs []skillRef
	_ = filepath.WalkDir(abs, func(p string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		if d.Name() == ".git" || d.Name() == "node_modules" {
			return filepath.SkipDir
		}
		if rel, _ := filepath.Rel(abs, p); strings.Count(filepath.ToSlash(rel), "/") > 4 {
			return filepath.SkipDir
		}
		if p != abs {
			if _, err := os.Stat(filepath.Join(p, "SKILL.md")); err == nil {
				refs = append(refs, skillRef{Name: d.Name(), Dir: p})
				return filepath.SkipDir
			}
		}
		return nil
	})
	if len(refs) == 0 {
		return nil, fmt.Errorf("no SKILL.md found in %s or its subfolders", path)
	}
	sort.Slice(refs, func(i, j int) bool { return refs[i].Dir < refs[j].Dir })
	return refs, nil
}

// collectSkillRefs is the shared "which skills do you mean?" logic: --path
// folders, or installed skills in each selected repo.
func collectSkillRefs(_ *cobra.Command, f repoFlags, path, targetID string, names []string) ([]skillRef, error) {
	if path != "" {
		if len(names) > 0 {
			return nil, fmt.Errorf("--path cannot be combined with skill names")
		}
		return pathSkillRefs(path)
	}
	repos, err := resolveRepos(f)
	if err != nil {
		return nil, err
	}
	eng := engine.New(defaultCacheRoot())
	var all []skillRef
	for _, repo := range repos {
		refs, err := installedSkillRefs(eng, repo, targetID, names)
		if err != nil {
			if len(names) == 0 {
				continue // a repo without skills has nothing to check
			}
			return nil, repoErr(repos, repo, err)
		}
		all = append(all, refs...)
	}
	return all, nil
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// Marks shared by the scan, verify, report and test commands.
const (
	markOK   = "✓"
	markWarn = "!"
	markBad  = "✗"
)
