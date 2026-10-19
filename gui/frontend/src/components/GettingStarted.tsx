import type { ReactNode } from "react";
import { Link } from "react-router";
import { CheckCircle2, Circle, Plus, X } from "lucide-react";
import { useRepoStore } from "@/store";
import { useAsync } from "@/hooks/useAsync";
import { isRepoInitialized, listInstalled, listRegistry, skellPresent } from "@/lib/skell";
import { buildProjectRoute, getProjectDisplayName } from "@/lib/navigation";
import { SelectDirectory } from "../../bindings/skell-gui/app";
import { Hint } from "@/components/Hint";

interface Progress {
  cli: boolean;
  project: boolean;
  setup: boolean;
  source: boolean;
  installed: boolean;
}

const attempt = async <T,>(work: () => Promise<T>, fallback: T): Promise<T> => {
  try { return await work(); } catch { return fallback; }
};

/** The project the guide tracks: the selected one, else the first added. */
export function guideProject(repos: string[], selectedRepo: string): string {
  const real = repos.filter((repo) => repo && repo !== "global");
  return real.includes(selectedRepo) ? selectedRepo : real[0] ?? "";
}

function Step({ done, current, title, children }: { done: boolean; current: boolean; title: ReactNode; children?: ReactNode }) {
  return (
    <li className="flex gap-3" aria-current={current ? "step" : undefined}>
      {done
        ? <CheckCircle2 size={18} className="mt-0.5 shrink-0 text-emerald-400" aria-label="done" />
        : <Circle size={18} className={`mt-0.5 shrink-0 ${current ? "text-brand-400" : "text-slate-600"}`} aria-label="to do" />}
      <div className="min-w-0 flex-1">
        <p className={done ? "text-sm text-slate-400 line-through decoration-slate-600" : current ? "text-sm font-medium text-slate-200" : "text-sm text-slate-400"}>{title}</p>
        {current && children && <div className="mt-2 space-y-2">{children}</div>}
      </div>
    </li>
  );
}

/**
 * A checklist driven by what actually exists on disk — CLI found, project
 * added, project set up, a source added, a skill installed — so it always
 * shows the next real step and a button that goes straight to it.
 */
export function GettingStarted({ embedded = false }: { embedded?: boolean }) {
  const { repos, selectedRepo, addRepo, gettingStartedDismissed, setGettingStartedDismissed } = useRepoStore();
  const project = guideProject(repos, selectedRepo);
  const name = project ? getProjectDisplayName(project) : "";
  const state = useAsync(`getting-started:${project}`, async (): Promise<Progress> => {
    const cli = await attempt(() => skellPresent(), true);
    if (!project) return { cli, project: false, setup: false, source: false, installed: false };
    const setup = await attempt(() => isRepoInitialized(project), false);
    if (!setup) return { cli, project: true, setup: false, source: false, installed: false };
    const [skills, installed] = await Promise.all([
      attempt(() => listRegistry(project), []),
      attempt(() => listInstalled(project), []),
    ]);
    return { cli, project: true, setup: true, source: skills.length > 0 || installed.length > 0, installed: installed.length > 0 };
  });

  if (!embedded && gettingStartedDismissed) return null;
  const progress = state.data;
  const order: (keyof Progress)[] = ["cli", "project", "setup", "source", "installed"];
  const doneCount = progress ? order.filter((key) => progress[key]).length : 0;
  const currentKey = progress ? order.find((key) => !progress[key]) : undefined;
  const allDone = progress !== undefined && currentKey === undefined;
  const catalogState = { installDestination: project };

  async function chooseProject() {
    const selected = await SelectDirectory();
    if (selected) addRepo(selected);
  }

  const list = (
    <ol className="space-y-4" aria-label="Getting started" data-tour={embedded ? "guide" : undefined}>
      <Step done={progress?.cli ?? true} current={currentKey === "cli"} title="Install the Skell command-line tool">
        <p className="text-xs text-slate-400">
          The app runs the <code className="font-mono">skell</code> CLI behind the scenes.{" "}
          <a href="https://github.com/aminmesbahi/skell#install" target="_blank" rel="noreferrer" className="underline hover:text-slate-200">Install instructions →</a>
        </p>
      </Step>
      <Step done={progress?.project ?? false} current={currentKey === "project"} title="Add a project folder">
        <p className="text-xs text-slate-400">Pick the folder of a codebase you work on. Skills are installed into it.</p>
        <button type="button" onClick={() => void chooseProject()} className="btn-primary text-sm"><Plus size={16} />Add Project</button>
      </Step>
      <Step done={progress?.setup ?? false} current={currentKey === "setup"} title={<>Set up {name || "the project"} for Skell <Hint term="setup" /></>}>
        <p className="text-xs text-slate-400">Choose which <Hint term="agent">agent</Hint> you use here, such as Claude Code or Cursor.</p>
        <Link to={`${project ? buildProjectRoute(project) : "/projects"}/skills`} className="btn-primary text-sm">Set up {name || "project"}</Link>
      </Step>
      <Step done={progress?.source ?? false} current={currentKey === "source"} title={<>Add a source of skills <Hint term="source" /></>}>
        <p className="text-xs text-slate-400">The Catalog lists well-known sources such as Anthropic's — add one with a single click.</p>
        <Link to="/catalog" state={catalogState} className="btn-primary text-sm">Choose a source</Link>
      </Step>
      <Step done={progress?.installed ?? false} current={currentKey === "installed"} title={<>Install your first <Hint term="skill">skill</Hint></>}>
        <p className="text-xs text-slate-400">Preview a skill, then install it. Skills that ship scripts ask for your review first.</p>
        <Link to="/catalog" state={catalogState} className="btn-primary text-sm">Browse skills</Link>
      </Step>
    </ol>
  );

  if (embedded) return list;

  return (
    <section className="rounded-2xl border border-[var(--palette-1a1f35)] bg-[var(--palette-0f1324)] p-5" aria-labelledby="getting-started-heading" data-tour="guide">
      <div className="mb-4 flex items-center justify-between gap-3">
        <div>
          <h2 id="getting-started-heading" className="text-sm font-semibold text-slate-200">Getting started</h2>
          <p className="mt-0.5 text-xs text-slate-500">
            {allDone ? "You're all set." : progress ? `${doneCount} of ${order.length} steps done` : "Checking your setup…"}
          </p>
        </div>
        <button type="button" onClick={() => setGettingStartedDismissed(true)} className="btn-subtle" aria-label="Hide the getting started guide" title="Hide this guide. You can bring it back from Home.">
          {allDone ? "Done" : <X size={14} />}
        </button>
      </div>
      {progress && !allDone && list}
    </section>
  );
}
