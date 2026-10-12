package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const configWithPolicy = `[sources]
shared = "https://github.com/acme/skills"
extra = "https://github.com/acme/extra"

[policy]
allowed-registries = ["https://github.com/acme/skills"]
block-unlisted = true
require-validation = true
`

func writeSourcesConfig(t *testing.T, content string) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("SKELL_HOME", home)
	path := filepath.Join(home, "config.toml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func TestRemoveSkillSourceKeepsPolicy(t *testing.T) {
	path := writeSourcesConfig(t, configWithPolicy)
	app := NewApp()

	require.NoError(t, app.RemoveSkillSource("extra"))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "[policy]")
	assert.Contains(t, string(data), "block-unlisted = true")
	assert.Contains(t, string(data), "require-validation = true")
	assert.Contains(t, string(data), "https://github.com/acme/skills")

	sources, err := app.ListSkillSources()
	require.NoError(t, err)
	require.Len(t, sources, 1)
	assert.Equal(t, "shared", sources[0].Alias)
}

func TestAddSkillSourceKeepsPolicyAndUnknownKeys(t *testing.T) {
	path := writeSourcesConfig(t, configWithPolicy+"\n[future]\nflag = true\n")
	app := NewApp()

	require.NoError(t, app.AddSkillSource("local", "/tmp/skills"))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "[policy]")
	assert.Contains(t, string(data), "[future]")
	sources, err := app.ListSkillSources()
	require.NoError(t, err)
	assert.Len(t, sources, 3)
}

func TestRemoveLastSourceKeepsPolicy(t *testing.T) {
	path := writeSourcesConfig(t, "[sources]\nonly = \"https://github.com/acme/skills\"\n\n[policy]\nblock-unlisted = true\n")
	require.NoError(t, NewApp().RemoveSkillSource("only"))
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "block-unlisted = true")
}

func TestRemoveSkillSourceWithoutConfigIsNoop(t *testing.T) {
	t.Setenv("SKELL_HOME", t.TempDir())
	require.NoError(t, NewApp().RemoveSkillSource("nothing"))
}

func TestAddSkillSourceRefusesCorruptConfig(t *testing.T) {
	path := writeSourcesConfig(t, "this is = = not toml")
	err := NewApp().AddSkillSource("a", "https://github.com/acme/skills")
	require.Error(t, err)
	data, _ := os.ReadFile(path)
	assert.Equal(t, "this is = = not toml", string(data), "a corrupt config must not be overwritten")
}
