package registry

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aminmesbahi/skell/internal/frontmatter"
	"github.com/aminmesbahi/skell/internal/model"
)

const gitTimeout = 2 * time.Minute

var ErrSkillNotFound = errors.New("registry: skill not found")

var allowedURLSchemes = map[string]bool{
	"https": true,
	"http":  true,
	"ssh":   true,
	"git":   true,
	"file":  true,
}

func validateRegistryURL(raw string) error {
	if raw == "" {
		return fmt.Errorf("registry: URL is empty")
	}
	if strings.HasPrefix(raw, "-") {
		return fmt.Errorf("registry: URL %q must not start with '-'", raw)
	}
	if isLocalPath(raw) {
		return nil
	}
	if u, err := url.Parse(raw); err == nil && u.Scheme != "" {
		if !allowedURLSchemes[strings.ToLower(u.Scheme)] {
			return fmt.Errorf("registry: URL scheme %q is not allowed", u.Scheme)
		}
		return nil
	}
	if isSCPStyleGitURL(raw) {
		return nil
	}
	return fmt.Errorf("registry: URL %q is not an absolute http(s)/ssh/git URL", raw)
}

func isLocalPath(raw string) bool {
	if strings.HasPrefix(raw, "/") {
		return true
	}
	if len(raw) >= 3 && raw[1] == ':' && (raw[2] == '/' || raw[2] == '\\') {
		c := raw[0]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			return true
		}
	}
	return false
}

func isSCPStyleGitURL(raw string) bool {
	if strings.Contains(raw, "://") {
		return false
	}
	colon := strings.Index(raw, ":")
	slash := strings.Index(raw, "/")
	return colon > 0 && (slash < 0 || colon < slash)
}

func validateAlias(alias string) error {
	if alias == "" {
		return fmt.Errorf("registry: alias is empty")
	}
	if strings.ContainsAny(alias, "/\\") || alias == "." || alias == ".." || strings.HasPrefix(alias, "-") {
		return fmt.Errorf("registry: alias %q is invalid", alias)
	}
	return nil
}

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

func (a *Adapter) cacheDir(alias string) string {
	return filepath.Join(a.cacheRoot, alias)
}

// IsLocalRegistryURL returns true if the registry points to a local filesystem path
// (absolute, relative, or file:// URL). These are treated as direct sources
// instead of git-based caches.
func IsLocalRegistryURL(raw string) bool {
	if isLocalPath(raw) {
		return true
	}
	if strings.HasPrefix(raw, "file:") {
		return true
	}
	return false
}

// sourceRoot returns the filesystem path to use as the skill root for a registry.
// For local directories it returns the path itself (resolved).
// For remote registries it returns the cache location.
func (a *Adapter) sourceRoot(reg Registry) string {
	u := reg.URL
	if strings.HasPrefix(u, "file://") {
		u = strings.TrimPrefix(u, "file://")
	}
	if isLocalPath(u) {
		if abs, err := filepath.Abs(u); err == nil {
			return abs
		}
		return u
	}
	return a.cacheDir(reg.Alias)
}

func runGit(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), gitTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Stdin = nil
	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_ASKPASS=",
		"SSH_ASKPASS=",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out)), nil
}

// Fetch performs a git clone or pull on the cached clone of the given registry.
// For local directory sources this is a no-op (after existence check).
func (a *Adapter) Fetch(reg Registry) error {
	if err := validateAlias(reg.Alias); err != nil {
		return err
	}
	if err := validateRegistryURL(reg.URL); err != nil {
		return err
	}

	if IsLocalRegistryURL(reg.URL) {
		root := a.sourceRoot(reg)
		if _, err := os.Stat(root); err != nil {
			return fmt.Errorf("local skill source not found: %s (%w)", root, err)
		}
		return nil
	}

	// Remote git registry — use cache
	dir := a.cacheDir(reg.Alias)

	if _, err := os.Stat(filepath.Join(dir, ".git")); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(dir), 0755); err != nil {
			return fmt.Errorf("registry: failed to create cache root: %w", err)
		}
		// "--" terminates option parsing so a URL or path beginning with '-'
		// cannot be interpreted as a git flag.
		if _, err := runGit("clone", "--depth=1", "--config", "core.autocrlf=false", "--", reg.URL, dir); err != nil {
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
// walking the cache (or local directory) for SKILL.md files.
func (a *Adapter) ListSkills(reg Registry) ([]model.RegistrySkill, error) {
	root := a.sourceRoot(reg)

	if IsLocalRegistryURL(reg.URL) {
		if _, err := os.Stat(root); err != nil {
			return nil, fmt.Errorf("local source not found: %s", root)
		}
	} else {
		if _, err := os.Stat(root); os.IsNotExist(err) {
			if err := a.Fetch(reg); err != nil {
				return nil, err
			}
		}
	}

	return walkSkills(root)
}

// walkSkills recursively finds all SKILL.md files under root and returns the
// parsed skills. Each skill's name is the name of the directory containing
// the SKILL.md file.
func walkSkills(root string) ([]model.RegistrySkill, error) {
	var skills []model.RegistrySkill
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		if d.IsDir() || d.Name() != "SKILL.md" {
			return nil
		}
		rs, parseErr := frontmatter.Parse(path)
		if parseErr != nil {
			return nil
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
		if d.IsDir() && d.Name() == ".git" {
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

// GetSkill returns the metadata for a single skill from the registry (cache or local dir).
func (a *Adapter) GetSkill(reg Registry, name string) (*model.RegistrySkill, error) {
	root := a.sourceRoot(reg)

	if IsLocalRegistryURL(reg.URL) {
		if _, err := os.Stat(root); err != nil {
			return nil, fmt.Errorf("local source not found: %s", root)
		}
	} else {
		if _, err := os.Stat(root); os.IsNotExist(err) {
			if err := a.Fetch(reg); err != nil {
				return nil, err
			}
		}
	}

	skillDir := findSkillDir(root, name)
	if skillDir == "" {
		return nil, fmt.Errorf("%w: %q in registry %q", ErrSkillNotFound, name, reg.Alias)
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

// CopySkillTo copies the skill directory from the source (cache or local folder)
// into the destination path.
func (a *Adapter) CopySkillTo(reg Registry, name, _ string, destPath string) error {
	root := a.sourceRoot(reg)

	if !IsLocalRegistryURL(reg.URL) {
		if _, err := os.Stat(root); os.IsNotExist(err) {
			if err := a.Fetch(reg); err != nil {
				return err
			}
		}
	}

	srcDir := findSkillDir(root, name)
	if srcDir == "" {
		return fmt.Errorf("%w: %q in registry %q", ErrSkillNotFound, name, reg.Alias)
	}

	if destPath == "" || destPath == "/" || destPath == "." {
		return fmt.Errorf("registry: refusing to copy into unsafe destination %q", destPath)
	}
	if err := os.RemoveAll(destPath); err != nil {
		return fmt.Errorf("registry: failed to clear destination %q: %w", destPath, err)
	}

	if err := copyDir(srcDir, destPath); err != nil {
		return fmt.Errorf("registry: failed to copy %q: %w", name, err)
	}
	return nil
}

// copyDir recursively copies src into dst. Symlinks are preserved as symlinks
// (rather than dereferenced) so a malicious registry cannot exfiltrate files
// from outside its own tree by linking to absolute paths.
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
		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			return os.Symlink(link, target)
		}
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

// CacheStatus returns a human-readable summary of cached (remote) registries.
// Local folder sources are not shown here as they are used directly.
func (a *Adapter) CacheStatus() (string, error) {
	if _, err := os.Stat(a.cacheRoot); os.IsNotExist(err) {
		return "cache is empty (no remote registries fetched)", nil
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
		skillList, _ := walkSkills(regDir)
		_, _ = fmt.Fprintf(&sb, "  %-20s  %3d skills  last fetched %s\n",
			entry.Name(), len(skillList), info.ModTime().Format(time.RFC3339))
		count++
	}
	if count == 0 {
		return "cache is empty (no remote registries fetched)", nil
	}
	return "remote registry cache status:\n" + sb.String(), nil
}

// CacheClear removes all cached registry data.
func (a *Adapter) CacheClear() error {
	if err := os.RemoveAll(a.cacheRoot); err != nil {
		return fmt.Errorf("registry: failed to clear cache: %w", err)
	}
	return nil
}

// CacheRefresh fetches the latest from the given registry (no-op for local folders).
func (a *Adapter) CacheRefresh(reg Registry) error {
	if IsLocalRegistryURL(reg.URL) {
		return nil // local sources are always live
	}
	return a.Fetch(reg)
}
