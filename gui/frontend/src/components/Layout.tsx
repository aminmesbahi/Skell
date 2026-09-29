import { useEffect } from "react";
import { useNavigate, useLocation } from "react-router";
import { ThemeControl } from "./ThemeControl";
import { Sidebar } from "./Sidebar";
import { NotificationToast } from "./NotificationToast";
import { Outlet } from "react-router";
import { isMac } from "@/lib/platform";

// IS_MAC is true on macOS (and iOS devices) so we can reserve space for the
// traffic-light buttons under TitleBarHiddenInset() and provide a drag strip.
const IS_MAC = isMac;

export function Layout() {
  const navigate = useNavigate();
  const location = useLocation();
  useEffect(() => {
    const focusSearch = () => document.querySelector<HTMLInputElement>('main [data-search], main input[placeholder*="Search"]')?.focus();
    if ((location.state as { focusSearch?: boolean } | null)?.focusSearch) focusSearch();
    const handle = (event: KeyboardEvent) => {
      if (event.isComposing || document.querySelector('[role="dialog"]')) return;
      const element = event.target as HTMLElement;
      const typing = element.matches("input, textarea, select") || element.isContentEditable;
      if (((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === "k") || (!typing && !event.ctrlKey && !event.metaKey && !event.altKey && event.key === "/")) {
        event.preventDefault();
        if (document.querySelector('main [data-search], main input[placeholder*="Search"]')) focusSearch();
        else navigate("/catalog", { state: { focusSearch: true } });
      }
    };
    window.addEventListener("keydown", handle);
    return () => window.removeEventListener("keydown", handle);
  }, [navigate, location]);
  return (
    <div className="flex h-screen overflow-hidden bg-[var(--palette-0a0c14)] relative">
      {IS_MAC && (
        <div
          aria-hidden
          className="fixed top-0 left-0 right-0 h-7 z-50"
          style={{ "--wails-draggable": "drag" } as React.CSSProperties}
        />
      )}
      <Sidebar />
      <main className="flex-1 flex flex-col overflow-hidden">
        {/* Draggable strip across the top of the main pane (macOS only via
            CSS); zero-height on other platforms. Lets users drag the window
            from anywhere along the top, not just the sidebar header. */}
        <div className="app-drag mac-titlebar-strip shrink-0" />
        <div className="flex items-center justify-between border-b border-[var(--palette-1a1f35)] bg-[var(--palette-0f1221)]/85 px-6 py-2.5 backdrop-blur">
          <div className="text-xs font-medium uppercase tracking-[0.14em] text-slate-600">Workspace</div>
          <div className="flex items-center gap-2">
            <button
              className="btn-ghost h-9 text-xs"
              onClick={() => navigate("/catalog", { state: { focusSearch: true } })}
            >
              Search · Ctrl/Cmd+K
            </button>
            <ThemeControl />
          </div>
        </div>
        <div className="flex-1 overflow-y-auto">
          <Outlet />
        </div>
      </main>
      <NotificationToast />
    </div>
  );
}
