import { useState } from "react";
import { Link } from "react-router";
import { useProject } from "@/hooks/useProject";
import { useAsync } from "@/hooks/useAsync";
import { getProjectDisplayName, skillRoute } from "@/lib/navigation";
import { listInstalled, getStatus, listSupportedTargets, targetFromInstalledPath, validateSkills, initRepo, isRepoInitialized, upgradeSkill, removeSkill, pinSkill, unpinSkill } from "@/lib/skell";
import type { InstalledSkill, SkillStatus } from "@/lib/types";
import { useUIStore } from "@/store";
import { SkillBadge, STATUS_CONFIG } from "@/components/Badges";
import { ValidationBadge } from "@/components/ValidationBadge";
import { ProjectPageHeader } from "@/components/ProjectPageHeader";
import { AddSkillButton } from "@/components/AddSkillButton";
import { AddFromURLDialog } from "@/components/AddFromURLDialog";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import { LoadState, MissingProject } from "@/components/LoadState";
import { UpgradeDiffDialog } from "@/components/UpgradeDiffDialog";

const identity = (skill: InstalledSkill) => `${targetFromInstalledPath(skill.installed_path)}:${skill.name}`;
type Operation = "upgrade" | "remove";
export function ProjectSkillsPage() {
  const project = useProject();
  return project ? <ProjectSkills key={project} project={project} /> : <MissingProject />;
}
function ProjectSkills({ project }: { project: string }) {
  const notify = useUIStore((s) => s.notify);
  const [search, setSearch] = useState("");
  const [target, setTarget] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const [selected, setSelected] = useState<string[]>([]);
  const [addOpen, setAddOpen] = useState(false);
  const [busy, setBusy] = useState(false);
  const [confirm, setConfirm] = useState<{ action: Operation; skills: InstalledSkill[] } | null>(null);
  const [results, setResults] = useState<string[]>([]);
  const [diffFor, setDiffFor] = useState<InstalledSkill | null>(null);
  const state = useAsync(project, async () => {
    const skills = await listInstalled(project);
    const targets = [...new Set(skills.map((s) => targetFromInstalledPath(s.installed_path)))];
    const [supported, initialized, checks] = await Promise.all([
      listSupportedTargets(), isRepoInitialized(project),
      Promise.allSettled(targets.map(async (id) => ({ id, statuses: await getStatus(project, id || undefined) }))),
    ]);
    const statuses: Record<string, SkillStatus> = {};
    const changed: Record<string, boolean> = {};
    const errors: string[] = [];
    for (const check of checks) {
      if (check.status === "rejected") errors.push(String(check.reason));
      else for (const entry of check.value.statuses ?? []) {
        statuses[`${check.value.id}:${entry.name}`] = entry.status;
        if (entry.changed) changed[`${check.value.id}:${entry.name}`] = true;
      }
    }
    return { skills, statuses, changed, errors, targets: supported, initialized };
  });
  const validation = useAsync(`${project}:${state.data?.skills.map(identity).join(",") ?? ""}`, async () => state.data ? validateSkills(project, "", false) : []);
  const skills = state.data?.skills ?? [];
  const filtered = skills.filter((s) => (!search || `${s.name} ${s.registry}`.toLowerCase().includes(search.toLowerCase())) && (!target || targetFromInstalledPath(s.installed_path) === target) && (!statusFilter || (s.pinned ? "pinned" : state.data?.statuses[identity(s)]) === statusFilter));
  const outdated = skills.filter((s) => !s.pinned && state.data?.statuses[identity(s)] === "outdated");
  const refresh = () => { void state.refresh(); void validation.refresh(); };
  async function run(action: Operation, items: InstalledSkill[], force = false) {
    setConfirm(null); setBusy(true); setResults([]);
    const report: string[] = [];
    for (const skill of items) {
      try {
        const opts = { repo: project, skillName: skill.name, target: targetFromInstalledPath(skill.installed_path) };
        const result = await (action === "upgrade" ? upgradeSkill(force ? { ...opts, force } : opts) : removeSkill(opts));
        report.push(`${skill.name} (${opts.target}): ${result.success ? action === "upgrade" ? "Upgrade finished" : "Removed" : "Failed"} — ${result.success ? result.stdout.trim() : result.stderr}`);
      } catch (error) { report.push(`${skill.name}: Failed — ${String(error)}`); }
      setResults([...report]);
    }
    setBusy(false); setSelected([]); refresh();
  }
  async function togglePin(skill: InstalledSkill) {
    setBusy(true);
    try {
      const result = await (skill.pinned ? unpinSkill : pinSkill)({ repo: project, skillName: skill.name, target: targetFromInstalledPath(skill.installed_path) });
      if (!result.success) throw new Error(result.stderr);
      refresh();
    } catch (error) { notify({ kind: "error", title: "Could not change pin", detail: String(error) }); }
    finally { setBusy(false); }
  }
  return <div className="mx-auto max-w-6xl px-6 py-8 space-y-6">
    <ProjectPageHeader projectPath={project} title={getProjectDisplayName(project)} subtitle={`Skills for ${getProjectDisplayName(project)}.`} breadcrumb="Skills" actions={<><AddSkillButton projectPath={project} onRefresh={refresh} /><button className="btn-ghost" disabled={state.loading || busy} onClick={refresh}>Refresh</button></>} />
    <LoadState loading={state.loading} error={state.error || state.data?.errors.join("\n") || undefined} retry={refresh} />
    {validation.error && <LoadState loading={false} error={`Validation unavailable: ${validation.error}`} retry={() => void validation.refresh()} />}
    {state.data?.initialized === false && <div className="card"><p>Initialize this project to manage skills.</p><button className="btn-primary" disabled={busy} onClick={async () => { setBusy(true); try { const result = await initRepo(project); if (!result.success) throw new Error(result.stderr); refresh(); } catch (error) { notify({ kind: "error", title: "Initialization failed", detail: String(error) }); } finally { setBusy(false); } }}>Initialize now</button></div>}
    <div className="surface-bar">
      <div className="flex flex-wrap items-center gap-3">
        <input data-search aria-label="Search installed skills" className="input max-w-xs" placeholder="Search skills… (/ or Ctrl+K)" value={search} onChange={(e) => setSearch(e.target.value)} />
        <select aria-label="Filter by agent" className="input w-auto min-w-36" value={target} onChange={(e) => setTarget(e.target.value)}><option value="">All agents</option>{state.data?.targets.map((t) => <option key={t.id} value={t.id}>{t.displayName}</option>)}</select>
        <select aria-label="Filter by status" className="input w-auto min-w-40" value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)}><option value="">All statuses</option>{Object.entries(STATUS_CONFIG).map(([value, config]) => <option key={value} value={value}>{config.label}</option>)}</select>
        <button className="btn-primary" disabled={busy || state.loading || !outdated.length} onClick={() => setConfirm({ action: "upgrade", skills: outdated })}>Upgrade all outdated ({outdated.length})</button>
        <button className="btn-danger" disabled={busy || state.loading || !selected.length} onClick={() => setConfirm({ action: "remove", skills: skills.filter((s) => selected.includes(identity(s))) })}>Remove selected ({selected.length})</button>
      </div>
    </div>
    {results.length > 0 && <section className="card" aria-live="polite"><h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-slate-300">Operation results {busy ? "(in progress)" : ""}</h2><ul className="space-y-2 text-sm text-slate-400">{results.map((r, i) => <li key={i}>{r}</li>)}</ul></section>}
    {!state.loading && !state.error && !skills.length ? <div className="card text-center"><h2>No skills installed yet</h2><div className="flex justify-center gap-3 mt-4"><Link className="btn-primary" to="/catalog" state={{ installDestination: project }}>Browse Catalog</Link><button className="btn-ghost" onClick={() => setAddOpen(true)}>Add from Repository</button></div></div> : !!skills.length && !filtered.length ? <div className="card"><p>No skills match your filters.</p><button className="btn-ghost" onClick={() => { setSearch(""); setTarget(""); setStatusFilter(""); }}>Clear filters</button></div> : filtered.length > 0 && <div className="card p-0 overflow-x-auto"><table className="data-table"><thead><tr><th><input type="checkbox" aria-label="Select all matching skills" disabled={busy} checked={filtered.length > 0 && filtered.every((s) => selected.includes(identity(s)))} onChange={(e) => setSelected(e.target.checked ? [...new Set([...selected, ...filtered.map(identity)])] : selected.filter((id) => !filtered.some((s) => identity(s) === id)))} /></th><th>Skill</th><th>Version</th><th>Status</th><th>Validation</th><th>Agent</th><th>Actions</th></tr></thead><tbody>{filtered.map((skill) => {
      const id = identity(skill), agent = targetFromInstalledPath(skill.installed_path), status = skill.pinned ? "pinned" : state.data?.statuses[id] ?? "unknown";
      return <tr key={id}><td><input type="checkbox" disabled={busy} aria-label={`Select ${skill.name} for ${agent}`} checked={selected.includes(id)} onChange={(e) => setSelected(e.target.checked ? [...selected, id] : selected.filter((s) => s !== id))} /></td><td><Link className="font-medium text-brand-400 hover:text-brand-300" to={skillRoute(project, skill.name, agent)}>{skill.name}</Link></td><td>{skill.version || "—"}{state.data?.changed[id] && status === "outdated" && <span className="ml-2 text-xs text-amber-400" title="The source files changed without a version bump">content changed</span>}</td><td><SkillBadge status={status} /></td><td><Link to={`${skillRoute(project, skill.name, agent)}&tab=validate`}><ValidationBadge loading={validation.loading} result={validation.data?.find((v) => v.name === skill.name && (!v.target || v.target === agent))} /></Link></td><td>{state.data?.targets.find((t) => t.id === agent)?.displayName ?? agent}</td><td><div className="flex flex-wrap items-center gap-1">{status === "locally-modified" ? <button className="btn-subtle" disabled={busy || state.loading} onClick={() => setDiffFor(skill)} title="See your local edits and the latest version">Changes</button> : <button className="btn-subtle" disabled={busy || state.loading || status !== "outdated"} onClick={() => setDiffFor(skill)} title="Review the changes before upgrading">Upgrade</button>}<button className="btn-subtle" disabled={busy || state.loading} onClick={() => void togglePin(skill)}>{skill.pinned ? "Unpin" : "Pin"}</button><button className="btn-subtle text-red-300 hover:bg-red-500/10 hover:text-red-200" disabled={busy || state.loading} onClick={() => setConfirm({ action: "remove", skills: [skill] })}>Remove</button></div></td></tr>;
    })}</tbody></table></div>}
    <ConfirmDialog open={!!confirm} title={confirm?.action === "remove" ? "Remove selected installations?" : "Upgrade selected installations?"} description={`${confirm?.skills.map((s) => `${s.name} (${targetFromInstalledPath(s.installed_path)})`).join(", ") ?? ""}. ${confirm?.action === "remove" ? "This deletes their local files, including local changes." : "Pinned and locally modified skills are protected; upgrades will not force overwrites."}`} danger={confirm?.action === "remove"} confirmLabel={confirm?.action === "remove" ? "Remove" : "Upgrade"} onConfirm={() => { if (confirm) void run(confirm.action, confirm.skills); }} onCancel={() => setConfirm(null)} />
    <UpgradeDiffDialog skillName={diffFor?.name ?? null} repo={project} onCancel={() => setDiffFor(null)} onUpgrade={(force) => { const skill = diffFor; setDiffFor(null); if (skill) void run("upgrade", [skill], force); }} />
    <AddFromURLDialog open={addOpen} initialRepo={project} onClose={() => setAddOpen(false)} onSuccess={refresh} />
  </div>;
}
