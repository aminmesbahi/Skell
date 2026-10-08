import { Link } from "react-router";
import { Plus, Compass, Sparkles, CheckCircle2, Circle, Terminal } from "lucide-react";
import { useRepoStore } from "@/store";
import { useAsync } from "@/hooks/useAsync";
import { skellPresent } from "@/lib/skell";
import { FeaturedSources } from "@/components/FeaturedSources";
import { getProjectDisplayName } from "@/lib/navigation";

function Step({ done, title, children }: { done: boolean; title: string; children?: React.ReactNode }) {
  return (
    <li className="flex gap-3">
      {done ? <CheckCircle2 size={18} className="mt-0.5 shrink-0 text-emerald-400" aria-label="done" /> : <Circle size={18} className="mt-0.5 shrink-0 text-slate-600" aria-label="to do" />}
      <div className="min-w-0 flex-1">
        <p className={done ? "text-sm text-slate-400 line-through decoration-slate-600" : "text-sm font-medium text-slate-200"}>{title}</p>
        {!done && children && <div className="mt-2">{children}</div>}
      </div>
    </li>
  );
}

export function Home() {
  const { repos, selectedRepo } = useRepoStore();
  const cli = useAsync("skell-present", () => skellPresent());
  const cliMissing = cli.data === false;

  if (repos.length === 0) {
    return (
      <div className="mx-auto flex min-h-full max-w-3xl flex-col justify-center gap-6 px-6 py-16">
        <div className="rounded-2xl border border-[var(--palette-1a1f35)] bg-[var(--palette-0f1324)] p-8 shadow-2xl shadow-black/20">
          <div className="mb-6 flex h-12 w-12 items-center justify-center rounded-full bg-brand-600/20 text-brand-400">
            <Sparkles size={24} />
          </div>
          <h1 className="text-2xl font-bold text-slate-200">Manage agent skills across your projects</h1>
          <p className="mt-3 text-sm leading-6 text-slate-400">
            Skell finds, installs and updates Agent Skills (SKILL.md) for Claude Code, Codex, Copilot, Cursor and more — and keeps them identical for everyone on your team.
          </p>
          <ol className="mt-8 space-y-4" aria-label="Getting started">
            <Step done={!cliMissing} title="Install the Skell command-line tool">
              <p className="flex items-center gap-2 text-xs text-slate-400">
                <Terminal size={13} /> The app uses the <code className="font-mono">skell</code> CLI.{" "}
                <a href="https://github.com/aminmesbahi/skell#install" target="_blank" rel="noreferrer" className="underline hover:text-slate-200">Install instructions →</a>
              </p>
            </Step>
            <Step done={false} title="Add a project folder">
              <Link to="/projects" className="inline-flex items-center gap-2 rounded-lg bg-brand-600 px-4 py-2 text-sm font-medium text-white hover:bg-brand-500">
                <Plus size={16} />
                Add Project
              </Link>
            </Step>
            <Step done={false} title="Pick skills from a source and install them">
              <Link to="/catalog" className="inline-flex items-center gap-2 rounded-lg border border-[var(--palette-2a3353)] px-4 py-2 text-sm font-medium text-slate-300 hover:bg-white/5">
                <Compass size={16} />
                Browse Catalog
              </Link>
            </Step>
          </ol>
        </div>
        {!cliMissing && <FeaturedSources compact />}
      </div>
    );
  }

  const project = selectedRepo && selectedRepo !== "global" ? selectedRepo : "";
  return (
    <div className="mx-auto max-w-6xl space-y-8 px-6 py-8">
      <div>
        <h1 className="text-2xl font-bold text-slate-200">Home</h1>
        <p className="mt-2 text-sm text-slate-400">Actionable summaries for the projects you manage.</p>
      </div>
      {cliMissing && (
        <div className="rounded-xl border border-amber-500/40 bg-amber-500/10 px-4 py-3 text-sm text-amber-200" role="alert">
          The <code className="font-mono">skell</code> CLI was not found.{" "}
          <a href="https://github.com/aminmesbahi/skell#install" target="_blank" rel="noreferrer" className="underline">Install it</a> to manage skills.
        </div>
      )}
      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        <Link to="/projects" className="rounded-2xl border border-[var(--palette-1a1f35)] bg-[var(--palette-0f1324)] p-5 hover:bg-white/5">
          <p className="text-sm font-medium text-slate-200">Projects</p>
          <p className="mt-2 text-2xl font-semibold text-white">{repos.length}</p>
        </Link>
        <Link to="/catalog" className="rounded-2xl border border-[var(--palette-1a1f35)] bg-[var(--palette-0f1324)] p-5 hover:bg-white/5">
          <p className="text-sm font-medium text-slate-200">Browse the catalog</p>
          <p className="mt-2 text-sm text-slate-400">Discover skills for the tools you use.</p>
        </Link>
        <Link to="/activity" className="rounded-2xl border border-[var(--palette-1a1f35)] bg-[var(--palette-0f1324)] p-5 hover:bg-white/5">
          <p className="text-sm font-medium text-slate-200">Recent activity</p>
          <p className="mt-2 text-sm text-slate-400">Open the activity feed for recent changes.</p>
        </Link>
      </div>
      {!cliMissing && (
        <div>
          {project && <p className="mb-2 text-xs text-slate-500">Adding to {getProjectDisplayName(project)}</p>}
          <FeaturedSources repo={project || undefined} compact />
        </div>
      )}
    </div>
  );
}
