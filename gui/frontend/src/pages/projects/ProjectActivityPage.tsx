import { useEffect, useMemo, useState } from "react";
import { useParams } from "react-router";
import { Activity, ArrowUpRight } from "lucide-react";
import { useRepoStore } from "@/store";
import { getProjectDisplayName } from "@/lib/navigation";
import { getStatus } from "@/lib/skell";
import type { StatusEntry } from "@/lib/types";

export function ProjectActivityPage() {
  const { projectId: _projectId } = useParams();
  const { repos, selectedRepo } = useRepoStore();

  const projectPath = useMemo(() => {
    if (selectedRepo && selectedRepo !== "global") return selectedRepo;
    return repos[0] ?? "";
  }, [repos, selectedRepo]);

  const [events, setEvents] = useState<StatusEntry[]>([]);

  useEffect(() => {
    async function loadActivity() {
      if (!projectPath) return;
      const entries = await getStatus(projectPath).catch(() => [] as StatusEntry[]);
      setEvents(entries);
    }

    void loadActivity();
  }, [projectPath]);

  return (
    <div className="mx-auto max-w-6xl px-6 py-8 space-y-6">
      <div className="card">
        <h2 className="text-2xl font-bold text-slate-200">Activity for {getProjectDisplayName(projectPath)}</h2>
        <p className="mt-2 text-sm text-slate-400">Recent skill state and update status for this project.</p>
      </div>

      <div className="space-y-3">
        {events.length === 0 ? (
          <div className="card text-sm text-slate-500">No activity recorded yet.</div>
        ) : (
          events.map((entry) => (
            <div key={entry.name} className="card flex items-center justify-between gap-4">
              <div className="flex items-center gap-3">
                <div className="rounded-full bg-white/5 p-2 text-slate-400">
                  <Activity size={15} />
                </div>
                <div>
                  <p className="font-medium text-slate-200">{entry.name}</p>
                  <p className="text-sm text-slate-500">{entry.installed} → {entry.latest}</p>
                </div>
              </div>
              <div className="flex items-center gap-2 text-sm text-slate-400">
                <span className="rounded-full border border-slate-700 bg-slate-800/60 px-2.5 py-1 text-xs font-medium">{entry.status}</span>
                <ArrowUpRight size={14} />
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
