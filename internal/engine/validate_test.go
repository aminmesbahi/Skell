package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/model"
	"github.com/aminmesbahi/skell/internal/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstall_ValidationGate_RejectsInvalidSkill(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com")

	eng := newWithProvider(&fakeProvider{skill: &model.RegistrySkill{Name: "pdf"}})
	eng.SetRequireValidation(true)

	err := eng.Install(repo, "pdf", "default", "", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed validation")

	// The invalid skill must not be left on disk.
	_, statErr := os.Stat(filepath.Join(repo, ".claude", "skills", "pdf"))
	assert.True(t, os.IsNotExist(statErr), "rejected skill should not be installed")
}

func TestInstall_NoGate_AllowsUnvalidatedSkill(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com")

	eng := newWithProvider(&fakeProvider{skill: &model.RegistrySkill{Name: "pdf"}})
	// requireValidation defaults to false for test engines.
	require.NoError(t, eng.Install(repo, "pdf", "default", "", false))
	_, statErr := os.Stat(filepath.Join(repo, ".claude", "skills", "pdf"))
	assert.NoError(t, statErr)
}

func TestValidateSkill_ReportsInstalledSkill(t *testing.T) {
	repo := makeRepo(t)
	makeInstalledSkill(t, repo, "pdf", "---\nname: pdf\n---\n") // missing description
	res, err := newWithProvider(nil).ValidateSkill(context.Background(), repo, "pdf", validator.Options{})
	require.NoError(t, err)
	assert.True(t, res.HasErrors())
}

func TestValidateSkill_RejectsTraversalName(t *testing.T) {
	repo := makeRepo(t)
	_, err := newWithProvider(nil).ValidateSkill(context.Background(), repo, "../escape", validator.Options{})
	require.Error(t, err)
}

func TestValidateAll_ReturnsResultPerSkill(t *testing.T) {
	repo := makeRepo(t)
	makeInstalledSkill(t, repo, "alpha", validSkillMD("alpha"))
	makeInstalledSkill(t, repo, "beta", "---\nname: beta\n---\n")

	results, err := newWithProvider(nil).ValidateAll(context.Background(), repo, validator.Options{})
	require.NoError(t, err)
	require.Len(t, results, 2)
	// Stable, name-sorted order.
	assert.Equal(t, "alpha", results[0].Name)
	assert.Equal(t, "beta", results[1].Name)
}
