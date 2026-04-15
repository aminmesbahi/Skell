// Package model defines the shared domain types used across all internal packages.
package model

// Lifecycle represents the maturity state of a skill in the registry.
type Lifecycle string

const (
	LifecycleDraft        Lifecycle = "draft"
	LifecycleExperimental Lifecycle = "experimental"
	LifecycleStable       Lifecycle = "stable"
	LifecycleDeprecated   Lifecycle = "deprecated"
	LifecycleArchived     Lifecycle = "archived"
)

// SkillStatus represents the comparison result between a registry skill and a local install.
type SkillStatus string

const (
	StatusUpToDate        SkillStatus = "up-to-date"
	StatusOutdated        SkillStatus = "outdated"
	StatusPinned          SkillStatus = "pinned"
	StatusDeprecated      SkillStatus = "deprecated"
	StatusArchived        SkillStatus = "archived"
	StatusLocallyModified SkillStatus = "locally-modified"
	StatusUnknown         SkillStatus = "unknown"
	StatusMissingMetadata SkillStatus = "missing-metadata"
	StatusUnversioned     SkillStatus = "unversioned"
)

// SkillMetadata holds the Skell-specific fields from SKILL.md frontmatter.
type SkillMetadata struct {
	Version    string    `yaml:"version"`
	Owner      string    `yaml:"owner"`
	Lifecycle  Lifecycle `yaml:"lifecycle"`
	Scope      string    `yaml:"scope"`
	Tags       string    `yaml:"tags"`
	SourceRepo string    `yaml:"source_repo"`
}

// RegistrySkill is a skill as defined in a registry, parsed from SKILL.md frontmatter.
type RegistrySkill struct {
	Name        string
	Description string
	License     string
	Metadata    SkillMetadata
}

// InstalledSkill is the entry for a skill as recorded in skell.lock.
type InstalledSkill struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	Registry      string `json:"registry"`
	SourceRepo    string `json:"source_repo"`
	SourceRef     string `json:"source_ref"`
	InstalledPath string `json:"installed_path"`
	InstalledAt   string `json:"installed_at"`
	Pinned        bool   `json:"pinned"`
	ContentHash   string `json:"content_hash"`
}

// StatusEntry is the result of comparing a registry skill against a local install.
type StatusEntry struct {
	Name      string
	Installed string
	Latest    string
	Status    SkillStatus
}
