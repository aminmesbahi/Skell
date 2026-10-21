package engine

import (
	"bytes"
	"html/template"
	"path/filepath"
)

// FleetHTML renders the report as one self-contained page: no scripts from
// elsewhere, no network access, safe to attach to an email or archive.
func FleetHTML(rep *FleetReport) (string, error) {
	var buf bytes.Buffer
	if err := fleetTmpl.Execute(&buf, rep); err != nil {
		return "", err
	}
	return buf.String(), nil
}

var fleetTmpl = template.Must(template.New("fleet").Funcs(template.FuncMap{
	"base": filepath.Base,
	"sigLabel": func(s string) string {
		switch s {
		case "trusted":
			return "trusted"
		case "untrusted-key":
			return "untrusted key"
		case "unsigned":
			return "unsigned"
		}
		return s
	},
}).Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Skill fleet report</title>
<style>
:root{--bg:#f6f7fb;--card:#fff;--ink:#141a2e;--mute:#5d6683;--line:#e3e6f0;--brand:#4f46e5;--ok:#0f9d58;--warn:#b7791f;--bad:#c53030;--okbg:#e6f6ee;--warnbg:#fdf3dc;--badbg:#fde8e8;--chip:#eef0f8}
@media (prefers-color-scheme:dark){:root{--bg:#0b1020;--card:#12182e;--ink:#e8ebf7;--mute:#97a0be;--line:#232b49;--brand:#8b93ff;--ok:#4cd08a;--warn:#f0b955;--bad:#ff8a8a;--okbg:#10301f;--warnbg:#3a2d10;--badbg:#3c1818;--chip:#1b2340}}
*{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--ink);font:15px/1.5 system-ui,-apple-system,"Segoe UI",Roboto,sans-serif}
main{max-width:1100px;margin:0 auto;padding:32px 20px 64px}
h1{font-size:26px;margin:0 0 4px}
h2{font-size:17px;margin:36px 0 12px}
.sub{color:var(--mute);margin:0 0 24px}
.tiles{display:grid;grid-template-columns:repeat(auto-fit,minmax(150px,1fr));gap:12px}
.tile{background:var(--card);border:1px solid var(--line);border-radius:12px;padding:14px 16px}
.tile b{display:block;font-size:28px;line-height:1.1}
.tile span{color:var(--mute);font-size:13px}
.tile.warn b{color:var(--warn)}.tile.bad b{color:var(--bad)}.tile.ok b{color:var(--ok)}
.card{background:var(--card);border:1px solid var(--line);border-radius:12px;overflow:auto}
table{border-collapse:collapse;width:100%}
th,td{text-align:left;padding:10px 14px;border-bottom:1px solid var(--line);vertical-align:top}
th{font-size:12px;text-transform:uppercase;letter-spacing:.04em;color:var(--mute);background:var(--card);position:sticky;top:0}
tr:last-child td{border-bottom:0}
.pill{display:inline-block;padding:1px 9px;border-radius:99px;font-size:12px;background:var(--chip);color:var(--mute);white-space:nowrap}
.pill.ok{background:var(--okbg);color:var(--ok)}.pill.warn{background:var(--warnbg);color:var(--warn)}.pill.bad{background:var(--badbg);color:var(--bad)}
.mono{font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;font-size:13px}
.mute{color:var(--mute)}
input[type=search]{width:100%;max-width:340px;padding:8px 12px;border-radius:8px;border:1px solid var(--line);background:var(--card);color:var(--ink);font:inherit;margin-bottom:12px}
details{border-top:1px solid var(--line)}details:first-child{border-top:0}
summary{cursor:pointer;padding:12px 14px;display:flex;gap:12px;align-items:center;flex-wrap:wrap}
summary:hover{background:var(--chip)}
details>div{padding:0 14px 12px}
footer{margin-top:40px;color:var(--mute);font-size:13px}
</style>
</head>
<body>
<main>
<h1>Skill fleet report</h1>
<p class="sub">{{.Summary.Repos}} repositories · generated {{.GeneratedAt}}{{if .Root}} · {{.Root}}{{end}}</p>

<div class="tiles">
<div class="tile"><b>{{.Summary.Repos}}</b><span>repositories</span></div>
<div class="tile"><b>{{.Summary.UniqueSkills}}</b><span>different skills ({{.Summary.Installs}} installs)</span></div>
<div class="tile {{if .Summary.Outdated}}warn{{else}}ok{{end}}"><b>{{.Summary.Outdated}}</b><span>outdated</span></div>
<div class="tile {{if .Summary.Modified}}warn{{else}}ok{{end}}"><b>{{.Summary.Modified}}</b><span>edited by hand</span></div>
<div class="tile {{if .Summary.HighRisk}}bad{{else}}ok{{end}}"><b>{{.Summary.HighRisk}}</b><span>high-risk installs</span></div>
<div class="tile {{if .Summary.VersionSpread}}warn{{else}}ok{{end}}"><b>{{.Summary.VersionSpread}}</b><span>skills on mixed versions</span></div>
<div class="tile {{if .Summary.Unhealthy}}warn{{else}}ok{{end}}"><b>{{.Summary.Unhealthy}}</b><span>repos needing attention</span></div>
</div>

<h2>Skills across the fleet</h2>
<input type="search" id="q" placeholder="Filter skills or repositories…" aria-label="Filter">
<div class="card"><table id="matrix">
<thead><tr><th>Skill</th><th>Repos</th><th>Versions in use</th></tr></thead>
<tbody>
{{range .Matrix}}<tr data-t="{{.Name}}"><td><b>{{.Name}}</b></td><td>{{.Repos}}</td>
<td>{{range .Versions}}<span class="pill" title="{{range .Paths}}{{.}} {{end}}">{{.Version}} × {{.Repos}}</span> {{end}}{{if .Spread}}<span class="pill warn">mixed</span>{{end}}</td></tr>
{{else}}<tr><td colspan="3" class="mute">No skills installed.</td></tr>{{end}}
</tbody></table></div>

<h2>Repositories</h2>
<div class="card" id="repos">
{{range .Repos}}<details data-t="{{.Name}} {{.Path}}"><summary>
<b>{{.Name}}</b>
<span class="pill {{if eq .Health "healthy"}}ok{{else if eq .Health "problem"}}bad{{else}}warn{{end}}">{{.Health}}</span>
<span class="mute">{{len .Skills}} skills{{if .Target}} · {{.Target}}{{end}}{{if .Errors}} · {{.Errors}} errors{{end}}{{if .Warnings}} · {{.Warnings}} warnings{{end}}{{if .Note}} · {{.Note}}{{end}}</span>
</summary>
<div><div class="mono mute" style="margin-bottom:8px">{{.Path}}</div>
{{if .Skills}}<table><thead><tr><th>Skill</th><th>Version</th><th>Status</th><th>Risk</th><th>Signature</th></tr></thead><tbody>
{{range .Skills}}<tr><td>{{.Name}}</td><td class="mono">{{if .Version}}{{.Version}}{{else if .Commit}}@{{.Commit}}{{else}}-{{end}}</td>
<td><span class="pill {{if eq .Status "outdated"}}warn{{else if eq .Status "locally-modified"}}warn{{else if eq .Status "missing"}}bad{{end}}">{{.Status}}</span></td>
<td><span class="pill {{if or (eq .Risk "high") (eq .Risk "critical")}}bad{{else if eq .Risk "medium"}}warn{{else if eq .Risk "clean"}}ok{{end}}">{{.Risk}}{{if .RiskScore}} · {{.RiskScore}}{{end}}</span></td>
<td><span class="pill {{if eq .Signature "trusted"}}ok{{end}}">{{sigLabel .Signature}}</span>{{if .Publisher}} <span class="mute">{{.Publisher}}</span>{{end}}</td></tr>{{end}}
</tbody></table>{{else}}<span class="mute">No skills installed.</span>{{end}}</div></details>
{{end}}
</div>

{{if .Events}}<h2>Recent activity</h2>
<div class="card"><table><thead><tr><th>When</th><th>Who</th><th>Action</th><th>Skill</th><th>Repository</th></tr></thead><tbody>
{{range .Events}}<tr><td class="mono">{{.Timestamp}}</td><td>{{.User}}</td><td>{{.Action}}</td><td>{{.Skill}}</td><td>{{base .Repo}}</td></tr>{{end}}
</tbody></table></div>{{end}}

<footer>Created by <span class="mono">skell report</span>. Risk is a heuristic scan, not a guarantee.</footer>
</main>
<script>
document.getElementById('q').addEventListener('input',function(e){
  var q=e.target.value.toLowerCase();
  document.querySelectorAll('[data-t]').forEach(function(el){
    el.style.display=el.getAttribute('data-t').toLowerCase().indexOf(q)>-1?'':'none';
  });
});
</script>
</body>
</html>
`))
