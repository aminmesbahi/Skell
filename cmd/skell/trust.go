package skell

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/aminmesbahi/skell/internal/config"
	"github.com/aminmesbahi/skell/internal/signing"
	"github.com/spf13/cobra"
)

func trustStorePath() (string, error) {
	home, err := config.DefaultRoot()
	if err != nil {
		return "", err
	}
	return signing.StorePath(home), nil
}

func loadTrustStore() (*signing.TrustStore, error) {
	p, err := trustStorePath()
	if err != nil {
		return nil, err
	}
	return signing.LoadStore(p)
}

// ── keygen ───────────────────────────────────────────────────────────────

func newKeygenCmd() *cobra.Command {
	var out string
	cmd := &cobra.Command{
		Use:   "keygen <publisher>",
		Short: "Create a signing key for publishing skills",
		Long: `Creates an Ed25519 key pair for a publisher (your name, team or company):
<publisher>.key is private — keep it secret and never commit it — and
<publisher>.pub is the public half you hand to people who want to trust your
skills ('skell trust add').

Keys are written to --out, by default the 'keys' folder inside the Skell home.
An existing key is never overwritten.`,
		Example: `  skell keygen acme
  skell keygen acme --out ~/secure/keys`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if out == "" {
				home, err := config.DefaultRoot()
				if err != nil {
					return err
				}
				out = filepath.Join(home, "keys")
			}
			if err := engineValidName(args[0]); err != nil {
				return err
			}
			kp, err := signing.GenerateKey(args[0])
			if err != nil {
				return err
			}
			keyPath, pubPath, err := signing.WriteKeyFiles(out, kp)
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(w, "  %s  private key  %s   (keep it secret)\n", markOK, keyPath)
			_, _ = fmt.Fprintf(w, "  %s  public key   %s\n", markOK, pubPath)
			_, _ = fmt.Fprintf(w, "     key id        %s\n\n", signing.KeyID(kp.Public))
			_, _ = fmt.Fprintf(w, "     sign a skill:   skell sign ./my-skill --key %s --publisher %s\n", keyPath, args[0])
			_, _ = fmt.Fprintf(w, "     others trust it: skell trust add %s %s\n", args[0], pubPath)
			return nil
		},
	}
	cmd.Flags().StringVar(&out, "out", "", "Folder to write the key files to")
	return cmd
}

// engineValidName keeps publisher names usable as file names.
func engineValidName(name string) error {
	for _, r := range name {
		ok := r == '-' || r == '_' || r == '.' || (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
		if !ok {
			return fmt.Errorf("publisher %q: use letters, digits, '-', '_' or '.'", name)
		}
	}
	if name == "" || name[0] == '.' {
		return fmt.Errorf("publisher name is required")
	}
	return nil
}

// ── sign ─────────────────────────────────────────────────────────────────

type signResult struct {
	Skill     string `json:"skill"`
	Path      string `json:"path"`
	Publisher string `json:"publisher"`
	KeyID     string `json:"key_id"`
	Hash      string `json:"content_hash"`
}

func newSignCmd() *cobra.Command {
	var keyPath, publisher string
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "sign <path>",
		Short: "Sign a skill (or every skill in a folder) so others can verify who published it",
		Long: `Writes SKILL.sig into a skill folder. The signature covers every file in the
skill, so any later change — by anyone — makes 'skell verify' report it as
tampered. Re-sign after every edit, ideally as part of your release.

<path> is a skill folder, or a folder containing several skills (a source
repository), in which case all of them are signed.`,
		Example: `  skell sign ./release-notes --key ~/.skell/keys/acme.key --publisher acme
  skell sign ~/src/team-skills --key ~/.skell/keys/acme.key --publisher acme`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if keyPath == "" || publisher == "" {
				return fmt.Errorf("--key and --publisher are required (create a key with 'skell keygen <publisher>')")
			}
			key, err := signing.ReadPrivateKey(keyPath)
			if err != nil {
				return err
			}
			refs, err := pathSkillRefs(args[0])
			if err != nil {
				return err
			}
			var out []signResult
			for _, r := range refs {
				sig, err := signing.Sign(r.Dir, publisher, key)
				if err != nil {
					return fmt.Errorf("%s: %w", r.Name, err)
				}
				out = append(out, signResult{Skill: r.Name, Path: r.Dir, Publisher: publisher, KeyID: sig.KeyID, Hash: sig.ContentHash})
			}
			w := cmd.OutOrStdout()
			if jsonOut {
				return json.NewEncoder(w).Encode(out)
			}
			for _, s := range out {
				_, _ = fmt.Fprintf(w, "  %s  signed %s as %s (key %s)\n", markOK, s.Skill, s.Publisher, s.KeyID)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&keyPath, "key", "", "Private key file from 'skell keygen'")
	cmd.Flags().StringVar(&publisher, "publisher", "", "Publisher name recorded in the signature")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output results as JSON")
	return cmd
}

// ── verify ───────────────────────────────────────────────────────────────

type verifyEntry struct {
	Repo  string `json:"repo,omitempty"`
	Skill string `json:"skill"`
	signing.Result
}

func newVerifyCmd() *cobra.Command {
	var f repoFlags
	var path, targetID string
	var requireSigned bool
	cmd := &cobra.Command{
		Use:   "verify [skill-name...]",
		Short: "Check who published installed skills and that nobody has changed them",
		Long: `Checks each skill's signature against the publishers you trust
('skell trust list'). Results:

  trusted        signed by a publisher you trust, and unchanged
  untrusted-key  validly signed, but by someone you have not trusted
  tampered       files changed after signing
  invalid        the signature file is damaged or forged
  unsigned       no signature (most skills today)

Exits non-zero when a skill is tampered or invalid. With --require-signed it
also fails on unsigned and untrusted skills.`,
		Example: `  skell verify
  skell verify pdf --json
  skell verify --path ./my-skill
  skell verify --require-signed      # for CI in a locked-down team`,
		ValidArgsFunction: completeInstalledSkills,
		RunE: func(cmd *cobra.Command, args []string) error {
			refs, err := collectSkillRefs(cmd, f, path, targetID, args)
			if err != nil {
				return err
			}
			store, err := loadTrustStore()
			if err != nil {
				return err
			}
			entries := []verifyEntry{}
			failed := 0
			for _, r := range refs {
				res := signing.Verify(r.Dir, store)
				entries = append(entries, verifyEntry{Repo: r.Repo, Skill: r.Name, Result: res})
				switch res.Status {
				case signing.StatusTampered, signing.StatusInvalid:
					failed++
				case signing.StatusUnsigned, signing.StatusUntrusted:
					if requireSigned {
						failed++
					}
				}
			}
			w := cmd.OutOrStdout()
			if f.jsonOut {
				enc := json.NewEncoder(w)
				enc.SetIndent("", "  ")
				if err := enc.Encode(entries); err != nil {
					return err
				}
			} else {
				printVerify(w, entries)
			}
			if failed > 0 {
				return fmt.Errorf("%d skill(s) failed verification", failed)
			}
			return nil
		},
	}
	bindRepoFlags(cmd, &f)
	cmd.Flags().StringVar(&path, "path", "", "Verify a skill folder (or a folder of skills) instead of installed skills")
	cmd.Flags().StringVar(&targetID, "target", "", "Agent platform to check")
	cmd.Flags().BoolVar(&requireSigned, "require-signed", false, "Also fail on unsigned skills and untrusted publishers")
	return cmd
}

func printVerify(w io.Writer, entries []verifyEntry) {
	if len(entries) == 0 {
		_, _ = fmt.Fprintln(w, "  no skills to verify")
		return
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "  \tSKILL\tSIGNATURE\tPUBLISHER")
	for _, e := range entries {
		mark := markWarn
		switch e.Status {
		case signing.StatusTrusted:
			mark = markOK
		case signing.StatusTampered, signing.StatusInvalid:
			mark = markBad
		case signing.StatusUnsigned:
			mark = "-"
		}
		_, _ = fmt.Fprintf(tw, "  %s\t%s\t%s\t%s\n", mark, e.Skill, e.Status, dash(e.Publisher))
	}
	_ = tw.Flush()
	untrusted := false
	for _, e := range entries {
		switch e.Status {
		case signing.StatusTampered, signing.StatusInvalid:
			_, _ = fmt.Fprintf(w, "\n  %s %s: %s\n", markBad, e.Skill, e.Detail)
		case signing.StatusUntrusted:
			untrusted = true
		}
	}
	if untrusted {
		_, _ = fmt.Fprintln(w, "\n  Validly signed by a publisher you have not trusted yet — see 'skell trust add'.")
	}
}

// ── trust ────────────────────────────────────────────────────────────────

func newTrustCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trust",
		Short: "Choose which skill publishers you trust",
		Long: `Manages the publishers whose signatures Skell accepts. The list lives in
trusted-keys.json in the Skell home, so a company can ship it with the rest of
its Skell configuration. With 'require-signed = true' in config.toml, only
skills signed by a trusted publisher can be installed or upgraded.`,
	}
	cmd.AddCommand(newTrustAddCmd(), newTrustListCmd(), newTrustRemoveCmd())
	return cmd
}

func newTrustAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <publisher> <public-key-file | key>",
		Short: "Trust a publisher's public key",
		Example: `  skell trust add acme ./acme.pub
  skell trust add acme "skell-public-key AbC…=="`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			text := args[1]
			if data, err := os.ReadFile(args[1]); err == nil { //nolint:gosec
				text = string(data)
			}
			pub, err := signing.DecodePublic(text)
			if err != nil {
				return err
			}
			store, err := loadTrustStore()
			if err != nil {
				return err
			}
			k, err := store.Add(args[0], pub)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s  trusting %s (key %s)\n", markOK, k.Publisher, k.KeyID)
			return nil
		},
	}
}

func newTrustListCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Show trusted publishers",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := loadTrustStore()
			if err != nil {
				return err
			}
			keys := store.Keys
			if keys == nil {
				keys = []signing.TrustedKey{}
			}
			w := cmd.OutOrStdout()
			if jsonOut {
				return json.NewEncoder(w).Encode(keys)
			}
			if len(keys) == 0 {
				_, _ = fmt.Fprintln(w, "  no trusted publishers yet — add one with 'skell trust add <publisher> <key-file>'")
				return nil
			}
			tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
			_, _ = fmt.Fprintln(tw, "PUBLISHER\tKEY ID\tADDED")
			for _, k := range keys {
				_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\n", k.Publisher, k.KeyID, k.AddedAt)
			}
			return tw.Flush()
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func newTrustRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "remove <publisher | key-id>",
		Aliases: []string{"rm"},
		Short:   "Stop trusting a publisher",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := loadTrustStore()
			if err != nil {
				return err
			}
			n, err := store.Remove(args[0])
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s  no longer trusting %s (%d key)\n", markOK, args[0], n)
			return nil
		},
	}
}
