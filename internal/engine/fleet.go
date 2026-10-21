package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/aminmesbahi/skell/internal/audit"
	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/aminmesbahi/skell/internal/scanner"
	"github.com/aminmesbahi/skell/internal/signing"
	"github.com/aminmesbahi/skell/internal/skillscan"
	"github.com/aminmesbahi/skell/internal/target"
)

// FleetOptions controls what a fleet report gathers.
type FleetOptions struct {
	// Roots are repositories to include. Use ScanRoot to discover them.
	Roots []string
	// ScanRoot, when set, finds every repository under it.
	ScanRoot string
	// CheckUpdates asks each source whether a newer version exists. It needs
	// the network (or a warm cache) and is off by default.
	CheckUpdates bool
	// Events is how many recent audit-log entries to include (0 = none).
	Events int
}

// FleetSkill is one installed skill in one repository.
type FleetSkill struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Commit    string `json:"commit,omitempty"`
	Registry  string `json:"registry"`
	Status    string `json:"status"`
	Pinned    bool   `json:"pinned,omitempty"`
	Linked    bool   `json:"linked,omitempty"`
	Modified  bool   `json:"modified,omitempty"`
	Risk      string `json:"risk"`
	RiskScore int    `json:"risk_score"`
	Signature string `json:"signature"`
	Publisher string `json:"publisher,omitempty"`
}

// FleetRepo is one repository's skill setup.
type FleetRepo struct {
	Path    string       `json:"path"`
	Name    string       `json:"name"`
	Target  string       `json:"target,omitempty"`
	Mirrors []string     `json:"mirrors,omitempty"`
	Skills  []FleetSkill `json:"skills"`
	// Errors and Warnings count `skell doctor` findings.
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
	// Health is "healthy", "attention" or "problem".
	Health string `json:"health"`
	Note   string `json:"note,omitempty"`
}

// VersionCount is how many repositories run one version of a skill.
type VersionCount struct {
	Version string   `json:"version"`
	Repos   int      `json:"repos"`
	Paths   []string `json:"paths,omitempty"`
}

// FleetMatrixRow aggregates one skill across the fleet.
type FleetMatrixRow struct {
	Name     string         `json:"name"`
	Repos    int            `json:"repos"`
	Versions []VersionCount `json:"versions"`
	// Spread is true when repositories run different versions of the skill.
	Spread bool `json:"spread"`
}

// FleetSummary is the headline numbers.
type FleetSummary struct {
	Repos         int `json:"repos"`
	Installs      int `json:"installs"`
	UniqueSkills  int `json:"unique_skills"`
	Outdated      int `json:"outdated"`
	Modified      int `json:"modified"`
	Pinned        int `json:"pinned"`
	HighRisk      int `json:"high_risk"`
	Unsigned      int `json:"unsigned"`
	VersionSpread int `json:"version_spread"`
	Unhealthy     int `json:"unhealthy_repos"`
}

// FleetReport is the whole picture of skills across many repositories.
type FleetReport struct {
	GeneratedAt string           `json:"generated_at"`
	Root        string           `json:"root,omitempty"`
	Summary     FleetSummary     `json:"summary"`
	Repos       []FleetRepo      `json:"repos"`
	Matrix      []FleetMatrixRow `json:"matrix"`
	Events      []audit.Entry    `json:"events,omitempty"`
}

// Fleet gathers the report. It reads lock files and installed folders; it
// only touches the network when opts.CheckUpdates is set.
func (e *Engine) Fleet(opts FleetOptions) (*FleetReport, error) {
	roots := append([]string{}, opts.Roots...)
	if opts.ScanRoot != "" {
		found, err := scanner.ScanAll(opts.ScanRoot)
		if err != nil {
			return nil, err
		}
		for _, r := range found {
			roots = append(roots, r.RepoRoot)
		}
	}
	roots = dedupeStrings(roots)
	if len(roots) == 0 {
		return nil, fmt.Errorf("no repositories to report on — pass --repo or --all-repos <folder>")
	}

	rep := &FleetReport{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Root:        opts.ScanRoot,
		Repos:       []FleetRepo{},
		Matrix:      []FleetMatrixRow{},
	}
	trust := e.TrustStore()
	for _, root := range roots {
		rep.Repos = append(rep.Repos, e.fleetRepo(root, opts, trust))
	}
	sort.SliceStable(rep.Repos, func(i, j int) bool { return rep.Repos[i].Path < rep.Repos[j].Path })

	rep.Matrix = buildMatrix(rep.Repos)
	rep.Summary = summarise(rep)

	if opts.Events > 0 && e.homeRoot != "" {
		events, _ := audit.ReadRecent(filepath.Join(e.homeRoot, "audit.log"), 0)
		inFleet := map[string]bool{}
		for _, r := range rep.Repos {
			inFleet[filepath.Clean(r.Path)] = true
		}
		for _, ev := range events {
			if len(rep.Events) >= opts.Events {
				break
			}
			if inFleet[filepath.Clean(ev.Repo)] {
				rep.Events = append(rep.Events, ev)
			}
		}
	}
	return rep, nil
}

func (e *Engine) fleetRepo(root string, opts FleetOptions, trust *signing.TrustStore) FleetRepo {
	fr := FleetRepo{Path: root, Name: filepath.Base(root), Skills: []FleetSkill{}, Health: "healthy"}
	m, t, err := manifest.ResolveWithTarget(root)
	if err != nil {
		fr.Health, fr.Note = "attention", "not set up for Skell"
		return fr
	}
	fr.Target, fr.Mirrors = t.ID, m.Mirrors

	latest := map[string]model.SkillStatus{}
	if opts.CheckUpdates {
		if entries, err := e.StatusFor(root, t.ID); err == nil {
			for _, en := range entries {
				latest[en.Name] = en.Status
			}
		}
	}

	if lf, err := lockfile.Read(lockfile.PathFor(root, *t)); err == nil {
		for _, s := range lf.Skills {
			fr.Skills = append(fr.Skills, e.fleetSkill(root, *t, s, latest, trust))
		}
	}
	sort.Slice(fr.Skills, func(i, j int) bool { return fr.Skills[i].Name < fr.Skills[j].Name })

	if issues, err := e.Doctor(root); err == nil {
		for _, is := range issues {
			switch is.Severity {
			case SeverityError:
				fr.Errors++
			case SeverityWarning:
				fr.Warnings++
			}
		}
	}
	switch {
	case fr.Errors > 0:
		fr.Health = "problem"
	case fr.Warnings > 0:
		fr.Health = "attention"
	}
	return fr
}

func (e *Engine) fleetSkill(root string, t target.Target, s model.InstalledSkill, latest map[string]model.SkillStatus, trust *signing.TrustStore) FleetSkill {
	fs := FleetSkill{
		Name: s.Name, Version: s.Version, Commit: s.Commit, Registry: s.Registry,
		Pinned: s.Pinned, Linked: s.Linked, Status: "installed", Risk: string(skillscan.LevelClean), Signature: string(signing.StatusUnsigned),
	}
	dir := filepath.Join(t.SkillsDir(root), s.Name)
	if _, err := os.Stat(dir); err != nil {
		fs.Status = "missing"
		return fs
	}
	if locallyModified(dir, s) {
		fs.Modified = true
		fs.Status = string(model.StatusLocallyModified)
	}
	if st, ok := latest[s.Name]; ok && !fs.Modified {
		fs.Status = string(st)
	}
	switch {
	case s.Linked:
		fs.Status = string(model.StatusLinked)
	case s.Pinned && fs.Status == "installed":
		fs.Status = string(model.StatusPinned)
	}
	if rep, err := skillscan.ScanDir(dir); err == nil {
		fs.Risk, fs.RiskScore = string(rep.Level), rep.Score
	}
	res := signing.Verify(dir, trust)
	fs.Signature, fs.Publisher = string(res.Status), res.Publisher
	return fs
}

func buildMatrix(repos []FleetRepo) []FleetMatrixRow {
	type key struct{ name, version string }
	paths := map[key][]string{}
	names := map[string]bool{}
	for _, r := range repos {
		for _, s := range r.Skills {
			k := key{s.Name, model.VersionLabel(s.Version, s.Commit)}
			paths[k] = append(paths[k], r.Path)
			names[s.Name] = true
		}
	}
	var rows []FleetMatrixRow
	for name := range names {
		row := FleetMatrixRow{Name: name}
		for k, ps := range paths {
			if k.name != name {
				continue
			}
			row.Versions = append(row.Versions, VersionCount{Version: k.version, Repos: len(ps), Paths: ps})
			row.Repos += len(ps)
		}
		sort.Slice(row.Versions, func(i, j int) bool {
			if row.Versions[i].Repos != row.Versions[j].Repos {
				return row.Versions[i].Repos > row.Versions[j].Repos
			}
			return row.Versions[i].Version < row.Versions[j].Version
		})
		row.Spread = len(row.Versions) > 1
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Spread != rows[j].Spread {
			return rows[i].Spread
		}
		if rows[i].Repos != rows[j].Repos {
			return rows[i].Repos > rows[j].Repos
		}
		return rows[i].Name < rows[j].Name
	})
	return rows
}

func summarise(rep *FleetReport) FleetSummary {
	sum := FleetSummary{Repos: len(rep.Repos), UniqueSkills: len(rep.Matrix)}
	for _, r := range rep.Repos {
		if r.Health != "healthy" {
			sum.Unhealthy++
		}
		for _, s := range r.Skills {
			sum.Installs++
			if s.Status == string(model.StatusOutdated) {
				sum.Outdated++
			}
			if s.Modified {
				sum.Modified++
			}
			if s.Pinned {
				sum.Pinned++
			}
			if skillscan.Level(s.Risk).AtLeast(skillscan.LevelHigh) {
				sum.HighRisk++
			}
			if s.Signature != string(signing.StatusTrusted) {
				sum.Unsigned++
			}
		}
	}
	for _, row := range rep.Matrix {
		if row.Spread {
			sum.VersionSpread++
		}
	}
	return sum
}

// FleetCSV renders one row per installed skill per repository.
func FleetCSV(rep *FleetReport) string {
	var b strings.Builder
	b.WriteString("repo,skill,version,commit,registry,status,risk,risk_score,signature,publisher\n")
	for _, r := range rep.Repos {
		for _, s := range r.Skills {
			fields := []string{r.Path, s.Name, s.Version, s.Commit, s.Registry, s.Status, s.Risk, fmt.Sprint(s.RiskScore), s.Signature, s.Publisher}
			for i, f := range fields {
				fields[i] = csvCell(f)
			}
			b.WriteString(strings.Join(fields, ",") + "\n")
		}
	}
	return b.String()
}

// csvCell quotes a value, and defuses spreadsheet formulas (a skill name or
// path starting with = + - @ would otherwise run when the file is opened).
func csvCell(s string) string {
	if s != "" && strings.ContainsRune("=+-@\t\r", rune(s[0])) {
		s = "'" + s
	}
	if strings.ContainsAny(s, ",\"\n\r") {
		s = `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return s
}

// FleetMarkdown renders the report as Markdown.
func FleetMarkdown(rep *FleetReport) string {
	var b strings.Builder
	s := rep.Summary
	fmt.Fprintf(&b, "# Skill fleet report\n\n_Generated %s_\n\n", rep.GeneratedAt)
	fmt.Fprintf(&b, "| Repos | Installs | Unique skills | Outdated | Modified | High risk | Not trusted-signed | Version spread |\n|---|---|---|---|---|---|---|---|\n")
	fmt.Fprintf(&b, "| %d | %d | %d | %d | %d | %d | %d | %d |\n\n", s.Repos, s.Installs, s.UniqueSkills, s.Outdated, s.Modified, s.HighRisk, s.Unsigned, s.VersionSpread)

	b.WriteString("## Repositories\n\n| Repo | Agent | Skills | Health |\n|---|---|---|---|\n")
	for _, r := range rep.Repos {
		fmt.Fprintf(&b, "| %s | %s | %d | %s |\n", mdCell(r.Name), mdCell(r.Target), len(r.Skills), r.Health)
	}
	b.WriteString("\n## Skills across the fleet\n\n| Skill | Repos | Versions |\n|---|---|---|\n")
	for _, row := range rep.Matrix {
		var vs []string
		for _, v := range row.Versions {
			vs = append(vs, fmt.Sprintf("%s ×%d", v.Version, v.Repos))
		}
		fmt.Fprintf(&b, "| %s | %d | %s |\n", mdCell(row.Name), row.Repos, mdCell(strings.Join(vs, ", ")))
	}
	if len(rep.Events) > 0 {
		b.WriteString("\n## Recent activity\n\n| When | Who | Action | Skill | Repo |\n|---|---|---|---|---|\n")
		for _, ev := range rep.Events {
			fmt.Fprintf(&b, "| %s | %s | %s | %s | %s |\n", ev.Timestamp, mdCell(ev.User), ev.Action, mdCell(ev.Skill), mdCell(filepath.Base(ev.Repo)))
		}
	}
	return b.String()
}

func mdCell(s string) string {
	s = strings.ReplaceAll(s, "|", `\|`)
	return strings.ReplaceAll(strings.ReplaceAll(s, "\n", " "), "\r", " ")
}
