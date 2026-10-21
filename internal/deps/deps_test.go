package deps

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	got, err := Parse("pdf, docx>=1.2.0 ; xlsx ^2.0\nPDF")
	require.NoError(t, err)
	assert.Equal(t, []Dep{{"pdf", ""}, {"docx", ">=1.2.0"}, {"xlsx", "^2.0"}, {"PDF", ""}}, got)
	assert.Equal(t, "docx>=1.2.0", got[1].String())

	none, err := Parse("  ")
	require.NoError(t, err)
	assert.Empty(t, none)
}

func TestParseRejectsNonsense(t *testing.T) {
	for _, bad := range []string{">=1.0", "pdf>=", "pdf 1.0", "pdf!"} {
		_, err := Parse(bad)
		assert.Error(t, err, bad)
	}
}

func TestSatisfied(t *testing.T) {
	cases := []struct {
		dep, version string
		want         bool
	}{
		{"pdf", "", true},
		{"pdf", "0.1", true},
		{"pdf>=1.2.0", "1.2.0", true},
		{"pdf>=1.2.0", "1.10.0", true},
		{"pdf>=1.2.0", "1.1.9", false},
		{"pdf>=1.2.0", "", false},
		{"pdf>1.0", "1.0", false},
		{"pdf<2", "1.9", true},
		{"pdf<=2.0.0", "2.0.1", false},
		{"pdf=1.0", "v1.0.0", true},
		{"pdf^1.2", "1.9.0", true},
		{"pdf^1.2", "2.0.0", false},
		{"pdf~1.2", "1.2.7", true},
		{"pdf~1.2", "1.3.0", false},
	}
	for _, c := range cases {
		ds, err := Parse(c.dep)
		require.NoError(t, err, c.dep)
		assert.Equal(t, c.want, ds[0].Satisfied(c.version), "%s vs %s", c.dep, c.version)
	}
}

func TestCompare(t *testing.T) {
	assert.Equal(t, 0, Compare("1.0", "1.0.0"))
	assert.Equal(t, -1, Compare("1.9", "1.10"))
	assert.Equal(t, 1, Compare("2.0.0-beta", "1.9.9"))
}
