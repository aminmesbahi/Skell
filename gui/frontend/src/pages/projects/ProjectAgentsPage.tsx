import { useProject } from "@/hooks/useProject";
import { useAsync } from "@/hooks/useAsync";
import { detectRepoTargets, listSupportedTargets } from "@/lib/skell";
import { getProjectDisplayName } from "@/lib/navigation";
import { ProjectPageHeader } from "@/components/ProjectPageHeader";
import { LoadState, MissingProject } from "@/components/LoadState";
export function ProjectAgentsPage() {
  const project = useProject();
  const state = useAsync(project, async () => {
    if (!project) return [];
    const [detected, supported] = await Promise.all([detectRepoTargets(project), listSupportedTargets()]);
    return [...detected, ...supported.filter((t) => !detected.some((d) => d.id === t.id))];
  });
  if (!project) return <MissingProject />;
  return <div className="mx-auto max-w-6xl px-6 py-8 space-y-6">
    <ProjectPageHeader projectPath={project} title={`Agents for ${getProjectDisplayName(project)}`} breadcrumb="Agents" />
    <LoadState {...state} retry={() => void state.refresh()} />
    {state.data?.map((target) => <div className="card flex justify-between" key={target.id}><div><h2>{target.displayName}</h2><p>{target.dir}/skills/</p></div><span>{target.detected ? "Detected" : "Available"}</span></div>)}
  </div>;
}
