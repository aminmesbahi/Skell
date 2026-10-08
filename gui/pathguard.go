package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// skellHome returns the Skell home directory, honouring SKELL_HOME like the CLI.
func skellHome() (string, error) {
	if h := strings.TrimSpace(os.Getenv("SKELL_HOME")); h != "" {
		return h, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".skell"), nil
}

// pathGuard limits the filesystem paths the webview may read through the
// bound ReadFileContent/ListDirectory methods. The frontend renders untrusted
// registry content, so these bindings must not be an arbitrary-file-read
// primitive. Allowed: anything under the Skell home, any skills directory of a
// known agent layout (<repo>/.claude/skills/...), and roots explicitly
// granted at runtime (a folder the user picked, or a previewed local source).
type pathGuard struct {
	mu    sync.RWMutex
	roots []string
}

func (g *pathGuard) grant(root string) {
	if root == "" {
		return
	}
	root = canonical(root)
	g.mu.Lock()
	defer g.mu.Unlock()
	g.roots = append(g.roots, root)
}

// canonical returns an absolute, symlink-resolved, cleaned path. Paths that
// don't exist yet are resolved through their deepest existing ancestor, so a
// symlinked parent (e.g. macOS /var → /private/var) is resolved the same way
// for existing roots and not-yet-created files beneath them.
func canonical(p string) string {
	p = filepath.Clean(p)
	if abs, err := filepath.Abs(p); err == nil {
		p = abs
	}
	var rest []string
	for cur := p; ; {
		if resolved, err := filepath.EvalSymlinks(cur); err == nil {
			return filepath.Join(append([]string{resolved}, rest...)...)
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return p
		}
		rest = append([]string{filepath.Base(cur)}, rest...)
		cur = parent
	}
}

func within(root, p string) bool {
	rel, err := filepath.Rel(root, p)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

func (g *pathGuard) check(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("path is empty")
	}
	p := canonical(path)

	if home, err := skellHome(); err == nil && within(canonical(home), p) {
		return p, nil
	}
	if inAgentSkillsDir(p) {
		return p, nil
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, r := range g.roots {
		if within(r, p) {
			return p, nil
		}
	}
	return "", fmt.Errorf("access to %q is not permitted", path)
}

// inAgentSkillsDir reports whether p lies under <something>/<agentDir>/skills.
func inAgentSkillsDir(p string) bool {
	parts := strings.Split(filepath.ToSlash(p), "/")
	for i := 0; i+1 < len(parts); i++ {
		if parts[i+1] != "skills" {
			continue
		}
		for _, d := range agentDirs {
			if parts[i] == d {
				return true
			}
		}
	}
	return false
}
