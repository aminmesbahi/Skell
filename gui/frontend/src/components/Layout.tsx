import { Sidebar } from "./Sidebar";
import { NotificationToast } from "./NotificationToast";
import { Outlet } from "react-router-dom";

export function Layout() {
  return (
    <div className="flex h-screen overflow-hidden bg-[#0a0c14]">
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
