// Package registry fetches, caches, and indexes skills from remote git registries.
package registry

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aminmesbahi/skell/internal/frontmatter"
	"github.com/aminmesbahi/skell/internal/model"
)

// Registry holds connection details for a single remote skill registry.
type Registry struct {
	Alias string
	URL   string
}

// Adapter provides access to one or more remote skill registries via a local cache.
type Adapter struct {
	cacheRoot string
}

// NewAdapter creates an Adapter rooted at the given cache directory.
func NewAdapter(cacheRoot string) *Adapter {
	return &Adapter{cacheRoot: cacheRoot}
}

// cacheDir returns the local cache path for the given registry alias.
func (a *Adapter) cacheDir(alias string) string {
	return filepath.Join(a.cacheRoot, alias)
}

// runGit executes a git command and returns its combined output.
func runGit(args ...string) (string, error) {
	cmd := exec.CommandContext(context.Background(), "git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out)), nil
}

// Fetch performs a git clone or pull on the cached clone of the given registry.
func (a *Adapter) Fetch(reg Registry) error {
	dir := a.cacheDir(reg.Alias)

	if _, err := os.Stat(filepath.Join(dir, ".git")); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(dir), 0755); err != nil {
			return fmt.Errorf("registry: failed to create cache root: %w", err)
		}
		if _, err := runGit("clone", "--depth=1", "--config", "core.autocrlf=false", reg.URL, dir); err != nil {
			return fmt.Errorf("registry: clone %q failed: %w", reg.URL, err)
		}
		return nil
	}

	if _, err := runGit("-C", dir, "fetch", "--prune", "--depth=1"); err != nil {
		return fmt.Errorf("registry: fetch %q failed: %w", reg.Alias, err)
	}
	if _, err := runGit("-C", dir, "reset", "--hard", "FETCH_HEAD"); err != nil {
		return fmt.Errorf("registry: reset failed for %q: %w", reg.Alias, err)
	}
	return nil
}

// ListSkills returns all skills available in the given registry by recursively
// walking the cache for SKILL.md files. Skills may be nested at any depth
// (e.g. skills/<name>/SKILL.md or plugins/<plugin>/skills/<name>/SKILL.md).
func (a *Adapter) ListSkills(reg Registry) ([]model.RegistrySkill, error) {
	dir := a.cacheDir(reg.Alias)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := a.Fetch(reg); err != nil {
			return nil, err
		}
	}

	return walkSkills(dir)
}

// walkSkills recursively finds all SKILL.md files under root and returns the
// parsed skills. Each skill's name is the name of the directory containing
// the SKILL.md file.
func walkSkills(root string) ([]model.RegistrySkill, error) {
	var skills []model.RegistrySkill
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		// Skip hidden directories (e.g. .git).
		if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}
		if d.IsDir() || d.Name() != "SKILL.md" {
			return nil
		}
		rs, parseErr := frontmatter.Parse(path)
		if parseErr != nil {
			return nil // skip malformed SKILL.md
		}
		if rs.Name == "" {
			rs.Name = filepath.Base(filepath.Dir(path))
		}
		skills = append(skills, *rs)
		return nil
	})
	return skills, err
}

// findSkillDir searches recursively under root for a skill directory that
// either has a directory name matching skillName OR whose SKILL.md frontmatter
// declares that name. This handles registries where the directory name differs
// from the name field in SKILL.md.
func findSkillDir(root, skillName string) string {
	found := ""
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || found != "" {
			return nil
		}
		if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}
		// Only act on SKILL.md files so we can inspect each potential skill.
		if d.IsDir() || d.Name() != "SKILL.md" {
			return nil
		}
		dir := filepath.Dir(path)
		// Fast path: directory name matches exactly.
		if filepath.Base(dir) == skillName {
			found = dir
			return filepath.SkipAll
		}
		// Fallback: parse frontmatter and check the name field.
		rs, parseErr := frontmatter.Parse(path)
		if parseErr == nil && rs.Name == skillName {
			found = dir
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

// GetSkill returns the metadata for a single skill from the registry cache.
func (a *Adapter) GetSkill(reg Registry, name string) (*model.RegistrySkill, error) {
	dir := a.cacheDir(reg.Alias)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := a.Fetch(reg); err != nil {
			return nil, err
		}
	}

	skillDir := findSkillDir(dir, name)
	if skillDir == "" {
		return nil, fmt.Errorf("registry: skill %q not found in registry %q", name, reg.Alias)
	}

	rs, err := frontmatter.Parse(filepath.Join(skillDir, "SKILL.md"))
	if err != nil {
		return nil, fmt.Errorf("registry: failed to parse SKILL.md for %q: %w", name, err)
	}
	if rs.Name == "" {
		rs.Name = name
	}
	return rs, nil
}

// CopySkillTo copies the skill directory from the cache into the destination path.
func (a *Adapter) CopySkillTo(reg Registry, name, _ string, destPath string) error {
	regDir := a.cacheDir(reg.Alias)
	if _, err := os.Stat(regDir); os.IsNotExist(err) {
		if err := a.Fetch(reg); err != nil {
			return err
		}
	}

	srcDir := findSkillDir(regDir, name)
	if srcDir == "" {
		return fmt.Errorf("registry: skill %q not in cache for registry %q", name, reg.Alias)
	}

	if err := copyDir(srcDir, destPath); err != nil {
		return fmt.Errorf("registry: failed to copy %q: %w", name, err)
	}
	return nil
}

// copyDir recursively copies src into dst.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		return copyFile(path, target, info.Mode())
	})
}

// copyFile copies a single file, preserving its mode bits.
func copyFile(src, dst string, mode os.FileMode) (retErr error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close() //nolint:errcheck

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := out.Close(); cerr != nil && retErr == nil {
			retErr = cerr
		}
	}()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

// CacheStatus returns a human-readable summary of what is cached locally.
func (a *Adapter) CacheStatus() (string, error) {
	if _, err := os.Stat(a.cacheRoot); os.IsNotExist(err) {
		return "cache is empty (no registries fetched)", nil
	}

	entries, err := os.ReadDir(a.cacheRoot)
	if err != nil {
		return "", fmt.Errorf("registry: cannot read cache root: %w", err)
	}

	var sb strings.Builder
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		regDir := filepath.Join(a.cacheRoot, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}
		// Count skills recursively.
		skillList, _ := walkSkills(regDir)
		_, _ = fmt.Fprintf(&sb, "  %-20s  %3d skills  last fetched %s\n",
			entry.Name(), len(skillList), info.ModTime().Format(time.RFC3339))
		count++
	}
	if count == 0 {
		return "cache is empty (no registries fetched)", nil
	}
	return "cache status:\n" + sb.String(), nil
}

// CacheClear removes all cached registry data.
func (a *Adapter) CacheClear() error {
	if err := os.RemoveAll(a.cacheRoot); err != nil {
		return fmt.Errorf("registry: failed to clear cache: %w", err)
	}
	return nil
}

// CacheRefresh fetches the latest from the given registry, updating the local cache.
func (a *Adapter) CacheRefresh(reg Registry) error {
	return a.Fetch(reg)
}
