// Package frontmatter parses SKILL.md files according to the Agent Skills spec.
package frontmatter

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aminmesbahi/skell/internal/model"
	"gopkg.in/yaml.v3"
)

// maxSkillMDBytes bounds how much of a SKILL.md file is ever read into
// memory. SKILL.md files are ordinary markdown with a small YAML frontmatter
// block and are never legitimately large; without a cap, a malicious or
// misbehaving registry could publish a multi-gigabyte SKILL.md and exhaust
// memory in any command that touches it (list/search/info/install).
const maxSkillMDBytes = 5 << 20 // 5MB

type skillDoc struct {
	Name        string              `yaml:"name"`
	Description string              `yaml:"description"`
	License     string              `yaml:"license"`
	Metadata    model.SkillMetadata `yaml:"metadata"`
	// Top-level fields also supported by the open standard
	Paths                  string `yaml:"paths"`
	DisableModelInvocation bool   `yaml:"disable_model_invocation"`
	Compatibility          string `yaml:"compatibility"`
}

// Parse reads a SKILL.md file and extracts the RegistrySkill metadata from its YAML frontmatter.
func Parse(path string) (*model.RegistrySkill, error) {
	content, err := readNormalizedContent(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(content, "\n")
	yamlContent, err := extractYAMLFrontmatter(lines)
	if err != nil {
		return nil, err
	}
	var doc skillDoc
	if err := yaml.Unmarshal([]byte(yamlContent), &doc); err != nil {
		return nil, fmt.Errorf("frontmatter: %w", err)
	}
	sanitizeDoc(&doc)
	return buildRegistrySkill(doc), nil
}

func readNormalizedContent(path string) (string, error) {
	f, err := os.Open(path) //nolint:gosec
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	data, err := io.ReadAll(io.LimitReader(f, maxSkillMDBytes+1))
	if err != nil {
		return "", err
	}
	if len(data) > maxSkillMDBytes {
		return "", fmt.Errorf("frontmatter: %s exceeds the %d byte limit for a SKILL.md file", path, maxSkillMDBytes)
	}
	return strings.ReplaceAll(string(data), "\r\n", "\n"), nil
}

// sanitizeDoc strips ASCII control characters — including the ESC byte that
// begins ANSI/terminal escape sequences — from every scalar field parsed out
// of a skill's frontmatter. This content comes straight from a third-party
// registry's SKILL.md and is later written verbatim to the terminal
// (skell list/search/info/status) and rendered in the GUI, so without this a
// malicious registry could embed newlines or escape sequences to spoof or
// corrupt that output.
func sanitizeDoc(doc *skillDoc) {
	doc.Name = sanitizeField(doc.Name)
	doc.Description = sanitizeField(doc.Description)
	doc.License = sanitizeField(doc.License)
	doc.Paths = sanitizeField(doc.Paths)
	doc.Compatibility = sanitizeField(doc.Compatibility)
	doc.Metadata.Version = sanitizeField(doc.Metadata.Version)
	doc.Metadata.Owner = sanitizeField(doc.Metadata.Owner)
	doc.Metadata.Lifecycle = model.Lifecycle(sanitizeField(string(doc.Metadata.Lifecycle)))
	doc.Metadata.Scope = sanitizeField(doc.Metadata.Scope)
	doc.Metadata.Tags = sanitizeField(doc.Metadata.Tags)
	doc.Metadata.SourceRepo = sanitizeField(doc.Metadata.SourceRepo)
	doc.Metadata.Paths = sanitizeField(doc.Metadata.Paths)
	doc.Metadata.Compatibility = sanitizeField(doc.Metadata.Compatibility)
	doc.Metadata.License = sanitizeField(doc.Metadata.License)
}

// sanitizeField neutralizes ASCII control characters (0x00-0x1F, including
// newline, carriage return and ESC) and DEL (0x7F) in a single frontmatter
// value by turning them into spaces and then collapsing whitespace. Turning
// them into spaces rather than deleting them keeps a legitimate multi-line
// YAML block-scalar description readable as one line instead of running its
// words together; every value here is displayed as a single-line field
// throughout the CLI and GUI regardless. Everything else — including any
// non-ASCII text — passes through unchanged.
func sanitizeField(s string) string {
	mapped := strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return ' '
		}
		return r
	}, s)
	return strings.Join(strings.Fields(mapped), " ")
}

func extractYAMLFrontmatter(lines []string) (string, error) {
	if len(lines) == 0 || lines[0] != "---" {
		return "", errors.New("frontmatter: missing opening delimiter")
	}
	closeIdx := findClosingDelimiter(lines)
	if closeIdx == -1 {
		return "", errors.New("frontmatter: missing closing delimiter")
	}
	return strings.Join(lines[1:closeIdx], "\n"), nil
}

func findClosingDelimiter(lines []string) int {
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			return i
		}
	}
	return -1
}

func buildRegistrySkill(doc skillDoc) *model.RegistrySkill {
	rs := &model.RegistrySkill{
		Name:        doc.Name,
		Description: doc.Description,
		License:     doc.License,
		Metadata:    doc.Metadata,
	}
	mergeTopLevelFields(rs, doc)
	return rs
}

func mergeTopLevelFields(rs *model.RegistrySkill, doc skillDoc) {
	if rs.Metadata.Paths == "" && doc.Paths != "" {
		rs.Metadata.Paths = doc.Paths
	}
	if !rs.Metadata.DisableModelInvocation && doc.DisableModelInvocation {
		rs.Metadata.DisableModelInvocation = true
	}
	if rs.Metadata.Compatibility == "" && doc.Compatibility != "" {
		rs.Metadata.Compatibility = doc.Compatibility
	}
	if rs.Metadata.License == "" && doc.License != "" {
		rs.Metadata.License = doc.License
	}
}

// ParseDir scans a directory for a SKILL.md and returns the parsed skill.
func ParseDir(skillDir string) (*model.RegistrySkill, error) {
	return Parse(filepath.Join(skillDir, "SKILL.md"))
}
