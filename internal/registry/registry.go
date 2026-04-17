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

// ListSkills returns all skills available in the given registry by walking the cache.
// Each skill is expected to live in its own subdirectory containing a SKILL.md file.
func (a *Adapter) ListSkills(reg Registry) ([]model.RegistrySkill, error) {
	dir := a.cacheDir(reg.Alias)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := a.Fetch(reg); err != nil {
			return nil, err
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("registry: failed to read cache for %q: %w", reg.Alias, err)
	}

	var skills []model.RegistrySkill
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		skillMD := filepath.Join(dir, entry.Name(), "SKILL.md")
		if _, err := os.Stat(skillMD); os.IsNotExist(err) {
			continue
		}
		rs, err := frontmatter.Parse(skillMD)
		if err != nil {
			continue // skip malformed entries
		}
		if rs.Name == "" {
			rs.Name = entry.Name()
		}
		skills = append(skills, *rs)
	}
	return skills, nil
}

// GetSkill returns the metadata for a single skill from the registry cache.
func (a *Adapter) GetSkill(reg Registry, name string) (*model.RegistrySkill, error) {
	dir := a.cacheDir(reg.Alias)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := a.Fetch(reg); err != nil {
			return nil, err
		}
	}

	skillMD := filepath.Join(dir, name, "SKILL.md")
	if _, err := os.Stat(skillMD); os.IsNotExist(err) {
		return nil, fmt.Errorf("registry: skill %q not found in registry %q", name, reg.Alias)
	}

	rs, err := frontmatter.Parse(skillMD)
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

	srcDir := filepath.Join(regDir, name)
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
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
func copyFile(src, dst string, mode os.FileMode) error {
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
	defer out.Close() //nolint:errcheck

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
		// Count skills (subdirs with SKILL.md)
		skillCount := 0
		subEntries, _ := os.ReadDir(regDir)
		for _, se := range subEntries {
			if se.IsDir() && !strings.HasPrefix(se.Name(), ".") {
				if _, err := os.Stat(filepath.Join(regDir, se.Name(), "SKILL.md")); err == nil {
					skillCount++
				}
			}
		}
		_, _ = fmt.Fprintf(&sb, "  %-20s  %3d skills  last fetched %s\n",
			entry.Name(), skillCount, info.ModTime().Format(time.RFC3339))
		count++
	}
	if count == 0 {
		return "cache is empty (no registries fetched)", nil
	}
	return sb.String(), nil
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
