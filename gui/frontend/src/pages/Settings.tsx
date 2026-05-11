import { useEffect, useState } from "react";
import {
  Settings as SettingsIcon,
  RefreshCw,
  Download,
  Info,
  Plus,
  Trash2,
  FolderOpen,
  Globe,
} from "lucide-react";
import { useUIStore } from "@/store";
import {
  selfUpdateCheck,
  selfUpdate,
  getSkellVersion,
  listSkillSources,
  addSkillSource,
  removeSkillSource,
} from "@/lib/skell";
import type { SkillSource } from "@/lib/types";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import { SelectDirectory } from "../../wailsjs/go/main/App";

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
  const [newAlias, setNewAlias] = useState("");
  const [newURL, setNewURL] = useState("");
  const [adding, setAdding] = useState(false);
  const [confirmRemove, setConfirmRemove] = useState<string | null>(null);

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
    } catch (e: any) {
      notify({ kind: "error", title: "Failed to load sources", detail: e?.message || String(e) });
    } finally {
      setLoadingSources(false);
    }
  }

  async function handleAddSource(isLocal: boolean) {
    if (!newAlias.trim() || (!newURL.trim() && !isLocal)) {
      notify({ kind: "error", title: "Alias and URL/path are required" });
      return;
    }
    setAdding(true);
    try {
      let url = newURL.trim();
      if (isLocal && !url) {
        const selected = await SelectDirectory();
        if (!selected) {
          setAdding(false);
          return;
        }
        url = selected;
      }
      await addSkillSource(newAlias.trim(), url);
      notify({ kind: "success", title: `Added source "${newAlias}"` });
      setNewAlias("");
      setNewURL("");
      await loadSources();
    } catch (e: any) {
      notify({ kind: "error", title: "Failed to add source", detail: e?.message || String(e) });
    } finally {
      setAdding(false);
    }
  }

  async function handleRemoveSource(alias: string) {
    setConfirmRemove(null);
    try {
      await removeSkillSource(alias);
      notify({ kind: "success", title: `Removed source "${alias}"` });
      await loadSources();
    } catch (e: any) {
      notify({ kind: "error", title: "Failed to remove source", detail: e?.message || String(e) });
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
              Git repositories and local folders available across projects. Shared sources live in <code>~/.skell/config.toml</code>; each project still installs skills into its own target folder such as <code>.claude</code>, <code>.github</code>, or <code>.cursor</code>.
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
          <div className="mb-4">
            <div className="text-base font-semibold text-slate-200">Add a Source</div>
            <p className="text-sm text-slate-500 mt-1">
              Register a shared git source or point Skell at a local skills folder.
            </p>
          </div>

          <div className="grid grid-cols-1 xl:grid-cols-12 gap-4 items-end">
            <div className="xl:col-span-3 space-y-2">
              <label className="text-xs font-semibold text-slate-400 uppercase tracking-[0.16em]">
                Source Alias
              </label>
              <input
                type="text"
                placeholder="company-design"
                className="input h-12 border-[#1e2640] font-mono text-[15px] placeholder:text-slate-600"
                value={newAlias}
                onChange={(e) => setNewAlias(e.target.value)}
              />
            </div>

            <div className="xl:col-span-6 space-y-2">
              <label className="text-xs font-semibold text-slate-400 uppercase tracking-[0.16em]">
                Git Source URL
              </label>
              <input
                type="text"
                placeholder="https://github.com/org/skills or git@github.com:org/skills.git"
                className="input h-12 border-[#1e2640] font-mono text-[15px] placeholder:text-slate-600"
                value={newURL}
                onChange={(e) => setNewURL(e.target.value)}
              />
            </div>

            <div className="xl:col-span-3 space-y-2">
              <label className="text-xs font-semibold text-slate-400 uppercase tracking-[0.16em]">
                Actions
              </label>
              <div className="flex flex-col sm:flex-row xl:flex-col gap-2">
                <button
                  onClick={() => void handleAddSource(false)}
                  disabled={adding || !newAlias.trim() || !newURL.trim()}
                  className="btn-primary h-12 justify-center px-5 whitespace-nowrap disabled:opacity-50"
                >
                  <Plus size={15} /> Add Git Source
                </button>
                <button
                  onClick={() => void handleAddSource(true)}
                  disabled={adding || !newAlias.trim()}
                  className="btn-ghost h-12 justify-center px-5 border border-[#2d3348] whitespace-nowrap disabled:opacity-50"
                  title="Pick a local folder containing SKILL.md files"
                >
                  <FolderOpen size={15} /> Choose Folder
                </button>
              </div>
            </div>
          </div>
          <p className="text-xs text-slate-500 mt-3 leading-relaxed">
            Local folders stay live with no cache. They work well for shared skill collections or for skills you are still editing before publishing them to a git source.
          </p>
        </div>
      </div>

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
