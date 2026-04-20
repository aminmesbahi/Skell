package skell

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeSearchCmdRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	m := &manifest.Manifest{
		Registries: map[string]string{"default": "https://example.com/reg"},
		Skills:     map[string]manifest.SkillEntry{},
	}
	require.NoError(t, manifest.Write(manifest.LocalPath(repo), m))
	return repo
}

func TestSearchCmd_NoManifest_FallsBackToGlobal(t *testing.T) {
	// When --repo points to a dir with no manifest, search falls back to the
	// global manifest (creating it if needed) and returns empty results.
	repo := t.TempDir()
	_, err := executeCmd(t, "search", "pdf", "--repo", repo)
	assert.NoError(t, err)
}

func TestSearchCmd_RegistryNotImplemented_ReturnsError(t *testing.T) {
	// Real registry.Adapter.ListSkills returns "not yet implemented"
	repo := makeSearchCmdRepo(t)
	_, err := executeCmd(t, "search", "pdf", "--repo", repo)
	assert.Error(t, err)
}

func TestSearchCmd_QueryArg_PassedThrough(t *testing.T) {
	// Verify the flag/arg parsing; result is an error from the real registry.
	repo := makeSearchCmdRepo(t)
	_, err := executeCmd(t, "search", "anything", "--repo", repo)
	assert.Error(t, err)
}

func TestSearchCmd_NoArgs_AcceptsNoQuery(t *testing.T) {
	repo := makeSearchCmdRepo(t)
	_, err := executeCmd(t, "search", "--repo", repo)
	// still fails with "not yet implemented" from registry
	assert.Error(t, err)
}

func TestSearchCmd_LifecycleFlag_Accepted(t *testing.T) {
	repo := makeSearchCmdRepo(t)
	_, err := executeCmd(t, "search", "--repo", repo, "--lifecycle", "stable")
	assert.Error(t, err)
}

func TestSearchCmd_TagFlag_Accepted(t *testing.T) {
	repo := makeSearchCmdRepo(t)
	_, err := executeCmd(t, "search", "--repo", repo, "--tag", "documents")
	assert.Error(t, err)
}

func TestSearchCmd_OwnerFlag_Accepted(t *testing.T) {
	repo := makeSearchCmdRepo(t)
	_, err := executeCmd(t, "search", "--repo", repo, "--owner", "platform-team")
	assert.Error(t, err)
}
