// Package skillscan is a static security scanner for Agent Skills.
//
// A skill is a set of instructions an AI agent follows, and may ship scripts.
// The scanner looks for the patterns that make a skill dangerous — prompt
// injection, hidden text, piping downloads into a shell, credential access,
// obfuscation, persistence and destructive commands — and condenses them into
// a risk score and level. It is heuristic and advisory: it flags what deserves
// a human look, it does not prove a skill safe.
package skillscan

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// Severity ranks a single finding.
type Severity string

// Severities, from least to most serious.
const (
	SeverityInfo     Severity = "info"
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// Level is the overall risk of a skill.
type Level string

// Risk levels, from least to most serious.
const (
	LevelClean    Level = "clean"
	LevelLow      Level = "low"
	LevelMedium   Level = "medium"
	LevelHigh     Level = "high"
	LevelCritical Level = "critical"
)

var levelOrder = map[Level]int{LevelClean: 0, LevelLow: 1, LevelMedium: 2, LevelHigh: 3, LevelCritical: 4}

// ParseLevel parses a level name; the empty string means "no limit" and is
// reported as ok with an empty Level.
func ParseLevel(s string) (Level, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return "", true
	}
	l := Level(s)
	_, ok := levelOrder[l]
	return l, ok
}

// AtLeast reports whether l is as serious as or more serious than other.
func (l Level) AtLeast(other Level) bool { return levelOrder[l] >= levelOrder[other] }

// Finding is one thing the scanner noticed.
type Finding struct {
	Rule     string   `json:"rule"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
	File     string   `json:"file"`
	Line     int      `json:"line,omitempty"`
	Evidence string   `json:"evidence,omitempty"`
}

// Report is the outcome of scanning one skill folder.
type Report struct {
	Skill    string    `json:"skill"`
	Path     string    `json:"path,omitempty"`
	Score    int       `json:"score"`
	Level    Level     `json:"level"`
	Files    int       `json:"files"`
	Findings []Finding `json:"findings"`
}

// Counts returns the number of findings per severity.
func (r *Report) Counts() map[Severity]int {
	out := map[Severity]int{}
	for _, f := range r.Findings {
		out[f.Severity]++
	}
	return out
}

// Summary is a short human description, e.g. "high risk (score 52): 2 high, 1 medium".
func (r *Report) Summary() string {
	if len(r.Findings) == 0 {
		return "no findings"
	}
	if r.Level == LevelClean {
		return fmt.Sprintf("clean (%d note(s))", len(r.Findings))
	}
	c := r.Counts()
	var parts []string
	for _, s := range []Severity{SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow, SeverityInfo} {
		if c[s] > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", c[s], s))
		}
	}
	return fmt.Sprintf("%s risk (score %d): %s", r.Level, r.Score, strings.Join(parts, ", "))
}

const (
	maxTextBytes = 2 << 20 // text files larger than this are only checked by name
	maxFindings  = 200
)

type rule struct {
	id       string
	severity Severity
	message  string
	rx       *regexp.Regexp
	// textOnly limits a rule to documentation files (prompt text); code-only
	// limits it to scripts. Neither means every text file.
	textOnly bool
	codeOnly bool
}

func r(id string, sev Severity, msg, pattern string, opts ...string) rule {
	ru := rule{id: id, severity: sev, message: msg, rx: regexp.MustCompile(pattern)}
	for _, o := range opts {
		switch o {
		case "text":
			ru.textOnly = true
		case "code":
			ru.codeOnly = true
		}
	}
	return ru
}

var rules = []rule{
	// Prompt injection and concealment — aimed at the agent reading the skill.
	r("prompt-injection", SeverityHigh, "tries to override the agent's earlier instructions",
		`(?i)\b(ignore|disregard|forget|override)\b[^.\n]{0,30}\b(all |any |your |the )?(previous|prior|above|earlier|system|safety)\b[^.\n]{0,20}\b(instructions?|prompts?|rules|guidelines|policies)\b`, "text"),
	r("conceal-from-user", SeverityHigh, "tells the agent to hide something from the user",
		`(?i)\b(do not|don't|never|without)\b[^.\n]{0,25}\b(tell|inform|notify|mention|show|reveal|alert|ask)\b[^.\n]{0,25}\b(the )?(user|human|operator)\b`, "text"),
	r("persona-override", SeverityMedium, "tries to reassign the agent's role or disable its safeguards",
		`(?i)\b(you are now|from now on you|act as if you have no|jailbreak|developer mode|disable (your )?(safety|guardrails|restrictions))\b`, "text"),
	r("auto-approve", SeverityMedium, "asks the agent to skip confirmation or permission prompts",
		`(?i)\b(without (asking|confirmation|permission|approval)|skip (the )?(confirmation|permission)|auto-?approve|dangerously-skip-permissions|--yolo)\b`),

	// Remote code execution.
	r("pipe-to-shell", SeverityHigh, "downloads content and pipes it straight into a shell or interpreter",
		`(?i)\b(curl|wget|iwr|invoke-webrequest|invoke-restmethod|irm)\b[^\n|]*\|\s*(sudo\s+)?(ba|z|da|k)?sh\b|\b(curl|wget)\b[^\n|]*\|\s*(sudo\s+)?(python3?|node|perl|ruby)\b|\b(iwr|irm|invoke-webrequest|invoke-restmethod)\b[^\n|]*\|\s*(iex|invoke-expression)\b`),
	r("download-execute", SeverityMedium, "downloads a file and runs it",
		`(?i)\b(curl|wget)\b[^\n]*(-o|--output|>)\s*\S+[^\n]*(;|&&)\s*(chmod \+x|\./|sh |bash |python)`),
	r("dynamic-eval", SeverityMedium, "evaluates dynamically built code",
		`(?i)(\beval\s*\(|\bexec\s*\(\s*(base64|compile)|\bInvoke-Expression\b|\biex\s*\(|\bnew Function\s*\()`, "code"),

	// Obfuscation.
	r("base64-exec", SeverityHigh, "decodes base64 and runs the result",
		`(?i)(base64\s+(-d|--decode)[^\n]*\|\s*(ba)?sh|FromBase64String[^\n]*(iex|invoke-expression)|atob\([^)]*\)[^\n]*eval|b64decode[^\n]*(exec|eval))`),
	r("long-encoded-blob", SeverityMedium, "contains a long encoded blob that cannot be read at a glance",
		`[A-Za-z0-9+/]{400,}={0,2}`),
	r("char-code-assembly", SeverityLow, "builds strings from character codes (a common obfuscation trick)",
		`(String\.fromCharCode\([^)]{40,}\)|chr\(\d+\)\s*\+\s*chr\(\d+\)\s*\+\s*chr\(\d+\))`, "code"),

	// Credentials and exfiltration.
	r("credential-path", SeverityMedium, "reads credential or key files",
		`(?i)(~|\$HOME|%USERPROFILE%|\$env:USERPROFILE)?[/\\]\.(ssh[/\\](id_|authorized)|aws[/\\]credentials|npmrc|pypirc|netrc|docker[/\\]config\.json|kube[/\\]config|gnupg)|\b(id_rsa|id_ed25519)\b|chrome[/\\][^\n]{0,40}Login Data|keychain|wallet\.dat`),
	r("env-secrets", SeverityMedium, "reads secrets from the environment",
		`(?i)(\$\{?|%|\$env:)(AWS_SECRET_ACCESS_KEY|AWS_SESSION_TOKEN|GITHUB_TOKEN|GH_TOKEN|NPM_TOKEN|OPENAI_API_KEY|ANTHROPIC_API_KEY|AZURE_CLIENT_SECRET)\b|\b(printenv|env)\s*\|\s*(curl|nc)\b`),
	r("exfiltration", SeverityCritical, "sends local files or secrets to a remote server",
		`(?i)\b(curl|wget|iwr|invoke-webrequest|invoke-restmethod|irm|nc|ncat)\b[^\n]*(-d|--data(-binary)?|--upload-file|-F|-T|-body)\s*[@=]?[^\n]*(\.ssh|\.aws|\.env\b|id_rsa|credentials|/etc/passwd|\$\(cat|\$\(env|printenv|GITHUB_TOKEN|API_KEY|SECRET)`),
	r("reverse-shell", SeverityCritical, "opens a reverse shell",
		`(?i)(/dev/tcp/|\bnc(at)?\b[^\n]*\s-e\s|bash\s+-i\s*>&|socket\.socket\([^\n]*\n?[^\n]*subprocess|powershell[^\n]*TcpClient)`),
	r("raw-ip-url", SeverityMedium, "talks to a bare IP address instead of a named host",
		`(?i)https?://(\d{1,3}\.){3}\d{1,3}(:\d+)?(/|\b)`),
	r("url-shortener", SeverityLow, "links through a URL shortener",
		`(?i)https?://(bit\.ly|tinyurl\.com|t\.co|goo\.gl|is\.gd|cutt\.ly|rb\.gy|shorturl\.at)/`),
	r("paste-site", SeverityMedium, "fetches content from a paste site",
		`(?i)https?://(pastebin\.com|paste\.ee|hastebin\.com|ghostbin\.[a-z]+|transfer\.sh|file\.io)/`),

	// Destructive and persistent actions.
	r("destructive-delete", SeverityHigh, "deletes broad parts of the file system",
		`(?i)(\brm\s+(-[a-z]*r[a-z]*f?|-[a-z]*f[a-z]*r)\s+(/|~|\$HOME|\*|\.\.?)(\s|$|/\*)|\bRemove-Item\b[^\n]*-Recurse[^\n]*(C:\\|\$HOME|~)|\bformat\s+[a-z]:|\bmkfs(\.\w+)?\b|\bdd\s+if=[^\n]*of=/dev/|:\(\)\s*\{\s*:\|:&\s*\};:)`),
	r("persistence", SeverityMedium, "installs itself to run later (cron, shell profile, startup, scheduled task)",
		`(?i)(\bcrontab\b|>>\s*~?/?\.?(bashrc|zshrc|profile|bash_profile)|\bschtasks\b[^\n]*/create|CurrentVersion\\Run\b|\blaunchctl\s+load\b|systemctl\s+(--user\s+)?enable\b|authorized_keys)`),
	r("privilege-escalation", SeverityMedium, "escalates privileges",
		`(?i)(\bsudo\s+(-S\s+)?(su|bash|sh|chmod\s+[0-7]*7[0-7]*s?|chown)\b|\bchmod\s+[ug]?\+s\b|\bchmod\s+(-R\s+)?777\b|Start-Process[^\n]*-Verb\s+RunAs|NOPASSWD)`),
	r("security-tampering", SeverityHigh, "turns off security tooling or checks",
		`(?i)(Set-MpPreference\s+-Disable|Add-MpPreference\s+-ExclusionPath|setenforce\s+0|ufw\s+disable|NODE_TLS_REJECT_UNAUTHORIZED\s*=\s*['"]?0|git\s+config[^\n]*http\.sslVerify\s+false|curl\s[^\n]*\s(-k|--insecure)\b)`),
}

// Unicode that renders as nothing or reorders text: it can hide instructions
// from a human reviewer while the agent still reads them.
func hiddenRune(c rune) (string, bool) {
	switch {
	case c >= 0x202A && c <= 0x202E, c >= 0x2066 && c <= 0x2069:
		return "bidirectional-control", true
	case c == 0x200B, c == 0x200C, c == 0x200D, c == 0x2060:
		return "zero-width", true
	case c >= 0xE0000 && c <= 0xE007F:
		return "unicode-tag", true
	}
	return "", false
}

var codeExts = map[string]bool{
	".sh": true, ".bash": true, ".zsh": true, ".ps1": true, ".psm1": true, ".bat": true, ".cmd": true,
	".py": true, ".js": true, ".mjs": true, ".cjs": true, ".ts": true, ".rb": true, ".pl": true, ".php": true,
}

var binaryExts = map[string]bool{
	".exe": true, ".dll": true, ".so": true, ".dylib": true, ".jar": true, ".msi": true, ".scr": true,
	".com": true, ".bin": true, ".class": true, ".pyc": true, ".wasm": true,
}

var textExts = map[string]bool{
	".md": true, ".markdown": true, ".txt": true, ".yaml": true, ".yml": true, ".json": true, ".toml": true,
	".html": true, ".xml": true, ".csv": true, ".rst": true, "": true,
}

// ScanDir scans a skill folder and returns its report.
func ScanDir(dir string) (*Report, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a folder", dir)
	}
	rep := &Report{Skill: filepath.Base(dir), Path: dir, Findings: []Finding{}}

	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}
		rel, _ := filepath.Rel(dir, p)
		rel = filepath.ToSlash(rel)
		rep.Files++
		scanFile(rep, p, rel, d)
		return nil
	})

	rep.Findings = dedupe(rep.Findings)
	sort.SliceStable(rep.Findings, func(i, j int) bool {
		a, b := rep.Findings[i], rep.Findings[j]
		if sevRank(a.Severity) != sevRank(b.Severity) {
			return sevRank(a.Severity) > sevRank(b.Severity)
		}
		if a.File != b.File {
			return a.File < b.File
		}
		return a.Line < b.Line
	})
	if len(rep.Findings) > maxFindings {
		rep.Findings = rep.Findings[:maxFindings]
	}
	rep.Score, rep.Level = score(rep.Findings)
	return rep, nil
}

func sevRank(s Severity) int {
	switch s {
	case SeverityCritical:
		return 4
	case SeverityHigh:
		return 3
	case SeverityMedium:
		return 2
	case SeverityLow:
		return 1
	}
	return 0
}

func scanFile(rep *Report, path, rel string, d fs.DirEntry) {
	ext := strings.ToLower(filepath.Ext(rel))
	fi, err := d.Info()
	if err != nil {
		return
	}

	if binaryExts[ext] {
		rep.Findings = append(rep.Findings, Finding{
			Rule: "binary-executable", Severity: SeverityHigh, File: rel,
			Message: "ships a compiled binary or executable that cannot be reviewed as text",
		})
		return
	}
	isCode := codeExts[ext] || strings.HasPrefix(rel, "scripts/")
	if isCode {
		rep.Findings = append(rep.Findings, Finding{
			Rule: "script", Severity: SeverityInfo, File: rel,
			Message: "ships a script the agent may run",
		})
	}
	if !isCode && !textExts[ext] {
		// Unknown extension: sniff the content before treating it as text.
		if !looksText(path) {
			rep.Findings = append(rep.Findings, Finding{
				Rule: "opaque-file", Severity: SeverityLow, File: rel,
				Message: "contains a binary file of an unrecognised type",
			})
			return
		}
	}
	if fi.Size() > maxTextBytes {
		rep.Findings = append(rep.Findings, Finding{
			Rule: "oversized-file", Severity: SeverityLow, File: rel,
			Message: fmt.Sprintf("is %.1f MB, too large to scan line by line", float64(fi.Size())/(1<<20)),
		})
		return
	}
	data, err := os.ReadFile(path) //nolint:gosec
	if err != nil {
		return
	}
	if bytes.IndexByte(data, 0) >= 0 {
		rep.Findings = append(rep.Findings, Finding{
			Rule: "opaque-file", Severity: SeverityLow, File: rel,
			Message: "contains a binary file disguised with a text extension",
		})
		return
	}

	if strings.EqualFold(filepath.Base(rel), "SKILL.md") {
		scanFrontmatter(rep, rel, data)
	}

	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 0, 64*1024), 4<<20)
	line := 0
	hiddenReported := map[string]bool{}
	for sc.Scan() {
		line++
		text := sc.Text()
		for _, c := range text {
			if kind, ok := hiddenRune(c); ok && !hiddenReported[kind] {
				hiddenReported[kind] = true
				sev := SeverityMedium
				if kind != "zero-width" {
					sev = SeverityHigh
				}
				rep.Findings = append(rep.Findings, Finding{
					Rule: "hidden-text", Severity: sev, File: rel, Line: line,
					Message: fmt.Sprintf("contains invisible %s characters that can hide instructions from a reviewer", kind),
				})
			}
		}
		for _, ru := range rules {
			if ru.codeOnly && !isCode {
				continue
			}
			if ru.textOnly && isCode {
				continue
			}
			if loc := ru.rx.FindStringIndex(text); loc != nil {
				rep.Findings = append(rep.Findings, Finding{
					Rule: ru.id, Severity: ru.severity, File: rel, Line: line,
					Message: ru.message, Evidence: evidence(text, loc),
				})
			}
		}
	}
}

// scanFrontmatter flags tool grants broad enough to deserve attention.
func scanFrontmatter(rep *Report, rel string, data []byte) {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	if !strings.HasPrefix(text, "---\n") {
		return
	}
	end := strings.Index(text[4:], "\n---")
	if end < 0 {
		return
	}
	fm := text[4 : 4+end]
	for i, ln := range strings.Split(fm, "\n") {
		trim := strings.TrimSpace(ln)
		if !strings.HasPrefix(trim, "allowed-tools:") {
			continue
		}
		val := strings.TrimSpace(strings.TrimPrefix(trim, "allowed-tools:"))
		lv := strings.ToLower(val)
		switch {
		case lv == "*" || lv == `"*"` || lv == "all":
			rep.Findings = append(rep.Findings, Finding{Rule: "broad-tools", Severity: SeverityHigh, File: rel, Line: i + 2,
				Message: "pre-approves every tool", Evidence: trim})
		case strings.Contains(" "+lv+" ", " bash ") || strings.Contains(lv, "bash(*") || strings.Contains(lv, "powershell ") || strings.HasSuffix(lv, "powershell"):
			rep.Findings = append(rep.Findings, Finding{Rule: "broad-tools", Severity: SeverityMedium, File: rel, Line: i + 2,
				Message: "pre-approves unrestricted shell access", Evidence: trim})
		}
	}
}

func looksText(path string) bool {
	f, err := os.Open(path) //nolint:gosec
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()
	buf := make([]byte, 4096)
	n, _ := f.Read(buf)
	return bytes.IndexByte(buf[:n], 0) < 0
}

func evidence(line string, loc []int) string {
	start, end := loc[0], loc[1]
	if end-start > 140 {
		end = start + 140
	}
	pre := start - 20
	if pre < 0 {
		pre = 0
	}
	s := strings.TrimSpace(line[pre:end])
	s = strings.Map(func(c rune) rune {
		if unicode.IsControl(c) {
			return ' '
		}
		return c
	}, s)
	if len(line) > end && end-start >= 140 {
		s += "…"
	}
	return s
}

// dedupe collapses repeated findings of one rule in one file down to the
// first few, so a long script does not drown the report.
func dedupe(in []Finding) []Finding {
	counts := map[string]int{}
	var out []Finding
	for _, f := range in {
		key := f.Rule + "\x00" + f.File
		counts[key]++
		if counts[key] <= 3 {
			out = append(out, f)
		}
	}
	return out
}

// score turns findings into 0–100 and a level. Each rule counts at its full
// weight once and a diminishing share for repeats, so one noisy rule cannot
// outweigh several different ones.
func score(findings []Finding) (int, Level) {
	weight := map[Severity]float64{
		SeverityInfo: 0, SeverityLow: 3, SeverityMedium: 10, SeverityHigh: 25, SeverityCritical: 45,
	}
	seen := map[string]int{}
	total := 0.0
	maxSev := SeverityInfo
	for _, f := range findings {
		n := seen[f.Rule]
		seen[f.Rule]++
		w := weight[f.Severity]
		switch n {
		case 0:
		case 1:
			w *= 0.4
		default:
			w *= 0.15
		}
		total += w
		if sevRank(f.Severity) > sevRank(maxSev) {
			maxSev = f.Severity
		}
	}
	s := int(total + 0.5)
	if s > 100 {
		s = 100
	}
	// The worst single finding sets a floor, so one critical finding is never
	// averaged away.
	switch maxSev {
	case SeverityCritical:
		if s < 70 {
			s = 70
		}
	case SeverityHigh:
		if s < 40 {
			s = 40
		}
	}
	switch {
	case s >= 70:
		return s, LevelCritical
	case s >= 40:
		return s, LevelHigh
	case s >= 12:
		return s, LevelMedium
	case s > 0:
		return s, LevelLow
	}
	return 0, LevelClean
}

// Exceeds reports whether l is more serious than limit. An empty limit
// means "no limit".
func (l Level) Exceeds(limit Level) bool {
	if limit == "" {
		return false
	}
	return levelOrder[l] > levelOrder[limit]
}
