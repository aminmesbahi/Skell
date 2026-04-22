package skell

import (
	"fmt"
	"os"

	"github.com/aminmesbahi/skell/internal/selfupdate"
	"github.com/aminmesbahi/skell/internal/version"
	"github.com/spf13/cobra"
)

func newSelfUpdateCmd() *cobra.Command {
	var checkOnly bool

	cmd := &cobra.Command{
		Use:   "selfupdate",
		Short: "Upgrade skell to the latest release from GitHub",
		Long: `Checks GitHub Releases for a newer version of skell.
If a newer version is found, downloads the platform-specific binary and
replaces the running executable in-place.

Use --check to only report whether an update is available without applying it.`,
		Example: `  # Check if a newer version exists (no download)
  skell selfupdate --check

  # Download and apply the latest release
  skell selfupdate`,
		RunE: func(cmd *cobra.Command, args []string) error {
			u := selfupdate.New("aminmesbahi", "Skell")
			if override := os.Getenv("SKELL_SELFUPDATE_API_URL"); override != "" {
				u.APIBaseURL = override
			}
			w := cmd.OutOrStdout()

			_, _ = fmt.Fprintf(w, "  current version: %s\n", version.Version)
			_, _ = fmt.Fprintf(w, "  checking for updates...\n")

			rel, err := u.LatestRelease()
			if err != nil {
				return err
			}

			if !selfupdate.IsNewer(version.Version, rel.TagName) {
				_, _ = fmt.Fprintf(w, "  ✓  already up to date (%s)\n", version.Version)
				return nil
			}

			_, _ = fmt.Fprintf(w, "  new version available: %s → %s\n", version.Version, rel.TagName)

			if checkOnly {
				_, _ = fmt.Fprintf(w, "  run 'skell selfupdate' (without --check) to apply\n")
				return nil
			}

			asset := selfupdate.FindAsset(rel)
			if asset == nil {
				return fmt.Errorf("no release asset for this platform (%s) — check https://github.com/aminmesbahi/Skell/releases/%s",
					selfupdate.ExpectedAssetName(), rel.TagName)
			}

			tmpPath := selfupdate.TempPath(asset.Name)
			_, _ = fmt.Fprintf(w, "  downloading %s...\n", asset.Name)

			if err := u.Download(asset, tmpPath); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(w, "  applying update...\n")
			if applyErr := selfupdate.Apply(tmpPath); applyErr != nil {
				_ = os.Remove(tmpPath)
				return applyErr
			}

			_, _ = fmt.Fprintf(w, "  ✓  updated to %s — restart skell to use the new version\n", rel.TagName)
			return nil
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "Only report if an update is available; do not download")
	return cmd
}
