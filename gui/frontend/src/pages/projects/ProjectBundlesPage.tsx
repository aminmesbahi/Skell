import { useState } from "react";
import { Check, Package, PackagePlus, Plus, Trash2 } from "lucide-react";
import { useProject } from "@/hooks/useProject";
import { useAsync } from "@/hooks/useAsync";
import { useDialog } from "@/hooks/useDialog";
import { deleteBundle, installBundle, listBundles, listInstalled, saveBundle } from "@/lib/skell";
import { getProjectDisplayName } from "@/lib/navigation";
import { BUNDLE_NAME_RX, parseSkillList } from "@/lib/bundles";
import { LoadState, MissingProject } from "@/components/LoadState";
import { ProjectPageHeader } from "@/components/ProjectPageHeader";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import { useUIStore } from "@/store";
import type { BundleInfo } from "@/lib/types";

function CreateBundleDialog({ installed, existing, onClose, onCreate }: {
  installed: string[];
  existing: string[];
  onClose: () => void;
  onCreate: (name: string, description: string, skills: string[]) => Promise<void>;
}) {
  const ref = useDialog(true, onClose);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [picked, setPicked] = useState<string[]>([]);
  const [others, setOthers] = useState("");
  const [busy, setBusy] = useState(false);

  const skills = [...new Set([...picked, ...parseSkillList(others)])];
  const nameError = name && !BUNDLE_NAME_RX.test(name) ? "Use lowercase letters, digits and hyphens." : existing.includes(name) ? "A bundle with this name exists and would be replaced." : "";
  const canSave = BUNDLE_NAME_RX.test(name) && skills.length > 0 && !busy;

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!canSave) return;
    setBusy(true);
    try {
      await onCreate(name, description, skills);
    } finally {
      setBusy(false);
    }
  }

  return (
    <div ref={ref} role="dialog" aria-modal="true" aria-labelledby="new-bundle-title" tabIndex={-1} className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 cursor-pointer bg-black/60 backdrop-blur-sm" onClick={onClose} />
      <form onSubmit={(e) => void submit(e)} className="relative z-10 mx-4 flex max-h-[85vh] w-full max-w-lg flex-col rounded-2xl border border-[var(--palette-2d3348)] bg-[var(--palette-13162a)] shadow-2xl">
        <div className="border-b border-[var(--palette-1e2540)] p-5">
          <h2 id="new-bundle-title" className="text-lg font-bold text-slate-200">New bundle</h2>
          <p className="mt-1 text-sm text-slate-400">A bundle is a starter pack: one click installs everything in it.</p>
        </div>
        <div className="flex-1 space-y-4 overflow-y-auto p-5 text-sm">
          <label className="block">
            <span className="mb-1 block text-slate-300">Name</span>
            <input className="input" value={name} onChange={(e) => setName(e.target.value.trim())} placeholder="frontend" autoFocus aria-invalid={Boolean(nameError)} />
            {nameError && <span className="mt-1 block text-xs text-amber-400">{nameError}</span>}
          </label>
          <label className="block">
            <span className="mb-1 block text-slate-300">What is it for? <span className="text-slate-600">(optional)</span></span>
            <input className="input" value={description} onChange={(e) => setDescription(e.target.value)} placeholder="Everything a UI developer needs" />
          </label>
          <fieldset>
            <legend className="mb-1 text-slate-300">Skills installed here</legend>
            {installed.length === 0 ? (
              <p className="text-slate-500">No skills installed yet. Type names below.</p>
            ) : (
              <div className="grid max-h-40 grid-cols-1 gap-1 overflow-y-auto sm:grid-cols-2">
                {installed.map((s) => (
                  <label key={s} className="flex cursor-pointer items-center gap-2 rounded-md px-2 py-1 hover:bg-white/5">
                    <input type="checkbox" checked={picked.includes(s)} onChange={(e) => setPicked((p) => (e.target.checked ? [...p, s] : p.filter((x) => x !== s)))} />
                    <span className="truncate font-mono text-xs text-slate-300">{s}</span>
                  </label>
                ))}
              </div>
            )}
          </fieldset>
          <label className="block">
            <span className="mb-1 block text-slate-300">Other skills <span className="text-slate-600">(names, separated by commas)</span></span>
            <input className="input" value={others} onChange={(e) => setOthers(e.target.value)} placeholder="pdf, docx" />
          </label>
          <p className="text-xs text-slate-500">{skills.length === 0 ? "Choose at least one skill." : `${skills.length} skill${skills.length === 1 ? "" : "s"}: ${skills.join(", ")}`}</p>
        </div>
        <div className="flex justify-end gap-2 border-t border-[var(--palette-1e2540)] p-4">
          <button type="button" onClick={onClose} className="btn-ghost text-xs">Cancel</button>
          <button type="submit" disabled={!canSave} className="btn-primary text-xs">{busy ? "Saving…" : "Create bundle"}</button>
        </div>
      </form>
    </div>
  );
}

function BundleCard({ bundle, installedNames, busy, onInstall, onDelete }: {
  bundle: BundleInfo;
  installedNames: Set<string>;
  busy: boolean;
  onInstall: () => void;
  onDelete: () => void;
}) {
  const total = bundle.skills.length;
  const missing = total - bundle.installed;
  const pct = total === 0 ? 0 : Math.round((bundle.installed / total) * 100);
  return (
    <article className="card flex flex-col gap-3" aria-label={`Bundle ${bundle.name}`}>
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h3 className="flex items-center gap-2 font-semibold text-slate-200"><Package size={15} className="text-brand-400" />@{bundle.name}</h3>
          {bundle.description && <p className="mt-0.5 text-sm text-slate-400">{bundle.description}</p>}
        </div>
        <button className="rounded p-1.5 text-slate-500 transition hover:bg-red-500/10 hover:text-red-400" onClick={onDelete} aria-label={`Delete bundle ${bundle.name}`} title="Delete bundle (its skills stay installed)">
          <Trash2 size={15} />
        </button>
      </div>
      <div>
        <div className="mb-1 flex justify-between text-xs text-slate-500">
          <span>{bundle.installed} of {total} installed</span>
          <span>{pct}%</span>
        </div>
        <div className="h-1.5 overflow-hidden rounded-full bg-white/10" role="progressbar" aria-valuenow={pct} aria-valuemin={0} aria-valuemax={100} aria-label={`${bundle.name} installed`}>
          <div className="h-full rounded-full bg-brand-500 transition-all" style={{ width: `${pct}%` }} />
        </div>
      </div>
      <ul className="flex flex-wrap gap-1.5">
        {bundle.skills.map((s) => (
          <li key={s} className={installedNames.has(s) ? "inline-flex items-center gap-1 rounded-md bg-emerald-500/10 px-2 py-0.5 font-mono text-xs text-emerald-300" : "inline-flex items-center gap-1 rounded-md bg-white/5 px-2 py-0.5 font-mono text-xs text-slate-400"}>
            {installedNames.has(s) && <Check size={10} aria-label="installed" />}
            {s}
          </li>
        ))}
      </ul>
      <div className="mt-auto">
        {missing > 0 ? (
          <button className="btn-primary" disabled={busy} onClick={onInstall}><PackagePlus size={14} />{busy ? "Installing…" : `Install ${missing} missing skill${missing === 1 ? "" : "s"}`}</button>
        ) : (
          <p className="flex items-center gap-1.5 text-sm text-emerald-400"><Check size={14} /> Everything is installed</p>
        )}
      </div>
    </article>
  );
}

/** Bundles tab: starter packs of skills for a role or team. */
export function ProjectBundlesPage() {
  const project = useProject();
  const { notify } = useUIStore();
  const [creating, setCreating] = useState(false);
  const [busy, setBusy] = useState<string | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<string | null>(null);
  const state = useAsync(project, async () => {
    if (!project) return null;
    const [bundles, installed] = await Promise.all([listBundles(project), listInstalled(project)]);
    return { bundles, installed: installed.map((s) => s.name) };
  });
  if (!project) return <MissingProject />;

  const bundles = state.data?.bundles ?? [];
  const installedNames = new Set(state.data?.installed ?? []);

  async function create(name: string, description: string, skills: string[]) {
    const res = await saveBundle({ repo: project, name, description, skills });
    if (res.success) {
      notify({ kind: "success", title: `Bundle @${name} created`, detail: `${skills.length} skills` });
      setCreating(false);
      await state.refresh();
    } else {
      notify({ kind: "error", title: "Could not create the bundle", detail: res.stderr || res.stdout });
    }
  }

  async function install(name: string) {
    setBusy(name);
    try {
      const res = await installBundle(project, name);
      if (res.success) {
        notify({ kind: "success", title: `Installed @${name}`, detail: res.stdout.trim().split("\n").slice(-1)[0] });
      } else {
        notify({ kind: "error", title: `Could not install @${name}`, detail: (res.stderr || res.stdout).trim() });
      }
    } finally {
      setBusy(null);
      await state.refresh();
    }
  }

  async function remove(name: string) {
    setConfirmDelete(null);
    const res = await deleteBundle(project, name);
    notify(res.success ? { kind: "success", title: `Bundle @${name} deleted` } : { kind: "error", title: "Could not delete the bundle", detail: res.stderr });
    await state.refresh();
  }

  return (
    <div className="mx-auto max-w-6xl space-y-6 px-6 py-8">
      <ProjectPageHeader
        projectPath={project}
        title={`Bundles for ${getProjectDisplayName(project)}`}
        subtitle="Starter packs of skills for a role or team. Bundles live in skell.toml, so teammates get them too."
        actions={<button className="btn-primary" onClick={() => setCreating(true)}><Plus size={14} />New bundle</button>}
      />
      <LoadState loading={state.loading && !state.data} error={state.error} retry={() => void state.refresh()} />

      {state.data && bundles.length === 0 && (
        <div className="card text-center">
          <Package size={28} className="mx-auto mb-2 text-slate-600" />
          <p className="font-medium text-slate-200">No bundles yet</p>
          <p className="mx-auto mt-1 max-w-md text-sm text-slate-500">Group the skills a role needs, like “frontend” or “release manager”. New teammates install the whole set in one click, and dependencies come along automatically.</p>
          <button className="btn-primary mt-4" onClick={() => setCreating(true)}><Plus size={14} />Create your first bundle</button>
        </div>
      )}

      {bundles.length > 0 && (
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
          {bundles.map((b) => (
            <BundleCard key={b.name} bundle={b} installedNames={installedNames} busy={busy === b.name} onInstall={() => void install(b.name)} onDelete={() => setConfirmDelete(b.name)} />
          ))}
        </div>
      )}

      {creating && <CreateBundleDialog installed={state.data?.installed ?? []} existing={bundles.map((b) => b.name)} onClose={() => setCreating(false)} onCreate={create} />}
      <ConfirmDialog
        open={confirmDelete !== null}
        title="Delete bundle?"
        description={`Delete @${confirmDelete}? The skills in it stay installed; only the group is removed from skell.toml.`}
        confirmLabel="Delete bundle"
        danger
        onConfirm={() => confirmDelete && void remove(confirmDelete)}
        onCancel={() => setConfirmDelete(null)}
      />
    </div>
  );
}
