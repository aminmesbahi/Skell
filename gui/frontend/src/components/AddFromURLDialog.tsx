import { useState } from "react";
import { X, Link, Loader2 } from "lucide-react";
import { useRepoStore, useUIStore } from "@/store";
import { addSkillFromURL } from "@/lib/skell";

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

  const defaultRepo = initialRepo ?? (selectedRepo !== "global" ? selectedRepo : (repos[0] ?? ""));
  const [url, setUrl] = useState("");
  const [repo, setRepo] = useState(defaultRepo);
  const [dryRun, setDryRun] = useState(false);
  const [loading, setLoading] = useState(false);

  if (!open) return null;

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const trimmedURL = url.trim();
    if (!trimmedURL) return;

    setLoading(true);
    try {
      const results = await addSkillFromURL({
        url: trimmedURL,
        repo: repo || undefined,
        dryRun,
      });

      const result = results[0];
      if (result) {
        if (result.installed) {
          notify({
            kind: "success",
            title: `Skill "${result.skill_name}" installed`,
            detail: `Registry "${result.alias}" · ${repo}`,
          });
        } else if (result.registered) {
          notify({
            kind: "success",
            title: `Registry "${result.alias}" registered`,
            detail: repo,
          });
        } else if (result.skill_name) {
          notify({
            kind: "info",
            title: `[dry-run] Would install "${result.skill_name}"`,
            detail: `from registry "${result.alias}"`,
          });
        } else {
          notify({
            kind: "info",
            title: `[dry-run] Would register registry "${result.alias}"`,
          });
        }
      }

      setUrl("");
      onClose();
      if (!dryRun) onSuccess?.();
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
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/60 backdrop-blur-sm"
        onClick={onClose}
      />

      {/* Dialog */}
      <div className="relative z-10 w-full max-w-lg mx-4 bg-[#13162a] border border-[#2d3348] rounded-2xl shadow-2xl">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-[#2d3348]">
          <div className="flex items-center gap-2 text-slate-200 font-semibold">
            <Link size={18} className="text-indigo-400" />
            Add from URL
          </div>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-slate-500 hover:text-slate-300 hover:bg-white/5 transition-colors"
          >
            <X size={16} />
          </button>
        </div>

        {/* Body */}
        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          {/* URL input */}
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-slate-400 uppercase tracking-wider">
              GitHub URL
            </label>
            <input
              type="url"
              required
              autoFocus
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="https://github.com/owner/repo/tree/main/skills/my-skill"
              className="w-full px-3 py-2 bg-[#0e1120] border border-[#2d3348] rounded-lg text-sm text-slate-200 placeholder-slate-600 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500/40 transition-colors"
            />
            <p className="text-xs text-slate-500">
              Paste a skill directory URL to install it, or a skills root URL to register the registry.
            </p>
          </div>

          {/* Repo selector */}
          {repos.length > 0 && (
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-slate-400 uppercase tracking-wider">
                Target repository
              </label>
              <select
                value={repo}
                onChange={(e) => setRepo(e.target.value)}
                className="w-full px-3 py-2 bg-[#0e1120] border border-[#2d3348] rounded-lg text-sm text-slate-200 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500/40 transition-colors"
              >
                {repos.map((r) => (
                  <option key={r} value={r}>
                    {r}
                  </option>
                ))}
              </select>
            </div>
          )}

          {/* Dry-run toggle */}
          <label className="flex items-center gap-3 cursor-pointer select-none">
            <input
              type="checkbox"
              checked={dryRun}
              onChange={(e) => setDryRun(e.target.checked)}
              className="w-4 h-4 rounded border-[#2d3348] bg-[#0e1120] accent-indigo-500"
            />
            <span className="text-sm text-slate-300">Dry-run (preview without writing)</span>
          </label>

          {/* Actions */}
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
