//go:build !windows

package main

import "os"

func sameFileRobust(left, right string) bool {
	if left == "" || right == "" {
		return false
	}
	if fi1, err := os.Stat(left); err == nil {
		if fi2, err := os.Stat(right); err == nil {
			return os.SameFile(fi1, fi2)
		}
	}
	return false
}
