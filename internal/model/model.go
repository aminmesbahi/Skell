// Package model defines the shared domain types used across all internal packages.
package model

// Lifecycle represents the maturity state of a skill in the registry.
type Lifecycle string

// Lifecycle stage constants for a skill in the registry.
const (
	LifecycleDraft        Lifecycle = "draft"
	LifecycleExperimental Lifecycle = "experimental"
	LifecycleStable       Lifecycle = "stable"
	LifecycleDeprecated   Lifecycle = "deprecated"
	LifecycleArchived     Lifecycle = "archived"
)

// SkillStatus represents the comparison result between a registry skill and a local install.
type SkillStatus string

// SkillStatus comparison result constants between registry and local install.
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
	Version    string    `yaml:"version"    json:"version"`
	Owner      string    `yaml:"owner"      json:"owner"`
	Lifecycle  Lifecycle `yaml:"lifecycle"  json:"lifecycle"`
	Scope      string    `yaml:"scope"      json:"scope"`
	Tags       string    `yaml:"tags"       json:"tags"`
	SourceRepo string    `yaml:"source_repo" json:"source_repo"`
}

// RegistrySkill is a skill as defined in a registry, parsed from SKILL.md frontmatter.
type RegistrySkill struct {
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	License        string        `json:"license"`
	Metadata       SkillMetadata `json:"metadata"`
	RegistryAlias  string        `json:"registry_alias,omitempty"`
	RegistryURL    string        `json:"registry_url,omitempty"`
	// RegistrySource indicates whether this skill comes from the global manifest
	// ("global") or the local repo manifest ("local"). Empty when not relevant.
	RegistrySource string        `json:"registry_source,omitempty"`
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
	Name      string      `json:"name"`
	Installed string      `json:"installed"`
	Latest    string      `json:"latest"`
	Status    SkillStatus `json:"status"`
}

// InfoResult holds the full detail for a single skill, combining local and registry data.
type InfoResult struct {
	Skill  *RegistrySkill  // metadata from SKILL.md frontmatter (local or registry)
	Lock   *InstalledSkill // entry from skell.lock, nil if not locked
	Status SkillStatus
}
