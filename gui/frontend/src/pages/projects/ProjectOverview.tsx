import { useEffect, useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { useRepoStore } from "@/store";
import { createProjectId, getProjectDisplayName } from "@/lib/navigation";
import { getStatus, listInstalled, doctorCheck, detectRepoTargets, isRepoInitialized } from "@/lib/skell";
import type { StatusEntry, DiagnosticEntry, InstalledSkill } from "@/lib/types";
import { FolderKanban, Package, ArrowUp, AlertTriangle, CheckCircle2, ShieldCheck } from "lucide-react";
import type { AgentTarget } from "@/lib/skell";

export function ProjectOverview() {
  const { projectId } = useParams();
  const { repos, selectedRepo } = useRepoStore();

  const projectPath = useMemo(() => {
    if (selectedRepo && selectedRepo !== "global") return selectedRepo;
    return repos[0] ?? "";
  }, [repos, selectedRepo]);

  const projectName = getProjectDisplayName(projectPath);
  const routeId = projectId ?? createProjectId(projectPath);

  const [skills, setSkills] = useState<InstalledSkill[]>([]);
  const [statuses, setStatuses] = useState<StatusEntry[]>([]);
  const [issues, setIssues] = useState<DiagnosticEntry[]>([]);
  const [targets, setTargets] = useState<AgentTarget[]>([]);
  const [inited, setInited] = useState<boolean | null>(null);

  useEffect(() => {
    if (!projectPath) return;
    void Promise.all([
      listInstalled(projectPath).catch(() => [] as InstalledSkill[]),
      getStatus(projectPath).catch(() => [] as StatusEntry[]),
      doctorCheck(projectPath).catch(() => [] as DiagnosticEntry[]),
      detectRepoTargets(projectPath).catch(() => [] as AgentTarget[]),
      isRepoInitialized(projectPath).catch(() => false),
    ]).then(([sk, st, diag, tgt, init]) => {
      setSkills(sk);
      setStatuses(st);
      setIssues(diag);
      setTargets(tgt);
      setInited(init);
    });
  }, [projectPath]);

  const outdated = statuses.filter((s) => s.status === "outdated").length;
  const errors = issues.filter((d) => d.severity === "error").length;
  const warnings = issues.filter((d) => d.severity === "warning").length;
  const detectedTarget = targets.find((t) => t.detected);

  return (
    <div className="mx-auto max-w-6xl px-6 py-8 space-y-6">
      {/* Header */}
      <div className="card">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold text-slate-200">{projectName}</h1>
            <p className="mt-1 text-sm text-slate-400">{projectPath || "Select a project to view its details."}</p>
          </div>
          <Link to={`/projects/${routeId}/skills`} className="btn-primary text-sm">
            Manage Skills
          </Link>
        </div>
      </div>

      {/* Stats grid */}
      <div className="grid gap-4 md:grid-cols-3">
        <div className="card flex items-center gap-3">
          <div className="rounded-lg bg-brand-600/15 p-2 text-brand-400">
            <Package size={18} />
          </div>
          <div>
            <p className="text-2xl font-bold text-slate-200">{skills.length}</p>
            <p className="text-xs text-slate-500">Skills installed</p>
          </div>
        </div>
        <div className="card flex items-center gap-3">
          <div className="rounded-lg bg-amber-500/15 p-2 text-amber-400">
            <ArrowUp size={18} />
          </div>
          <div>
            <p className="text-2xl font-bold text-slate-200">{outdated}</p>
            <p className="text-xs text-slate-500">Outdated</p>
          </div>
        </div>
        <div className="card flex items-center gap-3">
          <div className={`rounded-lg p-2 ${errors > 0 ? "bg-red-500/15 text-red-400" : "bg-emerald-500/15 text-emerald-400"}`}>
            {errors > 0 ? <AlertTriangle size={18} /> : <CheckCircle2 size={18} />}
          </div>
          <div>
            <p className="text-2xl font-bold text-slate-200">{errors}</p>
            <p className="text-xs text-slate-500">Issues ({warnings} warnings)</p>
          </div>
        </div>
      </div>

      {/* Agent target & init status */}
      <div className="card">
        <div className="flex items-center gap-3">
          <div className="rounded-lg bg-teal-500/10 p-2 text-teal-400">
            <FolderKanban size={18} />
          </div>
          <div className="flex-1">
            <p className="font-medium text-slate-200">
              {detectedTarget
                ? `Initialized for ${detectedTarget.displayName}`
                : inited === false
                ? "Not initialized"
                : "Agent targets"}
            </p>
            <p className="text-xs text-slate-500 mt-0.5">
              {detectedTarget
                ? `Skills directory: ${detectedTarget.dir}/skills/`
                : targets.length > 0
                ? `${targets.filter((t) => t.detected).length} of ${targets.length} targets detected`
                : "No agent targets detected"}
            </p>
          </div>
          <Link to={`/projects/${routeId}/agents`} className="btn-ghost text-xs">
            View agents
          </Link>
        </div>
      </div>

      {/* Quick links */}
      <div className="grid gap-3 md:grid-cols-3">
        <Link to={`/projects/${routeId}/skills`} className="card hover:border-[#2d3a5a] transition-colors">
          <Package size={18} className="text-brand-400 mb-2" />
          <p className="font-medium text-slate-200">Skills</p>
          <p className="text-xs text-slate-500 mt-1">View and manage installed skills</p>
        </Link>
        <Link to={`/projects/${routeId}/health`} className="card hover:border-[#2d3a5a] transition-colors">
          <ShieldCheck size={18} className="text-emerald-400 mb-2" />
          <p className="font-medium text-slate-200">Health</p>
          <p className="text-xs text-slate-500 mt-1">Validation and doctor diagnostics</p>
        </Link>
        <Link to={`/projects/${routeId}/activity`} className="card hover:border-[#2d3a5a] transition-colors">
          <ArrowUp size={18} className="text-amber-400 mb-2" />
          <p className="font-medium text-slate-200">Activity</p>
          <p className="text-xs text-slate-500 mt-1">Status and update history</p>
        </Link>
      </div>
    </div>
  );
}

