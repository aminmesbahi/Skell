package scaffold

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkill_IsValidPerSpec(t *testing.T) {
	dir, err := Skill(t.TempDir(), SkillOptions{Name: "release-notes", Description: "Write release notes: use when releasing", WithScripts: true})
	require.NoError(t, err)
	res := validator.ValidateDir(context.Background(), dir, validator.Options{})
	assert.False(t, res.HasErrors(), "%+v", res.Findings)
	assert.DirExists(t, filepath.Join(dir, "scripts"))

	_, err = Skill(filepath.Dir(dir), SkillOptions{Name: "release-notes"})
	assert.Error(t, err, "must not overwrite")
}

func TestValidateName(t *testing.T) {
	for _, ok := range []string{"a", "pdf", "release-notes", "v2-api"} {
		assert.NoError(t, ValidateName(ok), ok)
	}
	for _, bad := range []string{"", "-a", "a-", "A", "a--b", "a_b", "../x"} {
		assert.Error(t, ValidateName(bad), bad)
	}
}

func TestRegistry_Layout(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "team-skills")
	require.NoError(t, Registry(dir, RegistryOptions{}))
	for _, f := range []string{"README.md", ".github/workflows/validate-skills.yml", "skills/example-skill/SKILL.md"} {
		assert.FileExists(t, filepath.Join(dir, filepath.FromSlash(f)))
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, "x"), nil, 0o644))
	assert.Error(t, Registry(dir, RegistryOptions{}), "non-empty dir")
}

func TestDefaultActionRef(t *testing.T) {
	if got := defaultActionRef("v0.13.0"); got != "aminmesbahi/skell@v0.13.0" {
		t.Errorf("release build: got %q", got)
	}
	for _, dev := range []string{"dev", "v0.13.1-0.20261008222456-739298d3008c+dirty", ""} {
		if got := defaultActionRef(dev); got != "aminmesbahi/skell@main" {
			t.Errorf("%q: got %q", dev, got)
		}
	}
}
