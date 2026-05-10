package engine

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ParsedSkillURL holds the components extracted from a GitHub tree URL or plain git URL.
type ParsedSkillURL struct {
	// GitURL is the canonical git clone URL (no /tree/... suffix).
	GitURL string
	// Alias is the registry alias derived from the repository name.
	Alias string
	// SubPath is the full path component after the branch name, e.g. "ai/claude" or
	// "skills/my-skill". Empty for plain repo URLs with no /tree/ segment.
	SubPath string
	// SkillName is non-empty when the URL points to a specific skill directory.
	// Empty means the URL refers to a registry root (register only, no install).
	SkillName string
}

// ParseSkillURL decomposes a URL into registry and optional skill components.
//
// Supported formats:
//   - GitHub tree URL pointing to a skill:
//     https://github.com/owner/repo/tree/branch/path/to/skill
//   - GitHub tree URL pointing to a skills directory:
//     https://github.com/owner/repo/tree/branch/path
//   - Plain git URL (no /tree/ segment):
//     https://github.com/owner/repo
//
// The heuristic for "is this a specific skill?": the subpath after the branch
// must have at least two segments (e.g. "skills/akka-testing-patterns").
// A single-segment subpath (e.g. "skills") is treated as a registry root.
func isLocalPathForAdd(raw string) bool {
	if raw == "" {
		return false
	}
	if strings.HasPrefix(raw, "file:") {
		return true
	}
	if strings.HasPrefix(raw, "/") {
		return true
	}
	if len(raw) >= 3 && raw[1] == ':' && (raw[2] == '\\' || raw[2] == '/') {
		return true // Windows drive letter
	}
	if strings.HasPrefix(raw, "~/") || raw == "~" {
		return true
	}
	return false
}

// ParseSkillURL decomposes a skill source URL or local path into registry and
// optional skill components used by `skell add`.
func ParseSkillURL(rawURL string) (ParsedSkillURL, error) {
	rawURL = strings.TrimSpace(rawURL)
	if isLocalPathForAdd(rawURL) {
		return parseLocalSkillURL(rawURL)
	}
	return parseRemoteSkillURL(rawURL)
}


func parseLocalSkillURL(rawURL string) (ParsedSkillURL, error) {
	abs, base := normalizeLocalSkillPath(rawURL)
	alias := strings.ToLower(strings.TrimSuffix(base, filepath.Ext(base)))

	return ParsedSkillURL{
		GitURL:    abs,
		Alias:     alias,
		SubPath:   "",
		SkillName: detectLocalSkillName(abs, base),
	}, nil
}

func normalizeLocalSkillPath(rawURL string) (abs string, base string) {
	clean := strings.TrimPrefix(rawURL, "file://")
	if strings.HasPrefix(clean, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			clean = filepath.Join(home, clean[2:])
		}
	}
	abs, err := filepath.Abs(clean)
	if err != nil {
		abs = clean
	}
	base = filepath.Base(abs)
	if base == "." || base == ".." || base == "/" {
		base = "local"
	}
	return abs, base
}

func detectLocalSkillName(abs, base string) string {
	info, err := os.Stat(filepath.Join(abs, "SKILL.md"))
	if err == nil && !info.IsDir() {
		return base
	}
	return ""
}

func parseRemoteSkillURL(rawURL string) (ParsedSkillURL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ParsedSkillURL{}, fmt.Errorf("invalid URL %q: %w", rawURL, err)
	}
	if u.Scheme == "" || u.Host == "" {
		return ParsedSkillURL{}, fmt.Errorf("URL must be absolute (include https://): %q", rawURL)
	}

	repoPath, subPath := splitRepoAndSubPath(strings.TrimRight(u.Path, "/"))
	subPath = normalizeRemoteSubPath(subPath)

	gitURL := u.Scheme + "://" + u.Host + repoPath
	alias := path.Base(strings.TrimSuffix(repoPath, ".git"))
	alias = strings.ToLower(strings.TrimSpace(alias))

	return ParsedSkillURL{
		GitURL:    gitURL,
		Alias:     alias,
		SubPath:   subPath,
		SkillName: inferRemoteSkillName(subPath),
	}, nil
}

func splitRepoAndSubPath(p string) (repoPath, subPath string) {
	sep := "/tree/"
	if idx := strings.Index(p, "/blob/"); idx >= 0 {
		sep = "/blob/"
	}
	if idx := strings.Index(p, sep); idx >= 0 {
		repoPath = p[:idx]
		afterSep := p[idx+len(sep):]
		if slash := strings.Index(afterSep, "/"); slash >= 0 {
			subPath = afterSep[slash+1:]
		}
		return repoPath, subPath
	}
	return p, ""
}

func normalizeRemoteSubPath(subPath string) string {
	if strings.HasSuffix(subPath, "SKILL.md") || strings.HasSuffix(subPath, ".md") {
		subPath = path.Dir(subPath)
		if subPath == "." {
			return ""
		}
	}
	return subPath
}

func inferRemoteSkillName(subPath string) string {
	if subPath == "" {
		return ""
	}
	parts := strings.Split(strings.Trim(subPath, "/"), "/")
	last := parts[len(parts)-1]
	containers := map[string]bool{"skills": true, "claude": true, "codex": true, ".github": true, "cursor": true, "ai": true}
	if last == "" {
		return ""
	}
	if len(parts) >= 2 || !containers[last] {
		return last
	}
	return ""
}
