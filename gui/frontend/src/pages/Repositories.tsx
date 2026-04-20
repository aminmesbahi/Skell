import { useEffect, useState, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { SelectDirectory } from "../../wailsjs/go/main/App";
import {
  FolderOpen,
  Plus,
  Trash2,
  RefreshCw,
  FilePlus,
  Globe,
  FolderClosed,
  ChevronRight,
  AlertTriangle,
  CheckCircle2,
  ArrowUp,
} from "lucide-react";
import { useRepoStore, useUIStore } from "@/store";
import {
  listInstalled,
  getStatus,
  doctorCheck,
  initRepo,
} from "@/lib/skell";
import type { DiagnosticEntry, StatusEntry } from "@/lib/types";
import { ConfirmDialog } from "@/components/ConfirmDialog";

interface RepoHealth {
  total: number;
  outdated: number;
  errors: number;
}

export function Repositories() {
  const navigate = useNavigate();
  const { repos, addRepo, removeRepo, setSelectedRepo } = useRepoStore();
  const { notify } = useUIStore();

  const [health, setHealth] = useState<Record<string, RepoHealth>>({});
  const [loading, setLoading] = useState(false);
  const [removing, setRemoving] = useState<string | null>(null);
  const [initialising, setInitialising] = useState<string | null>(null);

  const loadHealth = useCallback(async () => {
    setLoading(true);
    const entries = await Promise.all(
      repos.map(async (repo) => {
        const [skills, statuses, issues] = await Promise.all([
          listInstalled(repo).catch(() => []),
          getStatus(repo).catch(() => [] as StatusEntry[]),
          doctorCheck(repo).catch(() => [] as DiagnosticEntry[]),
        ]);
        return [
          repo,
          {
            total: skills.length,
            outdated: statuses.filter((s) => s.status === "outdated").length,
            errors: issues.filter((d) => d.severity === "error").length,
          },
        ] as [string, RepoHealth];
      })
    );
    setHealth(Object.fromEntries(entries));
    setLoading(false);
  }, [repos]);

  useEffect(() => {
    void loadHealth();
  }, [loadHealth]);

  async function handleAddRepo() {
    const selected = await SelectDirectory();
    if (selected) {
      addRepo(selected);
      notify({ kind: "success", title: "Repository added", detail: selected });
    }
  }

  async function handleInit(repo: string) {
    setInitialising(repo);
    try {
      const result = await initRepo(repo);
      if (result.success) {
        notify({ kind: "success", title: "Repo initialized", detail: result.stdout.trim() });
        void loadHealth();
      } else {
        notify({ kind: "error", title: "Init failed", detail: result.stderr });
      }
    } finally {
      setInitialising(null);
    }
  }

  function handleSelectAndNavigate(repo: string) {
    setSelectedRepo(repo);
    navigate("/skills");
  }

  return (
    <div className="p-6 space-y-6 max-w-4xl mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-200">Repositories</h1>
          <p className="text-sm text-slate-500 mt-0.5">
            Manage the local repositories Skell tracks
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button onClick={() => void loadHealth()} className="btn-ghost" disabled={loading}>
            <RefreshCw size={14} className={loading ? "animate-spin" : ""} />
          </button>
          <button onClick={handleAddRepo} className="btn-primary">
            <Plus size={16} />
            Add Repository
          </button>
        </div>
      </div>

      {/* Global entry */}
      <div className="card flex items-center gap-4">
        <div className="w-10 h-10 rounded-xl bg-indigo-500/15 flex items-center justify-center shrink-0">
          <Globe size={20} className="text-indigo-400" />
        </div>
        <div className="flex-1">
          <p className="font-medium text-slate-200">Global Skills</p>
          <p className="text-xs text-slate-500 mt-0.5">~/.skell/.claude/skell.toml</p>
        </div>
        <button
          onClick={() => {
            setSelectedRepo("global");
            navigate("/skills");
          }}
          className="btn-ghost text-xs"
        >
          View Skills
          <ChevronRight size={14} />
        </button>
      </div>

      {/* Repo list */}
      {repos.length === 0 ? (
        <div className="card flex flex-col items-center py-16 text-center">
          <FolderOpen size={40} className="text-slate-700 mb-4" />
          <h3 className="font-medium text-slate-400 mb-1">No repositories</h3>
          <p className="text-sm text-slate-600 mb-4">
            Add a repository folder to get started.
          </p>
          <button onClick={handleAddRepo} className="btn-primary">
            <Plus size={16} />
            Add Repository
          </button>
        </div>
      ) : (
        <div className="space-y-3">
          {repos.map((repo) => {
            const name = repo.split(/[/\\]/).at(-1) ?? repo;
            const h = health[repo];
            return (
              <div
                key={repo}
                className="card hover:border-[#2d3a5a] transition-colors"
              >
                <div className="flex items-center gap-4">
                  <div className="w-10 h-10 rounded-xl bg-teal-500/10 flex items-center justify-center shrink-0">
                    <FolderClosed size={20} className="text-teal-400" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <p className="font-medium text-slate-200">{name}</p>
                      {h && <HealthDot health={h} />}
                    </div>
                    <p className="text-xs text-slate-600 truncate mt-0.5">{repo}</p>
                  </div>
                  {h && (
                    <div className="hidden md:flex items-center gap-4 text-xs text-slate-500 shrink-0">
                      <span>{h.total} skills</span>
                      {h.outdated > 0 && (
                        <span className="flex items-center gap-1 text-amber-400">
                          <ArrowUp size={11} />
                          {h.outdated} outdated
                        </span>
                      )}
                      {h.errors > 0 && (
                        <span className="flex items-center gap-1 text-red-400">
                          <AlertTriangle size={11} />
                          {h.errors} errors
                        </span>
                      )}
                    </div>
                  )}
                  <div className="flex items-center gap-2 shrink-0">
                    <button
                      onClick={() => void handleInit(repo)}
                      disabled={initialising === repo}
                      className="btn-ghost text-xs"
                      title="Run skell init in this repo"
                    >
                      <FilePlus size={13} />
                      Init
                    </button>
                    <button
                      onClick={() => handleSelectAndNavigate(repo)}
                      className="btn-ghost text-xs"
                    >
                      Skills
                      <ChevronRight size={13} />
                    </button>
                    <button
                      onClick={() => setRemoving(repo)}
                      className="p-1.5 rounded-lg text-slate-600 hover:text-red-400 hover:bg-red-500/10 transition-colors"
                      title="Remove from Skell"
                    >
                      <Trash2 size={14} />
                    </button>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Remove confirm */}
      <ConfirmDialog
        open={removing !== null}
        title="Remove repository"
        description={`Remove "${removing?.split(/[/\\]/).at(-1)}" from Skell? No files will be deleted.`}
        confirmLabel="Remove"
        danger
        onConfirm={() => {
          if (removing) removeRepo(removing);
          setRemoving(null);
        }}
        onCancel={() => setRemoving(null)}
      />
    </div>
  );
}

function HealthDot({ health }: { health: RepoHealth }) {
  if (health.errors > 0)
    return <AlertTriangle size={13} className="text-red-400" />;
  if (health.outdated > 0)
    return <ArrowUp size={13} className="text-amber-400" />;
  return <CheckCircle2 size={13} className="text-emerald-400" />;
}
