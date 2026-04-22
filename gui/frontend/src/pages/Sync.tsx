import { useEffect, useState, useCallback } from "react";
import {
  RefreshCw,
  CheckCircle2,
  AlertTriangle,
  Plus,
  Minus,
  Play,
  Clock,
  Zap,
} from "lucide-react";
import { useRepoStore, useUIStore } from "@/store";
import { syncRepo } from "@/lib/skell";
import type { SyncReport } from "@/lib/types";

interface RepoSyncState {
  repo: string;
  name: string;
  report: SyncReport | null;
  /** True when `report` came from a dry-run, false when it came from an actual apply. */
  reportIsDryRun: boolean | null;
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
        reportIsDryRun: null,
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
          reportIsDryRun: dry ? s[repoPath].reportIsDryRun : false,
          // After a real apply, also refresh the preview slot so the card
          // shows the correct post-apply state regardless of the dry-run checkbox.
          dryRunReport: dry ? report : report,
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
            const hasPendingChanges = report !== null && !inSync;
            const totalChanges = report
              ? report.installed.length + report.removed.length
              : 0;

            // Status badge shown in card header
            let statusBadge: React.ReactNode = null;
            if (state.loading) {
              statusBadge = (
                <span className="flex items-center gap-1 text-xs text-slate-500 bg-slate-800 border border-slate-700 rounded-full px-2 py-0.5">
                  <div className="spinner w-3 h-3" />
                  Checking…
                </span>
              );
            } else if (state.error) {
              statusBadge = (
                <span className="flex items-center gap-1 text-xs text-red-400 bg-red-500/10 border border-red-500/20 rounded-full px-2 py-0.5">
                  <AlertTriangle size={11} />
                  Error
                </span>
              );
            } else if (report === null) {
              statusBadge = (
                <span className="flex items-center gap-1 text-xs text-slate-500 bg-slate-800 border border-slate-700 rounded-full px-2 py-0.5">
                  <Clock size={11} />
                  Not checked
                </span>
              );
            } else if (inSync && state.reportIsDryRun === false) {
              statusBadge = (
                <span className="flex items-center gap-1 text-xs text-emerald-400 bg-emerald-500/10 border border-emerald-500/20 rounded-full px-2 py-0.5">
                  <CheckCircle2 size={11} />
                  Applied
                </span>
              );
            } else if (inSync) {
              statusBadge = (
                <span className="flex items-center gap-1 text-xs text-emerald-400 bg-emerald-500/10 border border-emerald-500/20 rounded-full px-2 py-0.5">
                  <CheckCircle2 size={11} />
                  Up to date
                </span>
              );
            } else {
              statusBadge = (
                <span className="flex items-center gap-1 text-xs text-amber-400 bg-amber-500/10 border border-amber-500/20 rounded-full px-2 py-0.5">
                  <AlertTriangle size={11} />
                  {totalChanges} change{totalChanges !== 1 ? "s" : ""} pending
                </span>
              );
            }

            return (
              <div key={repo} className="card">
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-3 min-w-0">
                    <div className="min-w-0">
                      <p className="font-medium text-slate-200">{state.name}</p>
                      <p className="text-xs text-slate-600 truncate">{repo}</p>
                    </div>
                    {statusBadge}
                  </div>
                  <button
                    onClick={() => void runSync(repo, dryRun)}
                    disabled={state.loading}
                    className="btn-ghost text-xs shrink-0 ml-3"
                  >
                    <RefreshCw size={12} className={state.loading ? "animate-spin" : ""} />
                    {dryRun ? "Preview" : "Sync"}
                  </button>
                </div>

                {state.error && (
                  <div className="bg-red-500/10 border border-red-500/20 rounded-lg px-4 py-3 text-sm text-red-400">
                    <AlertTriangle size={14} className="inline mr-1.5" />
                    {state.error}
                  </div>
                )}

                {!state.loading && !state.error && report === null && (
                  <div className="flex items-center gap-2 text-sm text-slate-600 py-2 border-t border-slate-800 mt-1 pt-3">
                    <Clock size={14} className="shrink-0" />
                    Press &ldquo;{dryRun ? "Preview" : "Sync"}&rdquo; to check this repo.
                  </div>
                )}

                {!state.loading && !state.error && report !== null && (
                  <>
                    {/* After a real apply with no remaining changes */}
                    {inSync && state.reportIsDryRun === false && (
                      <div className="flex items-center gap-2 text-sm text-emerald-400 bg-emerald-500/8 border border-emerald-500/20 rounded-lg px-4 py-3 mt-1">
                        <CheckCircle2 size={16} className="shrink-0" />
                        <div>
                          <p className="font-medium">Sync applied successfully</p>
                          <p className="text-emerald-500/70 text-xs mt-0.5">All skills are now up to date with skell.toml</p>
                        </div>
                      </div>
                    )}

                    {/* Preview or real-apply: already in sync */}
                    {inSync && state.reportIsDryRun !== false && (
                      <div className="flex items-center gap-2 text-sm text-emerald-400 bg-emerald-500/8 border border-emerald-500/20 rounded-lg px-4 py-3 mt-1">
                        <CheckCircle2 size={16} className="shrink-0" />
                        <div>
                          <p className="font-medium">Already up to date</p>
                          <p className="text-emerald-500/70 text-xs mt-0.5">No changes needed — installed skills match skell.toml</p>
                        </div>
                      </div>
                    )}

                    {/* Pending changes */}
                    {hasPendingChanges && (
                      <div className="space-y-3 mt-1">
                        {/* Context banner */}
                        <div className="flex items-center justify-between bg-amber-500/8 border border-amber-500/20 rounded-lg px-4 py-2.5">
                          <div className="flex items-center gap-2 text-sm text-amber-300">
                            <AlertTriangle size={14} className="shrink-0" />
                            <span>
                              {dryRun
                                ? `Preview — ${totalChanges} change${totalChanges !== 1 ? "s" : ""} would be applied`
                                : `${totalChanges} change${totalChanges !== 1 ? "s" : ""} applied`}
                            </span>
                          </div>
                          {dryRun && (
                            <button
                              onClick={() => void runSync(repo, false)}
                              className="flex items-center gap-1.5 text-xs font-medium text-white bg-brand-600 hover:bg-brand-500 px-3 py-1.5 rounded-lg transition-colors"
                            >
                              <Zap size={11} />
                              Apply now
                            </button>
                          )}
                        </div>

                        {report.installed.length > 0 && (
                          <div>
                            <p className="text-xs font-semibold text-emerald-400 uppercase tracking-wider mb-1.5 flex items-center gap-1">
                              <Plus size={11} />
                              {dryRun ? "Will install" : "Installed"} ({report.installed.length})
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
                            <p className="text-xs font-semibold text-red-400 uppercase tracking-wider mb-1.5 flex items-center gap-1">
                              <Minus size={11} />
                              {dryRun ? "Will remove" : "Removed"} ({report.removed.length})
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
