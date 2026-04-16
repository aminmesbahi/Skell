// Package registry fetches, caches, and indexes skills from remote git registries.
package registry

import "github.com/aminmesbahi/skell/internal/model"

// Registry holds connection details for a single remote skill registry.
type Registry struct {
	Alias string
	URL   string
}

// Adapter provides access to one or more remote skill registries via a local cache.
type Adapter struct {
	cacheRoot string
}

// NewAdapter creates an Adapter rooted at the given cache directory.
func NewAdapter(cacheRoot string) *Adapter {
	// TODO: implement
	panic("not implemented")
}

// Fetch performs a git fetch on the cached clone of the given registry,
// cloning it first if the cache is empty.
func (a *Adapter) Fetch(reg Registry) error {
	// TODO: implement
	_ = a.cacheRoot
	panic("not implemented")
}

// ListSkills returns all skills available in the given registry.
func (a *Adapter) ListSkills(reg Registry) ([]model.RegistrySkill, error) {
	// TODO: implement
	panic("not implemented")
}

// GetSkill returns the metadata for a single skill from the registry.
func (a *Adapter) GetSkill(reg Registry, name string) (*model.RegistrySkill, error) {
	// TODO: implement
	panic("not implemented")
}

// CopySkillTo copies the skill directory from the cache into the destination path.
func (a *Adapter) CopySkillTo(reg Registry, name, version, destPath string) error {
	// TODO: implement
	panic("not implemented")
}

// CacheStatus returns a human-readable summary of what is cached locally.
func (a *Adapter) CacheStatus() (string, error) {
	// TODO: implement
	panic("not implemented")
}

// CacheClear removes all cached registry data.
func (a *Adapter) CacheClear() error {
	// TODO: implement
	panic("not implemented")
}
