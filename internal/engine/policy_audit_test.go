package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aminmesbahi/skell/internal/audit"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/aminmesbahi/skell/internal/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeBlockingPolicy returns a policy that blocks every registry URL.
func makeBlockingPolicy(allowed ...string) *policy.Config {
	return &policy.Config{
		AllowedRegistries: allowed,
		BlockUnlisted:     true,
	}
}

func TestInstall_PolicyBlocked_ReturnsError(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://blocked.example.com/registry")

	provider := &fakeProvider{
		skill: &model.RegistrySkill{Name: "my-skill", Metadata: model.SkillMetadata{Version: "1.0.0"}},
	}
	eng := newWithProvider(provider)
	eng.pol = makeBlockingPolicy("https://allowed.example.com")

	err := eng.Install(repo, "my-skill", "default", "", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "policy")
}

func TestInstall_PolicyAllows_Succeeds(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://allowed.example.com/registry")

	provider := &fakeProvider{
		skill: &model.RegistrySkill{Name: "my-skill", Metadata: model.SkillMetadata{Version: "1.0.0"}},
	}
	eng := newWithProvider(provider)
	eng.pol = makeBlockingPolicy("https://allowed.example.com/registry")

	require.NoError(t, eng.Install(repo, "my-skill", "default", "", false))
}

func TestInstall_WritesAuditLog(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com/registry")

	provider := &fakeProvider{
		skill: &model.RegistrySkill{Name: "pdf-skill", Metadata: model.SkillMetadata{Version: "2.0.0"}},
	}
	eng := newWithProvider(provider)

	// Override the logger to write to a temp file.
	logPath := filepath.Join(t.TempDir(), "audit.log")
	eng.logger = audit.NewLogger(logPath)

	require.NoError(t, eng.Install(repo, "pdf-skill", "default", "", false))

	data, err := os.ReadFile(logPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"action":"install"`)
	assert.Contains(t, string(data), `"skill":"pdf-skill"`)
}

func TestInstall_DryRun_DoesNotWriteAuditLog(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com/registry")

	provider := &fakeProvider{
		skill: &model.RegistrySkill{Name: "pdf-skill", Metadata: model.SkillMetadata{Version: "2.0.0"}},
	}
	eng := newWithProvider(provider)
	logPath := filepath.Join(t.TempDir(), "audit.log")
	eng.logger = audit.NewLogger(logPath)

	require.NoError(t, eng.Install(repo, "pdf-skill", "default", "", true))

	// Audit log must not exist on dry-run.
	_, err := os.Stat(logPath)
	assert.True(t, os.IsNotExist(err), "audit log should not be written on dry-run")
}

func TestUpgrade_PolicyBlocked_SkillIsSkipped(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://blocked.example.com/registry")
	// Use makePinnableSkill to create manifest + lock with the skill.
	makePinnableSkill(t, repo, "my-skill", "1.0.0")
	// Overwrite the manifest with the blocked registry URL.
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	m, err := manifest.Read(manifest.LocalPath(repo))
	require.NoError(t, err)
	m.Registries = map[string]string{"default": "https://blocked.example.com/registry"}
	require.NoError(t, manifest.Write(manifest.LocalPath(repo), m))

	provider := &fakeProvider{
		skill: &model.RegistrySkill{Name: "my-skill", Metadata: model.SkillMetadata{Version: "2.0.0"}},
	}
	eng := newWithProvider(provider)
	eng.pol = makeBlockingPolicy("https://allowed.example.com")

	report, err := eng.Upgrade(repo, "", false, false)
	require.NoError(t, err)
	require.Len(t, report.Skipped, 1)
	assert.Contains(t, report.Skipped[0], "blocked by policy")
}

func TestRemove_WritesAuditLog(t *testing.T) {
	repo := makeRepo(t)
	// Create the skill directory (needed for Remove to succeed).
	makeInstalledSkill(t, repo, "old-skill", "---\nname: old-skill\n---\n")
	// Create manifest + lock so Remove can update them.
	makePinnableSkill(t, repo, "old-skill", "1.0.0")

	eng := newWithProvider(nil)
	logPath := filepath.Join(t.TempDir(), "audit.log")
	eng.logger = audit.NewLogger(logPath)

	require.NoError(t, eng.Remove(repo, "old-skill", false))

	data, err := os.ReadFile(logPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"action":"remove"`)
	assert.Contains(t, string(data), `"skill":"old-skill"`)
}

func TestPin_WritesAuditLog(t *testing.T) {
	repo := makeRepo(t)
	makePinnableSkill(t, repo, "code-review", "2.0.0")

	eng := newWithProvider(nil)
	logPath := filepath.Join(t.TempDir(), "audit.log")
	eng.logger = audit.NewLogger(logPath)

	require.NoError(t, eng.Pin(repo, "code-review", ""))

	data, err := os.ReadFile(logPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"action":"pin"`)
}

func TestUnpin_WritesAuditLog(t *testing.T) {
	repo := makeRepo(t)
	makePinnableSkill(t, repo, "code-review", "2.0.0")

	eng := newWithProvider(nil)
	require.NoError(t, eng.Pin(repo, "code-review", ""))

	logPath := filepath.Join(t.TempDir(), "audit.log")
	eng.logger = audit.NewLogger(logPath)

	require.NoError(t, eng.Unpin(repo, "code-review"))

	data, err := os.ReadFile(logPath)
	require.NoError(t, err)
	assert.Contains(t, strings.TrimSpace(string(data)), `"action":"unpin"`)
}

