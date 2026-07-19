import { memo, useId, useState, type ReactNode } from "react";
import { ChevronDown } from "lucide-react";

interface CollapsibleSectionProps {
  title: ReactNode;
  count?: number;
  defaultOpen?: boolean;
  children: ReactNode;
}

export const CollapsibleSection = memo(function CollapsibleSection({
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
          className={`text-slate-500 shrink-0 transition-transform group-hover:text-slate-300 ${
            open ? "" : "-rotate-90"
          }`}
        />
        <div className="flex-1 min-w-0 flex items-center gap-2 text-slate-400 group-hover:text-slate-300 transition-colors">
          {title}
        </div>
        {typeof count === "number" && (
          <span className="text-[10px] font-medium px-1.5 py-0.5 rounded-full bg-slate-700/50 text-slate-400 shrink-0">
            {count}
          </span>
        )}
      </button>
      <div id={panelId} role="region" hidden={!open}>
        {open && children}
      </div>
    </div>
  );
});
