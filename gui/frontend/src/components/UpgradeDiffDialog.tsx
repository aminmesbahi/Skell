import { useEffect, useState } from "react";
import { ArrowUp, X } from "lucide-react";
import clsx from "clsx";
import { useDialog } from "@/hooks/useDialog";
import { diffSkill } from "@/lib/skell";
import type { SkillDiff } from "@/lib/types";

interface UpgradeDiffDialogProps {
  skillName: string | null;
  repo: string;
  onCancel: () => void;
  /** force is true when the user agreed to discard local edits. */
  onUpgrade: (force: boolean) => void;
}

function lineClass(line: string): string {
  if (line.startsWith("+++") || line.startsWith("---")) return "text-slate-400 font-semibold";
  if (line.startsWith("@@")) return "text-cyan-400";
  if (line.startsWith("diff ") || line.startsWith("index ") || line.startsWith("new file") || line.startsWith("deleted file")) return "text-slate-500";
  if (line.startsWith("+")) return "bg-emerald-500/10 text-emerald-300";
  if (line.startsWith("-")) return "bg-red-500/10 text-red-300";
  return "text-slate-400";
}

/** Shows installed → latest changes for a skill and lets the user upgrade. */
export function UpgradeDiffDialog({ skillName, repo, onCancel, onUpgrade }: UpgradeDiffDialogProps) {
  const dialogRef = useDialog(skillName !== null, onCancel);
  const [diff, setDiff] = useState<SkillDiff | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [discardLocal, setDiscardLocal] = useState(false);

  useEffect(() => {
    if (!skillName) return;
    let cancelled = false;
    setDiff(null); setError(null); setDiscardLocal(false);
    diffSkill({ skillName, repo, refresh: true })
      .then((d) => { if (!cancelled) setDiff(d); })
      .catch((e) => { if (!cancelled) setError(String(e)); });
    return () => { cancelled = true; };
  }, [skillName, repo]);

  if (!skillName) return null;
  const lines = diff?.patch ? diff.patch.replace(/\n$/, "").split("\n") : [];
  const added = lines.filter((l) => l.startsWith("+") && !l.startsWith("+++")).length;
  const removed = lines.filter((l) => l.startsWith("-") && !l.startsWith("---")).length;
  const blocked = Boolean(diff?.locally_modified) && !discardLocal;

  return (
    <div ref={dialogRef} role="dialog" aria-modal="true" aria-labelledby="diff-title" tabIndex={-1} className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm cursor-pointer" onClick={onCancel} />
      <div className="relative z-10 mx-4 flex max-h-[88vh] w-full max-w-4xl flex-col rounded-2xl border border-[var(--palette-2d3348)] bg-[var(--palette-13162a)] shadow-2xl">
        <div className="flex items-start justify-between gap-3 border-b border-[var(--palette-1e2540)] p-5">
          <div>
            <h2 id="diff-title" className="text-lg font-bold text-slate-200">Changes in {skillName}</h2>
            {diff && (
              <p className="mt-1 text-sm text-slate-400">
                <span className="font-mono">{diff.installed_version || diff.installed_commit?.slice(0, 7) || "installed"}</span>
                {" → "}
                <span className="font-mono">{diff.latest_version || diff.latest_commit?.slice(0, 7) || "latest"}</span>
                {lines.length > 0 && <> · <span className="text-emerald-400">+{added}</span> <span className="text-red-400">−{removed}</span></>}
              </p>
            )}
          </div>
          <button onClick={onCancel} className="btn-ghost shrink-0 p-1.5" aria-label="Close changes"><X size={16} /></button>
        </div>

        <div className="flex-1 overflow-auto p-5">
          {!diff && !error && <div className="flex justify-center py-12"><div className="spinner h-6 w-6" /></div>}
          {error && <p className="text-sm text-red-400">Could not load changes: {error}</p>}
          {diff && lines.length === 0 && <p className="text-sm text-emerald-400">No differences — the installed copy matches the source.</p>}
          {lines.length > 0 && (
            <pre className="overflow-x-auto rounded-xl border border-[var(--palette-1e2540)] bg-[var(--palette-0e1120)] p-3 text-xs leading-5" aria-label="Unified diff">
              {lines.map((line, i) => <div key={i} className={clsx("whitespace-pre px-1", lineClass(line))}>{line || " "}</div>)}
            </pre>
          )}
        </div>

        <div className="flex flex-wrap items-center justify-between gap-3 border-t border-[var(--palette-1e2540)] p-4">
          {diff?.locally_modified ? (
            <label className="flex items-center gap-2 text-xs text-amber-300">
              <input type="checkbox" checked={discardLocal} onChange={(e) => setDiscardLocal(e.target.checked)} />
              The installed copy has local edits (shown above as removed lines). Discard them and upgrade.
            </label>
          ) : <span />}
          <div className="flex gap-2">
            <button onClick={onCancel} className="btn-ghost text-xs">Cancel</button>
            <button onClick={() => onUpgrade(discardLocal)} disabled={!diff || blocked} title={blocked ? "Tick the box to discard local edits first" : undefined} className="btn-primary text-xs">
              <ArrowUp size={13} />
              Upgrade
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
