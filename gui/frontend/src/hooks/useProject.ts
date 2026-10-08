import { useParams } from "react-router";
import { useEffect } from "react";
import { useRepoStore } from "@/store";
import { createProjectId } from "@/lib/navigation";

/** The URL is authoritative; never substitute another project for a stale link. */
export function useProject() {
  const { projectId } = useParams();
  const repos = useRepoStore((s) => s.repos);
  const selected = useRepoStore((s) => s.selectedRepo);
  const select = useRepoStore((s) => s.setSelectedRepo);
  const path = repos.find((path) => createProjectId(path) === projectId) ?? "";
  useEffect(() => { if (path && path !== selected) select(path); }, [path, selected, select]);
  return path;
}
