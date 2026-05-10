package main

import (
	"os"
	"testing"
)

func TestMain_Smoke(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{"skell"}
	t.Cleanup(func() { os.Args = oldArgs })

	main()
}
