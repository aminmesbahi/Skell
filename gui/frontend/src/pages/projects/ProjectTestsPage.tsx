import { useState } from "react";
import { AlertTriangle, CheckCircle2, CircleSlash, FlaskConical, Play, XCircle } from "lucide-react";
import { Link } from "react-router";
import clsx from "clsx";
import { useProject } from "@/hooks/useProject";
import { useAsync } from "@/hooks/useAsync";
import { runSkillTests } from "@/lib/skell";
import { getProjectDisplayName } from "@/lib/navigation";
import { summarizeResult } from "@/lib/skilltests";
import { LoadState, MissingProject } from "@/components/LoadState";
import { ProjectPageHeader } from "@/components/ProjectPageHeader";
import { useRepoStore, useUIStore } from "@/store";
import type { SkillTestResult, TestCaseResult } from "@/lib/types";

function CaseLine({ c }: { c: TestCaseResult }) {
  const Icon = c.status === "pass" ? CheckCircle2 : c.status === "skipped" ? CircleSlash : c.status === "fail" ? XCircle : AlertTriangle;
  const tone = c.status === "pass" ? "text-emerald-400" : c.status === "skipped" ? "text-slate-500" : c.status === "fail" ? "text-red-400" : "text-amber-400";
  return (
    <li className="text-sm">
      <div className="flex items-center gap-2">
        <Icon size={14} className={clsx("shrink-0", tone)} aria-label={c.status} />
        <span className="text-slate-200">{c.name}</span>
        {c.status === "pass" && c.duration_ms > 0 && <span className="text-xs text-slate-600">{(c.duration_ms / 1000).toFixed(1)}s</span>}
      </div>
      {(c.status === "fail" || c.status === "error") && c.failures?.map((f) => <p key={f} className="ml-6 mt-0.5 text-xs text-slate-400">{f}</p>)}
    </li>
  );
}

/** Tests tab: the eval cases skills ship with, and how they do against your agent. */
export function ProjectTestsPage() {
  const project = useProject();
  const { notify } = useUIStore();
  const testAgent = useRepoStore((s) => s.testAgent);
  const [results, setResults] = useState<Record<string, SkillTestResult>>({});
  const [running, setRunning] = useState<string | null>(null);

  // Loading only checks the eval files (no agent), so it is instant and free.
  const state = useAsync(project, async () => (project ? runSkillTests({ repo: project }) : null));
  if (!project) return <MissingProject />;

  const base = state.data?.results ?? [];
  const merged = base.map((r) => results[r.skill] ?? r);
  const without = state.data?.without_evals ?? [];

  async function run(skill?: string) {
    if (!testAgent.trim()) return;
    setRunning(skill ?? "*");
    try {
      const out = await runSkillTests({ repo: project, skillName: skill, agent: testAgent });
      setResults((prev) => ({ ...prev, ...Object.fromEntries(out.results.map((r) => [r.skill, r])) }));
      const failed = out.results.filter((r) => r.failed + r.errored > 0).length;
      notify(failed
        ? { kind: "error", title: `${failed} skill${failed === 1 ? "" : "s"} have failing cases` }
        : { kind: "success", title: "All cases passed" });
    } catch (e) {
      notify({ kind: "error", title: "Could not run the tests", detail: String(e) });
    } finally {
      setRunning(null);
    }
  }

  return (
    <div className="mx-auto max-w-6xl space-y-6 px-6 py-8">
      <ProjectPageHeader
        projectPath={project}
        title={`Tests for ${getProjectDisplayName(project)}`}
        subtitle="Skills can ship example prompts and what a good answer looks like. Run them to see whether a skill still does its job."
        actions={
          <button className="btn-primary" disabled={!testAgent.trim() || running !== null || merged.length === 0} onClick={() => void run()}>
            <Play size={14} />{running === "*" ? "Running…" : "Run all cases"}
          </button>
        }
      />
      <LoadState loading={state.loading && !state.data} error={state.error} retry={() => void state.refresh()} />

      {!testAgent.trim() && state.data && merged.length > 0 && (
        <div className="rounded-xl border border-sky-500/30 bg-sky-500/10 p-4 text-sm text-sky-100" role="note">
          <p className="font-medium">Choose an agent to run the cases</p>
          <p className="mt-1 text-sky-100/80">
            Every case is a real call to your AI agent, so Skell only runs them when you set it up. Pick a command such as <code className="font-mono">claude -p</code> in{" "}
            <Link to="/settings" className="underline">Settings</Link>. Until then, only the eval files are checked.
          </p>
        </div>
      )}

      {state.data && merged.length === 0 && (
        <div className="card text-center">
          <FlaskConical size={28} className="mx-auto mb-2 text-slate-600" />
          <p className="font-medium text-slate-200">None of the installed skills ship test cases</p>
          <p className="mx-auto mt-1 max-w-lg text-sm text-slate-500">
            Skill authors add them in <code className="font-mono">evals/evals.yaml</code>. Writing your own skill? Create it with <code className="font-mono">skell new my-skill --evals</code> to start with an example.
          </p>
        </div>
      )}

      <div className="space-y-4">
        {merged.map((r) => {
          const bad = r.failed + r.errored > 0 || (r.static_problems?.length ?? 0) > 0;
          return (
            <section key={r.skill} className="card" aria-label={`Tests for ${r.skill}`}>
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <h3 className="flex items-center gap-2 font-semibold text-slate-200">
                    {bad ? <XCircle size={16} className="text-red-400" /> : r.passed > 0 ? <CheckCircle2 size={16} className="text-emerald-400" /> : <FlaskConical size={16} className="text-slate-500" />}
                    {r.skill}
                  </h3>
                  <p className="text-sm text-slate-500">{summarizeResult(r)}</p>
                </div>
                <button className="btn-subtle" disabled={!testAgent.trim() || running !== null} onClick={() => void run(r.skill)}>
                  <Play size={12} />{running === r.skill ? "Running…" : "Run"}
                </button>
              </div>
              {r.static_problems?.map((p) => <p key={p} className="mt-3 flex items-start gap-2 text-sm text-amber-300"><AlertTriangle size={14} className="mt-0.5 shrink-0" />{p}</p>)}
              <ul className="mt-3 space-y-2">
                {r.cases.map((c) => <CaseLine key={c.name} c={c} />)}
              </ul>
            </section>
          );
        })}
      </div>

      {without.length > 0 && merged.length > 0 && (
        <p className="text-sm text-slate-600">No test cases: {without.join(", ")}</p>
      )}
    </div>
  );
}
