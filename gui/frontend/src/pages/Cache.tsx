import { useEffect, useState } from "react";
import { Database, RefreshCw, Trash2 } from "lucide-react";
import { useRepoStore, useUIStore } from "@/store";
import { cacheStatus, cacheRefresh, cacheClear } from "@/lib/skell";
import { ConfirmDialog } from "@/components/ConfirmDialog";

export function Cache() {
  const { notify } = useUIStore();
  const { selectedRepo } = useRepoStore();
  const [statusText, setStatusText] = useState("");
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [clearing, setClearing] = useState(false);
  const [confirmClear, setConfirmClear] = useState(false);

  async function loadStatus() {
    setLoading(true);
    try {
      const result = await cacheStatus();
      setStatusText(result.success ? result.stdout : result.stderr);
    } finally {
      setLoading(false);
    }
  }

  async function handleRefresh() {
    setRefreshing(true);
    try {
      const repo = selectedRepo !== "global" ? selectedRepo : undefined;
      const result = await cacheRefresh(repo);
      if (result.success) {
        notify({ kind: "success", title: "Cache refreshed", detail: result.stdout.trim() });
        await loadStatus();
      } else {
        notify({ kind: "error", title: "Refresh failed", detail: result.stderr });
      }
    } finally {
      setRefreshing(false);
    }
  }

  async function handleClear() {
    setConfirmClear(false);
    setClearing(true);
    try {
      const result = await cacheClear();
      if (result.success) {
        notify({ kind: "success", title: "Cache cleared" });
        await loadStatus();
      } else {
        notify({ kind: "error", title: "Clear failed", detail: result.stderr });
      }
    } finally {
      setClearing(false);
    }
  }

  useEffect(() => {
    void loadStatus();
  }, []);

  return (
    <div className="p-6 space-y-6 max-w-3xl mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-200 flex items-center gap-2">
            <Database size={22} className="text-brand-400" />
            Cache
          </h1>
          <p className="text-sm text-slate-500 mt-0.5">
            Manage local clones of registry repositories (~/.skell/cache)
          </p>
        </div>
        <button onClick={() => void loadStatus()} className="btn-ghost" disabled={loading}>
          <RefreshCw size={14} className={loading ? "animate-spin" : ""} />
        </button>
      </div>

      {/* Status */}
      <div className="card">
        <h2 className="section-title text-base mb-3">Cache Status</h2>
        {loading ? (
          <div className="flex items-center gap-2 text-sm text-slate-500 py-2">
            <div className="spinner w-4 h-4" />
            Loading...
          </div>
        ) : statusText ? (
          <pre className="text-xs font-mono text-slate-300 whitespace-pre-wrap bg-[#0f1225] rounded-lg p-4 border border-[#1e2540]">
            {statusText}
          </pre>
        ) : (
          <p className="text-sm text-slate-600">No cache data available.</p>
        )}
      </div>

      {/* Actions */}
      <div className="card">
        <h2 className="section-title text-base mb-4">Actions</h2>
        <div className="space-y-3">
          <div className="flex items-center justify-between p-3 rounded-lg bg-white/[0.02] border border-[#1e2540]">
            <div>
              <p className="text-sm font-medium text-slate-300">Refresh Cache</p>
              <p className="text-xs text-slate-600 mt-0.5">
                Re-fetch all configured registries from GitHub
              </p>
            </div>
            <button
              onClick={() => void handleRefresh()}
              disabled={refreshing}
              className="btn-primary py-1.5 text-xs"
            >
              <RefreshCw size={13} className={refreshing ? "animate-spin" : ""} />
              Refresh
            </button>
          </div>

          <div className="flex items-center justify-between p-3 rounded-lg bg-red-500/5 border border-red-500/15">
            <div>
              <p className="text-sm font-medium text-red-300">Clear Cache</p>
              <p className="text-xs text-slate-600 mt-0.5">
                Delete all locally cached registry data. Registries will be re-fetched on next use.
              </p>
            </div>
            <button
              onClick={() => setConfirmClear(true)}
              disabled={clearing}
              className="btn-danger py-1.5 text-xs"
            >
              <Trash2 size={13} />
              Clear
            </button>
          </div>
        </div>
      </div>

      <ConfirmDialog
        open={confirmClear}
        title="Clear cache?"
        description="This deletes all locally cached registry data. Skell will re-fetch registries on next use."
        confirmLabel="Clear Cache"
        danger
        onConfirm={() => void handleClear()}
        onCancel={() => setConfirmClear(false)}
      />
    </div>
  );
}
