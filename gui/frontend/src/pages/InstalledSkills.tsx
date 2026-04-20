import { useEffect, useState, useCallback, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import {
  Package,
  RefreshCw,
  Search,
  ArrowUp,
  Trash2,
  Pin,
  PinOff,
  Info,
  Filter,
  Link,
} from "lucide-react";
import { useRepoStore, useUIStore } from "@/store";
import {
  listInstalled,
  getStatus,
  upgradeSkill,
  removeSkill,
  pinSkill,
  unpinSkill,
  listInstalledGlobal,
} from "@/lib/skell";
import type { InstalledSkill, StatusEntry, SkillStatus } from "@/lib/types";
import { SkillBadge, ScopeBadge } from "@/components/Badges";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import { AddFromURLDialog } from "@/components/AddFromURLDialog";

type Scope = "local" | "global";

interface SkillRow extends InstalledSkill {
  statusEntry?: StatusEntry;
  scope: Scope;
  repoPath: string;
}

const ALL_STATUSES: SkillStatus[] = [
  "up-to-date",
  "outdated",
  "pinned",
  "deprecated",
  "archived",
  "locally-modified",
  "unknown",
  "missing-metadata",
  "unversioned",
];

export function InstalledSkills() {
  const navigate = useNavigate();
  const { selectedRepo, repos } = useRepoStore();
  const { notify } = useUIStore();

  const [skills, setSkills] = useState<SkillRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<SkillStatus | "">("");
  const [removing, setRemoving] = useState<SkillRow | null>(null);
  const [acting, setActing] = useState<string | null>(null);
  const [addDialogOpen, setAddDialogOpen] = useState(false);

  const loadSkills = useCallback(async () => {
    setLoading(true);
    try {
      let rows: SkillRow[] = [];

      const buildRows = async (
        repoPath: string,
        scope: Scope,
        fetchFn: () => Promise<InstalledSkill[]>
      ) => {
        const [installed, statuses] = await Promise.all([
          fetchFn().catch(() => [] as InstalledSkill[]),
          getStatus(repoPath).catch(() => [] as StatusEntry[]),
        ]);
        return installed.map((sk) => ({
          ...sk,
          statusEntry: statuses.find((s) => s.name === sk.name),
          scope,
          repoPath,
        }));
      };

      if (selectedRepo === "global") {
        rows = await buildRows("global", "global", listInstalledGlobal);
      } else if (selectedRepo) {
        rows = await buildRows(selectedRepo, "local", () =>
          listInstalled(selectedRepo)
        );
      } else {
        // Show all repos combined
        const allRows = await Promise.all(
          repos.map((repo) =>
            buildRows(repo, "local", () => listInstalled(repo))
          )
        );
        rows = allRows.flat();
      }

      setSkills(rows);
    } finally {
      setLoading(false);
    }
  }, [selectedRepo, repos]);

  useEffect(() => {
    void loadSkills();
  }, [loadSkills]);

  const filtered = useMemo(() => {
    let list = skills;
    if (search) {
      const q = search.toLowerCase();
      list = list.filter(
        (s) =>
          s.name.toLowerCase().includes(q) ||
          s.registry.toLowerCase().includes(q)
      );
    }
    if (statusFilter) {
      list = list.filter((s) => {
        const status = s.statusEntry?.status;
        if (statusFilter === "pinned") return s.pinned;
        return status === statusFilter;
      });
    }
    return list;
  }, [skills, search, statusFilter]);

  async function handleUpgrade(sk: SkillRow) {
    setActing(sk.name);
    try {
      const result = await upgradeSkill({
        skillName: sk.name,
        repo: sk.repoPath,
      });
      if (result.success) {
        notify({ kind: "success", title: `Upgraded ${sk.name}` });
        void loadSkills();
      } else {
        notify({ kind: "error", title: "Upgrade failed", detail: result.stderr });
      }
    } finally {
      setActing(null);
    }
  }

  async function handlePin(sk: SkillRow) {
    setActing(sk.name);
    try {
      const result = await (sk.pinned
        ? unpinSkill({ skillName: sk.name, repo: sk.repoPath })
        : pinSkill({ skillName: sk.name, repo: sk.repoPath }));
      if (result.success) {
        notify({
          kind: "success",
          title: sk.pinned ? `Unpinned ${sk.name}` : `Pinned ${sk.name}`,
        });
        void loadSkills();
      } else {
        notify({ kind: "error", title: "Pin operation failed", detail: result.stderr });
      }
    } finally {
      setActing(null);
    }
  }

  async function doRemove(sk: SkillRow) {
    setRemoving(null);
    setActing(sk.name);
    try {
      const result = await removeSkill({ skillName: sk.name, repo: sk.repoPath });
      if (result.success) {
        notify({ kind: "success", title: `Removed ${sk.name}` });
        void loadSkills();
      } else {
        notify({ kind: "error", title: "Remove failed", detail: result.stderr });
      }
    } finally {
      setActing(null);
    }
  }

  const repoLabel =
    selectedRepo === "global"
      ? "Global"
      : selectedRepo
      ? selectedRepo.split(/[/\\]/).at(-1)
      : "All Repos";

  return (
    <div className="p-6 space-y-5 max-w-6xl mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-200">Installed Skills</h1>
          <p className="text-sm text-slate-500 mt-0.5">
            {repoLabel} · {filtered.length} skill{filtered.length !== 1 ? "s" : ""}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button onClick={() => setAddDialogOpen(true)} className="btn-primary">
            <Link size={14} />
            Add from URL
          </button>
          <button onClick={() => void loadSkills()} className="btn-ghost" disabled={loading}>
            <RefreshCw size={14} className={loading ? "animate-spin" : ""} />
            Refresh
          </button>
        </div>
      </div>

      {/* Filters */}
      <div className="flex items-center gap-3 flex-wrap">
        <div className="relative flex-1 min-w-48">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
          <input
            className="input pl-8"
            placeholder="Search skills..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <div className="flex items-center gap-2">
          <Filter size={14} className="text-slate-500" />
          <select
            className="input w-44"
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value as SkillStatus | "")}
          >
            <option value="">All statuses</option>
            {ALL_STATUSES.map((s) => (
              <option key={s} value={s}>
                {s}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Table */}
      {loading ? (
        <div className="flex justify-center py-20">
          <div className="spinner w-8 h-8" />
        </div>
      ) : filtered.length === 0 ? (
        <div className="card flex flex-col items-center py-16 text-center">
          <Package size={40} className="text-slate-700 mb-3" />
          <p className="text-slate-500 text-sm">
            {skills.length === 0
              ? "No skills installed in this repo."
              : "No skills match the current filter."}
          </p>
        </div>
      ) : (
        <div className="card p-0 overflow-hidden">
          <table className="data-table">
            <thead>
              <tr>
                <th>Skill</th>
                <th>Version</th>
                <th>Status</th>
                <th>Registry</th>
                <th>Scope</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((sk) => {
                const status =
                  sk.pinned ? "pinned" : sk.statusEntry?.status ?? "unknown";
                const isOutdated =
                  !sk.pinned && sk.statusEntry?.status === "outdated";
                const isBusy = acting === sk.name;

                return (
                  <tr key={`${sk.repoPath}-${sk.name}`}>
                    <td>
                      <button
                        onClick={() =>
                          navigate(`/skills/${encodeURIComponent(sk.name)}`, {
                            state: { repo: sk.repoPath },
                          })
                        }
                        className="font-medium text-brand-400 hover:text-brand-300 transition-colors"
                      >
                        {sk.name}
                      </button>
                    </td>
                    <td className="font-mono text-xs">
                      {sk.version || "—"}
                      {isOutdated && sk.statusEntry?.latest && (
                        <span className="ml-2 text-amber-400">
                          → {sk.statusEntry.latest}
                        </span>
                      )}
                    </td>
                    <td>
                      <SkillBadge status={status as typeof status} size="sm" />
                    </td>
                    <td className="text-slate-400 text-xs">{sk.registry || "—"}</td>
                    <td>
                      <ScopeBadge scope={sk.scope} />
                    </td>
                    <td>
                      <div className="flex items-center gap-1">
                        <button
                          onClick={() =>
                            navigate(`/skills/${encodeURIComponent(sk.name)}`, {
                              state: { repo: sk.repoPath },
                            })
                          }
                          className="p-1.5 rounded-lg text-slate-500 hover:text-blue-400 hover:bg-blue-500/10 transition-colors"
                          title="Info"
                          disabled={isBusy}
                        >
                          <Info size={13} />
                        </button>
                        {isOutdated && (
                          <button
                            onClick={() => void handleUpgrade(sk)}
                            className="p-1.5 rounded-lg text-slate-500 hover:text-amber-400 hover:bg-amber-500/10 transition-colors"
                            title="Upgrade"
                            disabled={isBusy}
                          >
                            <ArrowUp size={13} />
                          </button>
                        )}
                        <button
                          onClick={() => void handlePin(sk)}
                          className="p-1.5 rounded-lg text-slate-500 hover:text-blue-400 hover:bg-blue-500/10 transition-colors"
                          title={sk.pinned ? "Unpin" : "Pin"}
                          disabled={isBusy}
                        >
                          {sk.pinned ? <PinOff size={13} /> : <Pin size={13} />}
                        </button>
                        <button
                          onClick={() => setRemoving(sk)}
                          className="p-1.5 rounded-lg text-slate-500 hover:text-red-400 hover:bg-red-500/10 transition-colors"
                          title="Remove"
                          disabled={isBusy}
                        >
                          <Trash2 size={13} />
                        </button>
                      </div>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {/* Remove confirm */}
      <ConfirmDialog
        open={removing !== null}
        title={`Remove "${removing?.name}"?`}
        description="This will delete the skill files and remove it from skell.toml and skell.lock."
        confirmLabel="Remove"
        danger
        onConfirm={() => removing && void doRemove(removing)}
        onCancel={() => setRemoving(null)}
      />

      {/* Add from URL */}
      <AddFromURLDialog
        open={addDialogOpen}
        onClose={() => setAddDialogOpen(false)}
        onSuccess={() => void loadSkills()}
      />
    </div>
  );
}
