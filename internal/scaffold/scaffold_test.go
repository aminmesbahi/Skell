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
