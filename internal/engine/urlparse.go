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
	if len(raw) >= 2 && raw[1] == ':' && (raw[2] == '\\' || raw[2] == '/') {
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

	// Support local filesystem paths (absolute, ~, Windows, file://) — added for local skill folders
	if isLocalPathForAdd(rawURL) {
		clean := rawURL
		clean = strings.TrimPrefix(clean, "file://")
		// Expand ~
		if strings.HasPrefix(clean, "~/") {
			if home, err := os.UserHomeDir(); err == nil {
				clean = filepath.Join(home, clean[2:])
			}
		}
		abs, err := filepath.Abs(clean)
		if err != nil {
			abs = clean
		}

		base := filepath.Base(abs)
		if base == "." || base == ".." || base == "/" {
			base = "local"
		}
		alias := strings.ToLower(strings.TrimSuffix(base, filepath.Ext(base)))

		return ParsedSkillURL{
			GitURL:    abs,
			Alias:     alias,
			SubPath:   "",
			SkillName: "", // will be treated as a registry root; AddFromURL will handle single-skill case
		}, nil
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return ParsedSkillURL{}, fmt.Errorf("invalid URL %q: %w", rawURL, err)
	}
	if u.Scheme == "" || u.Host == "" {
		return ParsedSkillURL{}, fmt.Errorf("URL must be absolute (include https://): %q", rawURL)
	}

	// Normalize: drop trailing slash and any /blob/main/.../SKILL.md → treat parent dir as skill
	p := strings.TrimRight(u.Path, "/")

	// Support both /tree/ (directory view) and /blob/ (file view people often copy)
	sep := "/tree/"
	if idx := strings.Index(p, "/blob/"); idx >= 0 {
		sep = "/blob/"
	}

	var repoPath, subPath string
	if idx := strings.Index(p, sep); idx >= 0 {
		repoPath = p[:idx]
		afterSep := p[idx+len(sep):]
		// afterSep = "<branch>/<rest...>"
		// Split once to separate branch from the actual content path
		if slash := strings.Index(afterSep, "/"); slash >= 0 {
			subPath = afterSep[slash+1:]
		}
		// If nothing after branch, subPath stays ""
	} else {
		repoPath = p
	}

	// If someone linked directly to SKILL.md, treat the containing directory as the skill
	if strings.HasSuffix(subPath, "SKILL.md") || strings.HasSuffix(subPath, ".md") {
		subPath = path.Dir(subPath)
		if subPath == "." {
			subPath = ""
		}
	}

	gitURL := u.Scheme + "://" + u.Host + repoPath
	alias := path.Base(strings.TrimSuffix(repoPath, ".git"))
	alias = strings.ToLower(strings.TrimSpace(alias))

	// Skill name heuristic (preserves original behavior + supports /blob/):
	// - If subPath has 2+ segments, last segment is the skill.
	// - If exactly 1 segment and it is *not* a known container dir, treat it as skill.
	// - Otherwise (empty or pure container), SkillName stays "" → registry root.
	var skillName string
	if subPath != "" {
		parts := strings.Split(strings.Trim(subPath, "/"), "/")
		last := parts[len(parts)-1]
		containers := map[string]bool{"skills": true, "claude": true, "codex": true, ".github": true, "cursor": true, "ai": true}
		if len(parts) >= 2 || !containers[last] {
			if last != "" {
				skillName = last
			}
		}
	}

	return ParsedSkillURL{
		GitURL:    gitURL,
		Alias:     alias,
		SubPath:   subPath,
		SkillName: skillName,
	}, nil
}
