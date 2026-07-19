export type NavigationProjectKind = "project" | "personal";

export interface NavigationProject {
  id: string;
  path: string;
  name: string;
  kind: NavigationProjectKind;
}

export function isPersonalScope(value?: string | null): boolean {
  return value === "global" || value === "personal" || value === "";
}

export function getProjectDisplayName(path?: string | null): string {
  if (!path || isPersonalScope(path)) return "Personal Skills";
  return path.split(/[\\/]/).filter(Boolean).pop() ?? "Project";
}

export function createProjectId(path: string): string {
  const normalized = path.replace(/\\/g, "/").toLowerCase();
  let hash = 0;
  for (let i = 0; i < normalized.length; i += 1) {
    hash = (hash << 5) - hash + normalized.charCodeAt(i);
    hash |= 0;
  }
  return `project-${Math.abs(hash).toString(36)}`;
}

export function buildProjectRoute(path: string, subPath = "") {
  const id = createProjectId(path);
  return `/projects/${id}${subPath}`;
}

export function resolveProject(projects: NavigationProject[], id?: string | null) {
  return projects.find((project) => project.id === id) ?? null;
}
