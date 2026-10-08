import { useState } from "react";
import { Check, Plus, Sparkles } from "lucide-react";
import { useAsync } from "@/hooks/useAsync";
import { addSource, addSkillSource, listCatalog } from "@/lib/skell";
import { useUIStore } from "@/store";
import type { CatalogSource } from "@/lib/types";

interface FeaturedSourcesProps {
  /** Project to add the source to; when empty it is added as a shared source. */
  repo?: string;
  /** URLs already configured, shown as added. */
  configuredUrls?: string[];
  onAdded?: () => void;
  compact?: boolean;
}

function normalize(url: string): string {
  return url.trim().toLowerCase().replace(/\/+$/, "").replace(/\.git$/, "");
}

/** Well-known skill sources with one-click "Add". */
export function FeaturedSources({ repo, configuredUrls = [], onAdded, compact }: FeaturedSourcesProps) {
  const { notify } = useUIStore();
  const catalog = useAsync("catalog", () => listCatalog());
  const [adding, setAdding] = useState<string | null>(null);
  const [added, setAdded] = useState<string[]>([]);
  const configured = new Set(configuredUrls.map(normalize));
  const sources: CatalogSource[] = catalog.data ?? [];
  if (!sources.length) return null;

  async function handleAdd(source: CatalogSource) {
    setAdding(source.id);
    try {
      if (repo && repo !== "global") {
        await addSource({ source: source.id, repo, alias: source.id });
      } else {
        await addSkillSource(source.id, source.url);
      }
      setAdded((prev) => [...prev, source.id]);
      notify({ kind: "success", title: `Added ${source.name}`, detail: "Its skills now appear in the catalog." });
      onAdded?.();
    } catch (e) {
      notify({ kind: "error", title: `Could not add ${source.name}`, detail: String(e) });
    } finally {
      setAdding(null);
    }
  }

  return (
    <section className="card" aria-labelledby="featured-sources-title">
      <div className="mb-3 flex items-center gap-2">
        <Sparkles size={14} className="text-brand-400" />
        <h2 id="featured-sources-title" className="text-sm font-semibold text-slate-200">Featured skill sources</h2>
        <span className="text-xs text-slate-500">{repo && repo !== "global" ? "added to this project" : "added as shared sources"}</span>
      </div>
      <div className={compact ? "grid gap-2 sm:grid-cols-2" : "grid gap-3 md:grid-cols-2"}>
        {sources.map((s) => {
          const isAdded = added.includes(s.id) || configured.has(normalize(s.url));
          return (
            <div key={s.id} className="flex items-start justify-between gap-3 rounded-xl border border-[var(--palette-1e2640)] bg-[var(--palette-0f1225)] p-3">
              <div className="min-w-0">
                <p className="truncate text-sm font-medium text-slate-200">{s.name}</p>
                {!compact && <p className="mt-0.5 line-clamp-2 text-xs text-slate-400">{s.description}</p>}
                <p className="mt-1 truncate font-mono text-[11px] text-slate-500">{s.url.replace(/^https:\/\//, "")}</p>
              </div>
              <button
                className={isAdded ? "btn-ghost shrink-0 text-xs" : "btn-primary shrink-0 text-xs"}
                disabled={isAdded || adding !== null}
                onClick={() => void handleAdd(s)}
                aria-label={isAdded ? `${s.name} added` : `Add ${s.name}`}
              >
                {isAdded ? <Check size={13} /> : <Plus size={13} />}
                {isAdded ? "Added" : adding === s.id ? "Adding…" : "Add"}
              </button>
            </div>
          );
        })}
      </div>
    </section>
  );
}
