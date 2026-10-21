package skell

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/spf13/cobra"
)

func newReportCmd() *cobra.Command {
	var f repoFlags
	var format, out, webhook string
	var checkUpdates bool
	var events int

	cmd := &cobra.Command{
		Use:   "report",
		Short: "See the skills of many repositories at once: versions, drift, risk, signatures, activity",
		Long: `Builds a fleet report over several repositories: which skills each one runs
and in which versions, who is outdated or hand-edited, the security risk and
signature status of every install, repositories that need attention, and
recent install/upgrade activity from the audit log.

By default it reads only local files (fast, works offline). Add
--check-updates to ask each source whether newer versions exist.

Formats: text (default), json, md, csv, html. The html report is a single
self-contained page you can open, attach to an email or archive.

--webhook posts the JSON report to an address you choose (https only, or
localhost). Nothing is sent anywhere unless you pass it.`,
		Example: `  # Every repo under ~/work, as a page you can share
  skell report --all-repos ~/work --format html --out fleet.html

  # Spreadsheet-friendly
  skell report --all-repos ~/work --format csv --out fleet.csv

  # Include update checks, ship to a dashboard
  skell report --all-repos ~/work --check-updates --webhook https://hooks.example.com/skell`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if f.jsonOut && format == "" {
				format = "json"
			}
			format = strings.ToLower(format)
			switch format {
			case "", "text", "json", "md", "markdown", "csv", "html":
			default:
				return fmt.Errorf("--format must be one of text, json, md, csv, html")
			}
			if webhook != "" {
				if err := checkWebhookURL(webhook); err != nil {
					return err
				}
			}

			eng := engine.New(defaultCacheRoot())
			opts := engine.FleetOptions{CheckUpdates: checkUpdates, Events: events}
			switch {
			case f.allRepos != "":
				opts.ScanRoot = f.allRepos
			default:
				repos, err := resolveRepos(f)
				if err != nil {
					return err
				}
				opts.Roots = repos
			}
			rep, err := eng.Fleet(opts)
			if err != nil {
				return err
			}

			body, err := renderFleet(rep, format)
			if err != nil {
				return err
			}
			if out != "" {
				if err := os.WriteFile(out, []byte(body), 0o644); err != nil { //nolint:gosec // a report is meant to be shared
					return err
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s  wrote %s (%d repos, %d installs)\n", markOK, out, rep.Summary.Repos, rep.Summary.Installs)
			} else {
				_, _ = io.WriteString(cmd.OutOrStdout(), body)
			}
			if webhook != "" {
				if err := postWebhook(webhook, rep); err != nil {
					return fmt.Errorf("report built, but sending it failed: %w", err)
				}
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  %s  sent report to %s\n", markOK, webhookHost(webhook))
			}
			return nil
		},
	}
	bindRepoFlags(cmd, &f)
	cmd.Flags().StringVar(&format, "format", "", "Output format: text | json | md | csv | html")
	cmd.Flags().StringVarP(&out, "out", "o", "", "Write the report to a file instead of the terminal")
	cmd.Flags().BoolVar(&checkUpdates, "check-updates", false, "Ask each source whether newer versions exist (uses the network)")
	cmd.Flags().IntVar(&events, "events", 20, "Recent audit-log entries to include (0 for none)")
	cmd.Flags().StringVar(&webhook, "webhook", "", "POST the JSON report to this https URL")
	return cmd
}

func renderFleet(rep *engine.FleetReport, format string) (string, error) {
	switch format {
	case "json":
		b, err := json.MarshalIndent(rep, "", "  ")
		return string(b) + "\n", err
	case "md", "markdown":
		return engine.FleetMarkdown(rep), nil
	case "csv":
		return engine.FleetCSV(rep), nil
	case "html":
		return engine.FleetHTML(rep)
	}
	var b bytes.Buffer
	printFleetText(&b, rep)
	return b.String(), nil
}

func printFleetText(w io.Writer, rep *engine.FleetReport) {
	s := rep.Summary
	_, _ = fmt.Fprintf(w, "\n  %d repositories · %d skills (%d installs)\n\n", s.Repos, s.UniqueSkills, s.Installs)
	_, _ = fmt.Fprintf(w, "  outdated %d   hand-edited %d   high risk %d   mixed versions %d   need attention %d\n\n",
		s.Outdated, s.Modified, s.HighRisk, s.VersionSpread, s.Unhealthy)

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "  \tREPOSITORY\tAGENT\tSKILLS\tHEALTH")
	for _, r := range rep.Repos {
		mark := markOK
		switch r.Health {
		case "problem":
			mark = markBad
		case "attention":
			mark = markWarn
		}
		health := r.Health
		if r.Note != "" {
			health += " (" + r.Note + ")"
		}
		_, _ = fmt.Fprintf(tw, "  %s\t%s\t%s\t%d\t%s\n", mark, r.Name, dash(r.Target), len(r.Skills), health)
	}
	_ = tw.Flush()

	var mixed []engine.FleetMatrixRow
	for _, row := range rep.Matrix {
		if row.Spread {
			mixed = append(mixed, row)
		}
	}
	if len(mixed) > 0 {
		_, _ = fmt.Fprintln(w, "\n  Skills on mixed versions:")
		for _, row := range limitRows(mixed, 8) {
			var vs []string
			for _, v := range row.Versions {
				vs = append(vs, fmt.Sprintf("%s ×%d", v.Version, v.Repos))
			}
			_, _ = fmt.Fprintf(w, "    %-24s %s\n", row.Name, strings.Join(vs, ", "))
		}
	}
	_, _ = fmt.Fprintln(w, "\n  Tip: --format html --out fleet.html makes a page you can share.")
	_, _ = fmt.Fprintln(w)
}

func limitRows(rows []engine.FleetMatrixRow, n int) []engine.FleetMatrixRow {
	if len(rows) > n {
		return rows[:n]
	}
	return rows
}

// checkWebhookURL accepts https URLs, and http only for localhost.
func checkWebhookURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return fmt.Errorf("--webhook: %q is not a valid URL", raw)
	}
	host := u.Hostname()
	local := host == "localhost" || host == "127.0.0.1" || host == "::1"
	if u.Scheme != "https" && !(u.Scheme == "http" && local) {
		return fmt.Errorf("--webhook must be an https URL (http is only allowed for localhost)")
	}
	return nil
}

func webhookHost(raw string) string {
	if u, err := url.Parse(raw); err == nil {
		return u.Host
	}
	return raw
}

func postWebhook(raw string, rep *engine.FleetReport) error {
	body, err := json.Marshal(map[string]any{"source": "skell", "report": rep})
	if err != nil {
		return err
	}
	client := &http.Client{
		Timeout: 15 * time.Second,
		// Never follow a redirect to some other host with the report.
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, raw, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "skell-report")
	resp, err := client.Do(req) //nolint:gosec // the destination is chosen explicitly by the person running the command
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("the server answered %s", resp.Status)
	}
	return nil
}
