import type { ReactNode } from "react";
import { ChevronRight } from "lucide-react";
import { Link, NavLink } from "react-router";
import clsx from "clsx";
import { getProjectDisplayName, createProjectId } from "@/lib/navigation";

interface ProjectPageHeaderProps {
  projectPath: string;
  title: string;
  subtitle?: string;
  actions?: ReactNode;
}

const TABS = [
  { to: "skills", label: "Skills" },
  { to: "agents", label: "Agents" },
  { to: "health", label: "Health" },
  { to: "activity", label: "Activity" },
];

/** Shared header for a project's pages: breadcrumb, title, and the tabs that switch between them. */
export function ProjectPageHeader({ projectPath, title, subtitle, actions }: ProjectPageHeaderProps) {
  const projectName = getProjectDisplayName(projectPath);
  const projectId = projectPath ? createProjectId(projectPath) : "";

  return (
    <div className="rounded-2xl border border-[var(--palette-1a1f35)] bg-[var(--palette-0f1324)] px-6 pt-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="min-w-0">
          <nav aria-label="Breadcrumb" className="flex items-center gap-2 text-sm text-slate-500">
            <Link to="/projects" className="hover:text-slate-300 transition-colors cursor-pointer">Projects</Link>
            <ChevronRight size={13} />
            <Link to={`/projects/${projectId}/skills`} className="font-medium text-slate-300 hover:text-slate-100 transition-colors cursor-pointer">{projectName}</Link>
          </nav>
          <h2 className="mt-2 text-2xl font-bold text-slate-200">{title}</h2>
          {subtitle && <p className="mt-1 text-sm text-slate-400">{subtitle}</p>}
          {projectPath && <p className="mt-2 truncate text-sm text-slate-600" title={projectPath}>{projectPath}</p>}
        </div>
        {actions && <div className="flex flex-wrap items-center gap-2">{actions}</div>}
      </div>
      {projectId && (
        <nav aria-label="Project sections" className="-mb-px mt-5 flex gap-1 overflow-x-auto">
          {TABS.map((tab) => (
            <NavLink
              key={tab.to}
              to={`/projects/${projectId}/${tab.to}`}
              className={({ isActive }) => clsx(
                "border-b-2 px-3 py-2.5 text-sm font-medium transition-colors",
                isActive ? "border-brand-500 text-slate-100" : "border-transparent text-slate-500 hover:text-slate-300",
              )}
            >
              {tab.label}
            </NavLink>
          ))}
        </nav>
      )}
      {!projectId && <div className="pb-6" />}
    </div>
  );
}
