package skell

import (
	"os"
	"sort"
	"strings"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/spf13/cobra"
)

// completionRepo returns the --repo flag value when set, else the cwd.
func completionRepo(cmd *cobra.Command) string {
	if fl := cmd.Flags().Lookup("repo"); fl != nil {
		if arr, err := cmd.Flags().GetStringArray("repo"); err == nil && len(arr) > 0 {
			return arr[0]
		}
		if s := fl.Value.String(); s != "" && s != "[]" {
			return strings.Trim(s, "[]")
		}
	}
	cwd, _ := os.Getwd()
	return cwd
}

func filterPrefix(names []string, prefix string, exclude []string) []string {
	skip := map[string]bool{}
	for _, e := range exclude {
		skip[e] = true
	}
	seen := map[string]bool{}
	var out []string
	for _, n := range names {
		if strings.HasPrefix(n, prefix) && !skip[n] && !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	sort.Strings(out)
	return out
}

// completeInstalledSkills completes names of skills installed in the repo.
func completeInstalledSkills(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	eng := engine.New(defaultCacheRoot())
	skills, err := eng.ListFor(completionRepo(cmd), "")
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	names := make([]string, 0, len(skills))
	for _, s := range skills {
		names = append(names, s.Name)
	}
	return filterPrefix(names, toComplete, args), cobra.ShellCompDirectiveNoFileComp
}

// completeRegistrySkills completes names of skills offered by the repo's
// configured sources (and global sources).
func completeRegistrySkills(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	m, err := manifest.Resolve(completionRepo(cmd))
	if err != nil {
		m = &manifest.Manifest{}
	}
	eng := engine.New(defaultCacheRoot())
	skills, err := eng.ListRegistry(m)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	names := make([]string, 0, len(skills))
	for _, s := range skills {
		names = append(names, s.Name)
	}
	return filterPrefix(names, toComplete, args), cobra.ShellCompDirectiveNoFileComp
}

// completeTargets completes agent target ids.
func completeTargets(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	ids := []string{"claude", "codex", "copilot", "cursor", "windsurf", "opencode", "cline", "grok"}
	return filterPrefix(ids, toComplete, args), cobra.ShellCompDirectiveNoFileComp
}

// completeCatalogIDs completes catalog source ids for 'skell add'.
func completeCatalogIDs(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var ids []string
	for _, s := range loadCatalog(false).Sources {
		ids = append(ids, s.ID)
	}
	return filterPrefix(ids, toComplete, args), cobra.ShellCompDirectiveDefault
}
