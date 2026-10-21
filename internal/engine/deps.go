package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aminmesbahi/skell/internal/deps"
	"github.com/aminmesbahi/skell/internal/frontmatter"
	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/aminmesbahi/skell/internal/registry"
)

// PlannedInstall is one skill to install, either asked for or pulled in as a
// dependency.
type PlannedInstall struct {
	Name string `json:"name"`
	// Registry is the source alias to install from; empty lets install find it.
	Registry string `json:"registry,omitempty"`
	// RequiredBy names the skills that need this one (empty for skills the
	// user asked for).
	RequiredBy []string `json:"required_by,omitempty"`
}

// InstallPlan is the full set of skills an install will add.
type InstallPlan struct {
	Items []PlannedInstall `json:"items"`
	// Cycles lists dependency loops that were found (they are harmless for a
	// flat install but usually a mistake).
	Cycles [][]string `json:"cycles,omitempty"`
}

// PlanInstall expands the requested skills with everything they require
// (transitively), skipping what is already installed. Skills the user named
// come first, then dependencies in the order they were discovered. A
// dependency that cannot be found, or whose version constraint no source can
// meet, is an error that names the skill that needs it.
func (e *Engine) PlanInstall(repoRoot string, names []string, registryAlias, registryURL string) (*InstallPlan, error) {
	m, t, err := manifest.ResolveWithTarget(repoRoot)
	if err != nil {
		// No manifest yet: nothing to resolve against; install reports the real error.
		items := make([]PlannedInstall, 0, len(names))
		for _, n := range names {
			items = append(items, PlannedInstall{Name: n, Registry: registryAlias})
		}
		return &InstallPlan{Items: items}, nil
	}
	installed := map[string]model.InstalledSkill{}
	if lf, err := lockfile.Read(lockfile.PathFor(repoRoot, *t)); err == nil {
		for _, s := range lf.Skills {
			installed[s.Name] = s
		}
	}

	plan := &InstallPlan{}
	index := map[string]int{}
	for _, n := range names {
		if _, dup := index[n]; dup {
			continue
		}
		index[n] = len(plan.Items)
		plan.Items = append(plan.Items, PlannedInstall{Name: n, Registry: registryAlias})
	}

	done := map[string]bool{}
	var visit func(name, regAlias, regURL string, stack []string) error
	visit = func(name, regAlias, regURL string, stack []string) error {
		for i, s := range stack {
			if s == name {
				plan.Cycles = append(plan.Cycles, append(append([]string{}, stack[i:]...), name))
				return nil
			}
		}
		if done[name] {
			return nil
		}
		done[name] = true

		alias, url, _, err := e.resolveInstallRegistry(m, name, regAlias, regURL)
		if err != nil {
			if len(stack) == 0 {
				return nil // a requested skill that cannot be found is reported by install itself
			}
			return fmt.Errorf("%s requires %q, which no configured source provides: %w", stack[len(stack)-1], name, err)
		}
		if e.pol.CheckRegistry(url) != nil {
			return nil
		}
		rs, err := e.provider.GetSkill(registry.Registry{Alias: alias, URL: url}, name)
		if err != nil {
			if len(stack) == 0 {
				return nil
			}
			return fmt.Errorf("%s requires %q: %w", stack[len(stack)-1], name, err)
		}
		if i, ok := index[name]; ok && plan.Items[i].Registry == "" {
			plan.Items[i].Registry = alias
		}
		requires, perr := deps.Parse(rs.Metadata.Requires)
		if perr != nil {
			return fmt.Errorf("%s: %w", name, perr)
		}
		for _, d := range requires {
			if cur, ok := installed[d.Name]; ok {
				if !d.Satisfied(cur.Version) {
					return fmt.Errorf("%s requires %s but version %s is installed — run 'skell upgrade %s'",
						name, d, model.VersionLabel(cur.Version, cur.Commit), d.Name)
				}
				continue
			}
			depAlias, depURL, _, derr := e.resolveInstallRegistry(m, d.Name, alias, "")
			if derr != nil {
				depAlias, depURL, _, derr = e.resolveInstallRegistry(m, d.Name, "", "")
			}
			if derr != nil {
				return fmt.Errorf("%s requires %q, which no configured source provides: %w", name, d.Name, derr)
			}
			drs, gerr := e.provider.GetSkill(registry.Registry{Alias: depAlias, URL: depURL}, d.Name)
			if gerr != nil {
				return fmt.Errorf("%s requires %q: %w", name, d.Name, gerr)
			}
			if !d.Satisfied(drs.Metadata.Version) {
				return fmt.Errorf("%s requires %s but the source offers version %s", name, d, dashVersion(drs.Metadata.Version))
			}
			if i, ok := index[d.Name]; ok {
				plan.Items[i].RequiredBy = appendUnique(plan.Items[i].RequiredBy, name)
			} else {
				index[d.Name] = len(plan.Items)
				plan.Items = append(plan.Items, PlannedInstall{Name: d.Name, Registry: depAlias, RequiredBy: []string{name}})
			}
			if err := visit(d.Name, depAlias, depURL, append(stack, name)); err != nil {
				return err
			}
		}
		return nil
	}

	for _, n := range names {
		if _, ok := installed[n]; ok {
			continue // already installed; install reports it
		}
		if err := visit(n, registryAlias, registryURL, nil); err != nil {
			return nil, err
		}
	}
	return plan, nil
}

func dashVersion(v string) string {
	if v == "" {
		return "(none)"
	}
	return v
}

func appendUnique(list []string, s string) []string {
	for _, x := range list {
		if x == s {
			return list
		}
	}
	return append(list, s)
}

// dependencyIssues reports installed skills whose requirements are missing or
// out of date, and dependency loops.
func dependencyIssues(skillsDir string, installed []model.InstalledSkill) []DiagnosticIssue {
	byName := map[string]model.InstalledSkill{}
	for _, s := range installed {
		byName[s.Name] = s
	}
	graph := map[string][]string{}
	var issues []DiagnosticIssue
	for _, s := range installed {
		dir := filepath.Join(skillsDir, s.Name)
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		rs, err := frontmatter.ParseDir(dir)
		if err != nil {
			continue
		}
		reqs, err := deps.Parse(rs.Metadata.Requires)
		if err != nil {
			issues = append(issues, DiagnosticIssue{
				Severity: SeverityWarning, Code: "bad-requires",
				Message: fmt.Sprintf("skill %q: %v", s.Name, err),
			})
			continue
		}
		for _, d := range reqs {
			graph[s.Name] = append(graph[s.Name], d.Name)
			have, ok := byName[d.Name]
			switch {
			case !ok:
				issues = append(issues, DiagnosticIssue{
					Severity: SeverityWarning, Code: "missing-dependency",
					Message: fmt.Sprintf("skill %q requires %q, which is not installed", s.Name, d.Name),
					Hint:    fmt.Sprintf("run 'skell install %s'", d.Name),
				})
			case !d.Satisfied(have.Version):
				issues = append(issues, DiagnosticIssue{
					Severity: SeverityWarning, Code: "dependency-version",
					Message: fmt.Sprintf("skill %q requires %s but %s is installed", s.Name, d, model.VersionLabel(have.Version, have.Commit)),
					Hint:    fmt.Sprintf("run 'skell upgrade %s'", d.Name),
				})
			}
		}
	}
	for _, cycle := range findCycles(graph) {
		issues = append(issues, DiagnosticIssue{
			Severity: SeverityInfo, Code: "dependency-cycle",
			Message: "skills require each other in a loop: " + strings.Join(cycle, " → "),
			Hint:    "break the loop by removing one of the requires entries",
		})
	}
	return issues
}

// findCycles returns each dependency loop once, as a path that ends where it began.
func findCycles(graph map[string][]string) [][]string {
	const (
		white = iota
		grey
		black
	)
	color := map[string]int{}
	var cycles [][]string
	seen := map[string]bool{}
	var dfs func(n string, path []string)
	dfs = func(n string, path []string) {
		color[n] = grey
		path = append(path, n)
		for _, next := range graph[n] {
			switch color[next] {
			case white:
				dfs(next, path)
			case grey:
				for i, p := range path {
					if p == next {
						cyc := append(append([]string{}, path[i:]...), next)
						if key := strings.Join(cyc, ">"); !seen[key] {
							seen[key] = true
							cycles = append(cycles, cyc)
						}
						break
					}
				}
			}
		}
		color[n] = black
	}
	names := make([]string, 0, len(graph))
	for n := range graph {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		if color[n] == white {
			dfs(n, nil)
		}
	}
	return cycles
}
