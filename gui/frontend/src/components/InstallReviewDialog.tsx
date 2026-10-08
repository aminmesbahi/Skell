import { useDialog } from "@/hooks/useDialog";
import { AlertTriangle, Download, FileCode2, Globe, ShieldCheck, Wrench, X } from "lucide-react";
import type { SkillReview } from "@/lib/types";

interface InstallReviewDialogProps {
  review: SkillReview | null;
  onConfirm: () => void;
  onCancel: () => void;
}

function formatBytes(n: number): string {
  if (n >= 1 << 20) return `${(n / (1 << 20)).toFixed(1)} MB`;
  if (n >= 1 << 10) return `${(n / (1 << 10)).toFixed(1)} KB`;
  return `${n} B`;
}

const MAX_LISTED = 12;

/**
 * Shows what a skill would bring into the project before it is installed —
 * scripts, pre-approved tools, external links — so the user can make an
 * informed decision. Mirrors the CLI's install review.
 */
export function InstallReviewDialog({ review, onConfirm, onCancel }: InstallReviewDialogProps) {
  const dialogRef = useDialog(review !== null, onCancel);
  if (!review) return null;
  const scripts = review.scripts ?? [];
  const tools = review.allowed_tools ?? [];
  const links = review.links ?? [];
  const warnings = review.warnings ?? [];

  return (
    <div ref={dialogRef} role="dialog" aria-modal="true" aria-labelledby="review-title" tabIndex={-1} className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm cursor-pointer" onClick={onCancel} />
      <div className="relative z-10 mx-4 flex max-h-[85vh] w-full max-w-2xl flex-col rounded-2xl border border-[var(--palette-2d3348)] bg-[var(--palette-13162a)] shadow-2xl">
        <div className="flex items-start justify-between gap-3 border-b border-[var(--palette-1e2540)] p-5">
          <div className="min-w-0">
            <h2 id="review-title" className="text-lg font-bold text-slate-200">Review before installing {review.skill.name}</h2>
            <p className="mt-1 text-sm text-slate-400">
              From <span className="font-mono">{review.registry}</span>
              {review.commit && <> at <span className="font-mono">{review.commit.slice(0, 7)}</span></>}
              {" · "}{review.files.length} files · {formatBytes(review.total_bytes)}
            </p>
          </div>
          <button onClick={onCancel} className="btn-ghost shrink-0 p-1.5" aria-label="Close review">
            <X size={16} />
          </button>
        </div>

        <div className="flex-1 space-y-4 overflow-y-auto p-5 text-sm">
          {warnings.length > 0 ? (
            <div className="rounded-xl border border-amber-500/30 bg-amber-500/10 p-3" role="alert">
              <p className="mb-1 flex items-center gap-2 font-medium text-amber-300"><AlertTriangle size={14} /> Worth a look</p>
              <ul className="list-disc space-y-0.5 pl-5 text-amber-200/90">
                {warnings.map((w) => <li key={w}>{w}</li>)}
              </ul>
              <p className="mt-2 text-xs text-amber-200/70">Skills are instructions your agent follows. Only install skills from sources you trust.</p>
            </div>
          ) : (
            <p className="flex items-center gap-2 text-emerald-400"><ShieldCheck size={14} /> No scripts or broad tool permissions.</p>
          )}

          {tools.length > 0 && (
            <section>
              <h3 className="mb-1 flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wider text-slate-500"><Wrench size={12} /> Pre-approved tools</h3>
              <div className="flex flex-wrap gap-1">
                {tools.map((t) => <span key={t} className="rounded-md bg-slate-700/50 px-1.5 py-0.5 font-mono text-xs text-slate-300">{t}</span>)}
              </div>
            </section>
          )}

          {scripts.length > 0 && (
            <section>
              <h3 className="mb-1 flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wider text-slate-500"><FileCode2 size={12} /> Scripts the agent may run ({scripts.length})</h3>
              <ul className="space-y-0.5 font-mono text-xs text-slate-300">
                {scripts.slice(0, MAX_LISTED).map((s) => <li key={s}>{s}</li>)}
                {scripts.length > MAX_LISTED && <li className="text-slate-500">… and {scripts.length - MAX_LISTED} more</li>}
              </ul>
            </section>
          )}

          {links.length > 0 && (
            <section>
              <h3 className="mb-1 flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wider text-slate-500"><Globe size={12} /> External links ({links.length})</h3>
              <ul className="space-y-0.5 break-all font-mono text-xs text-slate-400">
                {links.slice(0, MAX_LISTED).map((l) => <li key={l}>{l}</li>)}
                {links.length > MAX_LISTED && <li className="text-slate-500">… and {links.length - MAX_LISTED} more</li>}
              </ul>
            </section>
          )}
        </div>

        <div className="flex justify-end gap-2 border-t border-[var(--palette-1e2540)] p-4">
          <button onClick={onCancel} className="btn-ghost text-xs">Cancel</button>
          <button onClick={onConfirm} className="btn-primary text-xs">
            <Download size={13} />
            Install anyway
          </button>
        </div>
      </div>
    </div>
  );
}
