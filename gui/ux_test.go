package main

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidationOutputPreservesAgentIdentity(t *testing.T) {
	results, err := parseValidationOutput(`[{"name":"same-name","target":"cursor","result":{"errors":0,"warnings":1,"findings":[]}}]`)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "cursor", results[0].Target)
	require.Equal(t, 1, results[0].Warnings)
}

func TestFolderAndValidationRejectMissingPaths(t *testing.T) {
	app := NewApp()
	path := filepath.Join(t.TempDir(), "missing")
	require.Error(t, app.OpenSkillFolder(path))
	_, err := app.ValidateDirectory(path)
	require.Error(t, err)
}
