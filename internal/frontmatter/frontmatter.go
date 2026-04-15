// Package frontmatter parses SKILL.md files according to the Agent Skills spec.
package frontmatter

import "github.com/aminmesbahi/skell/internal/model"

// Parse reads a SKILL.md file and extracts the RegistrySkill metadata from its YAML frontmatter.
func Parse(path string) (*model.RegistrySkill, error) {
	// TODO: implement using github.com/adrg/frontmatter
	panic("not implemented")
}

// ParseDir scans a directory for a SKILL.md and returns the parsed skill.
func ParseDir(skillDir string) (*model.RegistrySkill, error) {
	// TODO: implement
	panic("not implemented")
}
