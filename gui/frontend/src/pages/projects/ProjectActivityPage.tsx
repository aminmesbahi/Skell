import { useProject } from "@/hooks/useProject";
import { useAsync } from "@/hooks/useAsync";
import { readAuditLog } from "@/lib/skell";
import { getProjectDisplayName } from "@/lib/navigation";
import { ProjectPageHeader } from "@/components/ProjectPageHeader";
import { LoadState, MissingProject } from "@/components/LoadState";
export function ProjectActivityPage() {
  const project = useProject();
  const normalize = (path: string) => path.replace(/\\/g, "/").replace(/\/$/, "");
  const state = useAsync(project, async () => project ? (await readAuditLog()).filter((event) => event.repo && normalize(event.repo) === normalize(project)) : []);
  if (!project) return <MissingProject />;
  return <div className="mx-auto max-w-6xl px-6 py-8 space-y-6">
    <ProjectPageHeader projectPath={project} title={`Activity for ${getProjectDisplayName(project)}`} breadcrumb="Activity" actions={<button className="btn-ghost" onClick={() => void state.refresh()} disabled={state.loading}>Refresh</button>} />
    <LoadState {...state} retry={() => void state.refresh()} />
    {state.data?.map((event, i) => <div className="card flex justify-between gap-4" key={`${event.timestamp}-${i}`}><div><h2>{event.skill}</h2><p>{event.action} {event.version}</p></div><time dateTime={event.timestamp}>{new Date(event.timestamp).toLocaleString()}</time></div>)}
    {!state.loading && !state.error && state.data?.length === 0 && <p className="card">No activity recorded yet.</p>}
  </div>;
}
