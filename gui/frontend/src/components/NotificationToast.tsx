import { useState } from "react";
import { X, CheckCircle, AlertCircle, Info } from "lucide-react";
import { useUIStore } from "@/store";
import clsx from "clsx";

export function NotificationToast() {
  const { notifications, dismissNotification } = useUIStore();

  const [copyMessage, setCopyMessage] = useState("");

  return (
    <div aria-live="polite" aria-relevant="additions" className="fixed bottom-4 right-4 z-50 flex flex-col gap-2 max-w-sm w-full">
      {copyMessage && <p role="status" className="card text-xs">{copyMessage}</p>}
      {notifications.map((n) => (
        <div
          key={n.id} role={n.kind === "error" ? "alert" : "status"}
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
            {n.detail && (n.kind === "error" ? <details className="text-xs mt-1"><summary>Technical details</summary><pre className="whitespace-pre-wrap max-h-40 overflow-auto break-words">{n.detail}</pre><button className="btn-ghost text-xs" onClick={async () => { try { await navigator.clipboard.writeText(`${n.title}\n${n.detail}`); setCopyMessage("Details copied"); } catch { setCopyMessage("Could not copy. Select the details to copy manually."); } }}>Copy details</button></details> : <p className="text-xs opacity-75 mt-0.5 break-words">{n.detail}</p>)}
          </div>
          <button
            aria-label="Dismiss notification" onClick={() => dismissNotification(n.id)}
            className="shrink-0 opacity-60 hover:opacity-100 transition-opacity"
          >
            <X size={14} />
          </button>
        </div>
      ))}
    </div>
  );
}
