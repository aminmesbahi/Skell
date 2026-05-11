import { useEffect, useState, useCallback } from "react";
import { RefreshCw, Stethoscope, CheckCircle2 } from "lucide-react";
import { useRepoStore, useUIStore } from "@/store";
import { doctorCheck } from "@/lib/skell";
import type { DiagnosticEntry } from "@/lib/types";
import { SeverityBadge } from "@/components/Badges";

interface RepoIssues {
  repo: string;
  name: string;
  issues: DiagnosticEntry[];
  loading: boolean;
  error: string | null;
}

export function Doctor() {
  const { repos, selectedRepo } = useRepoStore();
  const { notify } = useUIStore();
  const targets =
    selectedRepo && selectedRepo !== "global" ? [selectedRepo] : repos;

  const [repoIssues, setRepoIssues] = useState<RepoIssues[]>([]);
  const [running, setRunning] = useState(false);

  const init = useCallback(() => {
    setRepoIssues(
      targets.map((repo) => ({
        repo,
        name: repo.split(/[/\\]/).at(-1) ?? repo,
        issues: [],
        loading: false,
        error: null,
      }))
    );
  }, [targets]);

  useEffect(() => {
    init();
  }, [init]);

  async function runDoctor(repo: string) {
    setRepoIssues((prev) =>
      prev.map((r) => (r.repo === repo ? { ...r, loading: true, error: null } : r))
    );
    try {
      const issues = await doctorCheck(repo);
      setRepoIssues((prev) =>
        prev.map((r) =>
          r.repo === repo ? { ...r, loading: false, issues } : r
        )
      );
    } catch (e) {
      setRepoIssues((prev) =>
        prev.map((r) =>
          r.repo === repo ? { ...r, loading: false, error: String(e) } : r
        )
      );
    }
  }

  async function runAll() {
    setRunning(true);
    for (const repo of targets) {
      await runDoctor(repo);
    }
    setRunning(false);
    const total = repoIssues.reduce((s, r) => s + r.issues.length, 0);
    notify({
      kind: total === 0 ? "success" : "error",
      title: total === 0 ? "No issues found" : `${total} issue${total !== 1 ? "s" : ""} found`,
    });
  }

  const totalErrors = repoIssues.reduce(
    (s, r) => s + r.issues.filter((i) => i.severity === "error").length,
    0
  );
  const totalWarnings = repoIssues.reduce(
    (s, r) => s + r.issues.filter((i) => i.severity === "warning").length,
    0
  );

  return (
    <div className="p-6 space-y-6 max-w-4xl mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-200 flex items-center gap-2">
            <Stethoscope size={22} className="text-brand-400" />
            Doctor
          </h1>
          <p className="text-sm text-slate-500 mt-0.5">
            Diagnose manifest, lock file, and install problems
          </p>
        </div>
        <button
          onClick={() => void runAll()}
          disabled={running || targets.length === 0}
          className="btn-primary"
        >
          <Stethoscope size={14} className={running ? "animate-pulse" : ""} />
          Run Diagnostics
        </button>
      </div>

      {/* Summary */}
      {(totalErrors > 0 || totalWarnings > 0) && (
        <div className="flex items-center gap-4">
          {totalErrors > 0 && (
            <div className="flex items-center gap-2 text-sm text-red-400 bg-red-500/10 border border-red-500/20 rounded-lg px-4 py-2">
              <span className="font-bold">{totalErrors}</span> error{totalErrors !== 1 ? "s" : ""}
            </div>
          )}
          {totalWarnings > 0 && (
            <div className="flex items-center gap-2 text-sm text-amber-400 bg-amber-500/10 border border-amber-500/20 rounded-lg px-4 py-2">
              <span className="font-bold">{totalWarnings}</span> warning{totalWarnings !== 1 ? "s" : ""}
            </div>
          )}
        </div>
      )}

      {targets.length === 0 ? (
        <div className="card text-center py-12 text-slate-500 text-sm">
          No projects. Add one first.
        </div>
      ) : (
        <div className="space-y-4">
          {repoIssues.map((ri) => (
            <div key={ri.repo} className="card">
              <div className="flex items-center justify-between mb-3">
                <div>
                  <p className="font-medium text-slate-200">{ri.name}</p>
                  <p className="text-xs text-slate-600 truncate">{ri.repo}</p>
                </div>
                <button
                  onClick={() => void runDoctor(ri.repo)}
                  disabled={ri.loading}
                  className="btn-ghost text-xs"
                >
                  <RefreshCw size={12} className={ri.loading ? "animate-spin" : ""} />
                  Check
                </button>
              </div>

              {ri.loading && (
                <div className="flex items-center gap-2 text-sm text-slate-500 py-2">
                  <div className="spinner w-4 h-4" />
                  Checking...
                </div>
              )}

              {ri.error && (
                <div className="bg-red-500/10 border border-red-500/20 rounded-lg px-4 py-3 text-sm text-red-400">
                  {ri.error}
                </div>
              )}

              {!ri.loading && !ri.error && ri.issues.length === 0 && (
                <div className="flex items-center gap-2 text-sm text-emerald-400 py-1">
                  <CheckCircle2 size={15} />
                  No issues found — all good!
                </div>
              )}

              {!ri.loading && ri.issues.length > 0 && (
                <div className="space-y-2 mt-1">
                  {ri.issues.map((issue, idx) => (
                    <IssueRow key={idx} issue={issue} />
                  ))}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function IssueRow({ issue }: { issue: DiagnosticEntry }) {
  return (
    <div
      className={`rounded-lg border px-4 py-3 space-y-1 ${
        issue.severity === "error"
          ? "bg-red-500/5 border-red-500/20"
          : issue.severity === "warning"
          ? "bg-amber-500/5 border-amber-500/20"
          : "bg-blue-500/5 border-blue-500/20"
      }`}
    >
      <div className="flex items-center gap-3">
        <SeverityBadge severity={issue.severity} />
        <span className="text-xs font-mono text-slate-500">{issue.code}</span>
        <span className="text-sm text-slate-300">{issue.message}</span>
      </div>
      {issue.hint && (
        <p className="text-xs text-slate-500 pl-1">
          💡 {issue.hint}
        </p>
      )}
    </div>
  );
}
