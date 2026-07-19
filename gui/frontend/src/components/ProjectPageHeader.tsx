import type { ReactNode } from "react";
import { ArrowLeft, ChevronRight } from "lucide-react";
import { useNavigate, Link } from "react-router-dom";
import { getProjectDisplayName, createProjectId } from "@/lib/navigation";

interface ProjectPageHeaderProps {
  projectPath: string;
  title: string;
  subtitle?: string;
  breadcrumb?: string;
  backLabel?: string;
  actions?: ReactNode;
}

export function ProjectPageHeader({ projectPath, title, subtitle, breadcrumb, backLabel = "Back to Projects", actions }: ProjectPageHeaderProps) {
  const navigate = useNavigate();
  const projectName = getProjectDisplayName(projectPath);
  const projectId = projectPath ? createProjectId(projectPath) : "";

  return (
    <div className="rounded-2xl border border-[#1a1f35] bg-[#0f1324] p-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="min-w-0">
          <button onClick={() => navigate("/projects")} className="mb-3 inline-flex items-center gap-2 text-sm text-slate-400 hover:text-slate-200">
            <ArrowLeft size={14} />
            {backLabel}
          </button>
          <div className="flex items-center gap-2 text-sm text-slate-500">
            <Link to="/projects" className="hover:text-slate-300 transition-colors cursor-pointer">Projects</Link>
            <ChevronRight size={13} />
            <Link to={`/projects/${projectId}`} className="font-medium text-slate-300 hover:text-slate-100 transition-colors cursor-pointer">{projectName}</Link>
            {breadcrumb && (
              <>
                <ChevronRight size={13} />
                <span className="font-medium text-slate-300">{breadcrumb}</span>
              </>
            )}
          </div>
          <h2 className="mt-2 text-2xl font-bold text-slate-200">{title}</h2>
          {subtitle && <p className="mt-1 text-sm text-slate-400">{subtitle}</p>}
          {projectPath && <p className="mt-2 truncate text-sm text-slate-600" title={projectPath}>{projectPath}</p>}
        </div>
        {actions && <div className="flex flex-wrap items-center gap-2">{actions}</div>}
      </div>
    </div>
  );
}
