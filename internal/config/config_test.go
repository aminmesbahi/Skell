package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGlobalSources_ReadsConfig(t *testing.T) {
	home := t.TempDir()
	setHomeEnv(t, home)
	setArgv0(t, "skell")

	configPath := filepath.Join(home, ".skell", "config.toml")
	require.NoError(t, os.MkdirAll(filepath.Dir(configPath), 0755))
	require.NoError(t, os.WriteFile(configPath, []byte("[sources]\ndefault = \"https://example.com/skills\"\nlocal = \"C:/skills\"\n"), 0600))

	sources, err := GlobalSources()
	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"default": "https://example.com/skills",
		"local":   "C:/skills",
	}, sources)
}

func TestGlobalSources_ReturnsEmptyWhenMissing(t *testing.T) {
	home := t.TempDir()
	setHomeEnv(t, home)
	setArgv0(t, "skell")

	sources, err := GlobalSources()
	require.NoError(t, err)
	assert.Empty(t, sources)
}

func TestGlobalSources_SkipsTestBinary(t *testing.T) {
	setArgv0(t, "unit.test")

	sources, err := GlobalSources()
	require.NoError(t, err)
	assert.Empty(t, sources)
}

func TestGlobalSources_InvalidTOMLReturnsError(t *testing.T) {
	home := t.TempDir()
	setHomeEnv(t, home)
	setArgv0(t, "skell")

	configPath := filepath.Join(home, ".skell", "config.toml")
	require.NoError(t, os.MkdirAll(filepath.Dir(configPath), 0755))
	require.NoError(t, os.WriteFile(configPath, []byte(":::bad toml:::"), 0600))

	_, err := GlobalSources()
	assert.Error(t, err)
}

func TestPath_UsesHomeDir(t *testing.T) {
	home := t.TempDir()
	setHomeEnv(t, home)

	path, err := Path()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, ".skell", "config.toml"), path)
}

func setHomeEnv(t *testing.T, home string) {
	t.Helper()
	oldHome := os.Getenv("HOME")
	oldUserProfile := os.Getenv("USERPROFILE")
	require.NoError(t, os.Setenv("HOME", home))
	require.NoError(t, os.Setenv("USERPROFILE", home))
	t.Cleanup(func() {
		_ = os.Setenv("HOME", oldHome)
		_ = os.Setenv("USERPROFILE", oldUserProfile)
	})
}

func setArgv0(t *testing.T, argv0 string) {
	t.Helper()
	old := os.Args
	os.Args = []string{argv0}
	t.Cleanup(func() { os.Args = old })
}