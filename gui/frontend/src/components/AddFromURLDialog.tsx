import { useState } from "react";
import { X, Link, Loader2, CheckCircle2 } from "lucide-react";
import { useRepoStore, useUIStore } from "@/store";
import { addSkillFromURL } from "@/lib/skell";
import type { AddResult } from "@/lib/types";

interface AddFromURLDialogProps {
  open: boolean;
  onClose: () => void;
  /** Called after a successful add so callers can refresh their data. */
  onSuccess?: () => void;
  /** Pre-select a specific repo path; defaults to the globally selected repo. */
  initialRepo?: string;
}

export function AddFromURLDialog({
  open,
  onClose,
  onSuccess,
  initialRepo,
}: AddFromURLDialogProps) {
  const { repos, selectedRepo } = useRepoStore();
  const { notify } = useUIStore();

  const defaultRepo = initialRepo ?? (selectedRepo === "global" ? "global" : (selectedRepo || repos[0] || ""));
  const [url, setUrl] = useState("");
  const [repo, setRepo] = useState(defaultRepo);
  const [dryRun, setDryRun] = useState(false);
  const [loading, setLoading] = useState(false);
  const [dryRunResult, setDryRunResult] = useState<AddResult | null>(null);

  if (!open) return null;

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const trimmedURL = url.trim();
    if (!trimmedURL) return;

    setLoading(true);
    setDryRunResult(null);
    try {
      const results = await addSkillFromURL({
        url: trimmedURL,
        repo: repo || undefined,
        dryRun,
      });

      const result = results[0];
      if (!result) return;

      const repoLabel = repo === "global" ? "Shared Library (~/.skell)" : repo;

      if (dryRun) {
        // Keep the dialog open and show the result inline — the auto-dismiss
        // toast disappears too fast for the user to read.
        setDryRunResult(result);
      } else if (result.installed) {
        notify({
          kind: "success",
          title: `Skill "${result.skill_name}" installed`,
          detail: `Registry "${result.alias}" · ${repoLabel}`,
        });
        setUrl("");
        onClose();
        onSuccess?.();
      } else if (result.registered) {
        notify({
          kind: "success",
          title: `Registry "${result.alias}" registered`,
          detail: repoLabel,
        });
        setUrl("");
        onClose();
        onSuccess?.();
      }
    } catch (err) {
      notify({
        kind: "error",
        title: "Add failed",
        detail: err instanceof Error ? err.message : String(err),
      });
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div
        className="absolute inset-0 bg-black/60 backdrop-blur-sm"
        onClick={onClose}
      />

      <div className="relative z-10 w-full max-w-lg mx-4 bg-[#13162a] border border-[#2d3348] rounded-2xl shadow-2xl">
        <div className="flex items-center justify-between px-6 py-4 border-b border-[#2d3348]">
          <div className="flex items-center gap-2 text-slate-200 font-semibold">
            <Link size={18} className="text-indigo-400" />
            Add Skill or Source
          </div>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-slate-500 hover:text-slate-300 hover:bg-white/5 transition-colors"
          >
            <X size={16} />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-slate-400 uppercase tracking-wider">
              GIT URL OR LOCAL PATH
            </label>
            <input
              type="text"
              required
              autoFocus
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="https://github.com/owner/repo/tree/main/skills  or  /Users/you/my-skills-folder"
              className="w-full px-3 py-2 bg-[#0e1120] border border-[#2d3348] rounded-lg text-sm text-slate-200 placeholder-slate-600 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500/40 transition-colors"
            />
            <p className="text-xs text-slate-500">
              Paste a GitHub tree URL, or a local folder path containing <code>SKILL.md</code> files. Local folders are supported and always fresh.
            </p>
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-medium text-slate-400 uppercase tracking-wider">
              Target project
            </label>
            <select
              value={repo}
              onChange={(e) => setRepo(e.target.value)}
              className="w-full px-3 py-2 bg-[#0e1120] border border-[#2d3348] rounded-lg text-sm text-slate-200 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500/40 transition-colors"
            >
              <option value="global">Shared Library (~/.skell)</option>
              {repos.map((r) => (
                <option key={r} value={r}>
                  {r}
                </option>
              ))}
            </select>
          </div>

          <label className="flex items-center gap-3 cursor-pointer select-none">
            <input
              type="checkbox"
              checked={dryRun}
              onChange={(e) => { setDryRun(e.target.checked); setDryRunResult(null); }}
              className="w-4 h-4 rounded border-[#2d3348] bg-[#0e1120] accent-indigo-500"
            />
            <span className="text-sm text-slate-300">Dry-run (preview without writing)</span>
          </label>

          {dryRunResult && (
            <div className="rounded-lg border border-indigo-500/30 bg-indigo-500/10 px-4 py-3 space-y-1">
              <div className="flex items-center gap-2 text-xs font-semibold text-indigo-300 uppercase tracking-wider">
                <CheckCircle2 size={13} />
                Preview result
              </div>
              {dryRunResult.skill_name ? (
                <p className="text-sm text-slate-200">
                  Would install skill <span className="font-mono text-indigo-300">"{dryRunResult.skill_name}"</span>{" "}
                  from registry <span className="font-mono text-indigo-300">"{dryRunResult.alias}"</span>
                </p>
              ) : (
                <p className="text-sm text-slate-200">
                  Would register registry <span className="font-mono text-indigo-300">"{dryRunResult.alias}"</span>
                </p>
              )}
              <p className="text-xs text-slate-500">No files were written. Uncheck dry-run and click Add to apply.</p>
            </div>
          )}

          <div className="flex justify-end gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="btn-ghost"
              disabled={loading}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="btn-primary flex items-center gap-2"
              disabled={loading || !url.trim()}
            >
              {loading && <Loader2 size={14} className="animate-spin" />}
              {dryRun ? "Preview" : "Add"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
