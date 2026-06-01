package engine

import (
	"fmt"
	"strings"
)

// ValidateSkillName rejects skill names that are empty or that could be used to
// escape the skills directory when joined into a filesystem path. This is a
// defence-in-depth guard: several commands (notably `remove`, which deletes the
// directory directly without first consulting the registry) build a path from a
// raw, user-supplied name. Without this check `skell remove ../../something`
// would resolve outside the repository via filepath.Join's path cleaning.
//
// A valid skill name is a single path segment: no separators, no "." or ".."
// components, no NUL bytes, and not absolute.
func ValidateSkillName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("skill name is empty")
	}
	if name == "." || name == ".." {
		return fmt.Errorf("invalid skill name %q", name)
	}
	if strings.ContainsAny(name, "/\\") {
		return fmt.Errorf("invalid skill name %q: must not contain path separators", name)
	}
	if strings.ContainsRune(name, 0) {
		return fmt.Errorf("invalid skill name %q: contains NUL byte", name)
	}
	if strings.HasPrefix(name, "-") {
		return fmt.Errorf("invalid skill name %q: must not start with '-'", name)
	}
	return nil
}
