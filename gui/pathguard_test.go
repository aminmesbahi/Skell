package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPathGuard_AllowsSkillsDirAndGrantedRoots_RejectsOthers(t *testing.T) {
	var g pathGuard
	tmp := t.TempDir()
	t.Setenv("SKELL_HOME", filepath.Join(tmp, "home"))

	skills := filepath.Join(tmp, "repo", ".claude", "skills", "x")
	require.NoError(t, os.MkdirAll(skills, 0o755))
	_, err := g.check(filepath.Join(skills, "SKILL.md"))
	assert.NoError(t, err)

	_, err = g.check(filepath.Join(tmp, "repo", "secrets.txt"))
	assert.Error(t, err)
	_, err = g.check(filepath.Join(skills, "..", "..", "..", "secrets.txt"))
	assert.Error(t, err)
	_, err = g.check("")
	assert.Error(t, err)

	home := filepath.Join(tmp, "home")
	require.NoError(t, os.MkdirAll(home, 0o755))
	_, err = g.check(filepath.Join(home, "audit.log"))
	assert.NoError(t, err)

	local := filepath.Join(tmp, "local-src")
	require.NoError(t, os.MkdirAll(local, 0o755))
	_, err = g.check(local)
	assert.Error(t, err)
	g.grant(local)
	_, err = g.check(filepath.Join(local, "SKILL.md"))
	assert.NoError(t, err)
}

func TestApplyFrontmatterEdits_LiteralAndSafe(t *testing.T) {
	in := "---\nname: x\ndescription: old\n---\nbody\n"
	out := applyFrontmatterEdits(in, SkillMetadataFields{Description: "Costs $5 and ${1}\nevil: yes"})
	assert.Contains(t, out, `description: "Costs $5 and ${1} evil: yes"`)
	assert.NotContains(t, out, "\nevil:")
}

func TestSkellCacheDir_RejectsTraversal(t *testing.T) {
	for _, bad := range []string{"", "..", "a/b", `a\b`} {
		_, err := skellCacheDir(bad)
		assert.Error(t, err, bad)
	}
}
