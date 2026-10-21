import { useState } from "react";
import { ChevronDown, ChevronRight, FileDown, FolderOpen, RefreshCw, Radar } from "lucide-react";
import clsx from "clsx";
import { SelectDirectory } from "../../bindings/skell-gui/app";
import { useRepoStore, useUIStore } from "@/store";
import { useAsync } from "@/hooks/useAsync";
import { exportFleetReport, fleetReport } from "@/lib/skell";
import { filterMatrix, filterRepos, fleetTiles, joinPath, reportFileName, type FleetTile } from "@/lib/fleet";
import { LoadState } from "@/components/LoadState";
import { RiskBadge, SignatureBadge } from "@/components/Security";
import { SkillBadge } from "@/components/Badges";
import type { FleetRepo, FleetReport, SkillStatus } from "@/lib/types";

type Scope = { kind: "projects" } | { kind: "folder"; path: string };

const TILE_TONE: Record<FleetTile["tone"], string> = {
  neutral: "text-slate-200",
  ok: "text-emerald-400",
  warn: "text-amber-400",
  bad: "text-red-400",
};

const HEALTH: Record<FleetRepo["health"], { label: string; classes: string }> = {
  healthy: { label: "Healthy", classes: "bg-emerald-500/15 text-emerald-400 border-emerald-500/30" },
  attention: { label: "Needs attention", classes: "bg-amber-500/15 text-amber-400 border-amber-500/30" },
  problem: { label: "Problem", classes: "bg-red-500/15 text-red-400 border-red-500/30" },
};

const KNOWN_STATUS = new Set<SkillStatus>(["up-to-date", "outdated", "pinned", "deprecated", "archived", "locally-modified", "unknown", "missing-metadata", "unversioned", "linked"]);

function RepoRow({ repo, open, onToggle }: { repo: FleetRepo; open: boolean; onToggle: () => void }) {
  const health = HEALTH[repo.health];
  const panel = `fleet-repo-${repo.path}`;
  return (
    <li className="border-b border-[var(--palette-1a1f35)] last:border-0">
      <button type="button" onClick={onToggle} aria-expanded={open} aria-controls={panel} className="flex w-full flex-wrap items-center gap-x-4 gap-y-1 px-4 py-3 text-left transition-colors hover:bg-white/5">
        {open ? <ChevronDown size={14} className="shrink-0 text-slate-500" /> : <ChevronRight size={14} className="shrink-0 text-slate-500" />}
        <span className="min-w-0 flex-1 truncate font-medium text-slate-200" title={repo.path}>{repo.name}</span>
        <span className="text-xs text-slate-500">{repo.skills.length} skill{repo.skills.length === 1 ? "" : "s"}{repo.target ? ` · ${repo.target}` : ""}</span>
        <span className={clsx("rounded-full border px-2 py-0.5 text-xs font-medium", health.classes)}>{health.label}</span>
      </button>
      {open && (
        <div id={panel} className="px-4 pb-4 pl-10">
          <p className="mb-2 truncate font-mono text-xs text-slate-600" title={repo.path}>{repo.path}</p>
          {repo.note && <p className="mb-2 text-sm text-amber-300">{repo.note}</p>}
          {(repo.errors > 0 || repo.warnings > 0) && <p className="mb-2 text-sm text-slate-400">{repo.errors} error{repo.errors === 1 ? "" : "s"}, {repo.warnings} warning{repo.warnings === 1 ? "" : "s"} from the health check.</p>}
          {repo.skills.length === 0 ? (
            <p className="text-sm text-slate-500">No skills installed.</p>
          ) : (
            <table className="data-table">
              <thead><tr><th>Skill</th><th>Version</th><th>Status</th><th>Security</th><th>Signature</th></tr></thead>
              <tbody>
                {repo.skills.map((s) => (
                  <tr key={s.name}>
                    <td className="font-medium text-slate-200">{s.name}</td>
                    <td className="font-mono text-xs">{s.version || (s.commit ? `@${s.commit.slice(0, 7)}` : "-")}</td>
                    <td>{KNOWN_STATUS.has(s.status as SkillStatus) ? <SkillBadge status={s.status as SkillStatus} size="sm" /> : <span className="text-xs text-slate-500">{s.status}</span>}</td>
                    <td><RiskBadge level={s.risk} score={s.risk_score} size="sm" /></td>
                    <td><SignatureBadge status={s.signature} publisher={s.publisher} size="sm" /></td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}
    </li>
  );
}

function Report({ report, query }: { report: FleetReport; query: string }) {
  const [open, setOpen] = useState<string | null>(null);
  const matrix = filterMatrix(report.matrix, query);
  const repos = filterRepos(report.repos, query);
  return (
    <>
      <div className="grid grid-cols-2 gap-3 md:grid-cols-4 xl:grid-cols-7">
        {fleetTiles(report.summary).map((t) => (
          <div key={t.key} className="card !p-4" title={t.hint}>
            <p className={clsx("text-3xl font-bold", TILE_TONE[t.tone])}>{t.value}</p>
            <p className="text-sm text-slate-500">{t.label}</p>
          </div>
        ))}
      </div>

      <section aria-label="Skills across the fleet">
        <h2 className="mb-2 text-sm font-semibold uppercase tracking-wider text-slate-500">Skills across the fleet</h2>
        <div className="card !p-0 overflow-x-auto">
          {matrix.length === 0 ? (
            <p className="p-4 text-sm text-slate-500">{report.matrix.length === 0 ? "No skills are installed in these repositories." : "No skills match your filter."}</p>
          ) : (
            <table className="data-table">
              <thead><tr><th>Skill</th><th>Repos</th><th>Versions in use</th></tr></thead>
              <tbody>
                {matrix.map((row) => (
                  <tr key={row.name}>
                    <td className="font-medium text-slate-200">{row.name}</td>
                    <td>{row.repos}</td>
                    <td>
                      <div className="flex flex-wrap items-center gap-1.5">
                        {row.versions.map((v) => (
                          <span key={v.version} className="rounded-md bg-white/5 px-2 py-0.5 font-mono text-xs text-slate-300" title={v.paths?.join("\n")}>{v.version} × {v.repos}</span>
                        ))}
                        {row.spread && <span className="rounded-full border border-amber-500/30 bg-amber-500/15 px-2 py-0.5 text-xs font-medium text-amber-400">mixed</span>}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </section>

      <section aria-label="Repositories">
        <h2 className="mb-2 text-sm font-semibold uppercase tracking-wider text-slate-500">Repositories</h2>
        <div className="card !p-0 overflow-hidden">
          {repos.length === 0 ? <p className="p-4 text-sm text-slate-500">No repositories match your filter.</p> : (
            <ul>{repos.map((r) => <RepoRow key={r.path} repo={r} open={open === r.path} onToggle={() => setOpen(open === r.path ? null : r.path)} />)}</ul>
          )}
        </div>
      </section>

      {report.events && report.events.length > 0 && (
        <section aria-label="Recent activity">
          <h2 className="mb-2 text-sm font-semibold uppercase tracking-wider text-slate-500">Recent activity</h2>
          <div className="card !p-0 overflow-x-auto">
            <table className="data-table">
              <thead><tr><th>When</th><th>Who</th><th>Action</th><th>Skill</th><th>Repository</th></tr></thead>
              <tbody>
                {report.events.map((e, i) => (
                  <tr key={`${e.timestamp}-${i}`}>
                    <td className="font-mono text-xs">{e.timestamp.replace("T", " ").replace("Z", "")}</td>
                    <td>{e.user || "-"}</td>
                    <td>{e.action}</td>
                    <td>{e.skill}</td>
                    <td className="truncate" title={e.repo}>{(e.repo ?? "").split(/[\\/]/).filter(Boolean).pop() ?? "-"}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>
      )}
    </>
  );
}

/** Fleet: every project's skills at a glance — versions, drift, risk, signatures, activity. */
export function Fleet() {
  const repos = useRepoStore((s) => s.repos).filter((r) => r !== "global");
  const { notify } = useUIStore();
  const [scope, setScope] = useState<Scope>({ kind: "projects" });
  const [checkUpdates, setCheckUpdates] = useState(false);
  const [query, setQuery] = useState("");
  const [exporting, setExporting] = useState(false);

  const key = `${scope.kind === "folder" ? scope.path : repos.join("|")}::${checkUpdates}`;
  const empty = scope.kind === "projects" && repos.length === 0;
  const state = useAsync(key, async () => {
    if (empty) return null;
    return fleetReport(scope.kind === "folder" ? { root: scope.path, checkUpdates } : { repos, checkUpdates });
  });

  async function pickFolder() {
    const picked = await SelectDirectory();
    if (picked) setScope({ kind: "folder", path: picked });
  }

  async function save(format: "html" | "csv") {
    const dir = await SelectDirectory();
    if (!dir) return;
    setExporting(true);
    try {
      const out = joinPath(dir, reportFileName(format));
      const res = await exportFleetReport({ ...(scope.kind === "folder" ? { root: scope.path } : { repos }), format, out });
      notify(res.success ? { kind: "success", title: "Report saved", detail: out } : { kind: "error", title: "Could not save the report", detail: res.stderr });
    } finally {
      setExporting(false);
    }
  }

  return (
    <div className="mx-auto max-w-6xl space-y-6 px-6 py-8">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 className="flex items-center gap-2 text-2xl font-bold text-slate-200"><Radar size={22} className="text-brand-400" />Fleet</h1>
          <p className="mt-0.5 text-sm text-slate-500">Which skills run where: versions, hand edits, security risk and signatures across all your repositories.</p>
        </div>
        <div className="flex flex-wrap gap-2">
          <button className="btn-ghost" disabled={state.loading} onClick={() => void state.refresh()}><RefreshCw size={14} className={state.loading ? "animate-spin" : ""} />Refresh</button>
          <button className="btn-ghost" disabled={!state.data || exporting} onClick={() => void save("html")}><FileDown size={14} />Save as page</button>
          <button className="btn-ghost" disabled={!state.data || exporting} onClick={() => void save("csv")}><FileDown size={14} />Save as CSV</button>
        </div>
      </div>

      <div className="surface-bar flex flex-wrap items-center gap-3">
        <div className="inline-flex rounded-lg border border-[var(--palette-2d3348)] p-0.5" role="group" aria-label="Scope">
          <button className={clsx("rounded-md px-3 py-1.5 text-sm", scope.kind === "projects" ? "bg-brand-600 text-white" : "text-slate-400 hover:text-slate-200")} aria-pressed={scope.kind === "projects"} onClick={() => setScope({ kind: "projects" })}>My projects ({repos.length})</button>
          <button className={clsx("rounded-md px-3 py-1.5 text-sm", scope.kind === "folder" ? "bg-brand-600 text-white" : "text-slate-400 hover:text-slate-200")} aria-pressed={scope.kind === "folder"} onClick={() => void pickFolder()}><FolderOpen size={13} className="mr-1.5 inline" />{scope.kind === "folder" ? scope.path.split(/[\\/]/).filter(Boolean).pop() : "A folder…"}</button>
        </div>
        <label className="flex items-center gap-2 text-sm text-slate-400">
          <input type="checkbox" checked={checkUpdates} onChange={(e) => setCheckUpdates(e.target.checked)} />
          Check for updates <span className="text-xs text-slate-600">(asks each source, slower)</span>
        </label>
        <input className="input ml-auto max-w-xs" type="search" placeholder="Filter skills or repositories…" aria-label="Filter" value={query} onChange={(e) => setQuery(e.target.value)} />
      </div>

      {empty && (
        <div className="card text-center">
          <Radar size={28} className="mx-auto mb-2 text-slate-600" />
          <p className="font-medium text-slate-200">No projects yet</p>
          <p className="mx-auto mt-1 max-w-md text-sm text-slate-500">Add projects from the sidebar, or point Fleet at a folder that holds your repositories.</p>
          <button className="btn-primary mt-4" onClick={() => void pickFolder()}><FolderOpen size={14} />Choose a folder</button>
        </div>
      )}

      <LoadState loading={state.loading && !state.data} error={state.error} retry={() => void state.refresh()} />
      {state.data && <Report report={state.data} query={query} />}
    </div>
  );
}
