import { useAsync } from "@/hooks/useAsync";
import { LoadState } from "@/components/LoadState";
import { useEffect, useMemo, useState } from "react";
import { useLocation } from "react-router";
import { Search, Filter, GitBranchPlus } from "lucide-react";
import { listRegistry, installSkill, listInstalled, listSupportedTargets, activeRepoTarget, isRepoInitialized } from "@/lib/skell";
import { useRepoStore, useUIStore } from "@/store";
import type { RegistrySkill, InstalledSkill } from "@/lib/types";
import { getProjectDisplayName } from "@/lib/navigation";
import { SkillCard } from "@/components/SkillCard";
import { SkillPreviewModal } from "@/components/SkillPreviewModal";
import { inferRegistrySource, matchesRegistrySource, type RegistrySourceFilter, type NormalizedRegistrySource } from "@/lib/registry";
import { AddSkillSourceDialog } from "@/components/AddSkillSourceDialog";
import { CollapsibleSection } from "@/components/CollapsibleSection";

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

function sourceGroupLabel(skill: RegistrySkill): string {
  if (skill.registry_alias?.trim()) return skill.registry_alias.trim();
  if (skill.registry_url?.trim()) return skill.registry_url.trim();
  return "Unknown source";
}

function sourceGroupSubLabel(skill: RegistrySkill): string {
  if (skill.registry_url?.trim()) return skill.registry_url.trim();
  if (skill.registry_alias?.trim()) return "Alias only";
  return "No source URL available";
}

interface CatalogSourceGroup {
  key: string;
  sourceType: NormalizedRegistrySource;
  label: string;
  subLabel: string;
  skills: RegistrySkill[];
}

export function Catalog() {
  const location = useLocation();
  const { selectedRepo, repos, setSelectedRepo } = useRepoStore();
  const { notify } = useUIStore();
  const [queryInput, setQueryInput] = useState("");
  const [query, setQuery] = useState(""); // debounced
  const [sourceFilter, setSourceFilter] = useState<RegistrySourceFilter>("all");
  const [previewTarget, setPreviewTarget] = useState<RegistrySkill | null>(null);
  const [installing, setInstalling] = useState<string | null>(null);
  const [sourceDialogOpen, setSourceDialogOpen] = useState(false);
  const [installResults, setInstallResults] = useState<string[]>([]);
  const installDestination = (location.state as { installDestination?: string } | null)?.installDestination ?? (selectedRepo && selectedRepo !== "global" ? selectedRepo : "");
  const [destination, setDestination] = useState(installDestination);
  const [selectedTargets, setSelectedTargets] = useState<string[] | null>(null);
  const repo = destination && destination !== "global" ? destination : undefined;
  const agents = useAsync(destination, async () => {
    const [targets, active] = await Promise.all([listSupportedTargets(), repo ? activeRepoTarget(repo) : Promise.resolve("")]);
    return { targets, active };
  });
  const availableTargets = agents.data?.targets ?? [];
  const targets = selectedTargets ?? (agents.data?.active ? [agents.data.active] : []);
  useEffect(() => { setDestination(installDestination); setSelectedTargets(null); }, [installDestination]);
  useEffect(() => { const timer = setTimeout(() => setQuery(queryInput), 300); return () => clearTimeout(timer); }, [queryInput]);
  const data = useAsync(JSON.stringify([destination, targets]), async () => {
    const [skills, initialized, installations] = await Promise.all([
      listRegistry(), destination ? isRepoInitialized(destination) : Promise.resolve(false),
      Promise.all(targets.map(async (target) => ({ target, skills: destination ? await listInstalled(destination, target || undefined) : [] }))),
    ]);
    return { skills, initialized, installed: Object.fromEntries(installations.map((i) => [i.target, indexInstalled(i.skills)])) };
  });
  const skills = data.data?.skills ?? [];
  const loading = data.loading;
  const loadData = data.refresh;
  const installedFor = (skill: RegistrySkill, target: string) => Boolean(data.data?.installed[target]?.[skill.name]);
  const allInstalled = (skill: RegistrySkill) => targets.length > 0 && targets.every((target) => installedFor(skill, target));
  const disabledReason = !destination ? "Choose an installation destination" : data.error || agents.error ? "Resolve the loading error before installing" : loading || agents.loading ? "Checking installation destination" : !data.data?.initialized ? "Initialize this project first" : !targets.length ? "Choose at least one agent" : installing ? "Wait for the current installation to finish" : undefined;
  const queryMatches = (skill: RegistrySkill) => !query.trim() || [skill.name, skill.description, skill.metadata?.tags, skill.metadata?.owner, skill.compatibility, skill.registry_alias, skill.registry_url].filter(Boolean).join(" ").toLowerCase().includes(query.trim().toLowerCase());
  const filtered = useMemo(() => {
    return skills.filter((skill) => {
      const matchesQuery = queryMatches(skill);
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

  const groupedBySource = useMemo(() => {
    const byKey = new Map<string, CatalogSourceGroup>();
    for (const skill of filtered) {
      const sourceType = inferRegistrySource(skill);
      const label = sourceGroupLabel(skill);
      const subLabel = sourceGroupSubLabel(skill);
      const key = `${sourceType}::${skill.registry_alias ?? ""}::${skill.registry_url ?? ""}`;

      const existing = byKey.get(key);
      if (existing) {
        existing.skills.push(skill);
      } else {
        byKey.set(key, { key, sourceType, label, subLabel, skills: [skill] });
      }
    }

    const groups = Array.from(byKey.values());
    groups.sort((a, b) => {
      if (a.sourceType !== b.sourceType) return a.sourceType.localeCompare(b.sourceType);
      return a.label.localeCompare(b.label);
    });
    return groups;
  }, [filtered]);

  const sourceCounts = useMemo(() => {
    const counts: Record<NormalizedRegistrySource, number> = { global: 0, local: 0, unknown: 0 };
    for (const skill of skills.filter(queryMatches)) {
      counts[inferRegistrySource(skill)] += 1;
    }
    return counts;
  }, [skills, query]);

  async function handleInstall(skill: RegistrySkill) {
    if (disabledReason) return;
    setInstalling(skill.name); setInstallResults([]);
    const report: string[] = [];
    for (const target of targets) {
      const label = availableTargets.find((t) => t.id === target)?.displayName ?? (target || "Default agent");
      if (installedFor(skill, target)) { report.push(label + ": Already installed"); continue; }
      try {
        const result = await installSkill({ skillName: skill.name, repo: destination, registry: skill.registry_alias || undefined, registryURL: skill.registry_url || undefined, target: target || undefined });
        report.push(label + (result.success ? ": Installed" : ": Failed — " + result.stderr.split(/\r?\nUsage:/, 1)[0].trim()));
      } catch (error) { report.push(label + ": Failed — " + String(error)); }
      setInstallResults([...report]);
    }
    setInstallResults(report);
    notify({ kind: report.some((r) => r.includes(": Failed")) ? "error" : "success", title: "Installation results: " + skill.name, detail: report.join("\n") });
    setInstalling(null); await loadData();
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

      <LoadState loading={loading || agents.loading} error={data.error || agents.error} retry={() => { void loadData(); void agents.refresh(); }} />
      {installResults.length > 0 && <div className="card" aria-live="polite"><h2>Installation results</h2>{installResults.map((result, i) => <p key={i}>{result}</p>)}</div>}
      {/* Destination selector */}
      <div className="card">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <p className="text-sm font-medium text-slate-200">Destination</p>
            <p className="mt-1 text-sm text-slate-400">
              {destination
                ? `${getProjectDisplayName(destination)}`
                : "Choose a project"}
            </p>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <fieldset disabled={!!installing} className="flex flex-wrap gap-3"><legend className="text-sm mb-2">Install for these agents</legend>{availableTargets.map((t) => <label key={t.id} className="flex items-center gap-1 text-sm"><input type="checkbox" checked={targets.includes(t.id)} onChange={(e) => setSelectedTargets(e.target.checked ? [...new Set([...targets.filter(Boolean), t.id])] : targets.filter((id) => id !== t.id))} />{t.displayName}</label>)}</fieldset>
            <select
              value={destination}
              disabled={!!installing}
              onChange={(e) => {
                const next = e.target.value;
                setDestination(next); setSelectedTargets(null);
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
            <input data-search autoFocus={Boolean((location.state as { focusSearch?: boolean } | null)?.focusSearch)} value={queryInput} onChange={(e) => setQueryInput(e.target.value)} placeholder="Search" className="w-full bg-transparent outline-none text-slate-200 placeholder-slate-500" aria-label="Search skills" />
          </label>
          <div className="flex flex-wrap items-center gap-2">
            {([
              { value: "all", label: "All", count: sourceCounts.global + sourceCounts.local + sourceCounts.unknown },
              { value: "global", label: "Shared", count: sourceCounts.global },
              { value: "local", label: "Project", count: sourceCounts.local },
              { value: "unknown", label: "Other", count: sourceCounts.unknown },
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
      {loading && !data.data ? (
        <div className="flex justify-center py-20">
          <div className="spinner w-8 h-8" />
        </div>
      ) : data.error && !data.data ? null : filtered.length === 0 ? (
        <div className="card flex flex-col items-center py-16 text-center">
          <Search size={40} className="text-slate-700 mb-3" />
          <p className="text-slate-500 text-sm max-w-xl leading-6">
            {skills.length === 0
              ? "No skills found. Add a source and refresh."
              : "No skills match. Try Search or a source filter."}
          </p>
        </div>
      ) : sourceFilter === "all" ? (
        <div className="space-y-6">
          {(Object.keys(SOURCE_LABELS) as NormalizedRegistrySource[]).map((sourceType) => {
            const sourceGroups = groupedBySource.filter((g) => g.sourceType === sourceType);
            if (sourceGroups.length === 0) return null;

            return (
              <section key={sourceType} className="space-y-3">
                <div className="flex items-center justify-between gap-3">
                  <h2 className="text-sm font-semibold text-slate-300">{SOURCE_LABELS[sourceType]}</h2>
                  <span className="text-xs text-slate-500">{grouped[sourceType].length}</span>
                </div>

                <div className="space-y-4 rounded-xl border border-[var(--palette-1e2640)] bg-[var(--palette-0f1225)] p-4">
                  {sourceGroups.map((group) => (
                    <CollapsibleSection
                      key={group.key}
                      defaultOpen
                      count={group.skills.length}
                      title={(
                        <div className="min-w-0">
                          <div className="truncate text-sm font-medium text-slate-200">{group.label}</div>
                          <div className="truncate text-xs text-slate-500">{group.subLabel}</div>
                        </div>
                      )}
                    >
                      <div className="grid gap-4 md:grid-cols-2">
                        {group.skills.map((skill) => (
                          <SkillCard
                            key={`${group.key}:${skill.name}`}
                            skill={skill}
                            installing={installing === skill.name}
                            installed={allInstalled(skill)}
                            canInstall={!disabledReason} disabledReason={disabledReason}
                            onInstall={() => void handleInstall(skill)}
                            onPreview={() => setPreviewTarget(skill)} onTag={(tag) => { setQueryInput(tag); setQuery(tag); }}
                          />
                        ))}
                      </div>
                    </CollapsibleSection>
                  ))}
                </div>
              </section>
            );
          })}
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2">
          {filtered.map((skill) => (
            <SkillCard
              key={`${skill.registry_alias}:${skill.registry_url}:${skill.name}`}
              skill={skill}
              installing={installing === skill.name}
              installed={allInstalled(skill)}
              canInstall={!disabledReason} disabledReason={disabledReason}
              onInstall={() => void handleInstall(skill)}
              onPreview={() => setPreviewTarget(skill)} onTag={(tag) => { setQueryInput(tag); setQuery(tag); }}
            />
          ))}
        </div>
      )}

      {previewTarget && (
        <SkillPreviewModal
          skill={previewTarget}
          installed={allInstalled(previewTarget)}
          canInstall={!disabledReason} disabledReason={disabledReason}
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
