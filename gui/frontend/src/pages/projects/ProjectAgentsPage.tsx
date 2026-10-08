import { useState } from "react";
import { useProject } from "@/hooks/useProject";
import { useAsync } from "@/hooks/useAsync";
import { activeRepoTarget, detectRepoTargets, listMirrors, listSupportedTargets, setMirror } from "@/lib/skell";
import { getProjectDisplayName } from "@/lib/navigation";
import { ProjectPageHeader } from "@/components/ProjectPageHeader";
import { LoadState, MissingProject } from "@/components/LoadState";
import { useUIStore } from "@/store";

export function ProjectAgentsPage() {
  const project = useProject();
  const { notify } = useUIStore();
  const [busy, setBusy] = useState<string | null>(null);
  const state = useAsync(project, async () => {
    if (!project) return null;
    const [detected, supported, primary, mirrors] = await Promise.all([
      detectRepoTargets(project),
      listSupportedTargets(),
      Promise.resolve(activeRepoTarget(project)).catch(() => ""),
      Promise.resolve(listMirrors(project)).catch(() => [] as string[]),
    ]);
    const targets = [...(detected ?? []), ...(supported ?? []).filter((t) => !(detected ?? []).some((d) => d.id === t.id))];
    return { targets, primary: primary ?? "", mirrors: mirrors ?? [] };
  });
  if (!project) return <MissingProject />;

  async function toggle(id: string, enabled: boolean) {
    if (!project) return;
    setBusy(id);
    try {
      await setMirror(project, id, enabled);
      notify({ kind: "success", title: enabled ? `Skills now also provided to ${id}` : `Stopped providing skills to ${id}` });
      await state.refresh();
    } catch (e) {
      notify({ kind: "error", title: "Could not update mirrors", detail: String(e) });
    } finally {
      setBusy(null);
    }
  }

  const primary = state.data?.primary ?? "";
  const mirrors = state.data?.mirrors ?? [];
  return <div className="mx-auto max-w-6xl px-6 py-8 space-y-6">
    <ProjectPageHeader projectPath={project} title={`Agents for ${getProjectDisplayName(project)}`} breadcrumb="Agents" />
    <p className="text-sm text-slate-400">
      Using more than one agent in this project? Turn an agent on to give it a copy of every skill — Skell keeps the copies in sync on install, upgrade, remove and sync.
    </p>
    <LoadState loading={state.loading} error={state.error} retry={() => void state.refresh()} />
    {state.data?.targets.map((target) => {
      const isPrimary = target.id === primary;
      const isMirror = mirrors.includes(target.id);
      return <div className="card flex flex-wrap items-center justify-between gap-3" key={target.id}>
        <div>
          <h2 className="text-sm font-medium text-slate-200">{target.displayName}</h2>
          <p className="font-mono text-xs text-slate-500">{target.dir}/skills/</p>
        </div>
        {isPrimary ? (
          <span className="rounded-full border border-brand-500/30 bg-brand-500/10 px-2 py-0.5 text-xs text-brand-300" title="This agent's folder holds skell.toml and skell.lock">Primary</span>
        ) : primary ? (
          <label className="flex items-center gap-2 text-xs text-slate-300">
            <input
              type="checkbox"
              checked={isMirror}
              disabled={busy !== null}
              onChange={(e) => void toggle(target.id, e.target.checked)}
              aria-label={`Provide skills to ${target.displayName}`}
            />
            {busy === target.id ? "Updating…" : isMirror ? "Receives skills" : target.detected ? "Detected — not managed" : "Off"}
          </label>
        ) : (
          <span className="text-xs text-slate-500">{target.detected ? "Detected" : "Available"}</span>
        )}
      </div>;
    })}
  </div>;
}
