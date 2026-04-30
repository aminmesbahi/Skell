package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/aminmesbahi/skell/internal/target"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInstall_RespectsTargetLayout verifies that engine.Install writes skill
// files and the lock file under the directory dictated by the manifest's
// active target (e.g. .codex, .github, .cursor) rather than always .claude.
func TestInstall_RespectsTargetLayout(t *testing.T) {
	cases := []struct {
		targetID string
		dir      string
	}{
		{"codex", ".codex"},
		{"copilot", ".github"},
		{"cursor", ".cursor"},
	}

	for _, tc := range cases {
		t.Run(tc.targetID, func(t *testing.T) {
			repo := makeRepo(t)
			tg, err := target.Lookup(tc.targetID)
			require.NoError(t, err)

			// Init manifest in the chosen target's directory.
			require.NoError(t, os.MkdirAll(filepath.Join(repo, tg.Dir), 0o755))
			m := &manifest.Manifest{
				Target:     tc.targetID,
				Registries: map[string]string{"default": "https://example.com"},
				Skills:     map[string]manifest.SkillEntry{},
			}
			require.NoError(t, manifest.Write(manifest.LocalPathFor(repo, tg), m))

			provider := &fakeProvider{
				skill: &model.RegistrySkill{
					Name:     "pdf-processing",
					Metadata: model.SkillMetadata{Version: "1.0.0"},
				},
			}
			eng := newWithProvider(provider)

			require.NoError(t, eng.Install(repo, "pdf-processing", "default", "", false))

			// Skill file written into the target's skills/ folder.
			_, err = os.Stat(filepath.Join(repo, tg.Dir, "skills", "pdf-processing", "SKILL.md"))
			assert.NoError(t, err)

			// Lock file written into the same target dir.
			lockPath := lockfile.PathFor(repo, tg)
			lf, err := lockfile.Read(lockPath)
			require.NoError(t, err)
			require.Len(t, lf.Skills, 1)
			assert.Equal(t, filepath.Join(tg.Dir, "skills", "pdf-processing"), lf.Skills[0].InstalledPath)

			// Manifest auto-fills target ID when empty.
			m2, err := manifest.Read(manifest.LocalPathFor(repo, tg))
			require.NoError(t, err)
			assert.Equal(t, tc.targetID, m2.Target)
		})
	}
}

func TestResolveTarget_DetectsExistingFolders(t *testing.T) {
	repo := makeRepo(t)
	// No layout: default to claude.
	assert.Equal(t, "claude", ResolveTarget(repo).ID)

	// Pre-existing .cursor/skills folder takes precedence.
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".cursor", "skills"), 0o755))
	assert.Equal(t, "cursor", ResolveTarget(repo).ID)

	// A manifest in .codex wins over the cursor skills folder.
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".codex"), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(repo, ".codex", "skell.toml"),
		[]byte("target = \"codex\"\n[registries]\n[skills]\n"),
		0o600,
	))
	assert.Equal(t, "codex", ResolveTarget(repo).ID)
}
