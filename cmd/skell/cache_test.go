package skell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCacheStatusCmd_ReturnsError(t *testing.T) {
	// cache status is not yet implemented in the registry adapter; it should
	// propagate the "not yet implemented" error rather than panic or silently succeed.
	_, err := executeCmd(t, "cache", "status")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not yet implemented")
}

func TestCacheClearCmd_ReturnsError(t *testing.T) {
	_, err := executeCmd(t, "cache", "clear")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not yet implemented")
}

func TestCacheRefreshCmd_NoManifest_ReturnsError(t *testing.T) {
	repo := t.TempDir()
	_, err := executeCmd(t, "cache", "refresh", "--repo", repo)
	assert.Error(t, err)
}
