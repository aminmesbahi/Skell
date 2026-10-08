package engine

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/aminmesbahi/skell/internal/registry"
)

// Review summarises what installing a skill would bring into a repository, so
// a user can decide before any file is written. Skills are instructions that
// run inside an AI agent; scripts and broad tool permissions deserve a look.
type Review struct {
	Skill        *model.RegistrySkill `json:"skill"`
	Registry     string               `json:"registry"`
	RegistryURL  string               `json:"registry_url"`
	Commit       string               `json:"commit,omitempty"`
	Files        []string             `json:"files"`
	Scripts      []string             `json:"scripts,omitempty"`
	AllowedTools []string             `json:"allowed_tools,omitempty"`
	Links        []string             `json:"links,omitempty"`
	TotalBytes   int64                `json:"total_bytes"`
	Warnings     []string             `json:"warnings,omitempty"`
}

// NeedsConfirmation is true when the skill ships executable code or asks for
// broad tool access.
func (r *Review) NeedsConfirmation() bool { return len(r.Warnings) > 0 }

var (
	scriptExts = map[string]bool{
		".sh": true, ".bash": true, ".zsh": true, ".ps1": true, ".psm1": true, ".bat": true, ".cmd": true,
		".py": true, ".js": true, ".mjs": true, ".cjs": true, ".ts": true, ".rb": true, ".pl": true,
		".exe": true, ".dll": true, ".so": true, ".dylib": true, ".jar": true,
	}
	linkRx = regexp.MustCompile(`https?://[^\s)>\]"'` + "`" + `]+`)
)

// ReviewSkill inspects a registry skill without installing it. Registry
// resolution mirrors install: registryAlias (default "default") must be
// configured, or registryURL must be supplied.
func (e *Engine) ReviewSkill(repoRoot, skillName, registryAlias, registryURL string) (*Review, error) {
	if err := ValidateSkillName(skillName); err != nil {
		return nil, err
	}
	m, _ := manifest.Resolve(repoRoot)
	if m == nil {
		m = &manifest.Manifest{}
	}
	alias, url, _, err := e.resolveInstallRegistry(m, skillName, registryAlias, registryURL)
	if err != nil {
		return nil, err
	}
	if err := e.pol.CheckRegistry(url); err != nil {
		return nil, err
	}
	reg := registry.Registry{Alias: alias, URL: url}
	rs, err := e.provider.GetSkill(reg, skillName)
	if err != nil {
		return nil, fmt.Errorf("could not fetch skill %q from registry %q: %w", skillName, alias, err)
	}
	rev := &Review{Skill: rs, Registry: alias, RegistryURL: url}
	rev.AllowedTools = strings.Fields(rs.AllowedTools)

	if src := e.skillSource(reg, skillName); src != nil {
		rev.Commit = src.Commit
		inspectSkillDir(src.Dir, rev)
	}
	rev.Warnings = append(rev.Warnings, toolWarnings(rev.AllowedTools)...)
	if len(rev.Scripts) > 0 {
		rev.Warnings = append(rev.Warnings, fmt.Sprintf("ships %d script/executable file(s) the agent may run", len(rev.Scripts)))
	}
	return rev, nil
}

func inspectSkillDir(dir string, rev *Review) {
	links := map[string]bool{}
	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(dir, p)
		rel = filepath.ToSlash(rel)
		rev.Files = append(rev.Files, rel)
		info, err := d.Info()
		if err != nil {
			return nil
		}
		rev.TotalBytes += info.Size()
		ext := strings.ToLower(filepath.Ext(p))
		if scriptExts[ext] || strings.HasPrefix(rel, "scripts/") || info.Mode()&0o111 != 0 {
			rev.Scripts = append(rev.Scripts, rel)
		}
		if (ext == ".md" || ext == ".txt") && info.Size() < 2<<20 {
			if data, err := os.ReadFile(p); err == nil { //nolint:gosec
				for _, l := range linkRx.FindAllString(string(data), -1) {
					links[strings.TrimRight(l, ".,;:")] = true
				}
			}
		}
		return nil
	})
	for l := range links {
		rev.Links = append(rev.Links, l)
	}
	sort.Strings(rev.Files)
	sort.Strings(rev.Scripts)
	sort.Strings(rev.Links)
}

// toolWarnings flags allowed-tools entries that grant broad power.
func toolWarnings(tools []string) []string {
	var out []string
	for _, t := range tools {
		lt := strings.ToLower(t)
		switch {
		case lt == "*" || lt == "all":
			out = append(out, "requests access to all tools")
		case lt == "bash" || lt == "bash(*)" || lt == "bash(*:*)" || lt == "shell" || lt == "powershell":
			out = append(out, fmt.Sprintf("pre-approves unrestricted shell access (%s)", t))
		case strings.HasPrefix(lt, "write") || strings.HasPrefix(lt, "edit"):
			if !strings.Contains(lt, "(") {
				out = append(out, fmt.Sprintf("pre-approves unrestricted file edits (%s)", t))
			}
		case strings.HasPrefix(lt, "webfetch") && !strings.Contains(lt, "("):
			out = append(out, fmt.Sprintf("pre-approves fetching any URL (%s)", t))
		}
	}
	return out
}

// ResolveAdd parses a 'skell add' argument and picks the registry alias that
// AddFromURLWith would use, without writing anything. Callers use it to
// review a skill before adding it.
func (e *Engine) ResolveAdd(repoRoot, rawURL, alias string) (ParsedSkillURL, error) {
	parsed, err := ParseSkillURL(rawURL)
	if err != nil {
		return ParsedSkillURL{}, err
	}
	m, err := manifest.Resolve(repoRoot)
	if err != nil {
		return ParsedSkillURL{}, fmt.Errorf("no manifest found in %s — run 'skell init' first: %w", repoRoot, err)
	}
	parsed.Alias, err = e.chooseAlias(m, parsed, alias)
	return parsed, err
}
