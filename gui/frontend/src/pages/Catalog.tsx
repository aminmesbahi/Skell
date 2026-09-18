import { useEffect, useMemo, useState, useCallback, useRef } from "react";
import { useLocation } from "react-router";
import { Search, Filter, Monitor, GitBranchPlus } from "lucide-react";
import { listRegistry, installSkill, listInstalled, listInstalledGlobal, listSupportedTargets, activeRepoTarget, type AgentTarget } from "@/lib/skell";
import { useRepoStore, useUIStore } from "@/store";
import type { RegistrySkill, InstalledSkill } from "@/lib/types";
import { getProjectDisplayName } from "@/lib/navigation";
import { SkillCard } from "@/components/SkillCard";
import { SkillPreviewModal } from "@/components/SkillPreviewModal";
import { inferRegistrySource, matchesRegistrySource, type RegistrySourceFilter, type NormalizedRegistrySource } from "@/lib/registry";
import { AddSkillSourceDialog } from "@/components/AddSkillSourceDialog";

const SOURCE_LABELS: Record<NormalizedRegistrySource, string> = {
  global: "Shared",
  local: "Project",
  unknown: "Other",
};

function indexInstalled(skills: InstalledSkill[]): Record<string, InstalledSkill> {
  const map: Record<string, InstalledSkill> = {};
  for (const s of skills) map[s.name] = s;
  return map;
}

export function Catalog() {
  const location = useLocation();
  const { selectedRepo, repos, setSelectedRepo } = useRepoStore();
  const { notify } = useUIStore();
  const [queryInput, setQueryInput] = useState("");
  const [query, setQuery] = useState(""); // debounced
  const [sourceFilter, setSourceFilter] = useState<RegistrySourceFilter>("all");
  const [skills, setSkills] = useState<RegistrySkill[]>([]);
  const [installed, setInstalled] = useState<Record<string, InstalledSkill>>({});
  const [previewTarget, setPreviewTarget] = useState<RegistrySkill | null>(null);
  const [installing, setInstalling] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [sourceDialogOpen, setSourceDialogOpen] = useState(false);

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
    return skills.filter((skill) => {
      const matchesQuery = !needle || `${skill.name} ${skill.description ?? ""}`.toLowerCase().includes(needle);
      const matchesSource = matchesRegistrySource(skill, sourceFilter);
      return matchesQuery && matchesSource;
    });
  }, [query, sourceFilter, skills]);

  const grouped = useMemo(() => {
    const buckets: Record<NormalizedRegistrySource, RegistrySkill[]> = { global: [], local: [], unknown: [] };
    for (const skill of filtered) {
      buckets[inferRegistrySource(skill)].push(skill);
    }
    return buckets;
  }, [filtered]);

  const sourceCounts = useMemo(() => {
    const counts: Record<NormalizedRegistrySource, number> = { global: 0, local: 0, unknown: 0 };
    for (const skill of skills) {
      counts[inferRegistrySource(skill)] += 1;
    }
    return counts;
  }, [skills]);

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
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div className="min-w-0">
          <h1 className="text-2xl font-bold text-slate-200">Catalog</h1>
            <p className="mt-2 text-sm text-slate-400">Search skills and add a source when needed.</p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <button onClick={() => setSourceDialogOpen(true)} className="btn-primary text-sm">
            <GitBranchPlus size={14} />
            Add source
          </button>
        </div>
      </div>

      {/* Destination selector */}
      <div className="card">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <p className="text-sm font-medium text-slate-200">Destination</p>
            <p className="mt-1 text-sm text-slate-400">
              {destination
                ? `${getProjectDisplayName(destination)}${selectedTarget ? ` · ${availableTargets.find((t) => t.id === selectedTarget)?.displayName ?? selectedTarget}` : ""}`
                : "Choose a project"}
            </p>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <label className="flex items-center gap-1.5 input w-auto cursor-pointer">
              <Monitor size={14} className="text-slate-400" />
              <select
                value={selectedTarget}
                onChange={(e) => { setSelectedTarget(e.target.value); targetManuallySet.current = true; }}
                className="bg-transparent outline-none text-sm text-slate-200"
                title="Choose which AI agent to install the skill for"
                aria-label="Choose agent platform"
              >
                <option value="">Auto-detect</option>
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
              aria-label="Select destination project"
            >
              <option value="">Select project</option>
              {repos.map((repoPath) => (
                <option key={repoPath} value={repoPath}>{getProjectDisplayName(repoPath)}</option>
              ))}
            </select>
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="card">
        <div className="flex flex-wrap items-center gap-3">
          <label className="input flex min-w-0 flex-1 items-center gap-2">
            <Search size={16} className="text-slate-400 shrink-0" />
            <input value={queryInput} onChange={(e) => setQueryInput(e.target.value)} placeholder="Search" className="w-full bg-transparent outline-none text-slate-200 placeholder-slate-500" aria-label="Search skills" />
          </label>
          <div className="flex flex-wrap items-center gap-2">
            {([
              { value: "all", label: "All", count: filtered.length },
              { value: "global", label: "Shared", count: sourceCounts.global },
              { value: "local", label: "Project", count: sourceCounts.local },
            ] as const).map((item) => (
              <button
                key={item.value}
                type="button"
                onClick={() => setSourceFilter(item.value)}
                className={item.value === sourceFilter ? "btn-primary text-sm" : "btn-ghost text-sm"}
                aria-pressed={item.value === sourceFilter}
              >
                <Filter size={14} />
                {item.label}
                <span className="text-xs opacity-70">{item.count}</span>
              </button>
            ))}
          </div>
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
          <p className="text-slate-500 text-sm max-w-xl leading-6">
            {skills.length === 0
              ? "No skills found. Add a source and refresh."
              : "No skills match. Try Search or a source filter."}
          </p>
        </div>
      ) : (
        sourceFilter === "all" ? (
          <div className="space-y-6">
            {(Object.entries(grouped) as Array<[NormalizedRegistrySource, RegistrySkill[]]>).map(([source, groupSkills]) => (
              groupSkills.length > 0 ? (
                <section key={source} className="space-y-3">
                  <div className="flex items-center justify-between gap-3">
                    <h2 className="text-sm font-semibold text-slate-300">{SOURCE_LABELS[source]}</h2>
                    <span className="text-xs text-slate-500">{groupSkills.length}</span>
                  </div>
                  <div className="grid gap-4 md:grid-cols-2">
                    {groupSkills.map((skill) => (
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
                </section>
              ) : null
            ))}
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
        )
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

      <AddSkillSourceDialog
        open={sourceDialogOpen}
        onClose={() => setSourceDialogOpen(false)}
        onSuccess={() => void loadData()}
      />
    </div>
  );
}
