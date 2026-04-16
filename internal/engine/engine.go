// Package engine is the comparison and action layer.
// It coordinates registry, scanner, manifest, lockfile, and hasher to implement
// every skell command.
package engine

import (
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/aminmesbahi/skell/internal/registry"
)

// Engine wires together all internal subsystems.
type Engine struct {
	registry *registry.Adapter
}

// New creates a ready-to-use Engine.
func New(cacheRoot string) *Engine {
	// TODO: implement
	panic("not implemented")
}

// List returns all installed skills for the given repository root.
func (e *Engine) List(repoRoot string) ([]model.InstalledSkill, error) {
	// TODO: implement
	_ = e.registry
	panic("not implemented")
}

// ListRegistry returns all skills available in configured registries.
func (e *Engine) ListRegistry(m *manifest.Manifest) ([]model.RegistrySkill, error) {
	// TODO: implement
	panic("not implemented")
}

// Status returns the comparison between registry and local state for a repository.
func (e *Engine) Status(repoRoot string) ([]model.StatusEntry, error) {
	// TODO: implement
	panic("not implemented")
}

// Info returns the full detail for a single named skill.
func (e *Engine) Info(repoRoot, skillName, source string) (*model.StatusEntry, error) {
	// TODO: implement
	panic("not implemented")
}

// Install copies a skill from the registry into the target repository.
func (e *Engine) Install(repoRoot, skillName, registryAlias string, dryRun bool) error {
	// TODO: implement
	panic("not implemented")
}

// Upgrade updates one or all skills in a repository to the latest registry version.
func (e *Engine) Upgrade(repoRoot, skillName string, force, dryRun bool) error {
	// TODO: implement
	panic("not implemented")
}

// Remove deletes a skill from the target repository.
func (e *Engine) Remove(repoRoot, skillName string, dryRun bool) error {
	// TODO: implement
	panic("not implemented")
}

// Pin marks an installed skill as pinned in skell.toml and skell.lock.
func (e *Engine) Pin(repoRoot, skillName, version string) error {
	// TODO: implement
	panic("not implemented")
}

// Unpin removes the pinned flag from a skill.
func (e *Engine) Unpin(repoRoot, skillName string) error {
	// TODO: implement
	panic("not implemented")
}

// Sync applies skell.toml to the repository: installs missing, removes unlisted.
func (e *Engine) Sync(repoRoot string, checkOnly, dryRun bool) error {
	// TODO: implement
	panic("not implemented")
}

// Init creates a skell.toml from currently installed skills.
func (e *Engine) Init(repoRoot string) error {
	// TODO: implement
	panic("not implemented")
}

// Search queries configured registries for skills matching the criteria.
func (e *Engine) Search(m *manifest.Manifest, query, tag, lifecycle, owner string) ([]model.RegistrySkill, error) {
	// TODO: implement
	panic("not implemented")
}

// Doctor runs all diagnostic checks on a repository.
func (e *Engine) Doctor(repoRoot string) ([]DiagnosticIssue, error) {
	// TODO: implement
	panic("not implemented")
}
