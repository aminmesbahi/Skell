package model

import "testing"

func TestVersionLabel(t *testing.T) {
	cases := []struct{ version, commit, want string }{
		{"1.2.0", "dbd4588f9e1033efb41dad4bef2f7947c8993d44", "1.2.0"}, // a version wins
		{"", "dbd4588f9e1033efb41dad4bef2f7947c8993d44", "@dbd4588"},   // else the short commit
		{"", "abc", "@abc"}, // short commits are kept whole
		{"", "", "-"},       // local folders: nothing to show
	}
	for _, c := range cases {
		if got := VersionLabel(c.version, c.commit); got != c.want {
			t.Errorf("VersionLabel(%q, %q) = %q, want %q", c.version, c.commit, got, c.want)
		}
	}
}
