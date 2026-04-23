package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// SkellResult mirrors the JSON output contract used by the frontend.
type SkellResult struct {
	Stdout  string `json:"stdout"`
	Stderr  string `json:"stderr"`
	Success bool   `json:"success"`
}

// FileEntry represents a filesystem entry (file or directory).
type FileEntry struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Path  string `json:"path"`
}

// App is the Wails application struct. All exported methods are bound to the frontend.
type App struct {
	ctx context.Context
}

// NewApp creates a new App instance.
func NewApp() *App {
	return &App{}
}

// startup is called by Wails when the application starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// skellBin returns the path to the skell binary, searching PATH.
func skellBin() (string, error) {
	return exec.LookPath("skell")
}

// RunSkell executes the skell CLI with the provided arguments and returns stdout/stderr.
func (a *App) RunSkell(args []string) SkellResult {
	bin, err := skellBin()
	if err != nil {
		return SkellResult{
			Stderr:  "skell binary not found in PATH. Install skell first: https://github.com/aminmesbahi/Skell",
			Success: false,
		}
	}

	cmd := exec.Command(bin, args...) //nolint:gosec // args come from trusted frontend
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	return SkellResult{
		Stdout:  stdout.String(),
		Stderr:  stderr.String(),
		Success: err == nil,
	}
}

// ReadFileContent reads and returns the contents of a file.
func (a *App) ReadFileContent(path string) (string, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ListDirectory returns the immediate children of a directory.
func (a *App) ListDirectory(path string) ([]FileEntry, error) {
	entries, err := os.ReadDir(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	result := make([]FileEntry, 0, len(entries))
	for _, e := range entries {
		result = append(result, FileEntry{
			Name:  e.Name(),
			IsDir: e.IsDir(),
			Path:  filepath.Join(path, e.Name()),
		})
	}
	return result, nil
}

// SkellVersion returns the output of `skell version`.
func (a *App) SkellVersion() string {
	r := a.RunSkell([]string{"version"})
	return strings.TrimSpace(r.Stdout)
}

// SelectDirectory opens a native directory picker dialog and returns the selected path.
func (a *App) SelectDirectory() string {
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Repository",
	})
	if err != nil {
		return ""
	}
	return path
}

// AuditLogPath returns the platform-correct path to ~/.skell/audit.log.
func (a *App) AuditLogPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".skell", "audit.log")
}

// IsRepoInitialized returns true when the given directory contains a Skell
// manifest (.claude/skell.toml), meaning `skell init` has already been run.
func (a *App) IsRepoInitialized(repoPath string) bool {
	manifest := filepath.Join(filepath.Clean(repoPath), ".claude", "skell.toml")
	_, err := os.Stat(manifest)
	return err == nil
}

// GlobalRootDir returns the global Skell root directory (~/.skell) and ensures
// the global manifest exists so that `skell search --repo <path>` can resolve it.
func (a *App) GlobalRootDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	root := filepath.Join(home, ".skell")
	manifestDir := filepath.Join(root, ".claude")
	manifestPath := filepath.Join(manifestDir, "skell.toml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		if mkErr := os.MkdirAll(manifestDir, 0700); mkErr == nil {
			_ = os.WriteFile(manifestPath, []byte("[registries]\n[skills]\n"), 0600)
		}
	}
	return root
}

// ──────────────────────────────────────────────────────────────────────────────
// Metadata contribution feature
// ──────────────────────────────────────────────────────────────────────────────

// SkillMetadataFields holds the user-editable subset of SKILL.md frontmatter.
type SkillMetadataFields struct {
	Description string `json:"description"`
	Tags        string `json:"tags"`
	Lifecycle   string `json:"lifecycle"`
	Owner       string `json:"owner"`
}

// ContributeParams describes a metadata-contribution PR request.
type ContributeParams struct {
	InstalledPath string              `json:"installedPath"`
	SourceRepo    string              `json:"sourceRepo"`
	SkillName     string              `json:"skillName"`
	Metadata      SkillMetadataFields `json:"metadata"`
	GithubToken   string              `json:"githubToken"`
}

// ContributeResult is returned after attempting to open a PR.
type ContributeResult struct {
	PrURL   string `json:"prUrl"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// ReadSkillMetadata parses the SKILL.md in the given directory and returns
// the editable metadata fields.
func (a *App) ReadSkillMetadata(installedPath string) (SkillMetadataFields, error) {
	skillFile := filepath.Join(filepath.Clean(installedPath), "SKILL.md")
	data, err := os.ReadFile(skillFile)
	if err != nil {
		return SkillMetadataFields{}, err
	}
	return parseSkillMetadataFields(strings.ReplaceAll(string(data), "\r\n", "\n")), nil
}

// parseSkillMetadataFields extracts editable fields from SKILL.md YAML frontmatter.
func parseSkillMetadataFields(content string) SkillMetadataFields {
	var f SkillMetadataFields
	if !strings.HasPrefix(content, "---\n") {
		return f
	}
	endIdx := strings.Index(content[4:], "\n---")
	if endIdx == -1 {
		return f
	}
	fm := content[4 : 4+endIdx]
	for _, line := range strings.Split(fm, "\n") {
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		switch key {
		case "description":
			f.Description = val
		case "tags":
			f.Tags = val
		case "lifecycle":
			f.Lifecycle = val
		case "owner":
			f.Owner = val
		}
	}
	return f
}

// ContributeMetadata creates a GitHub PR to improve the metadata of a skill.
// It detects repo ownership automatically and forks if the authenticated user
// does not own the original repo.
func (a *App) ContributeMetadata(params ContributeParams) ContributeResult {
	token, err := resolveGitHubToken(params.GithubToken)
	if err != nil {
		return ContributeResult{Error: "no GitHub token: " + err.Error()}
	}

	ghURL, err := parseGitHubRepoURL(params.SourceRepo)
	if err != nil {
		return ContributeResult{Error: "invalid source repo URL: " + err.Error()}
	}

	login, err := ghGetLogin(token)
	if err != nil {
		return ContributeResult{Error: "GitHub auth failed: " + err.Error()}
	}

	repoData, err := ghAPIGet(token, fmt.Sprintf("https://api.github.com/repos/%s/%s", ghURL.Owner, ghURL.Repo))
	if err != nil {
		return ContributeResult{Error: "failed to get repo info: " + err.Error()}
	}
	defaultBranch, _ := repoData["default_branch"].(string)
	if defaultBranch == "" {
		defaultBranch = "main"
	}

	workOwner := ghURL.Owner
	if !strings.EqualFold(login, ghURL.Owner) {
		if _, err = ghAPIPost(token, fmt.Sprintf("https://api.github.com/repos/%s/%s/forks", ghURL.Owner, ghURL.Repo), map[string]any{}); err != nil {
			return ContributeResult{Error: "failed to fork repo: " + err.Error()}
		}
		workOwner = login
		time.Sleep(5 * time.Second) // wait for GitHub to provision the fork
	}

	sha, err := ghGetBranchSHA(token, workOwner, ghURL.Repo, defaultBranch)
	if err != nil {
		return ContributeResult{Error: "failed to get branch SHA: " + err.Error()}
	}

	branchName := "fix/metadata-" + params.SkillName
	// ignore 422 (branch already exists)
	_, _ = ghAPIPost(token, fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs", workOwner, ghURL.Repo), map[string]any{
		"ref": "refs/heads/" + branchName,
		"sha": sha,
	})

	filePath := "SKILL.md"
	if ghURL.SubPath != "" {
		filePath = ghURL.SubPath + "/SKILL.md"
	}
	fileContent, fileSHA, err := ghGetFileContent(token, workOwner, ghURL.Repo, filePath, defaultBranch)
	if err != nil {
		return ContributeResult{Error: "failed to read SKILL.md from GitHub: " + err.Error()}
	}

	updated := applyFrontmatterEdits(fileContent, params.Metadata)

	encoded := base64.StdEncoding.EncodeToString([]byte(updated))
	_, err = ghAPIPut(token,
		fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", workOwner, ghURL.Repo, filePath),
		map[string]any{
			"message": "fix(metadata): update metadata for " + params.SkillName,
			"content": encoded,
			"sha":     fileSHA,
			"branch":  branchName,
		})
	if err != nil {
		return ContributeResult{Error: "failed to commit changes: " + err.Error()}
	}

	prData, err := ghAPIPost(token,
		fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", ghURL.Owner, ghURL.Repo),
		map[string]any{
			"title": "fix(metadata): update metadata for " + params.SkillName,
			"body":  "Automated metadata improvement contributed via [Skell](https://github.com/aminmesbahi/Skell) GUI.\n\nImproves skill metadata: description, tags, lifecycle, and owner fields.",
			"head":  workOwner + ":" + branchName,
			"base":  defaultBranch,
		})
	if err != nil {
		return ContributeResult{Error: "failed to create PR: " + err.Error()}
	}

	prURL, _ := prData["html_url"].(string)
	if prURL == "" {
		return ContributeResult{Error: "PR created but no URL returned"}
	}
	return ContributeResult{PrURL: prURL, Success: true}
}

// ── GitHub API helpers ────────────────────────────────────────────────────────

type ghRepoURL struct {
	Owner   string
	Repo    string
	Branch  string
	SubPath string
}

func parseGitHubRepoURL(rawURL string) (ghRepoURL, error) {
	rawURL = strings.TrimSpace(rawURL)
	for _, prefix := range []string{"https://github.com/", "http://github.com/"} {
		rawURL = strings.TrimPrefix(rawURL, prefix)
	}
	parts := strings.Split(rawURL, "/")
	if len(parts) < 2 {
		return ghRepoURL{}, fmt.Errorf("expected github.com/owner/repo, got: %s", rawURL)
	}
	owner := parts[0]
	repo := strings.TrimSuffix(parts[1], ".git")
	var branch, subPath string
	if len(parts) > 3 && parts[2] == "tree" {
		branch = parts[3]
		if len(parts) > 4 {
			subPath = strings.Join(parts[4:], "/")
		}
	}
	return ghRepoURL{Owner: owner, Repo: repo, Branch: branch, SubPath: subPath}, nil
}

func resolveGitHubToken(provided string) (string, error) {
	if provided != "" {
		return provided, nil
	}
	cmd := exec.Command("git", "credential", "fill") //nolint:gosec
	cmd.Stdin = strings.NewReader("protocol=https\nhost=github.com\n\n")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git credential fill failed: %w", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		if after, ok := strings.CutPrefix(line, "password="); ok {
			return strings.TrimSpace(after), nil
		}
	}
	return "", fmt.Errorf("no password found via git credential fill")
}

func ghDo(method, apiURL, token string, body io.Reader) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(context.Background(), method, apiURL, body)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close() //nolint:errcheck
	data, err := io.ReadAll(resp.Body)
	return data, resp.StatusCode, err
}

func ghAPIGet(token, apiURL string) (map[string]any, error) {
	data, status, err := ghDo("GET", apiURL, token, nil)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	_ = json.Unmarshal(data, &result)
	if status < 200 || status >= 300 {
		if msg, ok := result["message"].(string); ok {
			return nil, fmt.Errorf("GitHub API (%d): %s", status, msg)
		}
		return nil, fmt.Errorf("GitHub API returned status %d", status)
	}
	return result, nil
}

func ghAPIPost(token, apiURL string, payload map[string]any) (map[string]any, error) {
	body, _ := json.Marshal(payload)
	data, status, err := ghDo("POST", apiURL, token, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	var result map[string]any
	_ = json.Unmarshal(data, &result)
	if status < 200 || status >= 300 {
		if msg, ok := result["message"].(string); ok {
			return nil, fmt.Errorf("GitHub API (%d): %s", status, msg)
		}
		return nil, fmt.Errorf("GitHub API returned status %d", status)
	}
	return result, nil
}

func ghAPIPut(token, apiURL string, payload map[string]any) (map[string]any, error) {
	body, _ := json.Marshal(payload)
	data, status, err := ghDo("PUT", apiURL, token, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	var result map[string]any
	_ = json.Unmarshal(data, &result)
	if status < 200 || status >= 300 {
		if msg, ok := result["message"].(string); ok {
			return nil, fmt.Errorf("GitHub API (%d): %s", status, msg)
		}
		return nil, fmt.Errorf("GitHub API returned status %d", status)
	}
	return result, nil
}

func ghGetLogin(token string) (string, error) {
	res, err := ghAPIGet(token, "https://api.github.com/user")
	if err != nil {
		return "", err
	}
	login, ok := res["login"].(string)
	if !ok {
		return "", fmt.Errorf("unexpected /user response")
	}
	return login, nil
}

func ghGetBranchSHA(token, owner, repo, branch string) (string, error) {
	res, err := ghAPIGet(token, fmt.Sprintf("https://api.github.com/repos/%s/%s/git/ref/heads/%s", owner, repo, branch))
	if err != nil {
		return "", err
	}
	obj, ok := res["object"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("unexpected ref response")
	}
	sha, ok := obj["sha"].(string)
	if !ok {
		return "", fmt.Errorf("no sha in ref response")
	}
	return sha, nil
}

func ghGetFileContent(token, owner, repo, filePath, ref string) (string, string, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s", owner, repo, filePath, ref)
	res, err := ghAPIGet(token, apiURL)
	if err != nil {
		return "", "", err
	}
	encoded, ok := res["content"].(string)
	if !ok {
		return "", "", fmt.Errorf("no content in file response")
	}
	fileSHA, _ := res["sha"].(string)
	encoded = strings.ReplaceAll(encoded, "\n", "")
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", fmt.Errorf("base64 decode: %w", err)
	}
	return string(decoded), fileSHA, nil
}

// ── Frontmatter editing ───────────────────────────────────────────────────────

var (
	rxFMDescription = regexp.MustCompile(`(?m)^description\s*:.*$`)
	rxFMOwner       = regexp.MustCompile(`(?m)^( {2}|\t)owner\s*:.*$`)
	rxFMLifecycle   = regexp.MustCompile(`(?m)^( {2}|\t)lifecycle\s*:.*$`)
	rxFMTags        = regexp.MustCompile(`(?m)^( {2}|\t)tags\s*:.*$`)
	rxFMMetaBlock   = regexp.MustCompile(`(?m)^metadata\s*:\s*$`)
	rxFMNameField   = regexp.MustCompile(`(?m)^name\s*:.*$`)
)

// applyFrontmatterEdits rewrites the YAML frontmatter block with the provided
// field values. Fields left empty are not modified.
func applyFrontmatterEdits(content string, fields SkillMetadataFields) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	if !strings.HasPrefix(content, "---\n") {
		return content
	}
	endIdx := strings.Index(content[4:], "\n---")
	if endIdx == -1 {
		return content
	}
	fm := content[4 : 4+endIdx]
	rest := content[4+endIdx:]

	if fields.Description != "" {
		fm = fmReplaceOrInsert(fm, rxFMDescription, "description", fields.Description, "")
	}
	if fields.Owner != "" {
		fm = fmReplaceOrInsert(fm, rxFMOwner, "owner", fields.Owner, "  ")
	}
	if fields.Lifecycle != "" {
		fm = fmReplaceOrInsert(fm, rxFMLifecycle, "lifecycle", fields.Lifecycle, "  ")
	}
	if fields.Tags != "" {
		fm = fmReplaceOrInsert(fm, rxFMTags, "tags", fields.Tags, "  ")
	}
	return "---\n" + fm + rest
}

// fmReplaceOrInsert replaces the field if found, otherwise inserts it in the
// right position (metadata block for indented fields, after name: for root fields).
func fmReplaceOrInsert(fm string, rx *regexp.Regexp, key, value, indent string) string {
	newLine := indent + key + ": " + value
	if rx.MatchString(fm) {
		return rx.ReplaceAllString(fm, newLine)
	}
	if indent != "" {
		// Indented field → put under metadata: block
		if rxFMMetaBlock.MatchString(fm) {
			return rxFMMetaBlock.ReplaceAllStringFunc(fm, func(s string) string {
				return s + "\n" + newLine
			})
		}
		return fm + "\nmetadata:\n" + newLine
	}
	// Root field → insert after name: line
	if rxFMNameField.MatchString(fm) {
		return rxFMNameField.ReplaceAllStringFunc(fm, func(s string) string {
			return s + "\n" + newLine
		})
	}
	return newLine + "\n" + fm
}

