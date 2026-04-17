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
}

// Parse reads a SKILL.md file and extracts the RegistrySkill metadata from its YAML frontmatter.
func Parse(path string) (*model.RegistrySkill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// Normalise CRLF → LF so the parser is platform-agnostic (e.g. files
	// checked out by git on Windows with core.autocrlf=true).
	content := strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || lines[0] != "---" {
		return nil, errors.New("frontmatter: missing opening delimiter")
	}
	closeIdx := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			closeIdx = i
			break
		}
	}
	if closeIdx == -1 {
		return nil, errors.New("frontmatter: missing closing delimiter")
	}
	yamlContent := strings.Join(lines[1:closeIdx], "\n")
	var doc skillDoc
	if err := yaml.Unmarshal([]byte(yamlContent), &doc); err != nil {
		return nil, fmt.Errorf("frontmatter: %w", err)
	}
	return &model.RegistrySkill{
		Name:        doc.Name,
		Description: doc.Description,
		License:     doc.License,
		Metadata:    doc.Metadata,
	}, nil
}

// ParseDir scans a directory for a SKILL.md and returns the parsed skill.
func ParseDir(skillDir string) (*model.RegistrySkill, error) {
	return Parse(filepath.Join(skillDir, "SKILL.md"))
}
