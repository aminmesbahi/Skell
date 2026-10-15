package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/target"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallTo_NewTargetInheritsSourcesAndFailureLeavesNothing(t *testing.T) {
	f := newGitRegistryFixture(t)
	require.NoError(t, f.eng.InstallTo(f.repo, "demo", "demo", "", "", false))

	// A failed install for a brand-new agent must not leave a manifest behind.
	err := f.eng.InstallTo(f.repo, "nope", "", "", "copilot", false)
	require.Error(t, err)
	assert.NoFileExists(t, target.MustLookup("copilot").ManifestPath(f.repo))

	// A successful one starts from the project's existing sources.
	require.NoError(t, f.eng.InstallTo(f.repo, "demo", "", "", "copilot", false))
	m, err := manifest.Read(target.MustLookup("copilot").ManifestPath(f.repo))
	require.NoError(t, err)
	primary, err := manifest.Read(target.MustLookup("claude").ManifestPath(f.repo))
	require.NoError(t, err)
	require.NotEmpty(t, primary.Registries["demo"])
	assert.Equal(t, primary.Registries["demo"], m.Registries["demo"])
	assert.FileExists(t, filepath.Join(f.repo, ".github", "skills", "demo", "SKILL.md"))
}

func TestInstallTo_DryRunWithNewTargetWritesNothing(t *testing.T) {
	f := newGitRegistryFixture(t)
	require.NoError(t, f.eng.InstallTo(f.repo, "demo", "demo", "", "", false))
	require.NoError(t, f.eng.InstallTo(f.repo, "demo", "", "", "copilot", true))
	assert.NoFileExists(t, target.MustLookup("copilot").ManifestPath(f.repo))
	_, err := os.Stat(filepath.Join(f.repo, ".github"))
	assert.True(t, os.IsNotExist(err), "dry run must not create the agent folder")
}

func TestInstallTo_MirrorTargetGoesThroughPrimary(t *testing.T) {
	f := newGitRegistryFixture(t)
	_, err := f.eng.SetMirrors(f.repo, []string{"cursor"}, nil)
	require.NoError(t, err)

	// Installing "for Cursor" installs via Claude and mirrors it; no second manifest.
	require.NoError(t, f.eng.InstallTo(f.repo, "demo", "demo", "", "cursor", false))
	assert.FileExists(t, filepath.Join(f.repo, ".claude", "skills", "demo", "SKILL.md"))
	assert.FileExists(t, filepath.Join(f.repo, ".cursor", "skills", "demo", "SKILL.md"))
	assert.NoFileExists(t, target.MustLookup("cursor").ManifestPath(f.repo))

	// Asking again is a no-op, not an "already installed" error.
	require.NoError(t, f.eng.InstallTo(f.repo, "demo", "demo", "", "cursor", false))
}
