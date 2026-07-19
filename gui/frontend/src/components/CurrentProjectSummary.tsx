import { FolderKanban, ChevronRight } from "lucide-react";

interface CurrentProjectSummaryProps {
  projectPath: string;
  collapsed?: boolean;
  onOpen?: () => void;
}

export function CurrentProjectSummary({ projectPath, collapsed = false, onOpen }: CurrentProjectSummaryProps) {
  if (!projectPath) return null;

  const displayName = projectPath.split(/[\\/]/).filter(Boolean).pop() ?? "Project";

  return (
    <button
      type="button"
      onClick={onOpen}
      disabled={!onOpen}
      className="w-full rounded-xl border border-[#1f2740] bg-[#11162a]/80 p-3 text-left transition-colors enabled:cursor-pointer enabled:hover:border-[#334268] enabled:hover:bg-[#151b32]"
      title={`Open ${displayName}\n${projectPath}`}
      aria-label={`Open current project ${displayName}`}
    >
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-2 min-w-0">
          <div className="rounded-lg bg-brand-600/15 p-1.5 text-brand-400">
            <FolderKanban size={14} />
          </div>
          {!collapsed && (
            <div className="min-w-0">
              <p className="text-[11px] uppercase tracking-wide text-slate-500">Current project</p>
              <p className="truncate text-sm font-medium text-slate-200" title={projectPath}>{displayName}</p>
            </div>
          )}
        </div>
        {!collapsed && onOpen && (
          <span className="rounded-md p-1 text-slate-500">
            <ChevronRight size={14} />
          </span>
        )}
      </div>
      {!collapsed && (
        <p className="mt-2 truncate text-xs text-slate-600" title={projectPath}>{projectPath}</p>
      )}
    </button>
  );
}
