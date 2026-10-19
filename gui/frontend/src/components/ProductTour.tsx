import { useCallback, useEffect, useLayoutEffect, useRef, useState } from "react";
import { useRepoStore, useUIStore } from "@/store";
import { TOUR_STEPS, type TourStep } from "@/lib/tour";
import { isMac } from "@/lib/platform";

const CARD_WIDTH = 320;
const GAP = 14;
const PAD = 6;

interface Placement {
  rect: DOMRect;
  top: number;
  left: number;
}

function place(rect: DOMRect, cardHeight: number): Placement {
  const vw = window.innerWidth;
  const vh = window.innerHeight;
  const clamp = (value: number, min: number, max: number) => Math.max(min, Math.min(value, max));
  // Prefer beside the target (sidebar items), else below, else above.
  if (rect.right + GAP + CARD_WIDTH <= vw - 8) {
    return { rect, left: rect.right + GAP, top: clamp(rect.top, 8, vh - cardHeight - 8) };
  }
  const left = clamp(rect.left + rect.width / 2 - CARD_WIDTH / 2, 8, vw - CARD_WIDTH - 8);
  if (rect.bottom + GAP + cardHeight <= vh - 8) return { rect, left, top: rect.bottom + GAP };
  return { rect, left, top: Math.max(8, rect.top - GAP - cardHeight) };
}

/**
 * A short coach-mark tour: dims the app, spotlights one element at a time and
 * explains it. Steps whose element isn't on the current page are skipped.
 * Driven by useUIStore().tourOpen; finishing or skipping marks it as seen.
 */
export function ProductTour() {
  const open = useUIStore((s) => s.tourOpen);
  const endTour = useUIStore((s) => s.endTour);
  const setTourSeen = useRepoStore((s) => s.setTourSeen);
  const [steps, setSteps] = useState<TourStep[]>([]);
  const [index, setIndex] = useState(0);
  const [placement, setPlacement] = useState<Placement | null>(null);
  const card = useRef<HTMLDivElement>(null);
  const next = useRef<HTMLButtonElement>(null);

  const finish = useCallback(() => {
    setTourSeen(true);
    endTour();
  }, [endTour, setTourSeen]);

  useEffect(() => {
    if (!open) { setSteps([]); setPlacement(null); return; }
    const available = TOUR_STEPS.filter((s) => document.querySelector(s.target));
    if (available.length === 0) { finish(); return; }
    const previous = document.activeElement as HTMLElement | null;
    setSteps(available);
    setIndex(0);
    return () => previous?.focus?.();
  }, [open, finish]);

  const step = steps[index];
  const last = index === steps.length - 1;

  useLayoutEffect(() => {
    if (!open || !step) return;
    const element = document.querySelector(step.target);
    if (!element) { setIndex((i) => Math.min(i + 1, steps.length - 1)); return; }
    element.scrollIntoView?.({ block: "nearest" });
    const measure = () => setPlacement(place(element.getBoundingClientRect(), card.current?.offsetHeight ?? 180));
    measure();
    window.addEventListener("resize", measure);
    window.addEventListener("scroll", measure, true);
    return () => {
      window.removeEventListener("resize", measure);
      window.removeEventListener("scroll", measure, true);
    };
  }, [open, step, steps.length]);

  useEffect(() => { if (open && step) next.current?.focus(); }, [open, step]);

  if (!open || !step) return null;

  const go = (delta: number) => {
    const target = index + delta;
    if (target < 0) return;
    if (target >= steps.length) finish();
    else setIndex(target);
  };

  function onKeyDown(event: React.KeyboardEvent) {
    if (event.key === "Escape") { event.stopPropagation(); finish(); }
    else if (event.key === "ArrowRight") go(1);
    else if (event.key === "ArrowLeft") go(-1);
    else if (event.key === "Tab") {
      const buttons = Array.from(card.current?.querySelectorAll<HTMLButtonElement>("button") ?? []);
      if (!buttons.length) return;
      const current = buttons.indexOf(document.activeElement as HTMLButtonElement);
      const nextIndex = event.shiftKey ? (current <= 0 ? buttons.length - 1 : current - 1) : (current === buttons.length - 1 ? 0 : current + 1);
      event.preventDefault();
      buttons[nextIndex].focus();
    }
  }

  const text = step.id === "search" && isMac ? step.text.replace("Ctrl+K", "⌘K") : step.text;

  return (
    <div className="fixed inset-0 z-[60]" onKeyDown={onKeyDown}>
      {/* Blocks clicks on the app underneath while the tour runs. */}
      <div className="absolute inset-0" aria-hidden="true" />
      {placement && (
        <div
          aria-hidden="true"
          className="pointer-events-none absolute rounded-xl ring-2 ring-brand-400"
          style={{
            top: placement.rect.top - PAD,
            left: placement.rect.left - PAD,
            width: placement.rect.width + PAD * 2,
            height: placement.rect.height + PAD * 2,
            boxShadow: "0 0 0 9999px rgba(8, 10, 20, 0.62)",
          }}
        />
      )}
      <div
        ref={card}
        role="dialog"
        aria-modal="true"
        aria-label="Product tour"
        tabIndex={-1}
        className="absolute rounded-xl border border-[var(--palette-26314d)] bg-[var(--palette-11162a)] p-4 shadow-2xl"
        style={{ width: CARD_WIDTH, top: placement?.top ?? 80, left: placement?.left ?? 80, opacity: placement ? 1 : 0 }}
      >
        <p className="text-[11px] font-medium uppercase tracking-wide text-slate-500">Step {index + 1} of {steps.length}</p>
        <h2 className="mt-1 text-sm font-semibold text-slate-100">{step.title}</h2>
        <p className="mt-1.5 text-xs leading-5 text-slate-300" aria-live="polite">{text}</p>
        <div className="mt-4 flex items-center justify-between gap-2">
          {last ? <span /> : <button type="button" className="btn-subtle" onClick={finish}>Skip tour</button>}
          <div className="flex gap-2">
            {index > 0 && <button type="button" className="btn-subtle" onClick={() => go(-1)}>Back</button>}
            <button ref={next} type="button" className="btn-primary h-8 px-3 text-xs" onClick={() => go(1)}>{last ? "Finish" : "Next"}</button>
          </div>
        </div>
      </div>
    </div>
  );
}
