// Package scaffold creates starter files for skill authors: a new skill
// folder with a spec-compliant SKILL.md, or a whole skill-source repository
// with CI validation.
package scaffold

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// nameRx follows the Agent Skills spec: lowercase letters, digits and
// hyphens, not starting or ending with a hyphen, at most 64 characters.
var nameRx = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,62}[a-z0-9])?$`)

// ValidateName checks a skill name against the Agent Skills naming rules.
func ValidateName(name string) error {
	if !nameRx.MatchString(name) || strings.Contains(name, "--") {
		return fmt.Errorf("invalid skill name %q: use lowercase letters, digits and single hyphens (max 64 chars), e.g. \"release-notes\"", name)
	}
	return nil
}

// SkillOptions configures a new skill.
type SkillOptions struct {
	Name        string
	Description string
	Owner       string
	WithScripts bool
}

// Skill creates <parent>/<name>/SKILL.md (and optional scripts/ and
// references/ folders). It refuses to overwrite an existing folder.
func Skill(parent string, opts SkillOptions) (string, error) {
	if err := ValidateName(opts.Name); err != nil {
		return "", err
	}
	dir := filepath.Join(parent, opts.Name)
	if _, err := os.Stat(dir); err == nil {
		return "", fmt.Errorf("%s already exists", dir)
	}
	desc := strings.TrimSpace(opts.Description)
	if desc == "" {
		desc = "Describe what this skill does and when the agent should use it (e.g. \"Use when the user asks to ...\")."
	}
	owner := strings.TrimSpace(opts.Owner)
	if owner == "" {
		owner = "your-team"
	}
	title := strings.ReplaceAll(opts.Name, "-", " ")
	if title != "" {
		title = strings.ToUpper(title[:1]) + title[1:]
	}

	files := map[string]string{
		"SKILL.md": fmt.Sprintf(`---
name: %s
description: %s
metadata:
  version: 0.1.0
  owner: %s
  lifecycle: draft
  tags: ""
---

# %s

## When to use

- Describe the situations where this skill applies.

## Instructions

1. Step-by-step guidance the agent should follow.
2. Keep instructions specific and imperative.

## Examples

Show a short input → output example.
`, opts.Name, yamlQuote(desc), yamlQuote(owner), title),
	}
	if opts.WithScripts {
		files["scripts/README.md"] = "Put helper scripts the skill can run here, and reference them from SKILL.md.\n"
		files["references/README.md"] = "Put longer reference material here; link to it from SKILL.md so it is loaded only when needed.\n"
	}
	if err := writeFiles(dir, files); err != nil {
		return "", err
	}
	return dir, nil
}

// RegistryOptions configures a new skill-source repository.
type RegistryOptions struct {
	Name string // used in the README title
	// ActionRef is the Skell GitHub Action reference for the CI workflow.
	ActionRef string
}

// Registry creates a skill-source repository layout in dir: README, an
// example skill under skills/, and a GitHub Actions workflow that validates
// every skill on pull requests. dir must be empty or not exist.
func Registry(dir string, opts RegistryOptions) error {
	if entries, err := os.ReadDir(dir); err == nil && len(entries) > 0 {
		return errors.New(dir + " is not empty")
	}
	name := opts.Name
	if name == "" {
		name = filepath.Base(dir)
	}
	ref := opts.ActionRef
	if ref == "" {
		ref = "aminmesbahi/skell@main"
	}
	files := map[string]string{
		"README.md": fmt.Sprintf(`# %s

A collection of [Agent Skills](https://agentskills.io) (SKILL.md) managed with
[Skell](https://github.com/aminmesbahi/skell).

## Use these skills

`+"```sh"+`
skell add <owner>/<this-repo>
skell search
skell install <skill-name>
`+"```"+`

## Add a skill

`+"```sh"+`
skell new my-skill --path skills
skell validate --path skills/my-skill
`+"```"+`

Every pull request is validated by CI (see .github/workflows/validate-skills.yml).
`, name),
		".github/workflows/validate-skills.yml": fmt.Sprintf(`name: Validate skills

on:
  pull_request:
  push:
    branches: [main]

permissions:
  contents: read

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      # Pin to a released tag (e.g. @v0.2.0) for reproducible CI.
      - uses: %s
        with:
          command: validate --path skills --strict
`, ref),
		".gitignore": ".DS_Store\n",
	}
	if err := writeFiles(dir, files); err != nil {
		return err
	}
	_, err := Skill(filepath.Join(dir, "skills"), SkillOptions{
		Name:        "example-skill",
		Description: "Example skill. Replace with a real description of what it does and when to use it.",
	})
	return err
}

func writeFiles(root string, files map[string]string) error {
	for rel, content := range files {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil { //nolint:gosec
			return err
		}
	}
	return nil
}

// yamlQuote double-quotes a scalar when it contains YAML-significant
// characters, keeping simple values readable.
func yamlQuote(s string) string {
	if s == "" || strings.ContainsAny(s, ":#[]{}&*!|>'\"%@`,") || strings.ContainsAny(s[:1], "-?") {
		return `"` + strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `"`, `\"`) + `"`
	}
	return s
}
