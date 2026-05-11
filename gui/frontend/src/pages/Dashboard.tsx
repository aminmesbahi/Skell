import { useEffect, useState, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import {
  Package,
  ArrowUp,
  Pin,
  FolderOpen,
  Stethoscope,
  RefreshCw,
  GitBranch,
  Zap,
  Search,
} from "lucide-react";
import { useRepoStore } from "@/store";
import {
  listInstalled,
  getStatus,
  doctorCheck,
  listInstalledGlobal,
} from "@/lib/skell";
import type { InstalledSkill, StatusEntry, DiagnosticEntry } from "@/lib/types";
import { SkillBadge } from "@/components/Badges";

interface RepoStats {
  path: string;
  name: string;
  total: number;
  outdated: number;
  pinned: number;
  modified: number;
  issues: number;
}

export function Dashboard() {
  const navigate = useNavigate();
  const { repos } = useRepoStore();
  const [repoStats, setRepoStats] = useState<RepoStats[]>([]);
  const [globalSkills, setGlobalSkills] = useState<InstalledSkill[]>([]);
  const [recentStatuses, setRecentStatuses] = useState<StatusEntry[]>([]);
  const [loading, setLoading] = useState(true);

  const loadStats = useCallback(async () => {
    setLoading(true);
    try {
      // Global skills
      const gSkills = await listInstalledGlobal().catch(() => [] as InstalledSkill[]);
      setGlobalSkills(gSkills);

      // Per-repo stats
      const stats = await Promise.all(
        repos.map(async (repo) => {
          const [skills, statuses, issues] = await Promise.all([
            listInstalled(repo).catch(() => [] as InstalledSkill[]),
            getStatus(repo).catch(() => [] as StatusEntry[]),
            doctorCheck(repo).catch(() => [] as DiagnosticEntry[]),
          ]);
          return {
            path: repo,
            name: repo.split(/[/\\]/).at(-1) ?? repo,
            total: skills.length,
            outdated: statuses.filter((s) => s.status === "outdated").length,
            pinned: skills.filter((s) => s.pinned).length,
            modified: statuses.filter((s) => s.status === "locally-modified").length,
            issues: issues.filter((d) => d.severity === "error").length,
          };
        })
      );
      setRepoStats(stats);

      // Recent statuses across all repos (first repo only for dashboard)
      if (repos[0]) {
        const recent = await getStatus(repos[0]).catch(() => [] as StatusEntry[]);
        setRecentStatuses(recent.slice(0, 8));
      }
    } finally {
      setLoading(false);
    }
  }, [repos]);

  useEffect(() => {
    void loadStats();
  }, [loadStats]);

  const totalSkills =
    repoStats.reduce((s, r) => s + r.total, 0) + globalSkills.length;
  const totalOutdated = repoStats.reduce((s, r) => s + r.outdated, 0);
  const totalPinned = repoStats.reduce((s, r) => s + r.pinned, 0) + globalSkills.filter((s) => s.pinned).length;
  const totalIssues = repoStats.reduce((s, r) => s + r.issues, 0);

  return (
    <div className="p-6 space-y-6 max-w-6xl mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-200">Dashboard</h1>
          <p className="text-sm text-slate-500 mt-0.5">
            {repos.length} projects managed
          </p>
        </div>
        <button onClick={() => void loadStats()} className="btn-ghost" disabled={loading}>
          <RefreshCw size={14} className={loading ? "animate-spin" : ""} />
          Refresh
        </button>
      </div>

      {/* Stat cards */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          icon={<Package size={20} className="text-blue-400" />}
          label="Total Skills"
          value={totalSkills}
          bg="bg-blue-500/10"
          onClick={() => navigate("/skills")}
        />
        <StatCard
          icon={<ArrowUp size={20} className="text-amber-400" />}
          label="Outdated"
          value={totalOutdated}
          bg="bg-amber-500/10"
          highlight={totalOutdated > 0}
          onClick={() => navigate("/skills")}
        />
        <StatCard
          icon={<Pin size={20} className="text-blue-400" />}
          label="Pinned"
          value={totalPinned}
          bg="bg-blue-500/10"
          onClick={() => navigate("/skills")}
        />
        <StatCard
          icon={<Stethoscope size={20} className="text-red-400" />}
          label="Doctor Issues"
          value={totalIssues}
          bg="bg-red-500/10"
          highlight={totalIssues > 0}
          onClick={() => navigate("/doctor")}
        />
      </div>

      {/* Quick Actions — makes the main flows obvious */}
      <div className="flex flex-wrap gap-3">
        <button
          onClick={() => navigate("/registry")}
          className="flex-1 min-w-[220px] flex items-center justify-center gap-3 rounded-xl border border-[#1e2640] bg-[#111827] px-5 py-3.5 text-sm font-medium text-slate-200 hover:border-brand-500/60 hover:bg-[#141c2e] active:bg-[#0f1726] transition"
        >
          <Search size={18} className="text-brand-400" />
          <span>Browse &amp; Install Skills</span>
          <span className="ml-1 text-xs px-2 py-0.5 rounded bg-brand-600/20 text-brand-400">Discover</span>
        </button>

        <button
          onClick={() => navigate("/sync")}
          className="flex items-center justify-center gap-2 rounded-xl border border-[#1e2640] bg-[#111827] px-5 py-3.5 text-sm font-medium text-slate-200 hover:border-slate-600 active:bg-[#0f1726] transition"
        >
          <RefreshCw size={17} />
          Sync Projects
        </button>

        <button
          onClick={() => navigate("/doctor")}
          className="flex items-center justify-center gap-2 rounded-xl border border-[#1e2640] bg-[#111827] px-5 py-3.5 text-sm font-medium text-slate-200 hover:border-slate-600 active:bg-[#0f1726] transition"
        >
          <Stethoscope size={17} />
          Run Diagnostics
        </button>
      </div>

      {/* Repos overview */}
      {repos.length > 0 ? (
        <div className="card">
          <h2 className="section-title flex items-center gap-2">
            <FolderOpen size={18} className="text-brand-400" />
            Projects Overview
          </h2>
          <div className="space-y-2">
            {repoStats.map((r) => (
              <RepoRow
                key={r.path}
                repo={r}
                onClick={() => navigate("/skills")}
              />
            ))}
          </div>
        </div>
      ) : (
        <div className="card flex flex-col items-center justify-center py-12 text-center">
          <div className="w-16 h-16 rounded-2xl bg-brand-600/15 flex items-center justify-center mb-4">
            <Zap size={28} className="text-brand-400" />
          </div>
          <h3 className="font-semibold text-slate-300 mb-2">Get started</h3>
          <p className="text-sm text-slate-500 max-w-xs mb-4">
            Add a project to begin managing skills.
          </p>
          <button className="btn-primary" onClick={() => navigate("/repositories")}>
            Add Project
          </button>
        </div>
      )}

      {/* Recent status */}
      {recentStatuses.length > 0 && (
        <div className="card">
          <h2 className="section-title flex items-center gap-2">
            <GitBranch size={18} className="text-brand-400" />
            Skill Status
            <span className="text-xs text-slate-600 font-normal ml-1">
              — {repos[0]?.split(/[/\\]/).at(-1)}
            </span>
          </h2>
          <table className="data-table">
            <thead>
              <tr>
                <th>Skill</th>
                <th>Installed</th>
                <th>Latest</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              {recentStatuses.map((s) => (
                <tr key={s.name}>
                  <td className="font-medium text-slate-200">{s.name}</td>
                  <td className="font-mono text-xs text-slate-400">{s.installed}</td>
                  <td className="font-mono text-xs text-slate-400">{s.latest}</td>
                  <td>
                    <SkillBadge status={s.status} size="sm" />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

function StatCard({
  icon,
  label,
  value,
  bg,
  highlight,
  onClick,
}: {
  icon: React.ReactNode;
  label: string;
  value: number;
  bg: string;
  highlight?: boolean;
  onClick?: () => void;
}) {
  return (
    <button
      onClick={onClick}
      className={`card text-left hover:border-brand-600/40 transition-colors cursor-pointer ${
        highlight ? "border-amber-500/30" : ""
      }`}
    >
      <div className={`w-10 h-10 rounded-xl ${bg} flex items-center justify-center mb-3`}>
        {icon}
      </div>
      <p className="text-2xl font-bold text-slate-200">{value}</p>
      <p className="text-xs text-slate-500 mt-0.5">{label}</p>
    </button>
  );
}

function RepoRow({
  repo,
  onClick,
}: {
  repo: RepoStats;
  onClick: () => void;
}) {
  const health =
    repo.issues > 0
      ? "text-red-400"
      : repo.outdated > 0
      ? "text-amber-400"
      : "text-emerald-400";

  return (
    <button
      onClick={onClick}
      className="w-full flex items-center gap-4 p-3 rounded-lg hover:bg-white/5 transition-colors text-left"
    >
      <FolderOpen size={16} className="text-slate-500 shrink-0" />
      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium text-slate-300 truncate">{repo.name}</p>
        <p className="text-xs text-slate-600 truncate">{repo.path}</p>
      </div>
      <div className="flex items-center gap-3 text-xs text-slate-500 shrink-0">
        <span>{repo.total} skills</span>
        {repo.outdated > 0 && (
          <span className="text-amber-400 font-medium">
            {repo.outdated} outdated
          </span>
        )}
        {repo.pinned > 0 && <span>{repo.pinned} pinned</span>}
        <div className={`w-2 h-2 rounded-full ${health} bg-current`} />
      </div>
    </button>
  );
}
