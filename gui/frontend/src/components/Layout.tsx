import { Sidebar } from "./Sidebar";
import { NotificationToast } from "./NotificationToast";
import { Outlet } from "react-router-dom";

// Detect macOS so we can render an invisible drag strip across the top of
// the window. With TitleBarHiddenInset() the OS does not provide a draggable
// title bar, so we add our own no-paint region that lets users move the
// window from anywhere along the top edge (not just the sidebar header).
const IS_MAC =
  typeof navigator !== "undefined" &&
  /Mac|iPhone|iPad|iPod/i.test(navigator.platform || navigator.userAgent || "");

export function Layout() {
  return (
    <div className="flex h-screen overflow-hidden bg-[#0a0c14] relative">
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
        <div className="flex-1 overflow-y-auto">
          <Outlet />
        </div>
      </main>
      <NotificationToast />
    </div>
  );
}
