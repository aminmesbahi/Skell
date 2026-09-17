package skell

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeValidateRepo(t *testing.T, skills map[string]string) string {
	t.Helper()
	repo := t.TempDir()
	claudeDir := filepath.Join(repo, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	require.NoError(t, manifest.Write(manifest.LocalPath(repo), &manifest.Manifest{
		Registries: map[string]string{},
		Skills:     map[string]manifest.SkillEntry{},
	}))

	lf := &lockfile.LockFile{}
	for name, content := range skills {
		dir := filepath.Join(claudeDir, "skills", name)
		require.NoError(t, os.MkdirAll(dir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0600))
		lf.Skills = append(lf.Skills, model.InstalledSkill{Name: name, Registry: "default"})
	}
	require.NoError(t, lockfile.Write(lockfile.Path(repo), lf))
	return repo
}

func TestValidateCmd_ValidSkill_Passes(t *testing.T) {
	repo := makeValidateRepo(t, map[string]string{
		"good": "---\nname: good\ndescription: A clear skill. Use when testing the validate command.\n---\n\n# Good\n\nDo the thing.\n",
	})
	out, err := executeCmd(t, "validate", "--repo", repo)
	require.NoError(t, err)
	assert.Contains(t, out, "passed")
}

func TestValidateCmd_InvalidSkill_FailsNonZero(t *testing.T) {
	repo := makeValidateRepo(t, map[string]string{
		"bad": "---\nname: bad\n---\n", // missing description
	})
	out, err := executeCmd(t, "validate", "--repo", repo)
	require.Error(t, err)
	assert.Contains(t, out, "error")
}

func TestValidateCmd_JSON(t *testing.T) {
	repo := makeValidateRepo(t, map[string]string{
		"bad": "---\nname: bad\n---\n",
	})
	out, _ := executeCmd(t, "validate", "--repo", repo, "--json")
	assert.Contains(t, out, `"findings"`)
	assert.Contains(t, out, `"severity"`)
}

func TestValidateCmd_JSON_MultipleRepos_IsSingleValidJSONDocument(t *testing.T) {
	// Uses passing skills (not "bad") so the command exits 0: executeCmd
	// combines stdout+stderr into one buffer for tests, and on a non-nil
	// RunE error cobra appends its own "Error: ..." plus full usage text
	// after the JSON already written — a test-harness artifact (in real use
	// that lands on stderr, a separate stream from the JSON on stdout), not
	// something this test is meant to exercise.
	good := "---\nname: good\ndescription: A clear skill. Use when testing.\n---\n\n# Good\n\nDo the thing.\n"
	repoA := makeValidateRepo(t, map[string]string{"good": good})
	repoB := makeValidateRepo(t, map[string]string{"good": good})

	out, err := executeCmd(t, "validate", "--repo", repoA, "--repo", repoB, "--json")
	require.NoError(t, err)

	var decoded []repoValidation
	require.NoError(t, json.Unmarshal([]byte(out), &decoded),
		"multi-repo --json output must be a single parseable JSON document, got: %s", out)
	require.Len(t, decoded, 2)
	assert.Equal(t, repoA, decoded[0].Repo)
	assert.Equal(t, repoB, decoded[1].Repo)
}
