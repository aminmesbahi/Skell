import { useEffect, useMemo, useState, useCallback, useRef } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { Search, Tag, Filter, Monitor } from "lucide-react";
import { listRegistry, installSkill, listInstalled, listInstalledGlobal, listSupportedTargets, activeRepoTarget, type AgentTarget } from "@/lib/skell";
import { useRepoStore, useUIStore } from "@/store";
import type { RegistrySkill, InstalledSkill, Lifecycle } from "@/lib/types";
import { getProjectDisplayName } from "@/lib/navigation";
import { SkillCard } from "@/components/SkillCard";
import { SkillPreviewModal } from "@/components/SkillPreviewModal";

const LIFECYCLES: Lifecycle[] = ["stable", "experimental", "draft", "deprecated", "archived"];

function indexInstalled(skills: InstalledSkill[]): Record<string, InstalledSkill> {
  const map: Record<string, InstalledSkill> = {};
  for (const s of skills) map[s.name] = s;
  return map;
}

export function Catalog() {
  const location = useLocation();
  const navigate = useNavigate();
  const { selectedRepo, repos, setSelectedRepo } = useRepoStore();
  const { notify } = useUIStore();
  const [queryInput, setQueryInput] = useState("");
  const [query, setQuery] = useState(""); // debounced
  const [lifecycle, setLifecycle] = useState<Lifecycle | "">("");
  const [owner, setOwner] = useState("");
  const [sourceFilter, setSourceFilter] = useState<"all" | "global" | "local">("all");
  const [skills, setSkills] = useState<RegistrySkill[]>([]);
  const [installed, setInstalled] = useState<Record<string, InstalledSkill>>({});
  const [previewTarget, setPreviewTarget] = useState<RegistrySkill | null>(null);
  const [installing, setInstalling] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const installDestination = (location.state as { installDestination?: string } | null)?.installDestination ?? (selectedRepo && selectedRepo !== "global" ? selectedRepo : undefined);
  const [destination, setDestination] = useState(installDestination ?? "");
  const repo = destination && destination !== "global" ? destination : undefined;

  // Agent target selection
  const [availableTargets, setAvailableTargets] = useState<AgentTarget[]>([]);
  const [selectedTarget, setSelectedTarget] = useState("");

  // Load available targets on mount
  useEffect(() => {
    listSupportedTargets().then(setAvailableTargets).catch(() => {});
  }, []);

  // Auto-detect active target when destination changes (only if user hasn't
  // already manually picked a target). Skip the initial mount to avoid a
  // double loadData call that causes a flash.
  const targetManuallySet = useRef(false);
  useEffect(() => {
    if (!repo || targetManuallySet.current) return;
    activeRepoTarget(repo).then((t) => {
      if (t) setSelectedTarget(t);
    }).catch(() => {});
  }, [repo]);

  // Debounce query input
  useEffect(() => {
    const t = setTimeout(() => setQuery(queryInput), 300);
    return () => clearTimeout(t);
  }, [queryInput]);

  useEffect(() => {
    setDestination(installDestination ?? "");
  }, [installDestination]);

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const [registrySkills, installedSkills] = await Promise.all([
        listRegistry().catch(() => [] as RegistrySkill[]),
        (repo ? listInstalled(repo, selectedTarget || undefined) : listInstalledGlobal()).catch(() => [] as InstalledSkill[]),
      ]);
      setSkills(registrySkills);
      setInstalled(indexInstalled(installedSkills));
    } finally {
      setLoading(false);
    }
  }, [repo, selectedTarget]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  const filtered = useMemo(() => {
    const needle = query.toLowerCase();
    const ownerNeedle = owner.toLowerCase();
    return skills.filter((skill) => {
      const matchesQuery = !needle || `${skill.name} ${skill.description ?? ""}`.toLowerCase().includes(needle);
      const matchesLifecycle = !lifecycle || skill.metadata.lifecycle === lifecycle;
      const matchesOwner = !ownerNeedle || (skill.metadata.owner ?? "").toLowerCase().includes(ownerNeedle);
      const matchesSource = sourceFilter === "all" || skill.registry_source === sourceFilter;
      return matchesQuery && matchesLifecycle && matchesOwner && matchesSource;
    });
  }, [query, lifecycle, owner, sourceFilter, skills]);

  async function handleInstall(skill: RegistrySkill) {
    if (!destination) {
      notify({ kind: "info", title: "Select a project first", detail: "Choose a project before installing a catalog skill." });
      return;
    }
    setInstalling(skill.name);
    try {
      const result = await installSkill({
        skillName: skill.name,
        repo: destination,
        registry: skill.registry_alias || undefined,
        registryURL: skill.registry_url || undefined,
        target: selectedTarget || undefined,
      });
      if (result.success) {
        const targetLabel = selectedTarget ? availableTargets.find((t) => t.id === selectedTarget)?.displayName ?? selectedTarget : "";
        notify({ kind: "success", title: `Installed ${skill.name}`, detail: `${getProjectDisplayName(destination)}${targetLabel ? ` · ${targetLabel}` : ""}` });
        const refreshed = destination === "global"
          ? await listInstalledGlobal()
          : await listInstalled(destination, selectedTarget || undefined);
        setInstalled(indexInstalled(refreshed));
      } else {
        const detail = result.stderr
          ? result.stderr.split(/\r?\nUsage:/, 1)[0].trim()
          : "Unable to install skill.";
        notify({ kind: "error", title: "Install failed", detail });
      }
    } catch (error) {
      notify({
        kind: "error",
        title: "Install failed",
        detail: error instanceof Error ? error.message : String(error),
      });
    } finally {
      setInstalling(null);
    }
  }

  return (
    <div className="mx-auto max-w-6xl px-6 py-8 space-y-6">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-slate-200">Catalog</h1>
          <p className="mt-2 text-sm text-slate-400">Browse skills from the registry and install them into the selected project.</p>
        </div>
      </div>

      {/* Destination selector */}
      <div className="card">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div>
            <p className="text-sm font-medium text-slate-200">Installing to</p>
            <p className="mt-1 text-sm text-slate-400">
              {destination
                ? `${getProjectDisplayName(destination)}${selectedTarget ? ` · ${availableTargets.find((t) => t.id === selectedTarget)?.displayName ?? selectedTarget}` : ""}`
                : "No project selected"}
            </p>
          </div>
          <div className="flex items-center gap-2 flex-wrap">
            <label className="flex items-center gap-1.5 input w-auto cursor-pointer">
              <Monitor size={14} className="text-slate-400" />
              <select
                value={selectedTarget}
                onChange={(e) => { setSelectedTarget(e.target.value); targetManuallySet.current = true; }}
                className="bg-transparent outline-none text-sm text-slate-200"
                title="Choose which AI agent to install the skill for"
              >
                <option value="">Auto-detect agent</option>
                {availableTargets.map((t) => (
                  <option key={t.id} value={t.id}>{t.displayName}</option>
                ))}
              </select>
            </label>
            <select
              value={destination}
              onChange={(e) => {
                const next = e.target.value;
                setDestination(next);
                if (next) setSelectedRepo(next);
              }}
              className="input w-auto"
            >
              <option value="">Select a project</option>
              {repos.map((repoPath) => (
                <option key={repoPath} value={repoPath}>{getProjectDisplayName(repoPath)}</option>
              ))}
            </select>
            <button onClick={() => navigate("/projects")} className="btn-ghost text-sm">Change</button>
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="card">
        <div className="grid gap-3 md:grid-cols-[2fr,1fr,1fr,1fr]">
          <label className="input flex items-center gap-2">
            <Search size={16} className="text-slate-400 shrink-0" />
            <input value={queryInput} onChange={(e) => setQueryInput(e.target.value)} placeholder="Search skills" className="w-full bg-transparent outline-none text-slate-200 placeholder-slate-500" />
          </label>
          <label className="input flex items-center gap-2">
            <Filter size={16} className="text-slate-400 shrink-0" />
            <select value={lifecycle} onChange={(e) => setLifecycle(e.target.value as Lifecycle | "")} className="w-full bg-transparent outline-none text-slate-200">
              <option value="">All lifecycles</option>
              {LIFECYCLES.map((lc) => (
                <option key={lc} value={lc}>{lc.charAt(0).toUpperCase() + lc.slice(1)}</option>
              ))}
            </select>
          </label>
          <label className="input flex items-center gap-2">
            <Tag size={16} className="text-slate-400 shrink-0" />
            <input value={owner} onChange={(e) => setOwner(e.target.value)} placeholder="Filter by owner" className="w-full bg-transparent outline-none text-slate-200 placeholder-slate-500" />
          </label>
          <label className="input flex items-center gap-2">
            <Filter size={16} className="text-slate-400 shrink-0" />
            <select value={sourceFilter} onChange={(e) => setSourceFilter(e.target.value as "all" | "global" | "local")} className="w-full bg-transparent outline-none text-slate-200">
              <option value="all">All sources</option>
              <option value="global">Shared</option>
              <option value="local">Project</option>
            </select>
          </label>
        </div>
      </div>

      {/* Skills grid */}
      {loading ? (
        <div className="flex justify-center py-20">
          <div className="spinner w-8 h-8" />
        </div>
      ) : filtered.length === 0 ? (
        <div className="card flex flex-col items-center py-16 text-center">
          <Search size={40} className="text-slate-700 mb-3" />
          <p className="text-slate-500 text-sm">
            {skills.length === 0
              ? "No skills found in the registry. Add a shared source in Settings to populate the catalog."
              : "No skills match the current filters."}
          </p>
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2">
          {filtered.map((skill) => (
            <SkillCard
              key={skill.name}
              skill={skill}
              installing={installing === skill.name}
              installed={Boolean(installed[skill.name])}
              canInstall={Boolean(destination)}
              onInstall={() => void handleInstall(skill)}
              onPreview={() => setPreviewTarget(skill)}
            />
          ))}
        </div>
      )}

      {previewTarget && (
        <SkillPreviewModal
          skill={previewTarget}
          installed={Boolean(installed[previewTarget.name])}
          canInstall={Boolean(destination)}
          onClose={() => setPreviewTarget(null)}
          onInstall={() => {
            const skill = previewTarget;
            setPreviewTarget(null);
            void handleInstall(skill);
          }}
        />
      )}
    </div>
  );
}
