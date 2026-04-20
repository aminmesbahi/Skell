import { create } from "zustand";
import { persist } from "zustand/middleware";

// ---------------------------------------------------------------------------
// Repo store — persisted across app restarts
// ---------------------------------------------------------------------------

interface RepoStore {
  /** Absolute paths of all managed repositories (local + global placeholder) */
  repos: string[];
  /** Currently focused repo (path or "global") */
  selectedRepo: string;
  /** Scroll positions keyed by page id so each view remembers where it was */
  scrollPositions: Record<string, number>;
  /** Skills table: which columns are visible */
  skillsColumns: string[];
  /** Last search query per page, keyed by page id */
  lastQueries: Record<string, string>;
  /** Collapsed state of sidebar sections */
  sidebarCollapsed: boolean;

  // Actions
  addRepo: (path: string) => void;
  removeRepo: (path: string) => void;
  setSelectedRepo: (path: string) => void;
  saveScrollPosition: (page: string, pos: number) => void;
  saveQuery: (page: string, query: string) => void;
  toggleSidebar: () => void;
}

export const useRepoStore = create<RepoStore>()(
  persist(
    (set) => ({
      repos: [],
      selectedRepo: "global",
      scrollPositions: {},
      skillsColumns: ["name", "version", "status", "registry", "pinned", "actions"],
      lastQueries: {},
      sidebarCollapsed: false,

      addRepo: (path) =>
        set((s) => ({
          repos: s.repos.includes(path) ? s.repos : [...s.repos, path],
          selectedRepo: path,
        })),

      removeRepo: (path) =>
        set((s) => {
          const repos = s.repos.filter((r) => r !== path);
          return {
            repos,
            selectedRepo:
              s.selectedRepo === path ? (repos[0] ?? "global") : s.selectedRepo,
          };
        }),

      setSelectedRepo: (path) => set({ selectedRepo: path }),

      saveScrollPosition: (page, pos) =>
        set((s) => ({
          scrollPositions: { ...s.scrollPositions, [page]: pos },
        })),

      saveQuery: (page, query) =>
        set((s) => ({
          lastQueries: { ...s.lastQueries, [page]: query },
        })),

      toggleSidebar: () =>
        set((s) => ({ sidebarCollapsed: !s.sidebarCollapsed })),
    }),
    {
      name: "skell-gui-repos",
      version: 1,
    }
  )
);

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
  loading: Record<string, boolean>;

  notify: (n: Omit<Notification, "id">) => void;
  dismissNotification: (id: string) => void;
  setLoading: (key: string, loading: boolean) => void;
  isLoading: (key: string) => boolean;
}

let _notifCounter = 0;

export const useUIStore = create<UIStore>()((set, get) => ({
  notifications: [],
  loading: {},

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

  setLoading: (key, loading) =>
    set((s) => ({ loading: { ...s.loading, [key]: loading } })),

  isLoading: (key) => !!get().loading[key],
}));
