package skell

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aminmesbahi/skell/internal/catalog"
	"github.com/aminmesbahi/skell/internal/config"
	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/spf13/cobra"
)

// stdinReader is shared so successive prompts don't lose buffered input.
var stdinReader = bufio.NewReader(os.Stdin)

// confirm asks a yes/no question on the terminal. Non-interactive sessions
// get def without prompting.
func confirm(cmd *cobra.Command, question string, def bool) bool {
	if !isInteractive() {
		return def
	}
	hint := "[y/N]"
	if def {
		hint = "[Y/n]"
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s %s ", question, hint)
	line, err := stdinReader.ReadString('\n')
	if err != nil {
		return def
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		return def
	}
}

// printReview renders what a skill would bring into the repo.
func printReview(w io.Writer, rev *engine.Review) {
	s := rev.Skill
	_, _ = fmt.Fprintf(w, "\n  %s", s.Name)
	if s.Metadata.Version != "" {
		_, _ = fmt.Fprintf(w, " %s", s.Metadata.Version)
	}
	_, _ = fmt.Fprintf(w, "  (from %s", rev.Registry)
	if rev.Commit != "" {
		_, _ = fmt.Fprintf(w, " @ %s", shortSHA(rev.Commit))
	}
	_, _ = fmt.Fprintln(w, ")")
	if s.Description != "" {
		_, _ = fmt.Fprintf(w, "  %s\n", truncate(s.Description, 200))
	}
	if s.Metadata.Owner != "" || s.Metadata.Lifecycle != "" || s.License != "" {
		_, _ = fmt.Fprintf(w, "  owner: %s  lifecycle: %s  license: %s\n",
			dash(s.Metadata.Owner), dash(string(s.Metadata.Lifecycle)), dash(s.License))
	}
	if len(rev.Files) > 0 {
		_, _ = fmt.Fprintf(w, "  files: %d (%s)\n", len(rev.Files), humanBytes(rev.TotalBytes))
	}
	if len(rev.AllowedTools) > 0 {
		_, _ = fmt.Fprintf(w, "  allowed tools: %s\n", strings.Join(rev.AllowedTools, " "))
	}
	if len(rev.Scripts) > 0 {
		_, _ = fmt.Fprintf(w, "  scripts: %s\n", strings.Join(limit(rev.Scripts, 8), ", "))
	}
	if len(rev.Links) > 0 {
		_, _ = fmt.Fprintf(w, "  links: %d external URL(s)\n", len(rev.Links))
	}
	for _, warn := range rev.Warnings {
		_, _ = fmt.Fprintf(w, "  !  %s\n", warn)
	}
	_, _ = fmt.Fprintln(w)
}

// reviewAndConfirm shows the review for risky skills (always, when verbose)
// and asks before installing on a terminal. It returns false when the user
// declines. assumeYes skips the question; non-interactive runs proceed.
func reviewAndConfirm(cmd *cobra.Command, rev *engine.Review, assumeYes, verbose bool) bool {
	if rev == nil {
		return true
	}
	switch {
	case verbose || (rev.NeedsConfirmation() && !assumeYes):
		printReview(cmd.OutOrStdout(), rev)
	case rev.NeedsConfirmation():
		// Already approved (--yes): a one-line note keeps output readable.
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  !  %s: %s\n", rev.Skill.Name, strings.Join(rev.Warnings, "; "))
	}
	if !rev.NeedsConfirmation() || assumeYes || !isInteractive() {
		return true
	}
	return confirm(cmd, fmt.Sprintf("Install %s?", rev.Skill.Name), false)
}

// loadCatalog returns the catalog of well-known sources, preferring a fresh
// download but always succeeding offline.
func loadCatalog(refresh bool) *catalog.Catalog {
	home, _ := config.DefaultRoot()
	return catalog.Load(context.Background(), catalog.Options{
		CacheDir: home,
		Refresh:  refresh,
		Offline:  os.Getenv("SKELL_OFFLINE") != "",
	})
}

// catalogLookup resolves a catalog id to its URL for engine.ExpandAddArg.
func catalogLookup(id string) (string, bool) {
	s, ok := loadCatalog(false).Find(id)
	return s.URL, ok
}

func shortSHA(s string) string {
	if len(s) > 7 {
		return s[:7]
	}
	return s
}

func dash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

func truncate(s string, n int) string {
	if len([]rune(s)) <= n {
		return s
	}
	return string([]rune(s)[:n-1]) + "…"
}

func limit(items []string, n int) []string {
	if len(items) <= n {
		return items
	}
	return append(append([]string{}, items[:n]...), fmt.Sprintf("… +%d more", len(items)-n))
}

func humanBytes(n int64) string {
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(n)/(1<<10))
	default:
		return fmt.Sprintf("%d B", n)
	}
}
