import { useProject } from "@/hooks/useProject";
import { useAsync } from "@/hooks/useAsync";
import { LoadState, MissingProject } from "@/components/LoadState";
import { Link, useParams } from "react-router";
import { createProjectId, getProjectDisplayName } from "@/lib/navigation";
import { getStatus, listInstalled, doctorCheck, detectRepoTargets, isRepoInitialized } from "@/lib/skell";
import { FolderKanban, Package, ArrowUp, AlertTriangle, CheckCircle2, ShieldCheck } from "lucide-react";

export function ProjectOverview() {
  const { projectId } = useParams();
  const projectPath = useProject();
  const projectName = getProjectDisplayName(projectPath);
  const routeId = projectId ?? createProjectId(projectPath);
  const state = useAsync(projectPath, async () => {
    if (!projectPath) return null;
    const [skills, statuses, issues, targets, inited] = await Promise.all([listInstalled(projectPath), getStatus(projectPath), doctorCheck(projectPath), detectRepoTargets(projectPath), isRepoInitialized(projectPath)]);
    return { skills, statuses, issues, targets, inited };
  });
  const { skills = [], statuses = [], issues = [], targets = [], inited = null } = state.data ?? {};
  if (!projectPath) return <MissingProject />;
  if (!state.data) return <div className="p-6"><LoadState {...state} retry={() => void state.refresh()} /></div>;

  const outdated = statuses.filter((s) => s.status === "outdated").length;
  const errors = issues.filter((d) => d.severity === "error").length;
  const warnings = issues.filter((d) => d.severity === "warning").length;
  const detectedTarget = targets.find((t) => t.detected);

  return (
    <div className="mx-auto max-w-6xl px-6 py-8 space-y-6">
      <LoadState {...state} retry={() => void state.refresh()} />
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
        <Link to={`/projects/${routeId}/skills`} className="card hover:border-[var(--palette-2d3a5a)] transition-colors">
          <Package size={18} className="text-brand-400 mb-2" />
          <p className="font-medium text-slate-200">Skills</p>
          <p className="text-xs text-slate-500 mt-1">View and manage installed skills</p>
        </Link>
        <Link to={`/projects/${routeId}/health`} className="card hover:border-[var(--palette-2d3a5a)] transition-colors">
          <ShieldCheck size={18} className="text-emerald-400 mb-2" />
          <p className="font-medium text-slate-200">Health</p>
          <p className="text-xs text-slate-500 mt-1">Validation and doctor diagnostics</p>
        </Link>
        <Link to={`/projects/${routeId}/activity`} className="card hover:border-[var(--palette-2d3a5a)] transition-colors">
          <ArrowUp size={18} className="text-amber-400 mb-2" />
          <p className="font-medium text-slate-200">Activity</p>
          <p className="text-xs text-slate-500 mt-1">Status and update history</p>
        </Link>
      </div>
    </div>
  );
}

