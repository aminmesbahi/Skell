package skell

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecute_Success(t *testing.T) {
	oldRoot := rootCmd
	t.Cleanup(func() { rootCmd = oldRoot })

	rootCmd = newRootCmd()
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{})

	Execute()
	assert.Contains(t, out.String(), "skell version")
}