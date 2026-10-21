import { useState } from "react";
import { ChevronDown, ChevronRight, ShieldCheck } from "lucide-react";
import { Link } from "react-router";
import clsx from "clsx";
import { useProject } from "@/hooks/useProject";
import { useAsync } from "@/hooks/useAsync";
import { scanSkills, verifySkills } from "@/lib/skell";
import { getProjectDisplayName, skillRoute } from "@/lib/navigation";
import { isHighRisk, joinSecurity, riskAdvice, signatureAdvice, type SecurityRow } from "@/lib/security";
import { LoadState, MissingProject } from "@/components/LoadState";
import { ProjectPageHeader } from "@/components/ProjectPageHeader";
import { FindingsList, RiskBadge, SignatureBadge } from "@/components/Security";

function Tile({ value, label, tone }: { value: number; label: string; tone?: "bad" | "warn" | "ok" }) {
  return (
    <div className="card !p-4">
      <p className={clsx("text-3xl font-bold", tone === "bad" && "text-red-400", tone === "warn" && "text-amber-400", tone === "ok" && "text-emerald-400", !tone && "text-slate-200")}>{value}</p>
      <p className="text-sm text-slate-500">{label}</p>
    </div>
  );
}

function SkillRow({ row, project, open, onToggle }: { row: SecurityRow; project: string; open: boolean; onToggle: () => void }) {
  const level = row.scan?.level ?? "clean";
  const sig = row.signature;
  const panelId = `security-${row.name}`;
  return (
    <li className="border-b border-[var(--palette-1a1f35)] last:border-0">
      <button
        type="button"
        onClick={onToggle}
        aria-expanded={open}
        aria-controls={panelId}
        className="flex w-full flex-wrap items-center gap-x-4 gap-y-2 px-4 py-3 text-left transition-colors hover:bg-white/5"
      >
        {open ? <ChevronDown size={14} className="shrink-0 text-slate-500" /> : <ChevronRight size={14} className="shrink-0 text-slate-500" />}
        <span className="min-w-0 flex-1 truncate font-medium text-slate-200">{row.name}</span>
        <RiskBadge level={level} score={row.scan?.score} />
        {sig && <SignatureBadge status={sig.status} publisher={sig.publisher} />}
      </button>
      {open && (
        <div id={panelId} className="space-y-3 px-4 pb-4 pl-10 text-sm">
          <p className="text-slate-400">{riskAdvice(level)}</p>
          {row.scan && <FindingsList report={row.scan} />}
          {sig && <p className="text-slate-500">{signatureAdvice(sig)}</p>}
          <Link to={skillRoute(project, row.name)} className="inline-block text-xs text-brand-400 hover:underline">Open skill details</Link>
        </div>
      )}
    </li>
  );
}

/** Security tab: scan results and signature checks for every installed skill. */
export function ProjectSecurityPage() {
  const project = useProject();
  const [open, setOpen] = useState<string | null>(null);
  const state = useAsync(project, async () => {
    if (!project) return null;
    const [scans, sigs] = await Promise.all([scanSkills(project), verifySkills(project)]);
    return joinSecurity(scans, sigs);
  });
  if (!project) return <MissingProject />;

  const rows = state.data ?? [];
  const high = rows.filter((r) => isHighRisk(r.scan?.level ?? "clean")).length;
  const broken = rows.filter((r) => r.signature && (r.signature.status === "tampered" || r.signature.status === "invalid")).length;
  const trusted = rows.filter((r) => r.signature?.status === "trusted").length;
  const needsAttention = rows.filter((r) => isHighRisk(r.scan?.level ?? "clean") || r.signature?.status === "tampered" || r.signature?.status === "invalid");

  return (
    <div className="mx-auto max-w-6xl space-y-6 px-6 py-8">
      <ProjectPageHeader
        projectPath={project}
        title={`Security for ${getProjectDisplayName(project)}`}
        subtitle="Skills are instructions your agent follows. This page checks what each one contains and who published it."
        actions={<button className="btn-ghost" disabled={state.loading} onClick={() => void state.refresh()}>Scan again</button>}
      />
      <LoadState loading={state.loading && !state.data} error={state.error} retry={() => void state.refresh()} />

      {state.data && rows.length === 0 && (
        <div className="card text-center">
          <ShieldCheck size={28} className="mx-auto mb-2 text-slate-600" />
          <p className="font-medium text-slate-200">No skills installed yet</p>
          <p className="mt-1 text-sm text-slate-500">Install a skill and its security check will appear here.</p>
        </div>
      )}

      {rows.length > 0 && (
        <>
          <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
            <Tile value={rows.length} label="skills checked" />
            <Tile value={high} label="high or critical risk" tone={high ? "bad" : "ok"} />
            <Tile value={broken} label="tampered signatures" tone={broken ? "bad" : "ok"} />
            <Tile value={trusted} label="signed by trusted publishers" tone={trusted ? "ok" : undefined} />
          </div>

          {needsAttention.length > 0 && (
            <div role="alert" className="rounded-xl border border-red-500/30 bg-red-500/10 p-4 text-sm text-red-200">
              <p className="font-medium">
                {needsAttention.length === 1 ? "1 skill needs a closer look" : `${needsAttention.length} skills need a closer look`}:{" "}
                {needsAttention.map((r) => r.name).join(", ")}
              </p>
              <p className="mt-1 text-red-200/80">Open a skill below to see exactly what was flagged. If you do not recognise it, remove the skill.</p>
            </div>
          )}

          <section className="card !p-0 overflow-hidden" aria-label="Skills">
            <ul>
              {rows.map((row) => (
                <SkillRow key={row.name} row={row} project={project} open={open === row.name} onToggle={() => setOpen(open === row.name ? null : row.name)} />
              ))}
            </ul>
          </section>

          <p className="text-xs text-slate-600">
            The scan is a heuristic: it points at what deserves a human look and cannot prove a skill safe. Signatures prove who published a skill and that nobody changed it afterwards.
            Choose publishers you trust in <Link to="/settings" className="text-brand-400 hover:underline">Settings</Link>.
          </p>
        </>
      )}
    </div>
  );
}
