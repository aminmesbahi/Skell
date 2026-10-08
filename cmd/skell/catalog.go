package skell

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newCatalogCmd() *cobra.Command {
	var jsonOut, refresh bool

	cmd := &cobra.Command{
		Use:   "catalog [query]",
		Short: "Browse well-known skill sources you can add",
		Long: `Lists curated, well-known skill sources (git repositories of SKILL.md skills).
Add one to the current project with 'skell add <id>', then browse its skills
with 'skell search'.

The catalog ships with Skell and is refreshed from the Skell repository at most
once a day. Set SKELL_OFFLINE=1 to never contact the network, or
SKELL_CATALOG_URL to use your organisation's own catalog.`,
		Example: `  # Show every catalog source
  skell catalog

  # Filter by keyword
  skell catalog dotnet

  # Add a source by id, then search it
  skell add anthropic
  skell search pdf`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := ""
			if len(args) == 1 {
				query = args[0]
			}
			c := loadCatalog(refresh)
			sources := c.Search(query)
			w := cmd.OutOrStdout()
			if jsonOut {
				enc := json.NewEncoder(w)
				enc.SetIndent("", "  ")
				return enc.Encode(sources)
			}
			if len(sources) == 0 {
				_, _ = fmt.Fprintf(w, "  no catalog sources match %q\n", query)
				return nil
			}
			tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
			_, _ = fmt.Fprintln(tw, "ID\tNAME\tTAGS\tURL")
			for _, s := range sources {
				_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", s.ID, s.Name, strings.Join(s.Tags, ","), s.URL)
			}
			_ = tw.Flush()
			_, _ = fmt.Fprintf(w, "\n  add one with: skell add %s\n", sources[0].ID)
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	cmd.Flags().BoolVar(&refresh, "refresh", false, "Download the latest catalog now")
	return cmd
}
