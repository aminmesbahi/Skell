package skell

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeRepoWithManifestCmd(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	require.NoError(t, manifest.Write(manifest.LocalPath(repo), &manifest.Manifest{
		Registries: map[string]string{},
		Skills:     map[string]manifest.SkillEntry{},
	}))
	return repo
}

func TestAddCmd_RequiresURLArg(t *testing.T) {
	_, err := executeCmd(t, "add")
	assert.Error(t, err)
}

func TestAddCmd_RejectsExtraPositionalArg(t *testing.T) {
	_, err := executeCmd(t, "add", "https://github.com/o/r", "extra")
	assert.Error(t, err)
}

func TestAddCmd_InvalidURL_ReturnsError(t *testing.T) {
	repo := makeRepoWithManifestCmd(t)
	_, err := executeCmd(t, "add", "not-a-url", "--repo", repo)
	require.Error(t, err)
}

func TestAddCmd_NoManifest_ReturnsError(t *testing.T) {
	repo := t.TempDir()
	_, err := executeCmd(t, "add", "https://github.com/owner/repo/tree/main/skills", "--repo", repo)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no manifest found")
}

func TestAddCmd_DryRunFlag_Accepted(t *testing.T) {
	repo := makeRepoWithManifestCmd(t)
	out, err := executeCmd(t, "add",
		"https://github.com/owner/myrepo/tree/main/skills",
		"--repo", repo,
		"--dry-run",
	)
	require.NoError(t, err)
	assert.Contains(t, out, "dry-run")
}

func TestAddCmd_JSONFlag_ValidJSON(t *testing.T) {
	repo := makeRepoWithManifestCmd(t)
	out, err := executeCmd(t, "add",
		"https://github.com/owner/myrepo/tree/main/skills",
		"--repo", repo,
		"--json",
	)
	require.NoError(t, err)

	var results []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(out), &results))
	require.Len(t, results, 1)
	assert.Equal(t, "myrepo", results[0]["alias"])
	assert.Equal(t, true, results[0]["registered"])
}

func TestAddCmd_RegistryRoot_AddsToManifest(t *testing.T) {
	repo := makeRepoWithManifestCmd(t)
	out, err := executeCmd(t, "add",
		"https://github.com/owner/awesome-skills/tree/main/skills",
		"--repo", repo,
	)
	require.NoError(t, err)
	assert.Contains(t, out, "registered registry")
	assert.Contains(t, out, "awesome-skills")

	m, err := manifest.Read(manifest.LocalPath(repo))
	require.NoError(t, err)
	assert.Equal(t, "https://github.com/owner/awesome-skills", m.Registries["awesome-skills"])
}

func TestAddCmd_RegistryRoot_AlreadyRegistered_NoError(t *testing.T) {
	repo := t.TempDir()
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	require.NoError(t, manifest.Write(manifest.LocalPath(repo), &manifest.Manifest{
		Registries: map[string]string{"awesome-skills": "https://github.com/owner/awesome-skills"},
		Skills:     map[string]manifest.SkillEntry{},
	}))

	_, err := executeCmd(t, "add",
		"https://github.com/owner/awesome-skills/tree/main/skills",
		"--repo", repo,
	)
	require.NoError(t, err) // already registered is not an error
}
