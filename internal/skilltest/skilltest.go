// Package skilltest runs the eval cases a skill author ships with a skill.
//
// A skill carries its cases in evals/evals.yaml:
//
//	cases:
//	  - name: summarises a PDF
//	    prompt: Summarise report.pdf in three bullet points.
//	    timeout: 180
//	    expect:
//	      contains: ["summary"]
//	      not_contains: ["I cannot"]
//	      matches: ["(?m)^- "]
//
// Each case is given to an AI agent running in a scratch workspace in which
// the skill is installed; the agent's answer is then checked against the
// expectations. Without an agent command the cases are only checked for
// mistakes (static checks), so the suite can still run in plain CI.
package skilltest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ErrNoSuite is returned by Load when a skill ships no eval file.
var ErrNoSuite = errors.New("this skill ships no evals (evals/evals.yaml)")

// SuiteFiles are the places a suite may live, in order of preference.
var SuiteFiles = []string{"evals/evals.yaml", "evals/evals.yml", "evals/evals.json"}

// Expect are the checks applied to the agent's answer.
type Expect struct {
	Contains    []string `yaml:"contains" json:"contains,omitempty"`
	NotContains []string `yaml:"not_contains" json:"not_contains,omitempty"`
	Matches     []string `yaml:"matches" json:"matches,omitempty"`
	NotMatches  []string `yaml:"not_matches" json:"not_matches,omitempty"`
}

// Case is one prompt and what a good answer looks like.
type Case struct {
	Name    string `yaml:"name" json:"name"`
	Prompt  string `yaml:"prompt" json:"prompt"`
	Timeout int    `yaml:"timeout" json:"timeout,omitempty"` // seconds
	Expect  Expect `yaml:"expect" json:"expect"`
}

// Suite is a skill's eval file.
type Suite struct {
	Path  string `yaml:"-" json:"path"`
	Cases []Case `yaml:"cases" json:"cases"`
}

// Load reads the suite shipped in a skill folder.
func Load(skillDir string) (*Suite, error) {
	for _, rel := range SuiteFiles {
		p := filepath.Join(skillDir, filepath.FromSlash(rel))
		data, err := os.ReadFile(p) //nolint:gosec
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		var s Suite
		if err := yaml.Unmarshal(data, &s); err != nil {
			return nil, fmt.Errorf("%s: %w", rel, err)
		}
		s.Path = rel
		return &s, nil
	}
	return nil, ErrNoSuite
}

// Check reports mistakes in the suite itself.
func (s *Suite) Check() []string {
	var problems []string
	if len(s.Cases) == 0 {
		return []string{"the suite has no cases"}
	}
	seen := map[string]bool{}
	for i, c := range s.Cases {
		label := c.Name
		if label == "" {
			label = fmt.Sprintf("case %d", i+1)
			problems = append(problems, label+": needs a name")
		}
		if seen[c.Name] && c.Name != "" {
			problems = append(problems, label+": duplicate name")
		}
		seen[c.Name] = true
		if strings.TrimSpace(c.Prompt) == "" {
			problems = append(problems, label+": needs a prompt")
		}
		if len(c.Expect.Contains)+len(c.Expect.NotContains)+len(c.Expect.Matches)+len(c.Expect.NotMatches) == 0 {
			problems = append(problems, label+": has no expectations, so it can never fail")
		}
		for _, rx := range append(append([]string{}, c.Expect.Matches...), c.Expect.NotMatches...) {
			if _, err := regexp.Compile(rx); err != nil {
				problems = append(problems, fmt.Sprintf("%s: bad pattern %q: %v", label, rx, err))
			}
		}
		if c.Timeout < 0 {
			problems = append(problems, label+": timeout cannot be negative")
		}
	}
	return problems
}

// Case outcomes.
const (
	StatusPass    = "pass"
	StatusFail    = "fail"
	StatusError   = "error"
	StatusSkipped = "skipped"
)

// CaseResult is the outcome of one case.
type CaseResult struct {
	Name       string   `json:"name"`
	Status     string   `json:"status"`
	Failures   []string `json:"failures,omitempty"`
	DurationMS int64    `json:"duration_ms"`
	Output     string   `json:"output,omitempty"`
}

// Result is the outcome of running a skill's suite.
type Result struct {
	Skill   string       `json:"skill"`
	Suite   string       `json:"suite,omitempty"`
	Agent   string       `json:"agent,omitempty"`
	Static  []string     `json:"static_problems,omitempty"`
	Cases   []CaseResult `json:"cases"`
	Passed  int          `json:"passed"`
	Failed  int          `json:"failed"`
	Errored int          `json:"errored"`
	Skipped int          `json:"skipped"`
}

// OK is true when nothing failed or errored and the suite is sound.
func (r *Result) OK() bool { return r.Failed == 0 && r.Errored == 0 && len(r.Static) == 0 }

// FailingCases lists the names of cases that failed or errored.
func (r *Result) FailingCases() []string {
	var out []string
	for _, c := range r.Cases {
		if c.Status == StatusFail || c.Status == StatusError {
			out = append(out, c.Name)
		}
	}
	return out
}

// Options configure a run.
type Options struct {
	// Agent is the command that answers a prompt, e.g. "claude -p". The prompt
	// is sent on standard input, or substituted for a {prompt} argument. An
	// empty Agent runs the static checks only.
	Agent string
	// SkillsDir is the agent's skills folder relative to the workspace root
	// (e.g. ".claude/skills").
	SkillsDir string
	// Timeout applies to cases that set none (default 120s).
	Timeout time.Duration
	// Only runs just the named cases.
	Only []string
}

// DefaultAgent returns a best-effort command for an agent target, or "".
func DefaultAgent(targetID string) string {
	switch targetID {
	case "claude":
		return "claude -p"
	case "codex":
		return "codex exec"
	}
	return ""
}

// Run executes a skill's suite.
func Run(ctx context.Context, name, skillDir string, suite *Suite, opts Options) *Result {
	res := &Result{Skill: name, Suite: suite.Path, Agent: opts.Agent, Cases: []CaseResult{}, Static: suite.Check()}
	skip := func(c Case, why string) {
		res.Cases = append(res.Cases, CaseResult{Name: c.Name, Status: StatusSkipped, Failures: []string{why}})
		res.Skipped++
	}
	cases := selectCases(suite.Cases, opts.Only)

	if len(res.Static) > 0 || strings.TrimSpace(opts.Agent) == "" {
		why := "no agent configured — pass --agent or set SKELL_TEST_AGENT"
		if len(res.Static) > 0 {
			why = "fix the suite first"
		}
		for _, c := range cases {
			skip(c, why)
		}
		return res
	}

	argv, err := splitCommand(opts.Agent)
	if err != nil || len(argv) == 0 {
		res.Static = append(res.Static, "agent command is not usable: "+fmt.Sprint(err))
		return res
	}
	work, err := os.MkdirTemp("", "skell-test-")
	if err != nil {
		res.Static = append(res.Static, err.Error())
		return res
	}
	defer func() { _ = os.RemoveAll(work) }()
	skillsDir := opts.SkillsDir
	if skillsDir == "" {
		skillsDir = ".claude/skills"
	}
	if err := copyTree(skillDir, filepath.Join(work, filepath.FromSlash(skillsDir), name)); err != nil {
		res.Static = append(res.Static, "could not stage the skill: "+err.Error())
		return res
	}

	for _, c := range cases {
		cr := runCase(ctx, c, argv, work, opts.Timeout)
		res.Cases = append(res.Cases, cr)
		switch cr.Status {
		case StatusPass:
			res.Passed++
		case StatusFail:
			res.Failed++
		default:
			res.Errored++
		}
	}
	return res
}

func selectCases(all []Case, only []string) []Case {
	if len(only) == 0 {
		return all
	}
	var out []Case
	for _, c := range all {
		for _, o := range only {
			if c.Name == o {
				out = append(out, c)
				break
			}
		}
	}
	return out
}

func runCase(ctx context.Context, c Case, argv []string, work string, def time.Duration) CaseResult {
	timeout := def
	if c.Timeout > 0 {
		timeout = time.Duration(c.Timeout) * time.Second
	}
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	args := make([]string, len(argv)-1)
	usesArg := false
	for i, a := range argv[1:] {
		if strings.Contains(a, "{prompt}") {
			usesArg = true
			a = strings.ReplaceAll(a, "{prompt}", c.Prompt)
		}
		args[i] = a
	}
	cmd := exec.CommandContext(cctx, argv[0], args...) //nolint:gosec // the agent command is chosen by the person running the tests
	cmd.Dir = work
	if !usesArg {
		cmd.Stdin = strings.NewReader(c.Prompt)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	start := time.Now()
	runErr := cmd.Run()
	cr := CaseResult{Name: c.Name, DurationMS: time.Since(start).Milliseconds(), Output: clip(stdout.String(), 4000)}
	switch {
	case cctx.Err() == context.DeadlineExceeded:
		cr.Status, cr.Failures = StatusError, []string{fmt.Sprintf("timed out after %s", timeout)}
	case runErr != nil:
		cr.Status = StatusError
		cr.Failures = []string{fmt.Sprintf("agent failed: %v: %s", runErr, clip(strings.TrimSpace(stderr.String()), 400))}
	default:
		cr.Failures = Evaluate(c.Expect, stdout.String())
		cr.Status = StatusPass
		if len(cr.Failures) > 0 {
			cr.Status = StatusFail
		}
	}
	return cr
}

// Evaluate checks an answer against expectations and returns what failed.
func Evaluate(e Expect, output string) []string {
	var fails []string
	lower := strings.ToLower(output)
	for _, s := range e.Contains {
		if !strings.Contains(lower, strings.ToLower(s)) {
			fails = append(fails, fmt.Sprintf("expected the answer to contain %q", s))
		}
	}
	for _, s := range e.NotContains {
		if strings.Contains(lower, strings.ToLower(s)) {
			fails = append(fails, fmt.Sprintf("expected the answer not to contain %q", s))
		}
	}
	for _, p := range e.Matches {
		if rx, err := regexp.Compile(p); err == nil && !rx.MatchString(output) {
			fails = append(fails, fmt.Sprintf("expected the answer to match /%s/", p))
		}
	}
	for _, p := range e.NotMatches {
		if rx, err := regexp.Compile(p); err == nil && rx.MatchString(output) {
			fails = append(fails, fmt.Sprintf("expected the answer not to match /%s/", p))
		}
	}
	return fails
}

// Regressions returns the failing cases in after that did not fail in before
// (or did not exist there): what an upgrade broke.
func Regressions(before, after *Result) []string {
	was := map[string]bool{}
	if before != nil {
		for _, n := range before.FailingCases() {
			was[n] = true
		}
	}
	var out []string
	for _, n := range after.FailingCases() {
		if !was[n] {
			out = append(out, n)
		}
	}
	return out
}

func clip(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// splitCommand splits a command line into words, honouring single and double
// quotes. It never starts a shell, so a prompt cannot inject commands.
func splitCommand(s string) ([]string, error) {
	var out []string
	var cur strings.Builder
	var quote rune
	started := false
	for _, r := range s {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
		case r == '"' || r == '\'':
			quote, started = r, true
		case r == ' ' || r == '\t':
			if started || cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
				started = false
			}
		default:
			cur.WriteRune(r)
		}
	}
	if quote != 0 {
		return nil, errors.New("unterminated quote")
	}
	if started || cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out, nil
}

// copyTree copies a folder, skipping symlinks and .git.
func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return os.MkdirAll(target, 0o755)
		}
		if d.Type()&fs.ModeSymlink != 0 || !d.Type().IsRegular() {
			return nil
		}
		data, err := os.ReadFile(p) //nolint:gosec
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

// StarterSuite is the evals file `skell new` writes for a new skill.
func StarterSuite(skillName string) string {
	return fmt.Sprintf(`# Eval cases for the %s skill. Run them with: skell test --path . --agent "claude -p"
# Each prompt is given to your agent with this skill installed; the answer is
# then checked against "expect". Delete the examples and write your own.
cases:
  - name: uses the skill for its main job
    prompt: "Replace this with a realistic request the skill should handle."
    timeout: 180
    expect:
      contains: ["a word the right answer must include"]
      not_contains: ["I cannot", "I don't have access"]
`, skillName)
}
