package skell

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallCmd_RequiresSkillNameArg(t *testing.T) {
	_, err := executeCmd(t, "install")
	assert.Error(t, err)
}

func TestInstallCmd_RegistryFlagParsed(t *testing.T) {
	// With no manifest the engine returns an error — this tests that the flag
	// is accepted and wired through (i.e. no flag-parsing error, only engine error).
	repo := t.TempDir()
	_, err := executeCmd(t, "install", "pdf-processing", "--repo", repo, "--registry", "public")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no manifest found", "error should come from engine, not flag parsing")
}

func TestInstallCmd_DryRunFlagAccepted(t *testing.T) {
	repo := t.TempDir()
	_, err := executeCmd(t, "install", "pdf-processing", "--repo", repo, "--dry-run")
	// No manifest → engine error, but the flag itself must not produce a parse error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no manifest found")
}

func TestInstallCmd_NoManifest_ReturnsError(t *testing.T) {
	repo := t.TempDir()
	_, err := executeCmd(t, "install", "pdf-processing", "--repo", repo)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no manifest found")
}

func TestInstallCmd_RegistryURLFlag_Parsed(t *testing.T) {
	repo := t.TempDir()
	_, err := executeCmd(t, "install", "pdf-processing",
		"--repo", repo,
		"--registry", "public",
		"--registry-url", "https://github.com/owner/skills",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no manifest found")
}
