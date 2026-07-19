import { useEffect, useMemo, useState, useCallback } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { AlertTriangle, FilePlus, RefreshCw, Search, Terminal, Library, GitBranchPlus, PackageOpen } from "lucide-react";
import { useRepoStore, useUIStore } from "@/store";
import { getProjectDisplayName } from "@/lib/navigation";
import { getStatus, initRepo, isRepoInitialized, listInstalled, skellPresent } from "@/lib/skell";
import type { InstalledSkill, StatusEntry } from "@/lib/types";
import { SkillBadge, ScopeBadge } from "@/components/Badges";
import { ProjectPageHeader } from "@/components/ProjectPageHeader";
import { AddSkillButton } from "@/components/AddSkillButton";

export function ProjectSkillsPage() {
  const { projectId: _projectId } = useParams();
  const navigate = useNavigate();
  const { repos, selectedRepo } = useRepoStore();
  const { notify } = useUIStore();

  const projectPath = useMemo(() => {
    if (selectedRepo && selectedRepo !== "global") return selectedRepo;
    return repos[0] ?? "";
  }, [repos, selectedRepo]);

  const [skills, setSkills] = useState<InstalledSkill[]>([]);
  const [statuses, setStatuses] = useState<StatusEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [repoInited, setRepoInited] = useState<boolean | null>(null);
  const [initRunning, setInitRunning] = useState(false);
  const [skellMissing, setSkellMissing] = useState(false);

  const loadSkills = useCallback(async () => {
    if (!projectPath) {
      setSkills([]);
      setStatuses([]);
      setLoading(false);
      return;
    }

    setLoading(true);
    try {
      const [installed, statusEntries] = await Promise.all([
        listInstalled(projectPath).catch(() => [] as InstalledSkill[]),
        getStatus(projectPath).catch(() => [] as StatusEntry[]),
      ]);
      setSkills(installed);
      setStatuses(statusEntries);
    } finally {
      setLoading(false);
    }
  }, [projectPath]);

  useEffect(() => {
    void loadSkills();
  }, [loadSkills]);

  useEffect(() => {
    if (!projectPath) return;
    setRepoInited(null);
    isRepoInitialized(projectPath)
      .then(setRepoInited)
      .catch(() => setRepoInited(false));
  }, [projectPath]);

  useEffect(() => {
    skellPresent()
      .then((present) => {
        if (!present) setSkellMissing(true);
      })
      .catch(() => {});
  }, []);

  async function handleInitHere() {
    if (!projectPath) return;
    setInitRunning(true);
    try {
      const result = await initRepo(projectPath);
      if (result.success) {
        notify({ kind: "success", title: "Project initialized", detail: result.stdout.trim() });
        setRepoInited(true);
        void loadSkills();
      } else {
        notify({ kind: "error", title: "Init failed", detail: result.stderr });
      }
    } finally {
      setInitRunning(false);
    }
  }

  const filtered = useMemo(() => {
    const q = search.toLowerCase();
    if (!q) return skills;
    return skills.filter((skill) => skill.name.toLowerCase().includes(q) || skill.registry.toLowerCase().includes(q));
  }, [search, skills]);

  const pageSubtitle = projectPath
    ? `Skills for ${getProjectDisplayName(projectPath)}.`
    : "Select a project to manage its skills.";

  return (
    <div className="mx-auto max-w-6xl px-6 py-8 space-y-6">
      <ProjectPageHeader
        projectPath={projectPath}
        title={`${getProjectDisplayName(projectPath)}${projectPath ? "" : ""}`}
        subtitle={pageSubtitle}
        breadcrumb="Skills"
        actions={
          <>
            <AddSkillButton projectPath={projectPath} onRefresh={() => void loadSkills()} />
            <button onClick={() => void loadSkills()} className="btn-ghost" disabled={loading}>
              <RefreshCw size={14} className={loading ? "animate-spin" : ""} />
              Refresh
            </button>
          </>
        }
      />

      {repoInited === false && (
        <div className="flex items-center gap-3 rounded-xl border border-amber-500/30 bg-amber-500/10 px-4 py-3 text-sm">
          <AlertTriangle size={16} className="text-amber-400 shrink-0" />
          <p className="flex-1 text-amber-300">This project has not been initialized yet. Initialize it to install skills.</p>
          <button onClick={() => void handleInitHere()} disabled={initRunning} className="shrink-0 flex items-center gap-1.5 text-xs font-medium px-3 py-1.5 rounded-lg bg-amber-500/20 hover:bg-amber-500/30 text-amber-200 border border-amber-500/30 transition-colors disabled:opacity-50">
            <FilePlus size={13} />
            {initRunning ? "Initializing…" : "Initialize now"}
          </button>
        </div>
      )}

      {skellMissing && (
        <div className="rounded-xl border border-amber-500/40 bg-amber-500/10 px-4 py-3 text-sm">
          <div className="flex items-start gap-3">
            <Terminal size={18} className="text-amber-400 mt-0.5 shrink-0" />
            <div className="flex-1 text-amber-200">
              <p className="font-medium">Skell CLI not found</p>
              <p className="mt-0.5 text-amber-300/90">Install the Skell CLI to manage and list skills.</p>
            </div>
            <button onClick={() => { setSkellMissing(false); void loadSkills(); }} className="shrink-0 rounded-lg border border-amber-500/40 bg-amber-500/20 px-3 py-1 text-xs text-amber-200 hover:bg-amber-500/30">Retry</button>
          </div>
        </div>
      )}

      <div className="flex items-center gap-3 flex-wrap">
        <div className="relative flex-1 min-w-48">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
          <input className="input pl-8" placeholder="Search skills..." value={search} onChange={(e) => setSearch(e.target.value)} />
        </div>
      </div>

      {loading ? (
        <div className="flex justify-center py-20">
          <div className="spinner w-8 h-8" />
        </div>
      ) : filtered.length === 0 ? (
        <div className="card flex flex-col items-center justify-center py-16 text-center">
          <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-2xl bg-brand-600/15 text-brand-400">
            <PackageOpen size={22} />
          </div>
          <h3 className="text-lg font-semibold text-slate-200">No skills installed yet</h3>
          <p className="mt-2 max-w-md text-sm text-slate-500">
            Add skills to give your coding agents reusable instructions, workflows, and supporting resources for this project.
          </p>
          <div className="mt-6 flex flex-wrap items-center justify-center gap-3">
            <button onClick={() => navigate("/catalog", { state: { installDestination: projectPath } })} className="btn-primary inline-flex items-center gap-2">
              <Library size={14} />
              Browse Catalog
            </button>
            <button onClick={() => { if (projectPath) { navigate("/catalog", { state: { installDestination: projectPath } }); } }} className="btn-ghost inline-flex items-center gap-2">
              <GitBranchPlus size={14} />
              Add from Repository
            </button>
          </div>
        </div>
      ) : (
        <div className="card p-0 overflow-hidden">
          <table className="data-table">
            <thead>
              <tr>
                <th>Skill</th>
                <th>Version</th>
                <th>Status</th>
                <th>Registry</th>
                <th>Scope</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((skill) => {
                const statusEntry = statuses.find((item) => item.name === skill.name);
                const status = statusEntry?.status ?? "unknown";
                return (
                  <tr key={skill.name}>
                    <td>
                      <button onClick={() => navigate(`/skills/${encodeURIComponent(skill.name)}`, { state: { repo: projectPath } })} className="font-medium text-brand-400 hover:text-brand-300 transition-colors cursor-pointer">
                        {skill.name}
                      </button>
                    </td>
                    <td className="font-mono text-xs">{skill.version || "—"}</td>
                    <td><SkillBadge status={status as typeof status} size="sm" /></td>
                    <td className="text-slate-400 text-xs">{skill.registry || "—"}</td>
                    <td><ScopeBadge scope="local" /></td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
