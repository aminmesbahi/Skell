import { ChevronRight, FolderOpen, Plus } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { useRepoStore } from "@/store";
import type { RecentProject } from "@/store";
import { createProjectId } from "@/lib/navigation";

interface RecentProjectsProps {
  collapsed?: boolean;
  onAddProject?: () => void;
}

export function RecentProjects({ collapsed = false, onAddProject }: RecentProjectsProps) {
  const navigate = useNavigate();
  const { recentProjects, selectedRepo, openProject } = useRepoStore();

  const visibleProjects = recentProjects.slice(0, 5);
  if (visibleProjects.length === 0) return null;

  function handleSelect(project: RecentProject) {
    openProject(project.path);
    navigate(`/projects/${createProjectId(project.path)}/skills`);
  }

  return (
    <div className="mt-4 flex min-h-0 flex-col">
      {!collapsed && (
        <div className="mb-2 flex items-center justify-between px-2">
          <p className="text-[11px] font-semibold uppercase tracking-[0.22em] text-slate-600">Recent Projects</p>
          <button
            onClick={() => navigate("/projects")}
            className="text-xs font-medium text-slate-500 hover:text-slate-300"
          >
            View all
          </button>
        </div>
      )}

      <div className="space-y-1 overflow-y-auto pr-1">
        {visibleProjects.map((project) => {
          const isActive = selectedRepo === project.path;
          return (
            <button
              key={project.id}
              onClick={() => handleSelect(project)}
              className={`flex w-full items-center gap-2 rounded-lg px-2 py-2 text-left transition-colors ${
                isActive
                  ? "bg-brand-600/15 text-brand-300"
                  : "text-slate-500 hover:bg-white/5 hover:text-slate-200"
              }`}
              title={`${project.displayName}\n${project.path}`}
            >
              <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg border border-[#23304d] bg-[#11162a] text-slate-400">
                <FolderOpen size={14} />
              </div>
              {!collapsed && (
                <div className="min-w-0 flex-1">
                  <div className="truncate text-sm font-medium">{project.displayName}</div>
                  <div className="truncate text-xs text-slate-600">{project.path}</div>
                </div>
              )}
              {!collapsed && <ChevronRight size={14} className="shrink-0 text-slate-600" />}
            </button>
          );
        })}
      </div>

      {!collapsed && (
        <button
          onClick={onAddProject}
          className="mt-2 flex items-center gap-2 rounded-lg px-2 py-2 text-sm text-slate-500 hover:bg-white/5 hover:text-slate-200"
        >
          <Plus size={14} />
          Add Project
        </button>
      )}
    </div>
  );
}
