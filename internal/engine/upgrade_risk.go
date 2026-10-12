package engine

import (
	"fmt"
	"strings"

	"github.com/aminmesbahi/skell/internal/frontmatter"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/aminmesbahi/skell/internal/registry"
)

// UpgradeRisk describes the extra risk an upgrade would bring compared with
// the copy that is already installed: scripts the installed version did not
// ship, or tool permissions it did not pre-approve. Upgrades that add nothing
// new are never reported, so routine updates stay quiet.
type UpgradeRisk struct {
	Skill           string   `json:"skill"`
	NewScripts      []string `json:"new_scripts,omitempty"`
	NewToolWarnings []string `json:"new_tool_warnings,omitempty"`
	// Review is the full review of the incoming version.
	Review *Review `json:"review"`
}

// Summary is a one-line, human-readable description of what is new.
func (r UpgradeRisk) Summary() string {
	var parts []string
	if len(r.NewScripts) > 0 {
		parts = append(parts, fmt.Sprintf("adds %d script/executable file(s): %s", len(r.NewScripts), strings.Join(firstN(r.NewScripts, 5), ", ")))
	}
	parts = append(parts, r.NewToolWarnings...)
	return strings.Join(parts, "; ")
}

func firstN(items []string, n int) []string {
	if len(items) <= n {
		return items
	}
	return append(append([]string{}, items[:n]...), fmt.Sprintf("… +%d more", len(items)-n))
}

// SetUpgradeGate installs a callback consulted before an upgrade that adds
// scripts or broad tool permissions is applied. Returning false skips that
// skill (it is listed in UpgradeReport.NeedsReview). Without a gate such
// upgrades proceed, which is what library callers and tests get.
func (e *Engine) SetUpgradeGate(gate func(UpgradeRisk) bool) { e.upgradeGate = gate }

// buildReview assembles a Review for a skill whose files live in dir.
func buildReview(rs *model.RegistrySkill, alias, url, dir, commit string) *Review {
	rev := &Review{Skill: rs, Registry: alias, RegistryURL: url, Commit: commit}
	rev.AllowedTools = strings.Fields(rs.AllowedTools)
	if dir != "" {
		inspectSkillDir(dir, rev)
	}
	rev.Warnings = append(rev.Warnings, toolWarnings(rev.AllowedTools)...)
	if len(rev.Scripts) > 0 {
		rev.Warnings = append(rev.Warnings, fmt.Sprintf("ships %d script/executable file(s) the agent may run", len(rev.Scripts)))
	}
	return rev
}

// upgradeRisk compares the incoming version (rs, src) with the installed copy
// in skillDir. It returns nil when the upgrade adds no scripts and no new
// broad tool permissions.
func upgradeRisk(name, skillDir string, rs *model.RegistrySkill, src *registry.SkillSource, reg registry.Registry) *UpgradeRisk {
	var dir, commit string
	if src != nil {
		dir, commit = src.Dir, src.Commit
	}
	incoming := buildReview(rs, reg.Alias, reg.URL, dir, commit)

	installed := &Review{}
	if cur, err := frontmatter.ParseDir(skillDir); err == nil {
		installed = buildReview(cur, reg.Alias, reg.URL, skillDir, "")
	}

	risk := &UpgradeRisk{Skill: name, Review: incoming}
	known := map[string]bool{}
	for _, s := range installed.Scripts {
		known[s] = true
	}
	for _, s := range incoming.Scripts {
		if !known[s] {
			risk.NewScripts = append(risk.NewScripts, s)
		}
	}
	oldWarn := map[string]bool{}
	for _, w := range toolWarnings(installed.AllowedTools) {
		oldWarn[w] = true
	}
	for _, w := range toolWarnings(incoming.AllowedTools) {
		if !oldWarn[w] {
			risk.NewToolWarnings = append(risk.NewToolWarnings, w)
		}
	}
	if len(risk.NewScripts) == 0 && len(risk.NewToolWarnings) == 0 {
		return nil
	}
	return risk
}
