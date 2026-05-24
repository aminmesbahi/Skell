import { Download, Eye } from "lucide-react";
import type { RegistrySkill } from "@/lib/types";
import { LifecycleBadge } from "./Badges";

interface SkillCardProps {
  skill: RegistrySkill;
  installing: boolean;
  installed: boolean;
  canInstall: boolean;
  onInstall: () => void;
  onPreview: () => void;
}

export function SkillCard({
  skill,
  installing,
  installed,
  canInstall,
  onInstall,
  onPreview,
}: SkillCardProps) {
  const tags = skill.metadata?.tags?.split(",").map((t) => t.trim()).filter(Boolean) ?? [];

  return (
    <div className="card hover:border-[#2d3a5a] transition-colors flex flex-col gap-3">
      <div className="flex items-start justify-between gap-2">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <p className="font-semibold text-slate-200 text-sm">{skill.name}</p>
            {skill.registry_source === "global" && (
              <span className="text-[10px] font-medium px-1.5 py-0.5 rounded-full bg-violet-500/20 text-violet-300 border border-violet-500/30">
                shared
              </span>
            )}
            {skill.registry_source === "local" && (
              <span className="text-[10px] font-medium px-1.5 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 border border-indigo-500/30">
                project
              </span>
            )}
          </div>
          {skill.description && (
            <p className="text-xs text-slate-500 mt-1 line-clamp-2">{skill.description}</p>
          )}
        </div>
        <div className="shrink-0">
          <LifecycleBadge lifecycle={skill.metadata?.lifecycle || "stable"} />
        </div>
      </div>

      {tags.length > 0 && (
        <div className="flex flex-wrap gap-1">
          {tags.map((tag) => (
            <span
              key={tag}
              className="text-xs px-1.5 py-0.5 rounded-full bg-slate-700/50 text-slate-400"
            >
              {tag}
            </span>
          ))}
        </div>
      )}
      <div className="flex items-center justify-between mt-auto pt-1">
        <div className="flex items-center gap-3 text-xs text-slate-600">
          {skill.metadata?.version && (
            <span className="font-mono">{skill.metadata.version}</span>
          )}
          {skill.metadata?.owner && <span>{skill.metadata.owner}</span>}
        </div>
        <div className="flex items-center gap-1.5">
          <button onClick={onPreview} className="btn-ghost py-1 text-xs" title="Preview metadata and SKILL.md">
            <Eye size={12} />
            Preview
          </button>
          <button
            onClick={onInstall}
            disabled={installing || !canInstall || installed}
            title={installed ? "This skill is already installed" : !canInstall ? "Initialize this project first" : undefined}
            className="btn-primary py-1 text-xs"
          >
            {installing ? <span className="spinner w-3 h-3" /> : <Download size={12} />}
            {installed ? "Installed" : "Install"}
          </button>
        </div>
      </div>
    </div>
  );
}

