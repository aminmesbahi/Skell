package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
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
	hideConsoleWindow(cmd)
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
// manifest in any supported AI-agent layout (.claude, .codex, .github, .cursor),
// meaning `skell init` has already been run for some target.
func (a *App) IsRepoInitialized(repoPath string) bool {
	root := filepath.Clean(repoPath)
	for _, dir := range agentDirs {
		if _, err := os.Stat(filepath.Join(root, dir, "skell.toml")); err == nil {
			return true
		}
	}
	return false
}

// AgentTarget describes a supported AI-agent platform layout for the GUI.
type AgentTarget struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Dir         string `json:"dir"`
	Detected    bool   `json:"detected"`
}

// agentDirs is the on-disk lookup order kept in sync with internal/target.
var agentDirs = []string{".claude", ".codex", ".github", ".cursor"}

var agentTargets = []AgentTarget{
	{ID: "claude", DisplayName: "Anthropic Claude Code", Dir: ".claude"},
	{ID: "codex", DisplayName: "OpenAI Codex", Dir: ".codex"},
	{ID: "copilot", DisplayName: "GitHub Copilot / VS Code", Dir: ".github"},
	{ID: "cursor", DisplayName: "Cursor", Dir: ".cursor"},
}

// SupportedTargets returns the static list of platforms the GUI offers as
// choices in the init dialog.
func (a *App) SupportedTargets() []AgentTarget {
	out := make([]AgentTarget, len(agentTargets))
	copy(out, agentTargets)
	return out
}

// DetectTargets returns the platform(s) already present in repoPath. A target
// is reported as "detected" when its directory contains either skell.toml or a
// skills/ subdirectory.
func (a *App) DetectTargets(repoPath string) []AgentTarget {
	root := filepath.Clean(repoPath)
	out := make([]AgentTarget, 0, len(agentTargets))
	for _, t := range agentTargets {
		t.Detected = false
		if _, err := os.Stat(filepath.Join(root, t.Dir, "skell.toml")); err == nil {
			t.Detected = true
		} else if info, err := os.Stat(filepath.Join(root, t.Dir, "skills")); err == nil && info.IsDir() {
			t.Detected = true
		}
		out = append(out, t)
	}
	return out
}

// ActiveTarget returns the id of the target currently in use for repoPath, or
// the empty string when none is detected.
func (a *App) ActiveTarget(repoPath string) string {
	root := filepath.Clean(repoPath)
	// Prefer a directory that already has skell.toml.
	for _, t := range agentTargets {
		if _, err := os.Stat(filepath.Join(root, t.Dir, "skell.toml")); err == nil {
			return t.ID
		}
	}
	for _, t := range agentTargets {
		if info, err := os.Stat(filepath.Join(root, t.Dir, "skills")); err == nil && info.IsDir() {
			return t.ID
		}
	}
	return ""
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
	Version     string `json:"version"`
}

// ContributeParams describes a metadata-contribution PR request.
type ContributeParams struct {
	InstalledPath string              `json:"installedPath"`
	SourceRepo    string              `json:"sourceRepo"`
	SkillName     string              `json:"skillName"`
	Metadata      SkillMetadataFields `json:"metadata"`
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

// ResolveSkillSourceRepoURL tries to upgrade a repo-root source URL into the
// nested skill folder URL by consulting the local registry cache.
func (a *App) ResolveSkillSourceRepoURL(sourceRepo, registryAlias, skillName string) string {
	sourceRepo = strings.TrimSpace(sourceRepo)
	if sourceRepo == "" || registryAlias == "" || skillName == "" || strings.Contains(sourceRepo, "/tree/") {
		return sourceRepo
	}

	ghURL, err := parseGitHubRepoURL(sourceRepo)
	if err != nil {
		return sourceRepo
	}

	branch := "main"
	if repoData, err := ghAPIGet(ghURL.Host, fmt.Sprintf("repos/%s/%s", ghURL.Owner, ghURL.Repo)); err == nil {
		if defaultBranch, ok := repoData["default_branch"].(string); ok && defaultBranch != "" {
			branch = defaultBranch
		}
	}

	cacheDir, err := skellCacheDir(registryAlias)
	if err != nil {
		return sourceRepo
	}
	skillDir := findCachedSkillDir(cacheDir, skillName)
	if skillDir == "" {
		return sourceRepo
	}
	rel, err := filepath.Rel(cacheDir, skillDir)
	if err != nil || rel == "." {
		return sourceRepo
	}
	return strings.TrimRight(sourceRepo, "/") + "/tree/" + branch + "/" + filepath.ToSlash(rel)
}

// parseSkillMetadataFields extracts editable fields from SKILL.md YAML frontmatter.
// It handles both root-level keys and keys inside the metadata: block (the common layout).
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

	inMeta := false
	for _, line := range strings.Split(fm, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "metadata:") {
			inMeta = true
			continue
		}
		if inMeta && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") && trim != "" {
			inMeta = false // left the metadata block
		}

		key, val, ok := strings.Cut(trim, ":")
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
		case "version":
			// version is almost always under metadata:, but support root too
			if inMeta || f.Version == "" {
				f.Version = strings.Trim(val, `"'`)
			}
		}
	}
	return f
}

func parseFrontmatterName(content string) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	if !strings.HasPrefix(content, "---\n") {
		return ""
	}
	endIdx := strings.Index(content[4:], "\n---")
	if endIdx == -1 {
		return ""
	}
	fm := content[4 : 4+endIdx]
	for _, line := range strings.Split(fm, "\n") {
		key, val, ok := strings.Cut(line, ":")
		if ok && strings.TrimSpace(key) == "name" {
			return strings.TrimSpace(val)
		}
	}
	return ""
}

func skellCacheDir(registryAlias string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".skell", "cache", registryAlias), nil
}

func findCachedSkillDir(root, skillName string) string {
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return ""
	}

	found := ""
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil || found != "" {
			return nil
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		if d.IsDir() || d.Name() != "SKILL.md" {
			return nil
		}
		dir := filepath.Dir(path)
		if filepath.Base(dir) == skillName {
			found = dir
			return filepath.SkipAll
		}
		data, readErr := os.ReadFile(path)
		if readErr == nil && parseFrontmatterName(string(data)) == skillName {
			found = dir
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

// ContributeMetadata creates a GitHub PR to improve the metadata of a skill.
// It relies on an existing GitHub CLI login and only forks when the current
// account cannot push a branch to the upstream repository.
func (a *App) ContributeMetadata(params ContributeParams) ContributeResult {
	if _, err := exec.LookPath("gh"); err != nil {
		return ContributeResult{
			Success: false,
			Error:   "GitHub CLI (gh) is required for contributions. Install it from https://cli.github.com and run 'gh auth login'.",
		}
	}

	ghURL, err := parseGitHubRepoURL(params.SourceRepo)
	if err != nil {
		return ContributeResult{Error: "invalid source repo URL: " + err.Error()}
	}

	login, err := ghGetLogin(ghURL.Host)
	if err != nil {
		return ContributeResult{Error: "GitHub CLI auth failed: " + err.Error()}
	}

	repoData, err := ghAPIGet(ghURL.Host, fmt.Sprintf("repos/%s/%s", ghURL.Owner, ghURL.Repo))
	if err != nil {
		return ContributeResult{Error: "failed to get repo info: " + err.Error()}
	}
	defaultBranch, _ := repoData["default_branch"].(string)
	if defaultBranch == "" {
		defaultBranch = "main"
	}
	baseBranch := ghURL.Branch
	if baseBranch == "" {
		baseBranch = defaultBranch
	}

	workOwner := ghURL.Owner
	branchName := fmt.Sprintf("skell/metadata-%s-%d", sanitizeBranchComponent(params.SkillName), time.Now().Unix())
	if err := ghCreateBranch(ghURL.Host, workOwner, ghURL.Repo, baseBranch, branchName); err != nil {
		if strings.EqualFold(login, ghURL.Owner) {
			return ContributeResult{Error: "failed to create branch in source repo: " + err.Error()}
		}
		if err := ghEnsureFork(ghURL.Host, ghURL.Owner, ghURL.Repo, login); err != nil {
			return ContributeResult{Error: "failed to fork repo: " + err.Error()}
		}
		workOwner = login
		if err := ghCreateBranch(ghURL.Host, workOwner, ghURL.Repo, baseBranch, branchName); err != nil {
			return ContributeResult{Error: "failed to create branch in fork: " + err.Error()}
		}
	}

	filePath := "SKILL.md"
	if ghURL.SubPath != "" {
		filePath = ghURL.SubPath + "/SKILL.md"
	}
	fileContent, fileSHA, err := ghGetFileContent(ghURL.Host, workOwner, ghURL.Repo, filePath, baseBranch)
	if err != nil {
		return ContributeResult{Error: "failed to read SKILL.md from GitHub: " + err.Error()}
	}

	updated := applyFrontmatterEdits(fileContent, params.Metadata)
	encoded := base64.StdEncoding.EncodeToString([]byte(updated))
	if _, err := ghAPIPut(
		ghURL.Host,
		fmt.Sprintf("repos/%s/%s/contents/%s", workOwner, ghURL.Repo, filePath),
		map[string]string{
			"message": "fix(metadata): update metadata for " + params.SkillName,
			"content": encoded,
			"sha":     fileSHA,
			"branch":  branchName,
		},
	); err != nil {
		return ContributeResult{Error: "failed to commit changes: " + err.Error()}
	}

	headRef := branchName
	if workOwner != ghURL.Owner {
		headRef = workOwner + ":" + branchName
	}
	prData, err := ghAPIPost(
		ghURL.Host,
		fmt.Sprintf("repos/%s/%s/pulls", ghURL.Owner, ghURL.Repo),
		map[string]string{
			"title": "fix(metadata): update metadata for " + params.SkillName,
			"body":  "Automated metadata improvement contributed via Skell GUI.",
			"head":  headRef,
			"base":  baseBranch,
		},
	)
	if err != nil {
		return ContributeResult{Error: "failed to create PR: " + err.Error()}
	}

	prURL, _ := prData["html_url"].(string)
	if prURL == "" {
		return ContributeResult{Error: "PR created but no URL returned"}
	}
	return ContributeResult{PrURL: prURL, Success: true}
}

// ── Global Skill Sources (for Settings page) ──────────────────────────────────

// SkillSource represents a persistent skill source (git or local folder).
type SkillSource struct {
	Alias   string `json:"alias"`
	URL     string `json:"url"`
	IsLocal bool   `json:"is_local"`
}

func sourcesConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".skell", "config.toml"), nil
}

// ListSkillSources returns the globally configured skill sources from ~/.skell/config.toml.
func (a *App) ListSkillSources() ([]SkillSource, error) {
	path, err := sourcesConfigPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return []SkillSource{}, nil // no config yet
	}

	type sourcesFile struct {
		Sources map[string]string `toml:"sources"`
	}
	var sf sourcesFile
	if _, err := toml.Decode(string(data), &sf); err != nil {
		return nil, err
	}

	out := make([]SkillSource, 0, len(sf.Sources))
	for alias, u := range sf.Sources {
		isLocal := strings.HasPrefix(u, "file:") || strings.HasPrefix(u, "/") || (len(u) > 2 && u[1] == ':') // windows drive
		out = append(out, SkillSource{Alias: alias, URL: u, IsLocal: isLocal})
	}
	return out, nil
}

// AddSkillSource adds or updates a global skill source (git URL or local folder).
func (a *App) AddSkillSource(alias, url string) error {
	if alias == "" || url == "" {
		return errors.New("alias and url/path are required")
	}
	path, err := sourcesConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	// Read existing
	type sourcesFile struct {
		Sources map[string]string `toml:"sources"`
		Policy  map[string]any    `toml:"policy,omitempty"`
	}
	var sf sourcesFile
	if data, err := os.ReadFile(path); err == nil {
		toml.Decode(string(data), &sf)
	}
	if sf.Sources == nil {
		sf.Sources = make(map[string]string)
	}
	sf.Sources[alias] = url

	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(sf); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0600)
}

// RemoveSkillSource removes a global skill source by alias.
func (a *App) RemoveSkillSource(alias string) error {
	path, err := sourcesConfigPath()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil // nothing to remove
	}

	type sourcesFile struct {
		Sources map[string]string `toml:"sources"`
	}
	var sf sourcesFile
	if _, err := toml.Decode(string(data), &sf); err != nil {
		return err
	}
	delete(sf.Sources, alias)

	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(sf); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0600)
}

// ── GitHub API helpers ────────────────────────────────────────────────────────

type ghRepoURL struct {
	Host    string
	Owner   string
	Repo    string
	Branch  string
	SubPath string
}

func parseGitHubRepoURL(rawURL string) (ghRepoURL, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ghRepoURL{}, fmt.Errorf("empty repository URL")
	}

	parsedInput := rawURL
	if !strings.Contains(parsedInput, "://") {
		parsedInput = "https://" + strings.TrimPrefix(parsedInput, "/")
	}
	parsed, err := url.Parse(parsedInput)
	if err != nil {
		return ghRepoURL{}, fmt.Errorf("parse URL: %w", err)
	}

	host := parsed.Host
	pathValue := strings.Trim(parsed.Path, "/")
	if host == "" {
		host = "github.com"
		pathValue = strings.Trim(rawURL, "/")
	}

	parts := strings.Split(pathValue, "/")
	if len(parts) < 2 {
		return ghRepoURL{}, fmt.Errorf("expected <host>/owner/repo, got: %s", rawURL)
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
	return ghRepoURL{Host: host, Owner: owner, Repo: repo, Branch: branch, SubPath: subPath}, nil
}

func sanitizeBranchComponent(value string) string {
	clean := strings.ToLower(strings.TrimSpace(value))
	clean = regexp.MustCompile(`[^a-z0-9._-]+`).ReplaceAllString(clean, "-")
	clean = strings.Trim(clean, "-./")
	if clean == "" {
		return "skill"
	}
	return clean
}

func ghBin() (string, error) {
	bin, err := exec.LookPath("gh")
	if err != nil {
		return "", fmt.Errorf("GitHub CLI not found in PATH; install 'gh' and run 'gh auth login'")
	}
	return bin, nil
}

func runGH(host string, args ...string) ([]byte, error) {
	bin, err := ghBin()
	if err != nil {
		return nil, err
	}
	fullArgs := append([]string{}, args...)
	if host != "" {
		fullArgs = append(fullArgs[:1], append([]string{"--hostname", host}, fullArgs[1:]...)...)
	}
	cmd := exec.Command(bin, fullArgs...) //nolint:gosec
	hideConsoleWindow(cmd)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("%s", msg)
	}
	return []byte(stdout.String()), nil
}

func ghAPIGet(host, endpoint string) (map[string]any, error) {
	data, err := runGH(host, "api", endpoint)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func ghAPIPost(host, endpoint string, payload map[string]string) (map[string]any, error) {
	args := []string{"api", "--method", "POST", endpoint}
	for key, value := range payload {
		args = append(args, "-f", key+"="+value)
	}
	data, err := runGH(host, args...)
	if err != nil {
		return nil, err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return map[string]any{}, nil
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func ghAPIPut(host, endpoint string, payload map[string]string) (map[string]any, error) {
	args := []string{"api", "--method", "PUT", endpoint}
	for key, value := range payload {
		args = append(args, "-f", key+"="+value)
	}
	data, err := runGH(host, args...)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func ghGetLogin(host string) (string, error) {
	data, err := runGH(host, "api", "user", "--jq", ".login")
	if err != nil {
		return "", err
	}
	login := strings.TrimSpace(string(data))
	if login == "" {
		return "", fmt.Errorf("no authenticated GitHub account found")
	}
	return login, nil
}

func ghGetBranchSHA(host, owner, repo, branch string) (string, error) {
	res, err := ghAPIGet(host, fmt.Sprintf("repos/%s/%s/git/ref/heads/%s", owner, repo, branch))
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

func ghCreateBranch(host, owner, repo, baseBranch, branchName string) error {
	sha, err := ghGetBranchSHA(host, owner, repo, baseBranch)
	if err != nil {
		return err
	}
	_, err = ghAPIPost(
		host,
		fmt.Sprintf("repos/%s/%s/git/refs", owner, repo),
		map[string]string{
			"ref": "refs/heads/" + branchName,
			"sha": sha,
		},
	)
	return err
}

func ghEnsureFork(host, owner, repo, login string) error {
	if _, err := ghAPIPost(host, fmt.Sprintf("repos/%s/%s/forks", owner, repo), map[string]string{}); err != nil {
		return err
	}
	for attempt := 0; attempt < 10; attempt++ {
		if _, err := ghAPIGet(host, fmt.Sprintf("repos/%s/%s", login, repo)); err == nil {
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("fork did not become available in time")
}

func ghGetFileContent(host, owner, repo, filePath, ref string) (string, string, error) {
	res, err := ghAPIGet(host, fmt.Sprintf("repos/%s/%s/contents/%s?ref=%s", owner, repo, filePath, ref))
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
	rxFMVersion     = regexp.MustCompile(`(?m)^( {2}|\t)version\s*:.*$`)
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
	if fields.Version != "" {
		fm = fmReplaceOrInsert(fm, rxFMVersion, "version", fields.Version, "  ")
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
