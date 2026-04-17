package skell

import (
	"fmt"
	"io"
	"os"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/spf13/cobra"
)

func newInfoCmd() *cobra.Command {
	var source string

	cmd := &cobra.Command{
		Use:   "info <skill-name>",
		Short: "Show full metadata for a skill",
		Long:  "Displays the full metadata for a skill, including frontmatter fields and lock file state.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := os.Getwd()
			if err != nil {
				return err
			}

			eng := engine.New(defaultCacheRoot())
			result, err := eng.Info(repo, args[0], source)
			if err != nil {
				return err
			}

			printInfo(cmd.OutOrStdout(), args[0], result)
			return nil
		},
	}

	cmd.Flags().StringVar(&source, "source", "", "Show only: registry | local")
	return cmd
}

func printInfo(w io.Writer, skillName string, r *model.InfoResult) {
	fw := &fmtWriter{w: w}
	fw.printf("  name       %s\n", skillName)

	if r.Skill != nil {
		if r.Skill.Metadata.Version != "" {
			fw.printf("  version    %s\n", r.Skill.Metadata.Version)
		}
		if r.Skill.Description != "" {
			fw.printf("  desc       %s\n", r.Skill.Description)
		}
		if r.Skill.Metadata.Owner != "" {
			fw.printf("  owner      %s\n", r.Skill.Metadata.Owner)
		}
		if r.Skill.Metadata.Lifecycle != "" {
			fw.printf("  lifecycle  %s\n", r.Skill.Metadata.Lifecycle)
		}
		if r.Skill.Metadata.Scope != "" {
			fw.printf("  scope      %s\n", r.Skill.Metadata.Scope)
		}
		if r.Skill.Metadata.Tags != "" {
			fw.printf("  tags       %s\n", r.Skill.Metadata.Tags)
		}
		if r.Skill.Metadata.SourceRepo != "" {
			fw.printf("  source     %s\n", r.Skill.Metadata.SourceRepo)
		}
		if r.Skill.License != "" {
			fw.printf("  license    %s\n", r.Skill.License)
		}
	}

	if r.Lock != nil {
		fw.printf("  installed  %s\n", r.Lock.Version)
		fw.printf("  locked at  %s\n", r.Lock.InstalledAt)
		fw.printf("  hash       %s\n", r.Lock.ContentHash)
		if r.Lock.Pinned {
			fw.printf("  pinned     true\n")
		}
	}

	if r.Status != "" {
		fw.printf("  status     %s\n", r.Status)
	}
}

// fmtWriter records the first write error so callers can check it after a sequence of prints.
type fmtWriter struct {
	w   io.Writer
	err error
}

func (fw *fmtWriter) printf(format string, args ...any) {
	if fw.err != nil {
		return
	}
	_, fw.err = fmt.Fprintf(fw.w, format, args...)
}
