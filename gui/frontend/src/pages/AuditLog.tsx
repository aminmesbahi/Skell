import { useEffect, useState } from "react";
import { ScrollText, RefreshCw, Search, Filter } from "lucide-react";
import { readAuditLog } from "@/lib/skell";
import type { AuditEntry } from "@/lib/types";

const ACTION_COLORS: Record<string, string> = {
  install: "text-emerald-400 bg-emerald-500/10 border-emerald-500/20",
  upgrade: "text-blue-400 bg-blue-500/10 border-blue-500/20",
  remove: "text-red-400 bg-red-500/10 border-red-500/20",
  pin: "text-purple-400 bg-purple-500/10 border-purple-500/20",
  unpin: "text-slate-400 bg-slate-500/10 border-slate-500/20",
};

const PAGE_SIZE = 50;

export function AuditLog() {
  const [entries, setEntries] = useState<AuditEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [search, setSearch] = useState("");
  const [actionFilter, setActionFilter] = useState("");
  const [page, setPage] = useState(1);

  async function load() {
    setLoading(true);
    try {
      const data = await readAuditLog();
      setEntries(data);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  const filtered = entries.filter((e) => {
    if (actionFilter && e.action !== actionFilter) return false;
    if (search) {
      const q = search.toLowerCase();
      return (
        e.skill?.toLowerCase().includes(q) ||
        e.repo?.toLowerCase().includes(q) ||
        e.registry?.toLowerCase().includes(q) ||
        e.action?.toLowerCase().includes(q)
      );
    }
    return true;
  });

  const paged = filtered.slice(0, page * PAGE_SIZE);
  const actions = [...new Set(entries.map((e) => e.action).filter(Boolean))];

  return (
    <div className="p-6 space-y-5 max-w-5xl mx-auto">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-200 flex items-center gap-2">
            <ScrollText size={22} className="text-brand-400" />
            Audit Log
          </h1>
          <p className="text-sm text-slate-500 mt-0.5">
            {filtered.length} entries — {entries.length} total
          </p>
        </div>
        <button onClick={() => void load()} className="btn-ghost" disabled={loading}>
          <RefreshCw size={14} className={loading ? "animate-spin" : ""} />
        </button>
      </div>

      {/* Filters */}
      <div className="flex gap-3 flex-wrap">
        <div className="relative flex-1 min-w-48">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
          <input
            className="input pl-8"
            placeholder="Search skill, project, registry..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1); }}
          />
        </div>
        <div className="flex items-center gap-2">
          <Filter size={14} className="text-slate-500 shrink-0" />
          <select
            className="input w-36"
            value={actionFilter}
            onChange={(e) => { setActionFilter(e.target.value); setPage(1); }}
          >
            <option value="">All actions</option>
            {actions.map((a) => (
              <option key={a} value={a}>{a}</option>
            ))}
          </select>
        </div>
      </div>

      {loading ? (
        <div className="flex justify-center py-20">
          <div className="spinner w-8 h-8" />
        </div>
      ) : filtered.length === 0 ? (
        <div className="card flex flex-col items-center py-16 text-center">
          <ScrollText size={40} className="text-slate-700 mb-3" />
          <p className="text-slate-500 text-sm">
            {entries.length === 0
              ? "No audit log entries yet. Actions are logged to ~/.skell/audit.log."
              : "No entries match the filter."}
          </p>
        </div>
      ) : (
        <>
          <div className="card p-0 overflow-hidden">
            <table className="data-table">
              <thead>
                <tr>
                  <th>Time</th>
                  <th>Action</th>
                  <th>Skill</th>
                  <th>Version</th>
                  <th>Registry</th>
                  <th>Project</th>
                  <th>User</th>
                </tr>
              </thead>
              <tbody>
                {paged.map((e, i) => (
                  <tr key={i}>
                    <td className="font-mono text-xs text-slate-500 whitespace-nowrap">
                      {new Date(e.timestamp).toLocaleString()}
                    </td>
                    <td>
                      <span
                        className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium border ${
                          ACTION_COLORS[e.action] ?? "text-slate-400 bg-slate-500/10 border-slate-500/20"
                        }`}
                      >
                        {e.action}
                      </span>
                    </td>
                    <td className="font-medium text-slate-200">{e.skill}</td>
                    <td className="font-mono text-xs text-slate-500">{e.version ?? "—"}</td>
                    <td className="text-xs text-slate-500">{e.registry ?? "—"}</td>
                    <td className="text-xs text-slate-500 max-w-[180px] truncate" title={e.repo}>
                      {e.repo ? e.repo.split(/[/\\]/).at(-1) : "—"}
                    </td>
                    <td className="text-xs text-slate-500">{e.user ?? "—"}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {paged.length < filtered.length && (
            <div className="flex justify-center">
              <button
                onClick={() => setPage((p) => p + 1)}
                className="btn-ghost"
              >
                Load more ({filtered.length - paged.length} remaining)
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
