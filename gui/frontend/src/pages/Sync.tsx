import { useEffect, useState, useCallback } from "react";
import {
  RefreshCw,
  CheckCircle2,
  AlertTriangle,
  Plus,
  Minus,
  Play,
} from "lucide-react";
import { useRepoStore, useUIStore } from "@/store";
import { syncRepo } from "@/lib/skell";
import type { SyncReport } from "@/lib/types";

interface RepoSyncState {
  repo: string;
  name: string;
  report: SyncReport | null;
  error: string | null;
  loading: boolean;
  dryRunReport: SyncReport | null;
}

export function Sync() {
  const { repos, selectedRepo } = useRepoStore();
  const { notify } = useUIStore();

  const targets =
    selectedRepo && selectedRepo !== "global"
      ? [selectedRepo]
      : repos;

  const [states, setStates] = useState<Record<string, RepoSyncState>>({});
  const [dryRun, setDryRun] = useState(true);
  const [running, setRunning] = useState(false);

  const initStates = useCallback(() => {
    const initial: Record<string, RepoSyncState> = {};
    for (const repo of targets) {
      initial[repo] = {
        repo,
        name: repo.split(/[/\\]/).at(-1) ?? repo,
        report: null,
        error: null,
        loading: false,
        dryRunReport: null,
      };
    }
    setStates(initial);
  }, [targets]);

  useEffect(() => {
    initStates();
  }, [initStates]);

  async function runSync(repoPath: string, dry: boolean) {
    setStates((s) => ({
      ...s,
      [repoPath]: { ...s[repoPath], loading: true, error: null },
    }));
    try {
      const report = await syncRepo({ repo: repoPath, dryRun: dry });
      setStates((s) => ({
        ...s,
        [repoPath]: {
          ...s[repoPath],
          loading: false,
          report: dry ? s[repoPath].report : report,
          dryRunReport: dry ? report : s[repoPath].dryRunReport,
          error: null,
        },
      }));
    } catch (e) {
      setStates((s) => ({
        ...s,
        [repoPath]: {
          ...s[repoPath],
          loading: false,
          error: String(e),
        },
      }));
    }
  }

  async function runAll() {
    setRunning(true);
    for (const repo of targets) {
      await runSync(repo, dryRun);
    }
    setRunning(false);
    notify({
      kind: dryRun ? "info" : "success",
      title: dryRun ? "Dry-run complete" : "Sync complete",
      detail: `${targets.length} repo${targets.length !== 1 ? "s" : ""} processed`,
    });
  }

  return (
    <div className="p-6 space-y-6 max-w-4xl mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-200">Sync</h1>
          <p className="text-sm text-slate-500 mt-0.5">
            Reconcile installed skills with skell.toml
          </p>
        </div>
        <div className="flex items-center gap-3">
          <label className="flex items-center gap-2 text-sm text-slate-400 cursor-pointer">
            <input
              type="checkbox"
              checked={dryRun}
              onChange={(e) => setDryRun(e.target.checked)}
              className="w-4 h-4 rounded border-slate-600 bg-slate-800 accent-brand-600"
            />
            Dry-run
          </label>
          <button
            onClick={() => void runAll()}
            disabled={running || targets.length === 0}
            className="btn-primary"
          >
            <Play size={14} className={running ? "animate-pulse" : ""} />
            {dryRun ? "Preview Sync" : "Apply Sync"}
          </button>
        </div>
      </div>

      {targets.length === 0 ? (
        <div className="card text-center py-12 text-slate-500 text-sm">
          No repositories selected. Add a repo first.
        </div>
      ) : (
        <div className="space-y-4">
          {targets.map((repo) => {
            const state = states[repo];
            if (!state) return null;
            const report = dryRun ? state.dryRunReport : state.report;
            const inSync =
              report !== null &&
              report.installed.length === 0 &&
              report.removed.length === 0;

            return (
              <div key={repo} className="card">
                <div className="flex items-center justify-between mb-3">
                  <div>
                    <p className="font-medium text-slate-200">{state.name}</p>
                    <p className="text-xs text-slate-600 truncate">{repo}</p>
                  </div>
                  <button
                    onClick={() => void runSync(repo, dryRun)}
                    disabled={state.loading}
                    className="btn-ghost text-xs"
                  >
                    <RefreshCw size={12} className={state.loading ? "animate-spin" : ""} />
                    {dryRun ? "Preview" : "Sync"}
                  </button>
                </div>

                {state.loading && (
                  <div className="flex items-center gap-2 text-sm text-slate-500 py-2">
                    <div className="spinner w-4 h-4" />
                    Running...
                  </div>
                )}

                {state.error && (
                  <div className="bg-red-500/10 border border-red-500/20 rounded-lg px-4 py-3 text-sm text-red-400">
                    <AlertTriangle size={14} className="inline mr-1.5" />
                    {state.error}
                  </div>
                )}

                {!state.loading && !state.error && report === null && (
                  <p className="text-sm text-slate-600 py-2">
                    Press "{dryRun ? "Preview" : "Sync"}" to check this repo.
                  </p>
                )}

                {!state.loading && report !== null && (
                  <>
                    {inSync ? (
                      <div className="flex items-center gap-2 text-sm text-emerald-400 py-1">
                        <CheckCircle2 size={16} />
                        Already in sync — no changes needed
                      </div>
                    ) : (
                      <div className="space-y-3 mt-2">
                        {report.installed.length > 0 && (
                          <div>
                            <p className="text-xs font-semibold text-emerald-400 uppercase tracking-wider mb-1">
                              <Plus size={11} className="inline mr-1" />
                              To install ({report.installed.length})
                            </p>
                            <ul className="space-y-1">
                              {report.installed.map((sk) => (
                                <li
                                  key={sk}
                                  className="flex items-center gap-2 text-sm text-slate-300 bg-emerald-500/5 border border-emerald-500/15 rounded-lg px-3 py-1.5"
                                >
                                  <Plus size={12} className="text-emerald-400 shrink-0" />
                                  {sk}
                                </li>
                              ))}
                            </ul>
                          </div>
                        )}
                        {report.removed.length > 0 && (
                          <div>
                            <p className="text-xs font-semibold text-red-400 uppercase tracking-wider mb-1">
                              <Minus size={11} className="inline mr-1" />
                              To remove ({report.removed.length})
                            </p>
                            <ul className="space-y-1">
                              {report.removed.map((sk) => (
                                <li
                                  key={sk}
                                  className="flex items-center gap-2 text-sm text-slate-300 bg-red-500/5 border border-red-500/15 rounded-lg px-3 py-1.5"
                                >
                                  <Minus size={12} className="text-red-400 shrink-0" />
                                  {sk}
                                </li>
                              ))}
                            </ul>
                          </div>
                        )}
                      </div>
                    )}
                  </>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
