import { Link } from "react-router";
import { Sparkles } from "lucide-react";
import { useRepoStore } from "@/store";
import { useAsync } from "@/hooks/useAsync";
import { skellPresent } from "@/lib/skell";
import { GettingStarted } from "@/components/GettingStarted";
import { FeaturedSources } from "@/components/FeaturedSources";
import { ProjectsAttention } from "@/components/ProjectsAttention";
import { getProjectDisplayName } from "@/lib/navigation";

export function Home() {
  const { repos, selectedRepo, gettingStartedDismissed, setGettingStartedDismissed } = useRepoStore();
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
          <div className="mt-8">
            <GettingStarted embedded />
          </div>
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
        <p className="mt-2 text-sm text-slate-400">
          What needs doing across your projects.
          {gettingStartedDismissed && (
            <>{" "}<button type="button" onClick={() => setGettingStartedDismissed(false)} className="underline hover:text-slate-200">Show the setup guide</button></>
          )}
        </p>
      </div>
      {cliMissing && (
        <div className="rounded-xl border border-amber-500/40 bg-amber-500/10 px-4 py-3 text-sm text-amber-200" role="alert">
          The <code className="font-mono">skell</code> CLI was not found.{" "}
          <a href="https://github.com/aminmesbahi/skell#install" target="_blank" rel="noreferrer" className="underline">Install it</a> to manage skills.
        </div>
      )}
      <GettingStarted />
      {!cliMissing && <ProjectsAttention repos={repos} />}
      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        <Link to="/projects" className="rounded-2xl border border-[var(--palette-1a1f35)] bg-[var(--palette-0f1324)] p-5 hover:bg-white/5">
          <p className="text-sm font-medium text-slate-200">Projects</p>
          <p className="mt-2 text-2xl font-semibold text-slate-200">{repos.length}</p>
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
          {project && <p className="mb-2 text-xs text-slate-500">Adding a source or skill here installs it into {getProjectDisplayName(project)}.</p>}
          <FeaturedSources repo={project || undefined} compact />
        </div>
      )}
    </div>
  );
}
