import { useEffect, useState, useCallback, useMemo, useRef } from "react";
import { Search, RefreshCw, Download, Filter, Globe, Link } from "lucide-react";
import { useRepoStore, useUIStore } from "@/store";
import { searchSkills, installSkill, getGlobalRootDir } from "@/lib/skell";
import type { RegistrySkill, Lifecycle } from "@/lib/types";
import { LifecycleBadge } from "@/components/Badges";
import { AddFromURLDialog } from "@/components/AddFromURLDialog";

const LIFECYCLES: Lifecycle[] = ["stable", "experimental", "draft", "deprecated", "archived"];

export function Registry() {
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

  // Install dialog state
  const [installTarget, setInstallTarget] = useState<RegistrySkill | null>(null);
  const [installRegistry, setInstallRegistry] = useState("");
  const [installRegistryURL, setInstallRegistryURL] = useState("");

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
      const results = await searchSkills({
        query: query || undefined,
        lifecycle: lifecycle || undefined,
        owner: owner || undefined,
        repo,
      });
      setSkills(results);
    } catch (e) {
      notify({ kind: "error", title: "Search failed", detail: String(e) });
    } finally {
      setLoading(false);
    }
  }, [query, lifecycle, owner, notify, selectedRepo]);

  useEffect(() => {
    void doSearch();
  }, [doSearch]);

  async function handleInstall() {
    if (!installTarget) return;
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
        registry: installRegistry || undefined,
        registryURL: installRegistryURL || undefined,
      });
      if (result.success) {
        notify({ kind: "success", title: `Installed ${installTarget.name}`, detail: result.stdout.trim() });
      } else {
        notify({ kind: "error", title: "Install failed", detail: result.stderr });
      }
    } finally {
      setInstalling(null);
      setInstallRegistry("");
      setInstallRegistryURL("");
    }
  }

  const grouped = useMemo(() => {
    const map = new Map<string, RegistrySkill[]>();
    for (const sk of skills) {
      const owner = sk.metadata?.owner || "Unknown";
      if (!map.has(owner)) map.set(owner, []);
      map.get(owner)!.push(sk);
    }
    return map;
  }, [skills]);

  return (
    <div className="p-6 space-y-5 max-w-5xl mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-200">Registry</h1>
          <p className="text-sm text-slate-500 mt-0.5">Browse and install skills from configured registries</p>
        </div>
        <div className="flex items-center gap-2">
          <button onClick={() => setAddDialogOpen(true)} className="btn-primary">
            <Link size={14} />
            Add from URL
          </button>
          <button onClick={() => void doSearch()} className="btn-ghost" disabled={loading}>
            <RefreshCw size={14} className={loading ? "animate-spin" : ""} />
          </button>
        </div>
      </div>

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
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" onClick={() => setInstallTarget(null)} />
          <div className="relative z-10 w-full max-w-md mx-4 bg-[#13162a] border border-[#2d3348] rounded-2xl p-6 shadow-2xl space-y-4">
            <h3 className="font-semibold text-slate-200 text-base">
              Install "{installTarget.name}"
            </h3>
            <div className="space-y-3">
              <div>
                <label className="text-xs text-slate-500 mb-1 block">Registry alias (optional)</label>
                <input
                  className="input"
                  placeholder="e.g. my-registry"
                  value={installRegistry}
                  onChange={(e) => setInstallRegistry(e.target.value)}
                />
              </div>
              <div>
                <label className="text-xs text-slate-500 mb-1 block">Registry URL (optional, to bootstrap)</label>
                <input
                  className="input"
                  placeholder="https://github.com/owner/repo"
                  value={installRegistryURL}
                  onChange={(e) => setInstallRegistryURL(e.target.value)}
                />
              </div>
              <p className="text-xs text-slate-600">
                Target: <span className="text-slate-400">
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
        onSuccess={() => void doSearch()}
      />
    </div>
  );
}

function SkillCard({
  skill,
  installing,
  onInstall,
}: {
  skill: RegistrySkill;
  installing: boolean;
  onInstall: () => void;
}) {
  const tags = skill.metadata?.tags?.split(",").map((t) => t.trim()).filter(Boolean) ?? [];

  return (
    <div className="card hover:border-[#2d3a5a] transition-colors flex flex-col gap-3">
      <div className="flex items-start justify-between gap-2">
        <div className="flex-1 min-w-0">
          <p className="font-semibold text-slate-200 text-sm">{skill.name}</p>
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
          disabled={installing}
          className="btn-primary py-1 text-xs"
        >
          {installing ? (
            <span className="spinner w-3 h-3" />
          ) : (
            <Download size={12} />
          )}
          Install
        </button>
      </div>
    </div>
  );
}
