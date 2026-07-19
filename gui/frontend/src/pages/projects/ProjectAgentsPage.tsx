import { useEffect, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import { useRepoStore } from "@/store";
import { getProjectDisplayName } from "@/lib/navigation";
import { detectRepoTargets, listSupportedTargets } from "@/lib/skell";
import type { AgentTarget } from "@/lib/skell";

export function ProjectAgentsPage() {
  const { projectId: _projectId } = useParams();
  const { repos, selectedRepo } = useRepoStore();

  const projectPath = useMemo(() => {
    if (selectedRepo && selectedRepo !== "global") return selectedRepo;
    return repos[0] ?? "";
  }, [repos, selectedRepo]);

  const [targets, setTargets] = useState<AgentTarget[]>([]);

  useEffect(() => {
    async function loadTargets() {
      if (!projectPath) return;
      const [detected, supported] = await Promise.all([detectRepoTargets(projectPath), listSupportedTargets()]);
      const combined = [...detected, ...supported.filter((item) => !detected.some((target) => target.id === item.id))];
      setTargets(combined);
    }

    void loadTargets();
  }, [projectPath]);

  return (
    <div className="mx-auto max-w-6xl px-6 py-8">
      <div className="card">
        <h2 className="text-2xl font-bold text-slate-200">Agents for {getProjectDisplayName(projectPath)}</h2>
        <p className="mt-2 text-sm text-slate-400">Detected and supported agent targets for this project.</p>
      </div>

      <div className="mt-6 space-y-3">
        {targets.map((target) => (
          <div key={target.id} className="card flex items-start justify-between gap-4">
            <div>
              <p className="font-medium text-slate-200">{target.displayName}</p>
              <p className="mt-1 text-sm text-slate-500">{target.dir}/skills/</p>
            </div>
            {target.detected ? (
              <span className="rounded-full border border-emerald-500/20 bg-emerald-500/10 px-2.5 py-1 text-xs font-medium text-emerald-400">detected</span>
            ) : (
              <span className="rounded-full border border-slate-700 bg-slate-800/60 px-2.5 py-1 text-xs font-medium text-slate-400">available</span>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
