import { useEffect, useRef, useState } from "react";
import { Download, X, FileText, Info } from "lucide-react";
import type { RegistrySkill, SkillPreview } from "@/lib/types";
import { previewRegistrySkill } from "@/lib/skell";
import { LifecycleBadge } from "./Badges";
import { MarkdownViewer } from "./MarkdownViewer";

interface SkillPreviewModalProps {
  skill: RegistrySkill;
  installed: boolean;
  canInstall: boolean;
  onClose: () => void;
  onInstall: () => void;
}

export function SkillPreviewModal({
  skill,
  installed,
  canInstall,
  onClose,
  onInstall,
}: SkillPreviewModalProps) {
  const [preview, setPreview] = useState<SkillPreview | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const closeBtnRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);
    previewRegistrySkill(
      skill.registry_alias ?? "",
      skill.registry_url ?? "",
      skill.name
    )
      .then((p) => {
        if (!cancelled) setPreview(p);
      })
      .catch((e) => {
        if (!cancelled) setError(String(e));
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [skill.name, skill.registry_alias, skill.registry_url]);

  useEffect(() => {
    closeBtnRef.current?.focus();
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  const tags = skill.metadata?.tags?.split(",").map((t) => t.trim()).filter(Boolean) ?? [];

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center"
      role="dialog"
      aria-modal="true"
      aria-labelledby="preview-title"
    >
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm cursor-pointer" onClick={onClose} />
      <div className="relative z-10 w-full max-w-3xl mx-4 bg-[#13162a] border border-[#2d3348] rounded-2xl shadow-2xl flex flex-col max-h-[85vh]">
        {/* Header */}
        <div className="flex items-start justify-between gap-3 p-5 border-b border-[#1e2540]">
          <div className="min-w-0">
            <div className="flex items-center gap-2 flex-wrap">
              <h2 id="preview-title" className="text-lg font-bold text-slate-200 truncate">
                {skill.name}
              </h2>
              <LifecycleBadge lifecycle={skill.metadata?.lifecycle || "stable"} />
            </div>
            {skill.description && (
              <p className="text-sm text-slate-400 mt-1">{skill.description}</p>
            )}
          </div>
          <button
            ref={closeBtnRef}
            onClick={onClose}
            className="btn-ghost p-1.5 shrink-0"
            aria-label="Close preview"
          >
            <X size={16} />
          </button>
        </div>

        {/* Scrollable body */}
        <div className="overflow-y-auto flex-1 p-5 space-y-5">
          {/* Metadata grid */}
          <section>
            <h3 className="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-2 flex items-center gap-1.5">
              <Info size={12} />
              Metadata
            </h3>
            <dl className="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-2 text-sm">
              <MetaRow label="Owner" value={skill.metadata?.owner} />
              <MetaRow label="Version" value={skill.metadata?.version} mono />
              <MetaRow label="Scope" value={skill.metadata?.scope} />
              <MetaRow label="License" value={skill.license} />
              <MetaRow label="Registry" value={skill.registry_alias} mono />
              <MetaRow
                label="Source"
                value={skill.registry_url || skill.metadata?.source_repo}
                mono
                breakAll
              />
              {tags.length > 0 && (
                <div className="sm:col-span-2">
                  <dt className="text-slate-600 text-xs mb-1">Tags</dt>
                  <dd className="flex flex-wrap gap-1">
                    {tags.map((tag) => (
                      <span
                        key={tag}
                        className="text-xs px-1.5 py-0.5 rounded-full bg-slate-700/50 text-slate-400"
                      >
                        {tag}
                      </span>
                    ))}
                  </dd>
                </div>
              )}
            </dl>
          </section>

          {/* README */}
          <section>
            <h3 className="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-2 flex items-center gap-1.5">
              <FileText size={12} />
              SKILL.md
            </h3>
            {loading ? (
              <div className="flex justify-center py-8">
                <div className="spinner w-6 h-6" />
              </div>
            ) : error ? (
              <p className="text-sm text-red-400">Failed to load SKILL.md: {error}</p>
            ) : preview?.found && preview.readme_content ? (
              <div className="rounded-xl border border-[#1e2540] bg-[#0e1120] p-4">
                <MarkdownViewer content={preview.readme_content} />
              </div>
            ) : (
              <p className="text-sm text-slate-500">
                SKILL.md not available for preview. The registry may not be cached yet — try
                refreshing the discovery list, or install the skill to inspect its files.
              </p>
            )}
          </section>
        </div>

        {/* Footer */}
        <div className="flex justify-end gap-2 p-4 border-t border-[#1e2540]">
          <button onClick={onClose} className="btn-ghost text-xs">
            Close
          </button>
          <button
            onClick={onInstall}
            disabled={!canInstall || installed}
            title={installed ? "Already installed" : !canInstall ? "Initialize this project first" : undefined}
            className="btn-primary text-xs"
          >
            <Download size={13} />
            {installed ? "Installed" : "Install"}
          </button>
        </div>
      </div>
    </div>
  );
}

function MetaRow({
  label,
  value,
  mono,
  breakAll,
}: {
  label: string;
  value?: string;
  mono?: boolean;
  breakAll?: boolean;
}) {
  if (!value) return null;
  return (
    <div className="flex gap-3">
      <dt className="text-slate-600 text-xs w-20 shrink-0 mt-0.5">{label}</dt>
      <dd
        className={`text-slate-300 text-xs flex-1 ${mono ? "font-mono" : ""} ${
          breakAll ? "break-all" : ""
        }`}
      >
        {value}
      </dd>
    </div>
  );
}
