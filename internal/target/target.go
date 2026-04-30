// Package target defines the supported AI agent platforms (targets) that
// Skell can install skills into and the per-target file-system layout.
//
// Each target maps to a different on-disk convention:
//
//	claude       → .claude/skills/<name>/SKILL.md   (Anthropic Claude Code, agentskills.io)
//	codex        → .codex/skills/<name>/SKILL.md    (OpenAI Codex CLI)
//	copilot      → .github/skills/<name>/SKILL.md   (VS Code & GitHub Copilot agent skills)
//	cursor       → .cursor/skills/<name>/SKILL.md   (Cursor)
//
// The skill content format is identical across platforms; only the directory
// where Skell places (or finds) the skills differs.
package target

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Target identifies a single AI-agent platform layout.
type Target struct {
	// ID is the canonical short identifier (e.g. "claude", "codex").
	ID string
	// DisplayName is a human-readable label.
	DisplayName string
	// Dir is the repo-relative directory that hosts skell.toml,
	// skell.lock and the skills/ subdirectory (e.g. ".claude").
	Dir string
	// Aliases are alternative IDs accepted on the CLI.
	Aliases []string
}

// SkillsSubdir is the conventional name of the skills directory inside Dir.
const SkillsSubdir = "skills"

// Default is the target used when none is configured and none can be detected.
// Picked for backward compatibility with prior Skell versions, which only
// supported the Claude layout.
const Default = "claude"

var builtins = []Target{
	{ID: "claude", DisplayName: "Anthropic Claude Code", Dir: ".claude", Aliases: []string{"claude-code", "anthropic"}},
	{ID: "codex", DisplayName: "OpenAI Codex", Dir: ".codex", Aliases: []string{"openai", "openai-codex"}},
	{ID: "copilot", DisplayName: "GitHub Copilot / VS Code", Dir: ".github", Aliases: []string{"github-copilot", "vscode", "vscode-copilot", "github"}},
	{ID: "cursor", DisplayName: "Cursor", Dir: ".cursor", Aliases: []string{}},
}

// All returns every built-in target in stable order.
func All() []Target {
	out := make([]Target, len(builtins))
	copy(out, builtins)
	return out
}

// IDs returns the canonical IDs of every built-in target.
func IDs() []string {
	out := make([]string, len(builtins))
	for i, t := range builtins {
		out[i] = t.ID
	}
	return out
}

// Lookup resolves an ID or alias (case-insensitive) to a Target.
// Returns an error when the value does not match any known target.
func Lookup(idOrAlias string) (Target, error) {
	needle := strings.ToLower(strings.TrimSpace(idOrAlias))
	if needle == "" {
		return Target{}, fmt.Errorf("target: empty identifier")
	}
	for _, t := range builtins {
		if t.ID == needle {
			return t, nil
		}
		for _, a := range t.Aliases {
			if a == needle {
				return t, nil
			}
		}
	}
	return Target{}, fmt.Errorf("target: unknown target %q (valid: %s)", idOrAlias, strings.Join(IDs(), ", "))
}

// MustLookup is Lookup that panics on error. Intended for static init only.
func MustLookup(id string) Target {
	t, err := Lookup(id)
	if err != nil {
		panic(err)
	}
	return t
}

// SkillsDir returns the absolute (or repo-relative) skills directory for a target.
func (t Target) SkillsDir(repoRoot string) string {
	return filepath.Join(repoRoot, t.Dir, SkillsSubdir)
}

// ManifestPath returns the location of skell.toml for a target.
func (t Target) ManifestPath(repoRoot string) string {
	return filepath.Join(repoRoot, t.Dir, "skell.toml")
}

// LockPath returns the location of skell.lock for a target.
func (t Target) LockPath(repoRoot string) string {
	return filepath.Join(repoRoot, t.Dir, "skell.lock")
}

// InstalledRelPath returns the repo-relative install path for a single skill,
// suitable for storage in the lock file.
func (t Target) InstalledRelPath(skillName string) string {
	return filepath.Join(t.Dir, SkillsSubdir, skillName)
}

// Detect inspects repoRoot and returns every built-in target that already has
// either a skell.toml manifest or a skills/ directory present.
// Order matches All().
func Detect(repoRoot string) []Target {
	var out []Target
	for _, t := range builtins {
		if hasManifest(t, repoRoot) || hasSkillsDir(t, repoRoot) {
			out = append(out, t)
		}
	}
	return out
}

// DetectPrimary returns the single best-matching target for a repo, or false
// when none can be detected. When multiple targets are present, the one that
// has a skell.toml takes priority; ties fall back to the All() ordering.
func DetectPrimary(repoRoot string) (Target, bool) {
	detected := Detect(repoRoot)
	if len(detected) == 0 {
		return Target{}, false
	}
	// Prefer a target with a manifest over one with only a skills folder.
	for _, t := range detected {
		if hasManifest(t, repoRoot) {
			return t, true
		}
	}
	return detected[0], true
}

func hasManifest(t Target, repoRoot string) bool {
	_, err := os.Stat(t.ManifestPath(repoRoot))
	return err == nil
}

func hasSkillsDir(t Target, repoRoot string) bool {
	info, err := os.Stat(t.SkillsDir(repoRoot))
	return err == nil && info.IsDir()
}

// SortByID sorts targets in-place by canonical ID.
func SortByID(ts []Target) {
	sort.Slice(ts, func(i, j int) bool { return ts[i].ID < ts[j].ID })
}
