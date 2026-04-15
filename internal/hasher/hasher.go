// Package hasher computes and compares SHA-256 content hashes for skill directories.
package hasher

// HashDir computes a deterministic SHA-256 hash over all files in a directory tree.
// Files are sorted by path before hashing to ensure reproducibility.
func HashDir(dirPath string) (string, error) {
	// TODO: implement using crypto/sha256
	panic("not implemented")
}

// Verify returns true if the SHA-256 hash of the directory matches the expected hash.
func Verify(dirPath, expectedHash string) (bool, error) {
	// TODO: implement
	panic("not implemented")
}
