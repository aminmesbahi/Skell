import { X, CheckCircle, AlertCircle, Info } from "lucide-react";
import { useUIStore } from "@/store";
import clsx from "clsx";

export function NotificationToast() {
  const { notifications, dismissNotification } = useUIStore();

  if (notifications.length === 0) return null;

  return (
    <div className="fixed bottom-4 right-4 z-50 flex flex-col gap-2 max-w-sm w-full">
      {notifications.map((n) => (
        <div
          key={n.id}
          className={clsx(
            "flex items-start gap-3 p-4 rounded-xl border shadow-xl backdrop-blur-sm animate-in slide-in-from-bottom-2",
            n.kind === "success" && "bg-emerald-950/90 border-emerald-800 text-emerald-200",
            n.kind === "error" && "bg-red-950/90 border-red-800 text-red-200",
            n.kind === "info" && "bg-blue-950/90 border-blue-800 text-blue-200"
          )}
        >
          {n.kind === "success" && <CheckCircle size={18} className="shrink-0 mt-0.5 text-emerald-400" />}
          {n.kind === "error" && <AlertCircle size={18} className="shrink-0 mt-0.5 text-red-400" />}
          {n.kind === "info" && <Info size={18} className="shrink-0 mt-0.5 text-blue-400" />}
          <div className="flex-1 min-w-0">
            <p className="font-medium text-sm">{n.title}</p>
            {n.detail && <p className="text-xs opacity-75 mt-0.5 break-words">{n.detail}</p>}
          </div>
          <button
            onClick={() => dismissNotification(n.id)}
            className="shrink-0 opacity-60 hover:opacity-100 transition-opacity"
          >
            <X size={14} />
          </button>
        </div>
      ))}
    </div>
  );
}
