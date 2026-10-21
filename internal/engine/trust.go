package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aminmesbahi/skell/internal/registry"
	"github.com/aminmesbahi/skell/internal/signing"
	"github.com/aminmesbahi/skell/internal/skillscan"
)

// TrustStore returns the publishers whose signatures this installation
// accepts. Without a Skell home (tests) it is empty.
func (e *Engine) TrustStore() *signing.TrustStore {
	if e.homeRoot == "" {
		return &signing.TrustStore{}
	}
	ts, err := signing.LoadStore(signing.StorePath(e.homeRoot))
	if err != nil {
		return &signing.TrustStore{}
	}
	return ts
}

// ScanSkillDir runs the security scanner over one skill folder.
func (e *Engine) ScanSkillDir(dir string) (*skillscan.Report, error) {
	return skillscan.ScanDir(dir)
}

// VerifySkillDir checks the signature of one skill folder against the trust store.
func (e *Engine) VerifySkillDir(dir string) signing.Result {
	return signing.Verify(dir, e.TrustStore())
}

// annotateReview adds the security scan and signature check for the skill in
// dir to a review, with a warning when either deserves a look.
func (e *Engine) annotateReview(rev *Review, dir string) {
	if rev == nil || dir == "" {
		return
	}
	if rep, err := skillscan.ScanDir(dir); err == nil {
		rev.Risk = rep
		if rep.Level.AtLeast(skillscan.LevelMedium) {
			rev.Warnings = append(rev.Warnings, "security scan: "+rep.Summary()+" — run 'skell scan' for details")
		}
	}
	res := e.VerifySkillDir(dir)
	rev.Signature = &res
	switch res.Status {
	case signing.StatusTampered, signing.StatusInvalid:
		rev.Warnings = append(rev.Warnings, "signature check failed: "+res.Detail)
	}
}

// newScanRisk describes how the incoming copy of a skill scores worse than
// the installed one, or returns "" when it does not.
func newScanRisk(installedDir, incomingDir string) string {
	if incomingDir == "" {
		return ""
	}
	in, err := skillscan.ScanDir(incomingDir)
	if err != nil || !in.Level.AtLeast(skillscan.LevelMedium) {
		return ""
	}
	if cur, err := skillscan.ScanDir(installedDir); err == nil && !in.Level.Exceeds(cur.Level) {
		return ""
	}
	return "security scan: " + in.Summary()
}

// enforceTrustPolicy refuses a skill that breaks the organisation's
// max-risk or require-signed policy. It is a no-op when neither is set. The
// skill is staged in a temporary folder and inspected before anything is
// written to the project.
func (e *Engine) enforceTrustPolicy(reg registry.Registry, name, version string) error {
	maxRisk, ok := skillscan.ParseLevel(e.pol.MaxRisk)
	if !ok {
		return fmt.Errorf("policy: max-risk %q is not one of low, medium, high, critical", e.pol.MaxRisk)
	}
	if maxRisk == "" && !e.pol.RequireSigned {
		return nil
	}
	tmp, err := os.MkdirTemp("", "skell-trust-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(tmp) }()
	stage := filepath.Join(tmp, name)
	if err := e.provider.CopySkillTo(reg, name, version, stage); err != nil {
		return err
	}
	if maxRisk != "" {
		rep, err := skillscan.ScanDir(stage)
		if err != nil {
			return err
		}
		if rep.Level.Exceeds(maxRisk) {
			return fmt.Errorf("policy: %q is %s — above the allowed risk level %q (see 'skell scan %s')",
				name, rep.Summary(), maxRisk, name)
		}
	}
	if e.pol.RequireSigned {
		if res := signing.Verify(stage, e.TrustStore()); !res.OK() {
			why := string(res.Status)
			if res.Detail != "" {
				why += ": " + res.Detail
			}
			return fmt.Errorf("policy: %q must be signed by a trusted publisher (%s)", name, strings.TrimSpace(why))
		}
	}
	return nil
}
