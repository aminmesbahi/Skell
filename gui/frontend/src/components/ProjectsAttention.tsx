import { Link } from "react-router";
import { AlertTriangle, ArrowUp, CheckCircle2 } from "lucide-react";
import { useAsync } from "@/hooks/useAsync";
import { getStatus, isRepoInitialized, listInstalled } from "@/lib/skell";
import { buildProjectRoute, getProjectDisplayName } from "@/lib/navigation";

interface ProjectSummary {
  path: string;
  total: number;
  outdated: number;
  modified: number;
  unknown: number;
  notSetUp?: boolean;
  error?: string;
}

const MAX_PROJECTS = 6;

/** What needs doing across the user's projects, with a link to each. */
export function ProjectsAttention({ repos }: { repos: string[] }) {
  const paths = repos.filter((repo) => repo && repo !== "global");
  const shown = paths.slice(0, MAX_PROJECTS);
  const state = useAsync(`attention:${shown.join("|")}`, () =>
    Promise.all(
      shown.map(async (path): Promise<ProjectSummary> => {
        try {
          // A folder that was never set up has no manifest to check; that is
          // something to do, not something that went wrong.
          if (!(await isRepoInitialized(path))) return { path, total: 0, outdated: 0, modified: 0, unknown: 0, notSetUp: true };
          const [skills, statuses] = await Promise.all([listInstalled(path), getStatus(path)]);
          const count = (status: string) => statuses.filter((s) => s.status === status).length;
          return { path, total: skills.length, outdated: count("outdated"), modified: count("locally-modified"), unknown: count("unknown") };
        } catch (error) {
          return { path, total: 0, outdated: 0, modified: 0, unknown: 0, error: String(error) };
        }
      })
    )
  );

  if (shown.length === 0) return null;
  const summaries = state.data ?? [];
  const needsAttention = summaries.filter((s) => s.error || s.notSetUp || s.outdated || s.modified || s.unknown);

  return (
    <section className="rounded-2xl border border-[var(--palette-1a1f35)] bg-[var(--palette-0f1324)] p-5" aria-labelledby="attention-heading">
      <div className="flex items-center justify-between gap-3">
        <h2 id="attention-heading" className="text-sm font-semibold text-slate-200">Needs attention</h2>
        {paths.length > shown.length && (
          <Link to="/projects" className="text-xs text-slate-500 hover:text-slate-300">
            + {paths.length - shown.length} more projects
          </Link>
        )}
      </div>
      {state.loading && !state.data ? (
        <p className="mt-3 text-sm text-slate-500" role="status">Checking your projects…</p>
      ) : needsAttention.length === 0 ? (
        <p className="mt-3 flex items-center gap-2 text-sm text-emerald-400">
          <CheckCircle2 size={16} />
          {summaries.length === 1 ? "Your project is up to date." : `All ${summaries.length} projects are up to date.`}
        </p>
      ) : (
        <ul className="mt-3 divide-y divide-[var(--palette-1a1f35)]">
          {needsAttention.map((summary) => (
            <li key={summary.path}>
              <Link to={`${buildProjectRoute(summary.path)}/skills`} className="flex flex-wrap items-center gap-x-4 gap-y-1 py-3 hover:bg-white/5">
                <span className="min-w-0 flex-1 truncate text-sm font-medium text-slate-200" title={summary.path}>{getProjectDisplayName(summary.path)}</span>
                {summary.notSetUp ? (
                  <span className="text-xs text-slate-400">Not set up yet</span>
                ) : summary.error ? (
                  <span className="inline-flex items-center gap-1 text-xs text-red-300" title={summary.error}><AlertTriangle size={13} /> Could not check</span>
                ) : (
                  <span className="flex flex-wrap items-center gap-3 text-xs">
                    {summary.outdated > 0 && <span className="inline-flex items-center gap-1 text-amber-300"><ArrowUp size={13} /> {summary.outdated} outdated</span>}
                    {summary.modified > 0 && <span className="text-sky-300">{summary.modified} edited locally</span>}
                    {summary.unknown > 0 && <span className="text-slate-400">{summary.unknown} unknown</span>}
                  </span>
                )}
              </Link>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
