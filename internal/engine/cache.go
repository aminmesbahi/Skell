package engine

import (
	"fmt"

	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/registry"
)

// CacheStatus returns a human-readable summary of the local registry cache.
func (e *Engine) CacheStatus() (string, error) {
	a := registry.NewAdapter(e.cacheRoot)
	return a.CacheStatus()
}

// CacheClear removes all locally cached registry data.
func (e *Engine) CacheClear() error {
	a := registry.NewAdapter(e.cacheRoot)
	return a.CacheClear()
}

// CacheRefresh fetches the latest from all registries configured in the manifest.
func (e *Engine) CacheRefresh(m *manifest.Manifest) error {
	a := registry.NewAdapter(e.cacheRoot)
	regs := e.effectiveRegistries(m)
	for _, alias := range sortedAliases(regs) {
		if e.pol.CheckRegistry(regs[alias]) != nil {
			continue
		}
		reg := registry.Registry{Alias: alias, URL: regs[alias]}
		if err := a.CacheRefresh(reg); err != nil {
			return fmt.Errorf("failed to refresh registry %q: %w", alias, err)
		}
	}
	return nil
}

// CacheRefreshAll refreshes every registry Skell knows about without requiring a
// repo manifest: the configured registries (global sources plus the optional
// manifest m, which may be nil) and any other already-cached clones on disk.
// This backs `skell cache refresh` as a global operation.
func (e *Engine) CacheRefreshAll(m *manifest.Manifest) error {
	a := registry.NewAdapter(e.cacheRoot)
	regs := e.effectiveRegistries(m)
	refreshed := make(map[string]bool, len(regs))
	for _, alias := range sortedAliases(regs) {
		refreshed[alias] = true // also keeps a policy-blocked clone out of RefreshCachedClones
		if e.pol.CheckRegistry(regs[alias]) != nil {
			continue
		}
		reg := registry.Registry{Alias: alias, URL: regs[alias]}
		if err := a.CacheRefresh(reg); err != nil {
			return fmt.Errorf("failed to refresh registry %q: %w", alias, err)
		}
	}
	// Refresh any other clones already in the cache (e.g. from a repo not
	// currently selected) so a global "refresh" updates everything cached.
	return a.RefreshCachedClones(refreshed)
}
