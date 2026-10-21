// Package deps parses and checks the dependencies a skill declares in its
// SKILL.md frontmatter:
//
//	metadata:
//	  requires: "pdf, docx>=1.2.0"
//
// Each entry is a skill name with an optional version constraint (>=, >, <=,
// <, =, ^ or ~).
package deps

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Dep is one required skill.
type Dep struct {
	Name       string `json:"name"`
	Constraint string `json:"constraint,omitempty"` // e.g. ">=1.2.0"
}

// String renders the dependency the way it is written in frontmatter.
func (d Dep) String() string { return d.Name + d.Constraint }

var entryRx = regexp.MustCompile(`^([A-Za-z0-9][A-Za-z0-9._-]*)\s*(>=|<=|>|<|=|\^|~)?\s*v?([0-9][0-9A-Za-z.+-]*)?$`)

// Parse reads a comma- or newline-separated requires value.
func Parse(requires string) ([]Dep, error) {
	var out []Dep
	seen := map[string]bool{}
	for _, raw := range strings.FieldsFunc(requires, func(r rune) bool { return r == ',' || r == '\n' || r == ';' }) {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		m := entryRx.FindStringSubmatch(raw)
		if m == nil || (m[2] != "" && m[3] == "") || (m[2] == "" && m[3] != "") {
			return nil, fmt.Errorf("invalid dependency %q — write a skill name, optionally with a constraint like pdf>=1.2.0", raw)
		}
		d := Dep{Name: m[1]}
		if m[2] != "" {
			d.Constraint = m[2] + m[3]
		}
		if seen[d.Name] {
			continue
		}
		seen[d.Name] = true
		out = append(out, d)
	}
	return out, nil
}

// Satisfied reports whether an installed or available version meets the
// constraint. A dependency without a constraint accepts anything; a
// constrained one cannot be satisfied by a skill that declares no version.
func (d Dep) Satisfied(version string) bool {
	if d.Constraint == "" {
		return true
	}
	version = strings.TrimPrefix(strings.TrimSpace(version), "v")
	if version == "" {
		return false
	}
	var op, want string
	for _, o := range []string{">=", "<=", ">", "<", "=", "^", "~"} {
		if strings.HasPrefix(d.Constraint, o) {
			op, want = o, strings.TrimPrefix(d.Constraint, o)
			break
		}
	}
	c := Compare(version, want)
	switch op {
	case ">=":
		return c >= 0
	case "<=":
		return c <= 0
	case ">":
		return c > 0
	case "<":
		return c < 0
	case "=":
		return c == 0
	case "^": // same major version, at least want
		return c >= 0 && major(version) == major(want)
	case "~": // same major.minor, at least want
		return c >= 0 && major(version) == major(want) && minor(version) == minor(want)
	}
	return false
}

func parts(v string) []int {
	v = strings.TrimPrefix(v, "v")
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		v = v[:i]
	}
	var out []int
	for _, p := range strings.Split(v, ".") {
		n, _ := strconv.Atoi(p)
		out = append(out, n)
	}
	return out
}

func at(p []int, i int) int {
	if i < len(p) {
		return p[i]
	}
	return 0
}

func major(v string) int { return at(parts(v), 0) }
func minor(v string) int { return at(parts(v), 1) }

// Compare orders two dotted versions numerically: -1, 0 or 1.
func Compare(a, b string) int {
	pa, pb := parts(a), parts(b)
	n := len(pa)
	if len(pb) > n {
		n = len(pb)
	}
	for i := 0; i < n; i++ {
		switch x, y := at(pa, i), at(pb, i); {
		case x < y:
			return -1
		case x > y:
			return 1
		}
	}
	return 0
}
