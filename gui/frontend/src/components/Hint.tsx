import { useEffect, useId, useRef, useState, type ReactNode } from "react";
import { CircleHelp } from "lucide-react";
import { GLOSSARY, type GlossaryTerm } from "@/lib/glossary";

const WIDTH = 260;
const MARGIN = 8;

/**
 * A small "?" button that explains a Skell term in one sentence. Pass children
 * to show the label next to it (`<Hint term="source">Source</Hint>`).
 * The popover is position: fixed so scrolling tables and cards never clip it.
 */
export function Hint({ term, children }: { term: GlossaryTerm; children?: ReactNode }) {
  const entry = GLOSSARY[term];
  const [pos, setPos] = useState<{ top: number; left: number } | null>(null);
  const button = useRef<HTMLButtonElement>(null);
  const id = useId();
  const open = pos !== null;

  function show() {
    const rect = button.current?.getBoundingClientRect();
    if (!rect) return;
    const left = Math.min(Math.max(MARGIN, rect.left + rect.width / 2 - WIDTH / 2), window.innerWidth - WIDTH - MARGIN);
    setPos({ top: rect.bottom + 6, left });
  }

  useEffect(() => {
    if (!open) return;
    const close = () => setPos(null);
    const onKey = (event: KeyboardEvent) => { if (event.key === "Escape") close(); };
    const onPointer = (event: PointerEvent) => { if (!button.current?.contains(event.target as Node)) close(); };
    document.addEventListener("keydown", onKey);
    document.addEventListener("pointerdown", onPointer);
    window.addEventListener("scroll", close, true);
    return () => {
      document.removeEventListener("keydown", onKey);
      document.removeEventListener("pointerdown", onPointer);
      window.removeEventListener("scroll", close, true);
    };
  }, [open]);

  return (
    <span className="inline-flex items-center gap-1">
      {children}
      <button
        ref={button}
        type="button"
        aria-label={`What is ${entry.label.toLowerCase()}?`}
        aria-describedby={open ? id : undefined}
        aria-expanded={open}
        onClick={() => (open ? setPos(null) : show())}
        onMouseEnter={show}
        onMouseLeave={() => setPos(null)}
        onBlur={() => setPos(null)}
        className="inline-flex shrink-0 rounded-full text-slate-500 hover:text-slate-300 focus:text-slate-300 focus:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/50"
      >
        <CircleHelp size={13} aria-hidden="true" />
      </button>
      {open && (
        <span
          id={id}
          role="tooltip"
          style={{ position: "fixed", top: pos.top, left: pos.left, width: WIDTH }}
          className="z-50 rounded-lg border border-[var(--palette-26314d)] bg-[var(--palette-11162a)] p-3 text-xs font-normal normal-case leading-5 tracking-normal text-slate-300 shadow-2xl"
        >
          <span className="mb-0.5 block font-semibold text-slate-100">{entry.label}</span>
          {entry.text}
        </span>
      )}
    </span>
  );
}
