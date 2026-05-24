import { useId, useState, type ReactNode } from "react";
import { ChevronDown } from "lucide-react";

interface CollapsibleSectionProps {
  title: string;
  count?: number;
  defaultOpen?: boolean;
  children: ReactNode;
}

export function CollapsibleSection({
  title,
  count,
  defaultOpen = true,
  children,
}: CollapsibleSectionProps) {
  const [open, setOpen] = useState(defaultOpen);
  const panelId = useId();

  return (
    <div>
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        aria-expanded={open}
        aria-controls={panelId}
        className="flex items-center gap-2 w-full text-left mb-3 group"
      >
        <ChevronDown
          size={14}
          className={`text-slate-500 transition-transform group-hover:text-slate-300 ${
            open ? "" : "-rotate-90"
          }`}
        />
        <h3 className="text-xs font-semibold text-slate-500 uppercase tracking-wider group-hover:text-slate-300 transition-colors truncate">
          {title}
        </h3>
        {typeof count === "number" && (
          <span className="text-[10px] font-medium px-1.5 py-0.5 rounded-full bg-slate-700/50 text-slate-400">
            {count}
          </span>
        )}
      </button>
      <div id={panelId} role="region" hidden={!open}>
        {open && children}
      </div>
    </div>
  );
}
