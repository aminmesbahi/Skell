package engine

import (
	"testing"

	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeSearchManifest(t *testing.T, repo string) *manifest.Manifest {
	t.Helper()
	makeManifestWithRegistry(t, repo, "default", "https://example.com/reg")
	m, err := manifest.Resolve(repo)
	require.NoError(t, err)
	return m
}

func TestSearch_ReturnsAllWhenNoFilter(t *testing.T) {
	repo := makeRepo(t)
	m := makeSearchManifest(t, repo)

	fp := &fakeProvider{listSkills: []model.RegistrySkill{
		{Name: "pdf-processing", Description: "PDF tool", Metadata: model.SkillMetadata{Version: "1.0.0", Lifecycle: model.LifecycleStable, Owner: "platform", Tags: "documents"}},
		{Name: "code-review", Description: "Review tool", Metadata: model.SkillMetadata{Version: "2.0.0", Lifecycle: model.LifecycleStable, Owner: "platform", Tags: "review"}},
	}}

	results, err := newWithProvider(fp).Search(m, "", "", "", "")
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestSearch_FiltersbyQuery(t *testing.T) {
	repo := makeRepo(t)
	m := makeSearchManifest(t, repo)

	fp := &fakeProvider{listSkills: []model.RegistrySkill{
		{Name: "pdf-processing", Description: "PDF tool"},
		{Name: "code-review", Description: "Code review"},
	}}

	results, err := newWithProvider(fp).Search(m, "pdf", "", "", "")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "pdf-processing", results[0].Name)
}

func TestSearch_FiltersByLifecycle(t *testing.T) {
	repo := makeRepo(t)
	m := makeSearchManifest(t, repo)

	fp := &fakeProvider{listSkills: []model.RegistrySkill{
		{Name: "stable-skill", Metadata: model.SkillMetadata{Lifecycle: model.LifecycleStable}},
		{Name: "draft-skill", Metadata: model.SkillMetadata{Lifecycle: model.LifecycleDraft}},
	}}

	results, err := newWithProvider(fp).Search(m, "", "", "stable", "")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "stable-skill", results[0].Name)
}

func TestSearch_FiltersByOwner(t *testing.T) {
	repo := makeRepo(t)
	m := makeSearchManifest(t, repo)

	fp := &fakeProvider{listSkills: []model.RegistrySkill{
		{Name: "platform-skill", Metadata: model.SkillMetadata{Owner: "platform-team"}},
		{Name: "other-skill", Metadata: model.SkillMetadata{Owner: "other-team"}},
	}}

	results, err := newWithProvider(fp).Search(m, "", "", "", "platform-team")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "platform-skill", results[0].Name)
}

func TestSearch_FiltersByTag(t *testing.T) {
	repo := makeRepo(t)
	m := makeSearchManifest(t, repo)

	fp := &fakeProvider{listSkills: []model.RegistrySkill{
		{Name: "doc-skill", Metadata: model.SkillMetadata{Tags: "documents, pdf"}},
		{Name: "review-skill", Metadata: model.SkillMetadata{Tags: "review, code"}},
	}}

	results, err := newWithProvider(fp).Search(m, "", "pdf", "", "")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "doc-skill", results[0].Name)
}

func TestSearch_NoRegistryConfigured_ReturnsEmpty(t *testing.T) {
	repo := makeRepo(t)
	makeManifestWithRegistry(t, repo, "default", "https://example.com/reg")
	m := &manifest.Manifest{Registries: map[string]string{}} // empty registries

	fp := &fakeProvider{}
	results, err := newWithProvider(fp).Search(m, "anything", "", "", "")
	require.NoError(t, err)
	assert.Empty(t, results)
}
