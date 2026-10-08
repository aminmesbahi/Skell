package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"

	"github.com/aminmesbahi/skell/internal/hasher"
	"github.com/aminmesbahi/skell/internal/lockfile"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/registry"
	"github.com/aminmesbahi/skell/internal/target"
)

// Mirrors let one manifest serve several agents in the same repository: the
// manifest's own target holds skell.toml, skell.lock and the real skill
// folders, and each mirror target's skills/ folder receives an identical copy
// (or link, for linked skills). Mirrors are derived state — 'skell sync'
// rebuilds them — so they never get their own manifest or lock file.

// mirrorTargets resolves the manifest's mirror ids to targets, skipping
// unknown ids and the manifest's own target.
func mirrorTargets(m *manifest.Manifest, primary target.Target) []target.Target {
	if m == nil {
		return nil
	}
	var out []target.Target
	for _, id := range m.Mirrors {
		mt, err := target.Lookup(id)
		if err != nil || mt.ID == primary.ID {
			continue
		}
		out = append(out, mt)
	}
	return out
}

// mirrorIDs returns the ids of mirror targets for the manifest found in
// repoRoot (empty when there is none). Used to keep mirrors out of
// multi-target aggregation so skills aren't listed twice.
func mirrorIDs(repoRoot string) map[string]bool {
	out := map[string]bool{}
	m, t, err := manifest.ResolveWithTarget(repoRoot)
	if err != nil {
		return out
	}
	for _, mt := range mirrorTargets(m, *t) {
		out[mt.ID] = true
	}
	return out
}

// mirrorSkill copies (or links) one installed skill into every mirror target.
func (e *Engine) mirrorSkill(repoRoot string, m *manifest.Manifest, t target.Target, name string) error {
	src := filepath.Join(t.SkillsDir(repoRoot), name)
	for _, mt := range mirrorTargets(m, t) {
		dst := filepath.Join(mt.SkillsDir(repoRoot), name)
		if err := os.MkdirAll(mt.SkillsDir(repoRoot), 0o755); err != nil {
			return err
		}
		if linkTarget, ok := linkDestination(src); ok {
			_ = os.RemoveAll(dst)
			if err := createDirLink(linkTarget, dst); err != nil {
				return fmt.Errorf("mirror %q to %s: %w", name, mt.ID, err)
			}
			continue
		}
		if err := registry.CopyTree(src, dst); err != nil {
			return fmt.Errorf("mirror %q to %s: %w", name, mt.ID, err)
		}
	}
	return nil
}

// unmirrorSkill removes a skill from every mirror target (best effort).
func (e *Engine) unmirrorSkill(repoRoot string, m *manifest.Manifest, t target.Target, name string) {
	for _, mt := range mirrorTargets(m, t) {
		_ = os.RemoveAll(filepath.Join(mt.SkillsDir(repoRoot), name))
	}
}

// mirrorAfterChange re-reads the manifest and mirrors one skill. Install and
// upgrade call it after writing the primary copy.
func (e *Engine) mirrorAfterChange(repoRoot string, t target.Target, name string) error {
	m, err := manifest.Read(manifest.LocalPathFor(repoRoot, t))
	if err != nil {
		return nil // no manifest at this target → nothing to mirror
	}
	return e.mirrorSkill(repoRoot, m, t, name)
}

// MirrorReport summarises a mirror refresh.
type MirrorReport struct {
	Targets []string `json:"targets"`
	Copied  []string `json:"copied"`
	Removed []string `json:"removed"`
}

// RefreshMirrors brings every mirror target in line with the primary: each
// skill in the lock file is copied, and skills Skell previously mirrored that
// are no longer installed are removed. Hand-authored folders in a mirror that
// Skell never managed are left alone.
func (e *Engine) RefreshMirrors(repoRoot string) (*MirrorReport, error) {
	m, t, err := manifest.ResolveWithTarget(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("no manifest found in %s — run 'skell init' first: %w", repoRoot, err)
	}
	rep := &MirrorReport{}
	mts := mirrorTargets(m, *t)
	for _, mt := range mts {
		rep.Targets = append(rep.Targets, mt.ID)
	}
	if len(mts) == 0 {
		return rep, nil
	}
	lf, err := lockfile.Read(lockfile.PathFor(repoRoot, *t))
	if err != nil {
		return rep, nil // nothing installed yet
	}
	present := map[string]bool{}
	for _, s := range lf.Skills {
		if _, err := os.Lstat(filepath.Join(t.SkillsDir(repoRoot), s.Name)); err != nil {
			continue
		}
		present[s.Name] = true
		if err := e.mirrorSkill(repoRoot, m, *t, s.Name); err != nil {
			return nil, err
		}
		rep.Copied = append(rep.Copied, s.Name)
	}
	// Remove mirrored copies of skills the manifest no longer has: anything
	// in a mirror that matches a manifest-declared or previously-locked name
	// but isn't present in the primary any more.
	for _, mt := range mts {
		entries, _ := os.ReadDir(mt.SkillsDir(repoRoot))
		for _, ent := range entries {
			name := ent.Name()
			if present[name] {
				continue
			}
			if _, declared := m.Skills[name]; declared || wasMirrored(mt, repoRoot, name, lf) {
				_ = os.RemoveAll(filepath.Join(mt.SkillsDir(repoRoot), name))
				rep.Removed = append(rep.Removed, mt.ID+"/"+name)
			}
		}
	}
	sort.Strings(rep.Copied)
	sort.Strings(rep.Removed)
	return rep, nil
}

// wasMirrored is a conservative check: a folder in a mirror is only treated as
// Skell-managed if the primary lock still mentions it.
func wasMirrored(_ target.Target, _ string, name string, lf *lockfile.LockFile) bool {
	return lf.FindSkill(name) != nil
}

// SetMirrors updates the manifest's mirror list (adding and removing target
// ids) and refreshes the mirror folders. Removed mirrors have the skills
// Skell placed there deleted.
func (e *Engine) SetMirrors(repoRoot string, add, remove []string) (*MirrorReport, error) {
	m, t, err := manifest.ResolveWithTarget(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("no manifest found in %s — run 'skell init' first: %w", repoRoot, err)
	}
	for _, id := range add {
		mt, err := target.Lookup(id)
		if err != nil {
			return nil, err
		}
		if mt.ID == t.ID {
			return nil, fmt.Errorf("%s is already this manifest's primary target", mt.ID)
		}
		if !slices.Contains(m.Mirrors, mt.ID) {
			m.Mirrors = append(m.Mirrors, mt.ID)
		}
	}
	lf, _ := lockfile.Read(lockfile.PathFor(repoRoot, *t))
	for _, id := range remove {
		mt, err := target.Lookup(id)
		if err != nil {
			return nil, err
		}
		m.Mirrors = slices.DeleteFunc(m.Mirrors, func(s string) bool { return s == mt.ID })
		if lf != nil {
			for _, s := range lf.Skills {
				_ = os.RemoveAll(filepath.Join(mt.SkillsDir(repoRoot), s.Name))
			}
		}
	}
	sort.Strings(m.Mirrors)
	if err := manifest.Write(manifest.LocalPathFor(repoRoot, *t), m); err != nil {
		return nil, err
	}
	return e.RefreshMirrors(repoRoot)
}

// staleMirrors lists "<target>/<skill>" entries whose mirror copy is missing
// or differs from the primary. Used by doctor.
func staleMirrors(repoRoot string, m *manifest.Manifest, t target.Target, lf *lockfile.LockFile) []string {
	var out []string
	for _, mt := range mirrorTargets(m, t) {
		for _, s := range lf.Skills {
			src := filepath.Join(t.SkillsDir(repoRoot), s.Name)
			dst := filepath.Join(mt.SkillsDir(repoRoot), s.Name)
			if _, err := os.Lstat(src); err != nil {
				continue
			}
			if _, ok := linkDestination(src); ok {
				if _, err := os.Lstat(dst); err != nil {
					out = append(out, mt.ID+"/"+s.Name)
				}
				continue
			}
			a, errA := hasher.HashDir(src)
			b, errB := hasher.HashDir(dst)
			if errA != nil || errB != nil || a != b {
				out = append(out, mt.ID+"/"+s.Name)
			}
		}
	}
	return out
}

// detectManaged is target.Detect minus the targets that are only mirrors of
// another target's manifest, so aggregated views list each skill once.
func detectManaged(repoRoot string) []target.Target {
	mirrors := mirrorIDs(repoRoot)
	var out []target.Target
	for _, t := range target.Detect(repoRoot) {
		if !mirrors[t.ID] {
			out = append(out, t)
		}
	}
	return out
}
