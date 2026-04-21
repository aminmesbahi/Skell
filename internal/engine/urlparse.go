package engine

import (
	"fmt"
	"net/url"
	"path"
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
func ParseSkillURL(rawURL string) (ParsedSkillURL, error) {
	rawURL = strings.TrimSpace(rawURL)
	u, err := url.Parse(rawURL)
	if err != nil {
		return ParsedSkillURL{}, fmt.Errorf("invalid URL %q: %w", rawURL, err)
	}
	if u.Scheme == "" || u.Host == "" {
		return ParsedSkillURL{}, fmt.Errorf("URL must be absolute (include https://): %q", rawURL)
	}

	p := strings.TrimRight(u.Path, "/")

	const treeSep = "/tree/"
	treeIdx := strings.Index(p, treeSep)

	var repoPath, subPath string
	if treeIdx >= 0 {
		repoPath = p[:treeIdx]
		afterTree := p[treeIdx+len(treeSep):]
		slashIdx := strings.Index(afterTree, "/")
		if slashIdx >= 0 {
			subPath = afterTree[slashIdx+1:]
		}
	} else {
		repoPath = p
	}

	gitURL := u.Scheme + "://" + u.Host + repoPath
	// Strip .git suffix for the alias but keep it in the git URL.
	alias := path.Base(strings.TrimSuffix(repoPath, ".git"))
	alias = strings.ToLower(alias)

	// ≥2 path segments after branch → last segment is the skill name.
	// ≤1 path segment → registry root only.
	var skillName string
	if subPath != "" {
		parts := strings.Split(strings.Trim(subPath, "/"), "/")
		if len(parts) >= 2 {
			skillName = parts[len(parts)-1]
		}
	}

	return ParsedSkillURL{
		GitURL:    gitURL,
		Alias:     alias,
		SubPath:   subPath,
		SkillName: skillName,
	}, nil
}
