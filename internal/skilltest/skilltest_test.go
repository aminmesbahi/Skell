package skilltest

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeSuite(t *testing.T, yaml string) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "evals"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: x\ndescription: x\n---\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "evals", "evals.yaml"), []byte(yaml), 0o644))
	return dir
}

const goodSuite = `cases:
  - name: greets
    prompt: say hello
    expect:
      contains: ["hello"]
      not_contains: ["goodbye"]
      matches: ["(?i)^hel+o"]
`

func TestLoadAndCheck(t *testing.T) {
	dir := writeSuite(t, goodSuite)
	s, err := Load(dir)
	require.NoError(t, err)
	assert.Equal(t, "evals/evals.yaml", s.Path)
	require.Len(t, s.Cases, 1)
	assert.Empty(t, s.Check())

	_, err = Load(t.TempDir())
	assert.ErrorIs(t, err, ErrNoSuite)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "evals", "evals.yaml"), []byte("cases: [oops"), 0o644))
	_, err = Load(dir)
	assert.Error(t, err)
}

func TestCheckFindsMistakes(t *testing.T) {
	s := &Suite{Cases: []Case{
		{Name: "a", Prompt: "p", Expect: Expect{Matches: []string{"("}}},
		{Name: "a", Prompt: "", Expect: Expect{}},
		{Prompt: "p", Expect: Expect{Contains: []string{"x"}}, Timeout: -1},
	}}
	problems := s.Check()
	assert.GreaterOrEqual(t, len(problems), 5)
	assert.Equal(t, []string{"the suite has no cases"}, (&Suite{}).Check())
}

func TestEvaluate(t *testing.T) {
	e := Expect{Contains: []string{"Hello"}, NotContains: []string{"bye"}, Matches: []string{`\d+`}, NotMatches: []string{`^err`}}
	assert.Empty(t, Evaluate(e, "hello world 42"))
	assert.Len(t, Evaluate(e, "err bye"), 4)
}

func TestRunWithoutAgentSkipsCases(t *testing.T) {
	dir := writeSuite(t, goodSuite)
	s, _ := Load(dir)
	res := Run(context.Background(), "x", dir, s, Options{})
	assert.Equal(t, 1, res.Skipped)
	assert.True(t, res.OK(), "static checks pass, nothing ran")
}

// helperAgent returns an agent command that echoes its stdin back, using the
// test binary itself so no external tools are needed.
func helperAgent(t *testing.T) string {
	t.Helper()
	exe, err := os.Executable()
	require.NoError(t, err)
	return `"` + exe + `" -test.run=TestHelperAgent -- agent`
}

func TestHelperAgent(t *testing.T) {
	if len(os.Args) < 2 || os.Args[len(os.Args)-1] != "agent" {
		t.Skip("helper process")
	}
	buf := make([]byte, 4096)
	n, _ := os.Stdin.Read(buf)
	prompt := string(buf[:n])
	if prompt == "fail please" {
		os.Stderr.WriteString("boom")
		os.Exit(3)
	}
	os.Stdout.WriteString("Hello! You said: " + prompt)
	os.Exit(0)
}

func TestRunWithAgent(t *testing.T) {
	if runtime.GOOS == "js" {
		t.Skip()
	}
	if _, err := exec.LookPath(os.Args[0]); err != nil && !filepath.IsAbs(os.Args[0]) {
		t.Log("running via relative test binary path")
	}
	dir := writeSuite(t, `cases:
  - name: passes
    prompt: say hello
    expect: {contains: ["hello"]}
  - name: fails
    prompt: say hello
    expect: {contains: ["nope"]}
  - name: errors
    prompt: fail please
    expect: {contains: ["x"]}
`)
	s, _ := Load(dir)
	res := Run(context.Background(), "x", dir, s, Options{Agent: helperAgent(t), Timeout: 30 * time.Second})
	assert.Equal(t, 1, res.Passed)
	assert.Equal(t, 1, res.Failed)
	assert.Equal(t, 1, res.Errored)
	assert.False(t, res.OK())
	assert.Equal(t, []string{"fails", "errors"}, res.FailingCases())
	assert.Contains(t, res.Cases[2].Failures[0], "boom")

	only := Run(context.Background(), "x", dir, s, Options{Agent: helperAgent(t), Only: []string{"passes"}})
	assert.Len(t, only.Cases, 1)
}

func TestRegressions(t *testing.T) {
	before := &Result{Cases: []CaseResult{{Name: "a", Status: StatusFail}, {Name: "b", Status: StatusPass}}}
	after := &Result{Cases: []CaseResult{{Name: "a", Status: StatusFail}, {Name: "b", Status: StatusFail}, {Name: "c", Status: StatusError}}}
	assert.Equal(t, []string{"b", "c"}, Regressions(before, after))
	assert.Equal(t, []string{"a", "b", "c"}, Regressions(nil, after))
}

func TestSplitCommand(t *testing.T) {
	got, err := splitCommand(`claude -p "two words" 'a b' plain`)
	require.NoError(t, err)
	assert.Equal(t, []string{"claude", "-p", "two words", "a b", "plain"}, got)
	_, err = splitCommand(`oops "unterminated`)
	assert.Error(t, err)
}

func TestStarterSuiteIsValid(t *testing.T) {
	dir := writeSuite(t, StarterSuite("demo"))
	s, err := Load(dir)
	require.NoError(t, err)
	assert.Empty(t, s.Check())
	assert.Equal(t, "claude -p", DefaultAgent("claude"))
	assert.Equal(t, "", DefaultAgent("cursor"))
}
