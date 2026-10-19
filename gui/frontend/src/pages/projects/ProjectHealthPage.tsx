import { useState } from "react";
import { Wrench } from "lucide-react";
import { useProject } from "@/hooks/useProject";
import { useAsync } from "@/hooks/useAsync";
import { doctorCheck, doctorFix, validateSkills } from "@/lib/skell";
import { getProjectDisplayName } from "@/lib/navigation";
import { LoadState, MissingProject } from "@/components/LoadState";
import { ProjectPageHeader } from "@/components/ProjectPageHeader";
import { ValidationReport } from "@/components/ValidationReport";
import { useUIStore } from "@/store";
import { describeIssue } from "@/lib/doctor";

/** Doctor issues that `skell doctor --fix` can repair without touching local edits. */
const FIXABLE = new Set(["missing-dir", "stale-mirror"]);

export function ProjectHealthPage() {
  const project = useProject();
  const { notify } = useUIStore();
  const [fixing, setFixing] = useState(false);
  const state = useAsync(project, async () => {
    if (!project) return null;
    const [doctor, validation] = await Promise.allSettled([doctorCheck(project), validateSkills(project, "", false)]);
    return { doctor, validation };
  });
  if (!project) return <MissingProject />;
  const issues = state.data?.doctor.status === "fulfilled" ? state.data.doctor.value ?? [] : [];
  const validations = state.data?.validation.status === "fulfilled" ? state.data.validation.value ?? [] : [];
  const failures = [state.data?.doctor, state.data?.validation].filter((r) => r?.status === "rejected");
  const errors = issues.filter((i) => i.severity === "error").length + validations.reduce((sum, v) => sum + v.errors, 0);
  const warnings = issues.filter((i) => i.severity === "warning").length + validations.reduce((sum, v) => sum + v.warnings, 0);
  const fixable = issues.filter((i) => FIXABLE.has(i.code));

  async function fix() {
    if (!project) return;
    setFixing(true);
    try {
      await doctorFix(project);
      notify({ kind: "success", title: "Repaired", detail: "Missing skills were reinstalled at their locked versions and mirrors refreshed." });
    } catch (e) {
      notify({ kind: "error", title: "Automatic fix failed", detail: String(e) });
    } finally {
      setFixing(false);
      await state.refresh();
    }
  }

  return <div className="mx-auto max-w-6xl px-6 py-8 space-y-6">
    <ProjectPageHeader projectPath={project} title={`Health for ${getProjectDisplayName(project)}`} actions={<div className="flex gap-2">
      {fixable.length > 0 && <button className="btn-primary" disabled={fixing || state.loading} onClick={() => void fix()}><Wrench size={14} />{fixing ? "Fixing…" : `Fix ${fixable.length} issue${fixable.length === 1 ? "" : "s"}`}</button>}
      <button className="btn-ghost" disabled={state.loading || fixing} onClick={() => void state.refresh()}>Refresh checks</button>
    </div>} />
    <LoadState loading={state.loading} error={state.error || (failures.length ? failures.map((r) => r?.status === "rejected" ? String(r.reason) : "").join("\n") : undefined)} retry={() => void state.refresh()} />
    {state.data && <>
      <div className="card"><p>{validations.length} skills checked · {errors} errors · {warnings} warnings</p>{failures.length > 0 && <p>Checks incomplete. These totals only include successful checks.</p>}</div>
      {issues.map((issue, i) => {
        const described = describeIssue(issue);
        return <div className="card" key={i}>
          <p className={issue.severity === "error" ? "text-red-400" : issue.severity === "warning" ? "text-amber-400" : "text-slate-300"}>
            <span className="font-semibold">{issue.severity === "error" ? "Error" : issue.severity === "warning" ? "Warning" : "Note"}:</span> {described.title}
            {FIXABLE.has(issue.code) && <span className="ml-2 text-xs text-emerald-400">fixable</span>}
          </p>
          {described.explanation && <p className="mt-1 text-sm text-slate-400">{described.explanation}</p>}
          <details className="mt-2 text-xs text-slate-500">
            <summary className="cursor-pointer hover:text-slate-300">Technical details</summary>
            <p className="mt-1 font-mono">{issue.code}: {issue.message}</p>
            {issue.hint && <p className="mt-1">{issue.hint}</p>}
          </details>
        </div>;
      })}
      {validations.map((result, i) => <section className="card" key={`${result.name}-${i}`}><h2 className="font-semibold mb-3">{result.name} {result.target && `(${result.target})`}</h2><ValidationReport result={result} /></section>)}
      {!state.loading && !failures.length && !errors && !warnings && <p className="card">No issues reported.</p>}
    </>}
  </div>;
}
