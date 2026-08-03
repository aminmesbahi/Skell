import { NavLink, useNavigate } from "react-router";
import { SelectDirectory } from "../../bindings/skell-gui/app";
import {
  Home,
  FolderKanban,
  Library,
  Activity,
  Settings as SettingsIcon,
  PanelLeftClose,
  PanelLeft,
} from "lucide-react";
import { useRepoStore } from "@/store";
import clsx from "clsx";
import { isMac } from "@/lib/platform";
import { createProjectId } from "@/lib/navigation";
import { CurrentProjectSummary } from "./CurrentProjectSummary";
import { RecentProjects } from "./RecentProjects";

const NAV_ITEMS = [
  { to: "/", icon: Home, label: "Home" },
  { to: "/projects", icon: FolderKanban, label: "Projects" },
  { to: "/catalog", icon: Library, label: "Catalog" },
  { to: "/activity", icon: Activity, label: "Activity" },
  { to: "/settings", icon: SettingsIcon, label: "Settings" },
];

// IS_MAC used for macOS titlebar padding and no-drag regions on traffic lights.
const IS_MAC = isMac;

export function Sidebar() {
  const { addRepo, sidebarCollapsed, toggleSidebar, selectedRepo } = useRepoStore();
  const navigate = useNavigate();

  async function handleAddRepo() {
    const selected = await SelectDirectory();
    if (selected) {
      addRepo(selected);
      navigate("/projects");
    }
  }

  return (
    <aside
      className={clsx(
        "flex flex-col h-full bg-[#0a0d1a] border-r border-[#1a1f35] transition-all duration-200",
        sidebarCollapsed ? "w-14" : "w-56"
      )}
    >
      {/* Logo + collapse toggle. On macOS, the Wails `TitleBarHiddenInset`
          style overlays traffic-light buttons in the top-left, so we reserve
          space with `mac-titlebar-pad` and make the strip draggable. */}
      <div className="app-drag mac-titlebar-pad flex items-center justify-between px-3 py-4 border-b border-[#1a1f35]">
        {!sidebarCollapsed && (
          <div
            className="flex items-center gap-2"
            style={IS_MAC ? ({ "--wails-draggable": "no-drag" } as React.CSSProperties) : undefined}
          >
            <div className="w-7 h-7 rounded-lg bg-gradient-to-br from-blue-500 to-indigo-600 flex items-center justify-center text-white font-bold text-sm">
              S
            </div>
            <span className="font-semibold text-slate-200 text-sm">Skell</span>
          </div>
        )}
        <button
          onClick={toggleSidebar}
          className="app-no-drag p-1.5 rounded-lg text-slate-500 hover:text-slate-300 hover:bg-white/5 transition-colors ml-auto"
          title={sidebarCollapsed ? "Expand sidebar" : "Collapse sidebar"}
          style={IS_MAC ? ({ "--wails-draggable": "no-drag" } as React.CSSProperties) : undefined}
        >
          {sidebarCollapsed ? <PanelLeft size={16} /> : <PanelLeftClose size={16} />}
        </button>
      </div>

      <div className="flex-1 overflow-y-auto px-2 py-3">
        <nav className="space-y-1">
          {NAV_ITEMS.map(({ to, icon: Icon, label }) => (
            <NavLink
              key={to}
              to={to}
              end={to === "/"}
              className={({ isActive }) =>
                clsx(
                  "flex items-center gap-3 rounded-lg px-2 py-2 text-sm transition-colors",
                  isActive
                    ? "bg-brand-600/20 text-brand-400 font-medium"
                    : "text-slate-500 hover:text-slate-300 hover:bg-white/5"
                )
              }
              title={sidebarCollapsed ? label : undefined}
            >
              <Icon size={16} className="shrink-0" />
              {!sidebarCollapsed && <span>{label}</span>}
            </NavLink>
          ))}
        </nav>

        {selectedRepo && selectedRepo !== "global" && (
          <div className="mt-4">
            <CurrentProjectSummary
              projectPath={selectedRepo}
              collapsed={sidebarCollapsed}
              onOpen={() => navigate(`/projects/${createProjectId(selectedRepo)}/skills`)}
            />
          </div>
        )}

        <RecentProjects collapsed={sidebarCollapsed} onAddProject={handleAddRepo} />
      </div>

      <div className="border-t border-[#1a1f35] px-2 py-3">
        <button
          onClick={handleAddRepo}
          className={clsx(
            "flex items-center rounded-lg px-2 py-2 text-sm transition-colors",
            sidebarCollapsed ? "justify-center" : "gap-2",
            "text-slate-400 hover:bg-white/5 hover:text-slate-100"
          )}
          title={sidebarCollapsed ? "Add Project" : undefined}
        >
          <span className="flex h-7 w-7 items-center justify-center rounded-lg bg-white/5 text-slate-300">
            <span className="text-base leading-none">+</span>
          </span>
          {!sidebarCollapsed && <span>Add Project</span>}
        </button>
      </div>
    </aside>
  );
}
