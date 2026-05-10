import { useEffect, useState, useCallback, useMemo, useRef } from "react";
import { useNavigate } from "react-router-dom";
import { Search, RefreshCw, Download, Filter, Globe, Link, AlertTriangle, FilePlus, Star } from "lucide-react";
import { useRepoStore, useUIStore } from "@/store";
import { searchSkills, installSkill, getGlobalRootDir, isRepoInitialized, initRepo, listInstalled, listInstalledGlobal } from "@/lib/skell";
import type { RegistrySkill, Lifecycle, InstalledSkill } from "@/lib/types";
import { LifecycleBadge } from "@/components/Badges";
import { AddFromURLDialog } from "@/components/AddFromURLDialog";

const LIFECYCLES: Lifecycle[] = ["stable", "experimental", "draft", "deprecated", "archived"];

export function Registry() {
  const navigate = useNavigate();
  const { selectedRepo } = useRepoStore();
  const { notify } = useUIStore();

  const [skills, setSkills] = useState<RegistrySkill[]>([]);
  const [loading, setLoading] = useState(false);
  const [query, setQuery] = useState("");
  const [lifecycle, setLifecycle] = useState<Lifecycle | "">("");
  const [owner, setOwner] = useState("");
  const [installing, setInstalling] = useState<string | null>(null);
  const globalRootRef = useRef<string | null>(null);
  const [addDialogOpen, setAddDialogOpen] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);
  const [repoInited, setRepoInited] = useState<boolean | null>(null);
  const [initRunning, setInitRunning] = useState(false);
  const [sourceFilter, setSourceFilter] = useState<"all" | "global" | "local">("all");
  const [installedSkills, setInstalledSkills] = useState<Record<string, InstalledSkill>>({});

  // Install dialog state — no longer asks for alias/URL (taken from the skill)
  const [installTarget, setInstallTarget] = useState<RegistrySkill | null>(null);

  const doSearch = useCallback(async () => {
    setLoading(true);
    try {
      let repo: string | undefined;
      if (selectedRepo && selectedRepo !== "global") {
        repo = selectedRepo;
      } else {
        // For global mode, resolve the global root dir (~/.skell) once and reuse it.
        if (!globalRootRef.current) {
          globalRootRef.current = await getGlobalRootDir();
        }
        repo = globalRootRef.current || undefined;
      }
      const [results, installed] = await Promise.all([
        searchSkills({
          query: query || undefined,
          lifecycle: lifecycle || undefined,
          owner: owner || undefined,
          repo,
        }),
        loadInstalledSkills(selectedRepo),
      ]);
      setSkills(results);
      setInstalledSkills(indexInstalledSkills(installed));
    } catch (e) {
      notify({ kind: "error", title: "Search failed", detail: String(e) });
    } finally {
      setLoading(false);
    }
  }, [query, lifecycle, owner, notify, selectedRepo, refreshKey]);

  useEffect(() => {
    void doSearch();
  }, [doSearch]);

  useEffect(() => {
    setRepoInited(null);
    isRepoInitialized(selectedRepo)
      .then(setRepoInited)
      .catch(() => setRepoInited(false));
  }, [selectedRepo]);

  async function handleInitHere() {
    if (!selectedRepo || selectedRepo === "global") return;
    setInitRunning(true);
    try {
      const result = await initRepo(selectedRepo);
      if (result.success) {
        notify({ kind: "success", title: "Repository initialized", detail: result.stdout.trim() });
        setRepoInited(true);
        setRefreshKey((k) => k + 1);
      } else {
        notify({ kind: "error", title: "Init failed", detail: result.stderr });
      }
    } finally {
      setInitRunning(false);
    }
  }

  async function handleInstall() {
    if (!installTarget) return;
    if (installedSkills[installTarget.name]) {
      notify({ kind: "info", title: `${installTarget.name} is already installed` });
      setInstallTarget(null);
      return;
    }
    const repo = selectedRepo === "global" ? undefined : selectedRepo;
    if (!repo && selectedRepo !== "global") {
      notify({ kind: "error", title: "Select a repository first" });
      return;
    }
    setInstalling(installTarget.name);
    setInstallTarget(null);
    try {
      const result = await installSkill({
        skillName: installTarget.name,
        repo: repo ?? "global",
        registry: installTarget.registry_alias || undefined,
        registryURL: installTarget.registry_url || undefined,
      });
      if (result.success) {
        notify({ kind: "success", title: `Installed ${installTarget.name}`, detail: result.stdout.trim() });
        setInstalledSkills((current) => ({
          ...current,
          [installTarget.name]: {
            name: installTarget.name,
            version: "",
            registry: installTarget.registry_alias ?? "",
            source_repo: installTarget.metadata?.source_repo ?? installTarget.registry_url ?? "",
            installed_path: "",
            installed_at: "",
            pinned: false,
            content_hash: "",
          },
        }));
      } else {
        notify({ kind: "error", title: "Install failed", detail: result.stderr });
      }
    } finally {
      setInstalling(null);
    }
  }

  const grouped = useMemo(() => {
    const map = new Map<string, RegistrySkill[]>();
    const filtered = sourceFilter === "all"
      ? skills
      : skills.filter((sk) => sk.registry_source === sourceFilter);
    for (const sk of filtered) {
      const key = sk.metadata?.owner || sk.registry_alias || "Unknown";
      if (!map.has(key)) map.set(key, []);
      map.get(key)!.push(sk);
    }
    return map;
  }, [skills, sourceFilter]);

  return (
    <div className="p-6 space-y-5 max-w-5xl mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-200">Discover Skills</h1>
          <p className="text-sm text-slate-500 mt-0.5">Browse and install skills from configured registries</p>
        </div>
        <div className="flex items-center gap-2">
          {/* Quick shortcut to manage favorite/global sources (git + local folders) */}
          <button
            onClick={() => navigate("/settings")}
            className="btn-ghost flex items-center gap-1.5"
            title="Manage your favorite global sources (git URLs and local skill folders)"
          >
            <Star size={14} />
            Sources
          </button>

          <button
            onClick={() => setAddDialogOpen(true)}
            className="btn-primary"
            disabled={repoInited === false}
            title={repoInited === false ? "Initialize this repository first" : undefined}
          >
            <Link size={14} />
            Add from URL or Path
          </button>
          <button onClick={() => void doSearch()} className="btn-ghost" disabled={loading}>
            <RefreshCw size={14} className={loading ? "animate-spin" : ""} />
          </button>
        </div>
      </div>

      {/* Not-initialized banner */}
      {repoInited === false && selectedRepo !== "global" && (
        <div className="flex items-center gap-3 rounded-xl border border-amber-500/30 bg-amber-500/10 px-4 py-3 text-sm">
          <AlertTriangle size={16} className="text-amber-400 shrink-0" />
          <p className="flex-1 text-amber-300">
            This repository hasn't been initialized yet — run{" "}
            <code className="font-mono text-amber-200 bg-amber-500/20 px-1 rounded">skell init</code>{" "}
            before adding registries or installing skills.
          </p>
          <button
            onClick={() => void handleInitHere()}
            disabled={initRunning}
            className="shrink-0 flex items-center gap-1.5 text-xs font-medium px-3 py-1.5 rounded-lg bg-amber-500/20 hover:bg-amber-500/30 text-amber-200 border border-amber-500/30 transition-colors disabled:opacity-50"
          >
            <FilePlus size={13} />
            {initRunning ? "Initializing…" : "Initialize now"}
          </button>
        </div>
      )}

      {/* Filters */}
      <div className="flex gap-3 flex-wrap">
        <div className="relative flex-1 min-w-52">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
          <input
            className="input pl-8"
            placeholder="Search by name, description, tags..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
        </div>
        <div className="flex items-center gap-2">
          <Filter size={14} className="text-slate-500 shrink-0" />
          <select
            className="input w-36"
            value={lifecycle}
            onChange={(e) => setLifecycle(e.target.value as Lifecycle | "")}
          >
            <option value="">All lifecycles</option>
            {LIFECYCLES.map((l) => (
              <option key={l} value={l}>{l}</option>
            ))}
          </select>
          <input
            className="input w-36"
            placeholder="Owner..."
            value={owner}
            onChange={(e) => setOwner(e.target.value)}
          />
        </div>
        {/* Source filter — only useful when local repo is selected (results include both sources) */}
        {selectedRepo !== "global" && (
          <div className="flex items-center gap-1 rounded-lg border border-[#2d3348] bg-[#0e1120] p-1">
            {(["all", "local", "global"] as const).map((s) => (
              <button
                key={s}
                onClick={() => setSourceFilter(s)}
                className={`px-3 py-1 rounded text-xs font-medium transition-colors ${
                  sourceFilter === s
                    ? "bg-indigo-600 text-white"
                    : "text-slate-400 hover:text-slate-200"
                }`}
              >
                {s === "all" ? "All" : s === "local" ? "Local" : "Global"}
              </button>
            ))}
          </div>
        )}
      </div>

      {/* Results */}
      {loading ? (
        <div className="flex justify-center py-20">
          <div className="spinner w-8 h-8" />
        </div>
      ) : skills.length === 0 ? (
        <div className="card flex flex-col items-center py-16 text-center">
          <Globe size={40} className="text-slate-700 mb-3" />
          <p className="text-slate-500 text-sm">No skills found. Try adjusting your search.</p>
        </div>
      ) : (
        <div className="space-y-6">
          {Array.from(grouped.entries()).map(([owner, ownerSkills]) => (
            <div key={owner}>
              <h3 className="text-xs font-semibold text-slate-600 uppercase tracking-wider mb-3">
                {owner}
              </h3>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                {ownerSkills.map((sk) => (
                  <SkillCard
                    key={sk.name}
                    skill={sk}
                    installing={installing === sk.name}
                    installed={Boolean(installedSkills[sk.name])}
                    canInstall={repoInited !== false}
                    onInstall={() => setInstallTarget(sk)}
                  />
                ))}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Install dialog */}
      {installTarget && (
        <div className="fixed inset-0 z-50 flex items-center justify-center" role="dialog" aria-modal="true" aria-labelledby="install-dialog-title">
          <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" onClick={() => setInstallTarget(null)} />
          <div className="relative z-10 w-full max-w-md mx-4 bg-[#13162a] border border-[#2d3348] rounded-2xl p-6 shadow-2xl space-y-4">
            <h3 id="install-dialog-title" className="font-semibold text-slate-200 text-base">
              Install "{installTarget.name}"
            </h3>
            <div className="space-y-2 text-sm text-slate-400">
              {installTarget.registry_alias && (
                <p>Registry: <span className="text-slate-300 font-mono">{installTarget.registry_alias}</span></p>
              )}
              <p>
                Target: <span className="text-slate-300">
                  {selectedRepo === "global" ? "Global" : selectedRepo?.split(/[/\\]/).at(-1) ?? "—"}
                </span>
              </p>
            </div>
            <div className="flex justify-end gap-3 pt-2">
              <button onClick={() => setInstallTarget(null)} className="btn-ghost">Cancel</button>
              <button onClick={() => void handleInstall()} className="btn-primary">
                <Download size={14} />
                Install
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Add from URL */}
      <AddFromURLDialog
        open={addDialogOpen}
        onClose={() => setAddDialogOpen(false)}
        onSuccess={() => setRefreshKey((k) => k + 1)}
      />
    </div>
  );
}

function SkillCard({
  skill,
  installing,
  installed,
  canInstall,
  onInstall,
}: {
  skill: RegistrySkill;
  installing: boolean;
  installed: boolean;
  canInstall: boolean;
  onInstall: () => void;
}) {
  const tags = skill.metadata?.tags?.split(",").map((t) => t.trim()).filter(Boolean) ?? [];

  return (
    <div className="card hover:border-[#2d3a5a] transition-colors flex flex-col gap-3">
      <div className="flex items-start justify-between gap-2">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <p className="font-semibold text-slate-200 text-sm">{skill.name}</p>
            {skill.registry_source === "global" && (
              <span className="text-[10px] font-medium px-1.5 py-0.5 rounded-full bg-violet-500/20 text-violet-300 border border-violet-500/30">
                global
              </span>
            )}
            {skill.registry_source === "local" && (
              <span className="text-[10px] font-medium px-1.5 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 border border-indigo-500/30">
                local
              </span>
            )}
          </div>
          {skill.description && (
            <p className="text-xs text-slate-500 mt-1 line-clamp-2">{skill.description}</p>
          )}
        </div>
        <div className="shrink-0">
          <LifecycleBadge lifecycle={skill.metadata?.lifecycle || "stable"} />
        </div>
      </div>
      {tags.length > 0 && (
        <div className="flex flex-wrap gap-1">
          {tags.map((tag) => (
            <span
              key={tag}
              className="text-xs px-1.5 py-0.5 rounded-full bg-slate-700/50 text-slate-400"
            >
              {tag}
            </span>
          ))}
        </div>
      )}
      <div className="flex items-center justify-between mt-auto pt-1">
        <div className="flex items-center gap-3 text-xs text-slate-600">
          {skill.metadata?.version && (
            <span className="font-mono">{skill.metadata.version}</span>
          )}
          {skill.metadata?.owner && <span>{skill.metadata.owner}</span>}
        </div>
        <button
          onClick={onInstall}
          disabled={installing || !canInstall || installed}
          title={installed ? "This skill is already installed" : !canInstall ? "Initialize this repository first" : undefined}
          className="btn-primary py-1 text-xs"
        >
          {installing ? (
            <span className="spinner w-3 h-3" />
          ) : installed ? (
            <Download size={12} />
          ) : (
            <Download size={12} />
          )}
          {installed ? "Installed" : "Install"}
        </button>
      </div>
    </div>
  );
}

async function loadInstalledSkills(selectedRepo: string): Promise<InstalledSkill[]> {
  try {
    if (!selectedRepo || selectedRepo === "global") {
      return await listInstalledGlobal();
    }
    return await listInstalled(selectedRepo);
  } catch {
    return [];
  }
}

function indexInstalledSkills(skills: InstalledSkill[]): Record<string, InstalledSkill> {
  return skills.reduce<Record<string, InstalledSkill>>((acc, skill) => {
    acc[skill.name] = skill;
    return acc;
  }, {});
}
