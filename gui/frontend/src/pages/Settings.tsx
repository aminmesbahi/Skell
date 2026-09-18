import { useEffect, useState } from "react";
import {
  Settings as SettingsIcon,
  RefreshCw,
  Download,
  Info,
  Trash2,
  FolderOpen,
  Globe,
  GitBranchPlus,
} from "lucide-react";
import { useUIStore } from "@/store";
import {
  selfUpdateCheck,
  selfUpdate,
  getSkellVersion,
  listSkillSources,
  removeSkillSource,
} from "@/lib/skell";
import type { SkillSource } from "@/lib/types";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import { AddSkillSourceDialog } from "@/components/AddSkillSourceDialog";

export function Settings() {
  const { notify } = useUIStore();
  const [skellVersion, setSkellVersion] = useState("");
  const [checking, setChecking] = useState(false);
  const [updating, setUpdating] = useState(false);
  const [updateInfo, setUpdateInfo] = useState("");
  const [confirmUpdate, setConfirmUpdate] = useState(false);

  // Skill Sources state
  const [sources, setSources] = useState<SkillSource[]>([]);
  const [loadingSources, setLoadingSources] = useState(true);
  const [confirmRemove, setConfirmRemove] = useState<string | null>(null);
  const [addSourceOpen, setAddSourceOpen] = useState(false);

  useEffect(() => {
    getSkellVersion()
      .then(setSkellVersion)
      .catch(() => setSkellVersion("unknown"));

    void loadSources();
  }, []);

  async function loadSources() {
    setLoadingSources(true);
    try {
      const list = await listSkillSources();
      setSources(Array.isArray(list) ? list : []);
    } catch (e: unknown) {
      notify({ kind: "error", title: "Failed to load sources", detail: e instanceof Error ? e.message : String(e) });
    } finally {
      setLoadingSources(false);
    }
  }

  async function handleRemoveSource(alias: string) {
    setConfirmRemove(null);
    try {
      await removeSkillSource(alias);
      notify({ kind: "success", title: `Removed source "${alias}"` });
      await loadSources();
    } catch (e: unknown) {
      notify({ kind: "error", title: "Failed to remove source", detail: e instanceof Error ? e.message : String(e) });
    }
  }

  async function handleCheck() {
    setChecking(true);
    setUpdateInfo("");
    try {
      const result = await selfUpdateCheck();
      setUpdateInfo(result.success ? result.stdout.trim() : result.stderr.trim());
    } finally {
      setChecking(false);
    }
  }

  async function handleUpdate() {
    setConfirmUpdate(false);
    setUpdating(true);
    try {
      const result = await selfUpdate();
      if (result.success) {
        notify({ kind: "success", title: "Skell updated!", detail: result.stdout.trim() });
        setSkellVersion(await getSkellVersion().catch(() => "unknown"));
      } else {
        notify({ kind: "error", title: "Update failed", detail: result.stderr });
      }
    } finally {
      setUpdating(false);
    }
  }

  return (
    <div className="p-6 space-y-6 max-w-4xl mx-auto">
      <div>
        <h1 className="text-2xl font-bold text-slate-200 flex items-center gap-2">
          <SettingsIcon size={22} className="text-brand-400" />
          Settings
        </h1>
        <p className="text-sm text-slate-500 mt-0.5">Configure Skell, manage shared sources, and update the CLI</p>
      </div>

      {/* === SHARED SKILL SOURCES === */}
      <div className="card">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="section-title text-base flex items-center gap-2">
              <Globe size={16} className="text-brand-400" />
              Shared Skill Sources
            </h2>
            <p className="text-sm text-slate-500 mt-0.5">
              Add one shared source from GitHub or a local folder, then browse those skills from Catalog.
            </p>
          </div>
          <button onClick={() => void loadSources()} className="btn-ghost" disabled={loadingSources}>
            <RefreshCw size={14} className={loadingSources ? "animate-spin" : ""} /> Refresh
          </button>
        </div>

        {/* Current sources list */}
        <div className="mb-6">
          {loadingSources ? (
            <div className="text-sm text-slate-500 py-4">Loading sources...</div>
          ) : sources.length === 0 ? (
            <div className="text-sm text-slate-500 py-4 border border-dashed border-[#1e2640] rounded-lg p-4 text-center">
              No shared skill sources configured yet. Add your first one below.
            </div>
          ) : (
            <div className="space-y-2">
              {sources.map((src) => (
                <div key={src.alias} className="flex items-center justify-between rounded-lg border border-[#1e2640] bg-[#0f1225] px-4 py-3">
                  <div className="flex items-center gap-3 min-w-0">
                    {src.is_local ? (
                      <FolderOpen size={16} className="text-emerald-400 shrink-0" />
                    ) : (
                      <Globe size={16} className="text-blue-400 shrink-0" />
                    )}
                    <div className="min-w-0">
                      <div className="font-mono text-sm text-slate-200 truncate">{src.alias}</div>
                      <div className="font-mono text-xs text-slate-500 truncate" title={src.url}>{src.url}</div>
                    </div>
                    <span className={`ml-2 text-[10px] px-1.5 py-0.5 rounded ${src.is_local ? "bg-emerald-500/15 text-emerald-400" : "bg-blue-500/15 text-blue-400"}`}>
                      {src.is_local ? "LOCAL" : "GIT"}
                    </span>
                  </div>
                  <button
                    onClick={() => setConfirmRemove(src.alias)}
                    className="p-1.5 text-slate-500 hover:text-red-400 hover:bg-red-500/10 rounded transition"
                    title="Remove source"
                  >
                    <Trash2 size={15} />
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Add new source form */}
        <div className="border-t border-[#1e2640] pt-4">
          <button onClick={() => setAddSourceOpen(true)} className="flex w-full items-center gap-3 rounded-xl border border-[#1e2640] bg-[#0f1225] px-4 py-4 text-left hover:border-[#334268] hover:bg-[#151b32] transition-colors">
            <div className="rounded-lg bg-blue-500/15 p-2 text-blue-400">
              <GitBranchPlus size={18} />
            </div>
            <div>
              <div className="font-medium text-slate-200">Add source</div>
              <div className="text-sm text-slate-500">Paste a repo URL, folder path, or file URI.</div>
            </div>
          </button>
        </div>
      </div>

      <AddSkillSourceDialog
        open={addSourceOpen}
        onClose={() => setAddSourceOpen(false)}
        onSuccess={() => void loadSources()}
      />

      {/* Confirm remove dialog */}
      <ConfirmDialog
        open={!!confirmRemove}
        title="Remove Skill Source?"
        description={`Remove "${confirmRemove}"? This only affects the shared source list — project skell.toml files are untouched.`}
        confirmLabel="Remove Source"
        onConfirm={() => confirmRemove && void handleRemoveSource(confirmRemove)}
        onCancel={() => setConfirmRemove(null)}
      />

      {/* Version info */}
      <div className="card">
        <h2 className="section-title text-base flex items-center gap-2">
          <Info size={16} className="text-brand-400" />
          About Skell
        </h2>
        <div className="space-y-2 text-sm">
          <div className="flex gap-6 text-slate-400">
            <span className="text-slate-600 w-24">CLI Version</span>
            <span className="font-mono">{skellVersion || "—"}</span>
          </div>
          <div className="flex gap-6 text-slate-400">
            <span className="text-slate-600 w-24">GUI Version</span>
            <span className="font-mono">0.2.0</span>
          </div>
          <div className="flex gap-6 text-slate-400">
            <span className="text-slate-600 w-24">GitHub Repo</span>
            <span className="font-mono text-brand-400">aminmesbahi/Skell</span>
          </div>
        </div>
      </div>

      {/* Self-update */}
      <div className="card">
        <h2 className="section-title text-base flex items-center gap-2">
          <Download size={16} className="text-brand-400" />
          CLI Self-Update
        </h2>
        <p className="text-sm text-slate-500 mb-4">
          Update the <code className="text-brand-400">skell</code> CLI binary to the latest GitHub release.
        </p>

        {updateInfo && (
          <pre className="text-xs font-mono text-slate-300 whitespace-pre-wrap bg-[#0f1225] rounded-lg p-3 border border-[#1e2540] mb-4">
            {updateInfo}
          </pre>
        )}

        <div className="flex items-center gap-3">
          <button
            onClick={() => void handleCheck()}
            disabled={checking || updating}
            className="btn-ghost"
          >
            <RefreshCw size={14} className={checking ? "animate-spin" : ""} />
            Check for Updates
          </button>
          <button
            onClick={() => setConfirmUpdate(true)}
            disabled={checking || updating}
            className="btn-primary"
          >
            <Download size={14} />
            {updating ? "Updating..." : "Update Now"}
          </button>
        </div>
      </div>

      {/* Storage paths */}
      <div className="card">
        <h2 className="section-title text-base mb-3">Storage Paths</h2>
        <div className="space-y-2 text-sm font-mono text-slate-500">
          {[
            ["Shared sources", "~/.skell/config.toml [sources]"],
            ["Shared manifest", "~/.skell/.claude/skell.toml (legacy global location)"],
            ["Cache root (remote only)", "~/.skell/cache/"],
            ["Audit log", "~/.skell/audit.log"],
          ].map(([label, path]) => (
            <div key={label} className="flex gap-4">
              <span className="text-slate-600 w-36 shrink-0 font-sans">{label}</span>
              <span className="text-slate-400">{path}</span>
            </div>
          ))}
        </div>
      </div>

      <ConfirmDialog
        open={confirmUpdate}
        title="Update Skell CLI?"
        description="This will download the latest skell binary and replace the current one in-place."
        confirmLabel="Update"
        onConfirm={() => void handleUpdate()}
        onCancel={() => setConfirmUpdate(false)}
      />
    </div>
  );
}
