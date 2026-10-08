package registry

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// SkillSource describes where a skill's files live inside a registry.
type SkillSource struct {
	// Dir is the on-disk skill directory (inside the cache clone, or inside a
	// local-folder registry).
	Dir string
	// RelPath is the skill directory relative to the registry root, using
	// forward slashes (e.g. "skills/pdf-processing").
	RelPath string
	// Commit is the registry HEAD commit for git registries. Empty for local
	// folders, which have no stable revision.
	Commit string
}

var commitRx = regexp.MustCompile(`^[0-9a-fA-F]{7,64}$`)

// ensureAvailable makes sure the registry's files exist locally, cloning a
// remote registry on first use.
func (a *Adapter) ensureAvailable(reg Registry) (string, error) {
	root := a.sourceRoot(reg)
	if IsLocalRegistryURL(reg.URL) {
		if _, err := os.Stat(root); err != nil {
			return "", fmt.Errorf("local source not found: %s", root)
		}
		return root, nil
	}
	if _, err := os.Stat(root); os.IsNotExist(err) {
		if err := a.Fetch(reg); err != nil {
			return "", err
		}
	}
	return root, nil
}

// Source locates a skill inside a registry and reports its directory, its
// path relative to the registry root and (for git registries) the commit the
// cache is currently checked out at.
func (a *Adapter) Source(reg Registry, name string) (*SkillSource, error) {
	root, err := a.ensureAvailable(reg)
	if err != nil {
		return nil, err
	}
	dir := findSkillDir(root, name)
	if dir == "" {
		return nil, fmt.Errorf("%w: %q in registry %q", ErrSkillNotFound, name, reg.Alias)
	}
	rel, err := filepath.Rel(root, dir)
	if err != nil {
		return nil, err
	}
	src := &SkillSource{Dir: dir, RelPath: filepath.ToSlash(rel)}
	if !IsLocalRegistryURL(reg.URL) {
		if head, err := runGit("-C", root, "rev-parse", "HEAD"); err == nil {
			src.Commit = head
		}
	}
	return src, nil
}

// CopySkillAt installs the skill exactly as it was at commit, which makes
// installs reproducible from skell.lock. relPath is the skill's location in
// the registry at that commit (from the lock file); when empty, the skill's
// current location is used. Local-folder registries and an empty commit fall
// back to copying the current files.
func (a *Adapter) CopySkillAt(reg Registry, name, commit, relPath, destPath string) error {
	if commit == "" || IsLocalRegistryURL(reg.URL) {
		return a.CopySkillTo(reg, name, "", destPath)
	}
	if !commitRx.MatchString(commit) {
		return fmt.Errorf("registry: invalid commit %q", commit)
	}
	if destPath == "" || destPath == "/" || destPath == "." {
		return fmt.Errorf("registry: refusing to copy into unsafe destination %q", destPath)
	}

	src, srcErr := a.Source(reg, name)
	if srcErr == nil && strings.EqualFold(src.Commit, commit) {
		return a.CopySkillTo(reg, name, "", destPath)
	}
	if relPath == "" {
		if srcErr != nil {
			return srcErr
		}
		relPath = src.RelPath
	}
	if err := validateRelPath(relPath); err != nil {
		return err
	}

	dir := a.cacheDir(reg.Alias)
	if _, err := runGit("-C", dir, "cat-file", "-e", commit+"^{commit}"); err != nil {
		if _, err := runGit("-C", dir, "fetch", "--depth=1", "origin", commit); err != nil {
			return fmt.Errorf("registry: commit %s is not available from %q: %w", shortCommit(commit), reg.Alias, err)
		}
	}

	return stageAndSwap(destPath, func(stage string) error {
		return extractCommitTree(dir, commit, relPath, stage)
	})
}

// validateRelPath rejects registry-relative paths that could escape the
// registry root once joined onto a filesystem path.
func validateRelPath(rel string) error {
	if rel == "" || rel == "." {
		return nil
	}
	clean := path.Clean(rel)
	if path.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, "../") || strings.HasPrefix(clean, "-") {
		return fmt.Errorf("registry: invalid skill path %q", rel)
	}
	return nil
}

func shortCommit(c string) string {
	if len(c) > 12 {
		return c[:12]
	}
	return c
}

// extractCommitTree writes the files under relPath at commit into dest.
func extractCommitTree(repoDir, commit, relPath, dest string) error {
	args := []string{"-C", repoDir, "archive", "--format=tar", commit}
	if relPath != "" && relPath != "." {
		args = append(args, "--", relPath)
	}
	data, err := runGitStdout(args...)
	if err != nil {
		return fmt.Errorf("registry: cannot read skill at %s: %w", shortCommit(commit), err)
	}
	prefix := ""
	if relPath != "" && relPath != "." {
		prefix = strings.TrimSuffix(relPath, "/") + "/"
	}
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	found := false
	tr := tar.NewReader(bytes.NewReader(data))
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("registry: invalid archive: %w", err)
		}
		if !strings.HasPrefix(hdr.Name, prefix) {
			continue
		}
		rel := strings.TrimPrefix(hdr.Name, prefix)
		if rel == "" {
			continue
		}
		if err := validateRelPath(rel); err != nil {
			return err
		}
		target := filepath.Join(dest, filepath.FromSlash(rel))
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			found = true
			if err := writeArchiveFile(tr, target, os.FileMode(hdr.Mode)&0777); err != nil {
				return err
			}
		case tar.TypeSymlink:
			linkPath := filepath.Join(dest, filepath.FromSlash(rel))
			if err := validateSymlinkTarget(dest, linkPath, hdr.Linkname); err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			if err := os.Symlink(hdr.Linkname, target); err != nil {
				return err
			}
		}
	}
	if !found {
		return fmt.Errorf("registry: %q does not exist at commit %s", relPath, shortCommit(commit))
	}
	return nil
}

func writeArchiveFile(r io.Reader, target string, mode os.FileMode) (retErr error) {
	if mode == 0 {
		mode = 0644
	}
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode) //nolint:gosec
	if err != nil {
		return err
	}
	defer func() {
		if cerr := out.Close(); cerr != nil && retErr == nil {
			retErr = cerr
		}
	}()
	_, err = io.Copy(out, io.LimitReader(r, 100<<20))
	return err
}

// stageAndSwap fills a staging directory next to destPath via fill and, only
// if that succeeds, replaces destPath with it. A failure never destroys an
// existing install.
func stageAndSwap(destPath string, fill func(stage string) error) error {
	parent := filepath.Dir(destPath)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return fmt.Errorf("registry: failed to create %q: %w", parent, err)
	}
	staging, err := os.MkdirTemp(parent, ".skell-copy-")
	if err != nil {
		return fmt.Errorf("registry: failed to create staging dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(staging) }()

	stage := filepath.Join(staging, "skill")
	if err := fill(stage); err != nil {
		return err
	}
	if err := os.RemoveAll(destPath); err != nil {
		return fmt.Errorf("registry: failed to clear destination %q: %w", destPath, err)
	}
	if err := os.Rename(stage, destPath); err != nil {
		return fmt.Errorf("registry: failed to move skill into place: %w", err)
	}
	return nil
}

// runGitStdout runs git and returns its raw stdout (stderr is only used for
// error messages). Used where output is binary, e.g. git archive.
func runGitStdout(args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), gitTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	hideSubprocessWindow(cmd)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GIT_ASKPASS=", "SSH_ASKPASS=")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, stderr.String())
	}
	return stdout.Bytes(), nil
}

// DiffDirs returns a unified diff from oldDir (the installed skill) to newDir
// (the registry version), using git's no-index diff so no repository is
// needed. An empty string means the trees are identical.
func DiffDirs(oldDir, newDir string, color bool) (string, error) {
	tmp, err := os.MkdirTemp("", "skell-diff-")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.RemoveAll(tmp) }()

	// Copy both sides under fixed names so the diff shows short, stable
	// paths ("installed/SKILL.md" vs "latest/SKILL.md").
	for _, p := range []struct{ src, name string }{{oldDir, "installed"}, {newDir, "latest"}} {
		dst := filepath.Join(tmp, p.name)
		if p.src == "" {
			if err := os.MkdirAll(dst, 0755); err != nil {
				return "", err
			}
			continue
		}
		if err := copyDir(p.src, dst); err != nil {
			return "", err
		}
	}

	colorFlag := "--no-color"
	if color {
		colorFlag = "--color=always"
	}
	ctx, cancel := context.WithTimeout(context.Background(), gitTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "-c", "core.quotepath=off", "diff", "--no-index", colorFlag,
		"--src-prefix=", "--dst-prefix=", "--", "installed", "latest")
	cmd.Dir = tmp
	hideSubprocessWindow(cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	var exitErr *exec.ExitError
	if err != nil && (!errors.As(err, &exitErr) || exitErr.ExitCode() != 1) {
		return "", fmt.Errorf("git diff: %w\n%s", err, stderr.String())
	}
	return stdout.String(), nil
}

// CopyTree replaces dst with a copy of the directory src, staging the copy
// first so a failure leaves dst untouched. Symlinks escaping src are rejected,
// as for registry copies.
func CopyTree(src, dst string) error {
	return stageAndSwap(dst, func(stage string) error { return copyDir(src, stage) })
}
