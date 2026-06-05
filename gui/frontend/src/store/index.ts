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
  /** Collapsed state of sidebar sections */
  sidebarCollapsed: boolean;

  // Actions
  addRepo: (path: string) => void;
  removeRepo: (path: string) => void;
  setSelectedRepo: (path: string) => void;
  toggleSidebar: () => void;
}

export const useRepoStore = create<RepoStore>()(
  persist(
    (set) => ({
      repos: [],
      selectedRepo: "global",
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
