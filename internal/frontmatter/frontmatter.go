// Package frontmatter parses SKILL.md files according to the Agent Skills spec.
package frontmatter

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aminmesbahi/skell/internal/model"
	"gopkg.in/yaml.v3"
)

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
	return buildRegistrySkill(doc), nil
}

func readNormalizedContent(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(string(data), "\r\n", "\n"), nil
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
