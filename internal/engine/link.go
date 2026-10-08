package engine

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/aminmesbahi/skell/internal/audit"
	"github.com/aminmesbahi/skell/internal/frontmatter"
	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/model"
)

// LinkResult describes a 'skell link'.
type LinkResult struct {
	Name   string `json:"name"`
	Source string `json:"source"`
	Path   string `json:"path"`
	Target string `json:"target"`
}

// Link makes a skill under development available to an agent without copying
// it: the skill folder is symlinked (a directory junction on Windows when
// symlinks need admin rights) into the repo's skills directory, so edits are
// live. The link is recorded in skell.lock as linked, is skipped by upgrade
// and status checks, is never pruned by sync, and is added to
// .git/info/exclude so the machine-specific link is not committed.
// Use 'skell remove <name>' to unlink.
func (e *Engine) Link(repoRoot, srcPath, targetID string) (*LinkResult, error) {
	abs, err := filepath.Abs(srcPath)
	if err != nil {
		return nil, err
	}
	if info, err := os.Stat(filepath.Join(abs, "SKILL.md")); err != nil || info.IsDir() {
		return nil, fmt.Errorf("%s does not contain a SKILL.md", abs)
	}
	name := filepath.Base(abs)
	version := ""
	if rs, err := frontmatter.ParseDir(abs); err == nil {
		if rs.Name != "" {
			name = rs.Name
		}
		version = rs.Metadata.Version
	}
	if err := ValidateSkillName(name); err != nil {
		return nil, err
	}

	m, t, err := resolveManifestFor(repoRoot, targetID)
	if err != nil {
		return nil, fmt.Errorf("no manifest found in %s — run 'skell init' first: %w", repoRoot, err)
	}
	dst := filepath.Join(t.SkillsDir(repoRoot), name)
	if _, err := os.Lstat(dst); err == nil {
		if cur, ok := linkDestination(dst); ok && sameDir(cur, abs) {
			return &LinkResult{Name: name, Source: abs, Path: dst, Target: t.ID}, nil
		}
		return nil, fmt.Errorf("skill %q already exists at %s; run 'skell remove %s' first", name, dst, name)
	}
	if err := os.MkdirAll(t.SkillsDir(repoRoot), 0o755); err != nil {
		return nil, err
	}
	if err := createDirLink(abs, dst); err != nil {
		return nil, fmt.Errorf("could not link %s: %w", abs, err)
	}

	lockPath := lockfile.PathFor(repoRoot, *t)
	lf, err := lockfile.Read(lockPath)
	if err != nil {
		lf = &lockfile.LockFile{SkellVersion: skellVersion(), Skills: []model.InstalledSkill{}}
	}
	now := time.Now().UTC().Format(time.RFC3339)
	lf.LockedAt = now
	lf.Upsert(model.InstalledSkill{
		Name:          name,
		Version:       version,
		SourceRepo:    abs,
		InstalledPath: t.InstalledRelPath(name),
		InstalledAt:   now,
		Linked:        true,
	})
	if err := lockfile.Write(lockPath, lf); err != nil {
		_ = os.Remove(dst)
		return nil, err
	}
	excludeFromGit(repoRoot, t.InstalledRelPath(name))
	for _, mt := range mirrorTargets(m, *t) {
		excludeFromGit(repoRoot, mt.InstalledRelPath(name))
	}
	if err := e.mirrorSkill(repoRoot, m, *t, name); err != nil {
		return nil, err
	}
	_ = e.logger.Log(audit.ActionInstall, name, version, "link", repoRoot)
	return &LinkResult{Name: name, Source: abs, Path: dst, Target: t.ID}, nil
}

// createDirLink creates a directory symlink at link pointing to target. On
// Windows, where symlinks usually need Developer Mode or admin rights, it falls
// back to a directory junction, which any user can create.
func createDirLink(target, link string) error {
	err := os.Symlink(target, link)
	if err == nil || runtime.GOOS != "windows" {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	out, jerr := exec.CommandContext(ctx, "cmd", "/c", "mklink", "/J", filepath.Clean(link), filepath.Clean(target)).CombinedOutput() //nolint:gosec
	if jerr != nil {
		return fmt.Errorf("symlink failed (%v) and junction failed: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// linkDestination reports where a symlink or junction points.
func linkDestination(p string) (string, bool) {
	info, err := os.Lstat(p)
	if err != nil || info.Mode()&(os.ModeSymlink|os.ModeIrregular) == 0 {
		return "", false
	}
	dst, err := os.Readlink(p)
	if err != nil {
		return "", false
	}
	if !filepath.IsAbs(dst) {
		dst = filepath.Join(filepath.Dir(p), dst)
	}
	return filepath.Clean(strings.TrimPrefix(dst, `\??\`)), true
}

func sameDir(a, b string) bool {
	ia, errA := os.Stat(a)
	ib, errB := os.Stat(b)
	return errA == nil && errB == nil && os.SameFile(ia, ib)
}

// excludeFromGit adds a repo-relative path to .git/info/exclude (once), so a
// machine-specific link never ends up in a commit. Best effort.
func excludeFromGit(repoRoot, rel string) {
	gitDir := filepath.Join(repoRoot, ".git")
	if info, err := os.Stat(gitDir); err != nil || !info.IsDir() {
		return
	}
	line := "/" + filepath.ToSlash(rel)
	path := filepath.Join(gitDir, "info", "exclude")
	if f, err := os.Open(path); err == nil { //nolint:gosec
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			if strings.TrimSpace(sc.Text()) == line {
				_ = f.Close()
				return
			}
		}
		_ = f.Close()
	}
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644) //nolint:gosec
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	_, _ = fmt.Fprintf(f, "\n# skell link (local only)\n%s\n", line)
}

// realSkillDir returns the destination of a linked skill (symlink or
// junction) so tools that don't follow links — like the spec validator —
// see the actual files; other paths are returned unchanged.
func realSkillDir(dir string) string {
	if dst, ok := linkDestination(dir); ok {
		return dst
	}
	return dir
}
