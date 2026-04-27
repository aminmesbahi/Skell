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
  ChevronDown,
  Globe,
  FolderClosed,
  PanelLeftClose,
  PanelLeft,
} from "lucide-react";
import { useRepoStore } from "@/store";
import clsx from "clsx";

const NAV_ITEMS = [
  { to: "/", icon: LayoutDashboard, label: "Dashboard" },
  { to: "/repositories", icon: FolderOpen, label: "Repositories" },
  { to: "/skills", icon: Package, label: "Installed Skills" },
  { to: "/registry", icon: Search, label: "Registry" },
  { to: "/sync", icon: RefreshCw, label: "Sync" },
  { to: "/doctor", icon: Stethoscope, label: "Doctor" },
  { to: "/cache", icon: Database, label: "Cache" },
  { to: "/audit", icon: ScrollText, label: "Audit Log" },
  { to: "/settings", icon: Settings, label: "Settings" },
];

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
          <div className="flex items-center gap-2">
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
      </nav>

      {/* Repo switcher */}
      {!sidebarCollapsed && (
        <div className="border-t border-[#1a1f35] px-2 py-3">
          <div className="flex items-center justify-between px-2 mb-2">
            <span className="text-xs font-semibold text-slate-600 uppercase tracking-wider flex items-center gap-1">
              <ChevronDown size={12} />
              Repos
            </span>
            <button
              onClick={handleAddRepo}
              className="p-1 rounded text-slate-600 hover:text-brand-400 hover:bg-brand-600/10 transition-colors"
              title="Add repository"
            >
              <Plus size={14} />
            </button>
          </div>

          {/* Global entry */}
          <button
            onClick={() => setSelectedRepo("global")}
            className={clsx(
              "w-full flex items-center gap-2 px-2 py-1.5 rounded-lg text-xs transition-colors text-left",
              selectedRepo === "global"
                ? "bg-indigo-600/20 text-indigo-400"
                : "text-slate-500 hover:text-slate-300 hover:bg-white/5"
            )}
          >
            <Globe size={12} className="shrink-0" />
            <span className="truncate">Global</span>
          </button>

          {/* Local repos */}
          {repos.map((repo) => {
            const short = repo.split(/[/\\]/).at(-1) ?? repo;
            return (
              <button
                key={repo}
                onClick={() => setSelectedRepo(repo)}
                className={clsx(
                  "w-full flex items-center gap-2 px-2 py-1.5 rounded-lg text-xs transition-colors text-left",
                  selectedRepo === repo
                    ? "bg-teal-600/20 text-teal-400"
                    : "text-slate-500 hover:text-slate-300 hover:bg-white/5"
                )}
                title={repo}
              >
                <FolderClosed size={12} className="shrink-0" />
                <span className="truncate">{short}</span>
              </button>
            );
          })}

          {repos.length === 0 && (
            <p className="text-xs text-slate-700 px-2 py-1">No repos added yet</p>
          )}
        </div>
      )}
    </aside>
  );
}
