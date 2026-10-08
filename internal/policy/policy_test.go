package policy_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRead_ValidPolicy(t *testing.T) {
	dir := t.TempDir()
	content := `
[policy]
allowed-registries = [
  "https://github.com/mycompany/skills-registry"
]
block-unlisted = true
`
	path := filepath.Join(dir, "config.toml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))

	cfg, err := policy.Read(path)
	require.NoError(t, err)
	assert.True(t, cfg.BlockUnlisted)
	assert.Contains(t, cfg.AllowedRegistries, "https://github.com/mycompany/skills-registry")
}

func TestCheckRegistry_AllowedURL(t *testing.T) {
	cfg := &policy.Config{
		AllowedRegistries: []string{"https://github.com/mycompany/skills-registry"},
		BlockUnlisted:     true,
	}
	assert.NoError(t, cfg.CheckRegistry("https://github.com/mycompany/skills-registry"))
}

func TestCheckRegistry_BlockedURL(t *testing.T) {
	cfg := &policy.Config{
		AllowedRegistries: []string{"https://github.com/mycompany/skills-registry"},
		BlockUnlisted:     true,
	}
	assert.Error(t, cfg.CheckRegistry("https://github.com/untrusted/other-registry"))
}

func TestCheckRegistry_BlockUnlistedFalse_AlwaysAllows(t *testing.T) {
	cfg := &policy.Config{
		AllowedRegistries: []string{},
		BlockUnlisted:     false,
	}
	assert.NoError(t, cfg.CheckRegistry("https://any.registry.example.com"))
}

func TestCheckRegistry_NormalizesURLs(t *testing.T) {
	c := &policy.Config{BlockUnlisted: true, AllowedRegistries: []string{"https://GitHub.com/org/repo.git/"}}
	assert.NoError(t, c.CheckRegistry("https://github.com/org/repo"))
	assert.Error(t, c.CheckRegistry("https://github.com/org/other"))
}

func TestInvalid_RefusesEverything(t *testing.T) {
	c := policy.Invalid(errors.New("boom"))
	err := c.CheckRegistry("https://github.com/org/repo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}
