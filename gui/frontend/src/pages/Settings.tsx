import { useEffect, useState } from "react";
import {
  Settings as SettingsIcon,
  RefreshCw,
  Download,
  Info,
} from "lucide-react";
import { useUIStore } from "@/store";
import { selfUpdateCheck, selfUpdate, getSkellVersion } from "@/lib/skell";
import { ConfirmDialog } from "@/components/ConfirmDialog";

export function Settings() {
  const { notify } = useUIStore();
  const [skellVersion, setSkellVersion] = useState("");
  const [checking, setChecking] = useState(false);
  const [updating, setUpdating] = useState(false);
  const [updateInfo, setUpdateInfo] = useState("");
  const [confirmUpdate, setConfirmUpdate] = useState(false);

  useEffect(() => {
    getSkellVersion()
      .then(setSkellVersion)
      .catch(() => setSkellVersion("unknown"));
  }, []);

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
    <div className="p-6 space-y-6 max-w-3xl mx-auto">
      <div>
        <h1 className="text-2xl font-bold text-slate-200 flex items-center gap-2">
          <SettingsIcon size={22} className="text-brand-400" />
          Settings
        </h1>
        <p className="text-sm text-slate-500 mt-0.5">Configure Skell and manage updates</p>
      </div>

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
            <span className="font-mono">0.1.0</span>
          </div>
          <div className="flex gap-6 text-slate-400">
            <span className="text-slate-600 w-24">Repository</span>
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
            ["Global manifest", "~/.skell/.claude/skell.toml"],
            ["Cache root", "~/.skell/cache/"],
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
