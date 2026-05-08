import { NavLink, useNavigate } from "react-router-dom";
import { SelectDirectory } from "../../wailsjs/go/main/App";
import {
  LayoutDashboard,
  FolderOpen,
  Package,
  Search,
  RefreshCw,
  Stethoscope,
  Database,
  ScrollText,
  Settings,
  Plus,
  Globe,
  FolderClosed,
  PanelLeftClose,
  PanelLeft,
  GitPullRequest,
} from "lucide-react";
import { useRepoStore } from "@/store";
import clsx from "clsx";

const NAV_ITEMS = [
  { to: "/", icon: LayoutDashboard, label: "Dashboard" },
  { to: "/repositories", icon: FolderOpen, label: "Projects" },
  { to: "/skills", icon: Package, label: "My Skills" },
  { to: "/registry", icon: Search, label: "Discover Skills" },
  { to: "/sync", icon: RefreshCw, label: "Sync" },
  { to: "/doctor", icon: Stethoscope, label: "Doctor" },
  { to: "/cache", icon: Database, label: "Cache" },
  { to: "/audit", icon: ScrollText, label: "Audit Log" },
  { to: "/settings", icon: Settings, label: "Settings" },
];

const CONTRIBUTE_ITEM = { to: "/contribute-info", icon: GitPullRequest, label: "Contribute" };

// Detect macOS so we can leave room for the traffic-light buttons that
// Wails renders on top of the window when using TitleBarHiddenInset().
const IS_MAC =
  typeof navigator !== "undefined" &&
  /Mac|iPhone|iPad|iPod/i.test(navigator.platform || navigator.userAgent || "");

export function Sidebar() {
  const { repos, selectedRepo, setSelectedRepo, addRepo, sidebarCollapsed, toggleSidebar } =
    useRepoStore();
  const navigate = useNavigate();

  async function handleAddRepo() {
    const selected = await SelectDirectory();
    if (selected) {
      addRepo(selected);
      navigate("/skills");
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

      {/* Navigation */}
      <nav className="flex-1 px-2 py-3 space-y-0.5 overflow-y-auto">
        {NAV_ITEMS.map(({ to, icon: Icon, label }) => (
          <NavLink
            key={to}
            to={to}
            end={to === "/"}
            className={({ isActive }) =>
              clsx(
                "flex items-center gap-3 px-2 py-2 rounded-lg text-sm transition-colors",
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

        {/* Contribute section — visually separated */}
        <div className="pt-2 mt-2 border-t border-[#1a1f35]">
          {!sidebarCollapsed && (
            <p className="px-2 pb-1 text-xs font-semibold text-slate-600 uppercase tracking-wider">
              Community
            </p>
          )}
          <NavLink
            to={CONTRIBUTE_ITEM.to}
            className={() =>
              clsx(
                "flex items-center gap-3 px-2 py-2 rounded-lg text-sm transition-colors",
                "text-indigo-500 hover:text-indigo-300 hover:bg-indigo-500/10"
              )
            }
            title={sidebarCollapsed ? CONTRIBUTE_ITEM.label : undefined}
          >
            <CONTRIBUTE_ITEM.icon size={16} className="shrink-0" />
            {!sidebarCollapsed && <span>{CONTRIBUTE_ITEM.label}</span>}
          </NavLink>
        </div>
      </nav>

      {/* Project Context Switcher - now more prominent */}
      {!sidebarCollapsed && (
        <div className="border-t border-[#1a1f35] px-2 py-3">
          <div className="flex items-center justify-between px-2 mb-1.5">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider flex items-center gap-1">
              Your Projects
            </span>
            <button
              onClick={handleAddRepo}
              className="p-1 rounded text-slate-500 hover:text-brand-400 hover:bg-brand-600/10 transition-colors"
              title="Add another project folder"
            >
              <Plus size={14} />
            </button>
          </div>
          <p className="px-2 text-[10px] text-slate-600 mb-1.5">Skills are installed into the selected project</p>

          {/* Global entry */}
          <button
            onClick={() => setSelectedRepo("global")}
            className={clsx(
              "w-full flex items-center justify-between px-2 py-1.5 rounded-lg text-xs transition-colors text-left group",
              selectedRepo === "global"
                ? "bg-indigo-600/25 text-indigo-400 ring-1 ring-indigo-500/30"
                : "text-slate-500 hover:text-slate-300 hover:bg-white/5"
            )}
          >
            <div className="flex items-center gap-2 truncate">
              <Globe size={12} className="shrink-0" />
              <span>Global</span>
            </div>
            {selectedRepo === "global" && <span className="text-[10px] opacity-70">active</span>}
          </button>

          {/* Local repos */}
          {repos.map((repo) => {
            const short = repo.split(/[/\\]/).at(-1) ?? repo;
            const isActive = selectedRepo === repo;
            return (
              <button
                key={repo}
                onClick={() => setSelectedRepo(repo)}
                className={clsx(
                  "w-full flex items-center justify-between px-2 py-1.5 rounded-lg text-xs transition-colors text-left group",
                  isActive
                    ? "bg-teal-600/25 text-teal-400 ring-1 ring-teal-500/30"
                    : "text-slate-500 hover:text-slate-300 hover:bg-white/5"
                )}
                title={repo}
              >
                <div className="flex items-center gap-2 truncate">
                  <FolderClosed size={12} className="shrink-0" />
                  <span>{short}</span>
                </div>
                {isActive && <span className="text-[10px] opacity-75">current</span>}
              </button>
            );
          })}

          {repos.length === 0 && (
            <p className="text-xs text-slate-700 px-2 py-1">No projects added yet</p>
          )}
        </div>
      )}
    </aside>
  );
}
