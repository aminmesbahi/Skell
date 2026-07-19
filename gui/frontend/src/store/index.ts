import { create } from "zustand";
import { persist } from "zustand/middleware";

export interface RecentProject {
  id: string;
  path: string;
  displayName: string;
  lastOpenedAt: string;
}

interface RepoStoreState {
  repos: string[];
  selectedRepo: string;
  recentProjects: RecentProject[];
  sidebarCollapsed: boolean;
}

interface RepoStore extends RepoStoreState {
  addRepo: (path: string) => void;
  removeRepo: (path: string) => void;
  setSelectedRepo: (path: string) => void;
  toggleSidebar: () => void;
  openProject: (path: string) => void;
}

function normalizeForComparison(path: string): string {
  return path.replace(/\\/g, "/").trim().toLowerCase();
}

function getDisplayName(path: string): string {
  if (!path) return "Project";
  const parts = path.split(/[\\/]/).filter(Boolean);
  const last = parts.at(-1);
  return last || "Project";
}

function createProjectId(path: string): string {
  const normalized = normalizeForComparison(path);
  let hash = 0;
  for (let i = 0; i < normalized.length; i += 1) {
    hash = (hash << 5) - hash + normalized.charCodeAt(i);
    hash |= 0;
  }
  return `project-${Math.abs(hash).toString(36)}`;
}

function coerceProjectRecord(path: string, lastOpenedAt?: string): RecentProject {
  return {
    id: createProjectId(path),
    path,
    displayName: getDisplayName(path),
    lastOpenedAt: lastOpenedAt ?? new Date().toISOString(),
  };
}

function dedupeProjects(projects: RecentProject[]): RecentProject[] {
  const seen = new Set<string>();
  return projects.reduce<RecentProject[]>((acc, project) => {
    const key = normalizeForComparison(project.path);
    if (seen.has(key)) return acc;
    seen.add(key);
    acc.push(project);
    return acc;
  }, []);
}

export function migrateRepoStoreState(input: Partial<RepoStoreState> & { repos?: string[]; selectedRepo?: string }): RepoStoreState {
  const seen = new Set<string>();
  const repos = Array.isArray(input.repos)
    ? input.repos.filter((repo): repo is string => typeof repo === "string").filter((repo) => {
        const key = normalizeForComparison(repo);
        if (seen.has(key)) return false;
        seen.add(key);
        return true;
      })
    : [];
  const selectedRepo = typeof input.selectedRepo === "string" ? input.selectedRepo : "global";
  const recentProjects = dedupeProjects(
    repos.filter((path) => path !== "global").map((path) => coerceProjectRecord(path, new Date().toISOString()))
  ).slice(0, 5);

  return {
    repos,
    selectedRepo,
    recentProjects,
    sidebarCollapsed: false,
  };
}

const createRepoStore = () =>
  create<RepoStore>()(
    persist(
      (set) => ({
        repos: [],
        selectedRepo: "global",
        recentProjects: [],
        sidebarCollapsed: false,

        addRepo: (path) =>
          set((s) => ({
            repos: s.repos.includes(path) ? s.repos : [...s.repos, path],
            selectedRepo: path,
            recentProjects: dedupeProjects([
              coerceProjectRecord(path),
              ...s.recentProjects.filter((project) => project.path !== path),
            ]).slice(0, 5),
          })),

        removeRepo: (path) =>
          set((s) => {
            const repos = s.repos.filter((r) => r !== path);
            return {
              repos,
              selectedRepo:
                s.selectedRepo === path ? (repos[0] ?? "global") : s.selectedRepo,
              recentProjects: s.recentProjects.filter((project) => project.path !== path),
            };
          }),

        setSelectedRepo: (path) => set((s) => ({
          selectedRepo: path,
          recentProjects: path && path !== "global"
            ? dedupeProjects([
                coerceProjectRecord(path, new Date().toISOString()),
                ...s.recentProjects.filter((project) => project.path !== path),
              ]).slice(0, 5)
            : s.recentProjects,
        })),

        toggleSidebar: () =>
          set((s) => ({ sidebarCollapsed: !s.sidebarCollapsed })),

        openProject: (path) =>
          set((s) => ({
            selectedRepo: path,
            recentProjects: dedupeProjects([
              coerceProjectRecord(path, new Date().toISOString()),
              ...s.recentProjects.filter((project) => project.path !== path),
            ]).slice(0, 5),
          })),
      }),
      {
        name: "skell-gui-repos",
        version: 2,
        migrate: (persistedState, version) => {
          if (version < 2) {
            return migrateRepoStoreState(persistedState as Partial<RepoStoreState> & { repos?: string[]; selectedRepo?: string });
          }
          return persistedState as RepoStoreState;
        },
      }
    )
  );

export const useRepoStore = createRepoStore();

// ---------------------------------------------------------------------------
// UI store — transient, not persisted
// ---------------------------------------------------------------------------

interface Notification {
  id: string;
  kind: "success" | "error" | "info";
  title: string;
  detail?: string;
}

interface UIStore {
  notifications: Notification[];

  notify: (n: Omit<Notification, "id">) => void;
  dismissNotification: (id: string) => void;
}

let _notifCounter = 0;

export const useUIStore = create<UIStore>()((set) => ({
  notifications: [],

  notify: (n) => {
    const id = String(++_notifCounter);
    set((s) => ({ notifications: [...s.notifications, { ...n, id }] }));
    // Auto-dismiss success/info after 4 s
    if (n.kind !== "error") {
      setTimeout(() => {
        set((s) => ({
          notifications: s.notifications.filter((x) => x.id !== id),
        }));
      }, 4000);
    }
  },

  dismissNotification: (id) =>
    set((s) => ({
      notifications: s.notifications.filter((n) => n.id !== id),
    })),
}));
