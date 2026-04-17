// Package registry fetches, caches, and indexes skills from remote git registries.
package registry

import (
	"errors"

	"github.com/aminmesbahi/skell/internal/model"
)

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
	return &Adapter{cacheRoot: cacheRoot}
}

// Fetch performs a git fetch on the cached clone of the given registry,
// cloning it first if the cache is empty.
func (a *Adapter) Fetch(reg Registry) error {
	// TODO: implement git fetch/clone
	_ = reg
	return errors.New("registry: git fetch not yet implemented")
}

// ListSkills returns all skills available in the given registry.
func (a *Adapter) ListSkills(reg Registry) ([]model.RegistrySkill, error) {
	// TODO: implement
	return nil, errors.New("registry: list skills not yet implemented")
}

// GetSkill returns the metadata for a single skill from the registry.
func (a *Adapter) GetSkill(reg Registry, name string) (*model.RegistrySkill, error) {
	// TODO: implement
	return nil, errors.New("registry: get skill not yet implemented")
}

// CopySkillTo copies the skill directory from the cache into the destination path.
func (a *Adapter) CopySkillTo(reg Registry, name, version, destPath string) error {
	// TODO: implement
	return errors.New("registry: copy skill not yet implemented")
}

// CacheStatus returns a human-readable summary of what is cached locally.
func (a *Adapter) CacheStatus() (string, error) {
	// TODO: implement
	return "", errors.New("registry: cache status not yet implemented")
}

// CacheClear removes all cached registry data.
func (a *Adapter) CacheClear() error {
	// TODO: implement
	return errors.New("registry: cache clear not yet implemented")
}

// CacheRefresh fetches the latest from the given registry, updating the local cache.
func (a *Adapter) CacheRefresh(reg Registry) error {
	// TODO: implement git fetch/clone
	_ = reg
	return errors.New("registry: cache refresh not yet implemented")
}
