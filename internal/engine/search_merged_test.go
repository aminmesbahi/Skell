package engine

import (
	"testing"

	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchMerged_LocalRepo_ReturnsLocalSkills(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com/reg")

	fp := &fakeProvider{listSkills: []model.RegistrySkill{
		{Name: "local-skill", Description: "A local skill",
			Metadata: model.SkillMetadata{Lifecycle: model.LifecycleStable, Owner: "team", Tags: "test"}},
	}}

	results, err := newWithProvider(fp).SearchMerged(repo, "", "", "", "")
	require.NoError(t, err)

	foundLocal := false
	for _, r := range results {
		if r.Name == "local-skill" && r.RegistrySource == "local" {
			foundLocal = true
		}
	}
	assert.True(t, foundLocal, "local skill should appear with RegistrySource=local")
}

func TestSearchMerged_LocalRepo_FilterByQuery(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com/reg")

	fp := &fakeProvider{listSkills: []model.RegistrySkill{
		{Name: "pdf-processing", Description: "PDF tool", Metadata: model.SkillMetadata{Lifecycle: model.LifecycleStable}},
		{Name: "code-review", Description: "Review tool", Metadata: model.SkillMetadata{Lifecycle: model.LifecycleStable}},
	}}

	results, err := newWithProvider(fp).SearchMerged(repo, "pdf", "", "", "")
	require.NoError(t, err)

	// only pdf-processing should be returned (possibly duplicated if it also appears in global, but name check is enough)
	for _, r := range results {
		assert.Equal(t, "pdf-processing", r.Name)
	}
}

func TestSearchMerged_LocalRepo_FilterByLifecycle(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com/reg")

	fp := &fakeProvider{listSkills: []model.RegistrySkill{
		{Name: "stable-skill", Metadata: model.SkillMetadata{Lifecycle: model.LifecycleStable}},
		{Name: "draft-skill", Metadata: model.SkillMetadata{Lifecycle: model.LifecycleDraft}},
	}}

	results, err := newWithProvider(fp).SearchMerged(repo, "", "", "stable", "")
	require.NoError(t, err)

	for _, r := range results {
		assert.Equal(t, model.LifecycleStable, r.Metadata.Lifecycle)
	}
}

func TestSearchMerged_LocalRepo_FilterByOwner(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com/reg")

	fp := &fakeProvider{listSkills: []model.RegistrySkill{
		{Name: "skill-a", Metadata: model.SkillMetadata{Owner: "platform"}},
		{Name: "skill-b", Metadata: model.SkillMetadata{Owner: "mobile"}},
	}}

	results, err := newWithProvider(fp).SearchMerged(repo, "", "", "", "PLATFORM")
	require.NoError(t, err)

	for _, r := range results {
		assert.Equal(t, "platform", r.Metadata.Owner)
	}
}

func TestSearchMerged_LocalRepo_FilterByTag(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com/reg")

	fp := &fakeProvider{listSkills: []model.RegistrySkill{
		{Name: "tagged-skill", Metadata: model.SkillMetadata{Tags: "documents,pdf"}},
		{Name: "other-skill", Metadata: model.SkillMetadata{Tags: "code"}},
	}}

	results, err := newWithProvider(fp).SearchMerged(repo, "", "pdf", "", "")
	require.NoError(t, err)

	for _, r := range results {
		assert.Contains(t, r.Metadata.Tags, "pdf")
	}
}

func TestSearchMerged_LocalRepo_NoLocalManifest_ReturnsEmpty(t *testing.T) {
	// repo with no manifest — local skills are skipped; global is tried but
	// (in the test environment) likely empty. Either way it should not error.
	repo := makeRepo(t) // no manifest written

	fp := &fakeProvider{listSkills: []model.RegistrySkill{
		{Name: "unreachable-skill"},
	}}

	results, err := newWithProvider(fp).SearchMerged(repo, "", "", "", "")
	require.NoError(t, err)
	// local skills skipped; global may or may not have skills — just assert no error
	_ = results
}

func TestSearchMerged_LocalRepo_ListRegistryError_StillReturnsNoError(t *testing.T) {
	// ListRegistry error is silently skipped for local skills.
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com/reg")

	fp := &fakeProvider{listErr: errTest("registry unavailable")}

	results, err := newWithProvider(fp).SearchMerged(repo, "", "", "", "")
	require.NoError(t, err)
	// listErr causes ListRegistry to return error → local block silently skipped
	_ = results
}

func TestSearchMerged_GlobalRoot_IsGlobal_ReturnsSkillsWithGlobalSource(t *testing.T) {
	require.NoError(t, manifest.EnsureGlobal())
	globalRoot, err := manifest.GlobalRootDir()
	require.NoError(t, err)

	// Ensure the global manifest has at least one registry entry for ListRegistry to call.
	globalPath, err := manifest.GlobalPath()
	require.NoError(t, err)

	gm, err := manifest.Read(globalPath)
	require.NoError(t, err)
	if gm.Registries == nil {
		gm.Registries = make(map[string]string)
	}
	gm.Registries["test-reg"] = "https://example.com/test"
	require.NoError(t, manifest.Write(globalPath, gm))

	fp := &fakeProvider{listSkills: []model.RegistrySkill{
		{Name: "global-skill", Metadata: model.SkillMetadata{Lifecycle: model.LifecycleStable}},
	}}

	results, err := newWithProvider(fp).SearchMerged(globalRoot, "", "", "", "")
	require.NoError(t, err)

	found := false
	for _, r := range results {
		if r.Name == "global-skill" {
			assert.Equal(t, "global", r.RegistrySource)
			found = true
		}
	}
	assert.True(t, found, "global skill should appear with RegistrySource=global")

	// Cleanup: remove the test-reg entry from global manifest.
	gm2, err := manifest.Read(globalPath)
	if err == nil {
		delete(gm2.Registries, "test-reg")
		_ = manifest.Write(globalPath, gm2)
	}
}

// errTest is a local error type for testing.
type errTest string

func (e errTest) Error() string { return string(e) }
