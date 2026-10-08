import type { SkillValidationResult } from "@/lib/types";

export function ValidationBadge({ result, loading = false }: { result?: SkillValidationResult; loading?: boolean }) {
  const label = loading ? "Checking…" : !result ? "Not checked" : result.errors ? `${result.errors} validation error${result.errors === 1 ? "" : "s"}` : result.warnings ? `${result.warnings} validation warning${result.warnings === 1 ? "" : "s"}` : "Validation passed";
  return <span className={`text-xs ${!result ? "text-slate-400" : result.errors ? "text-red-400" : result.warnings ? "text-amber-400" : "text-emerald-400"}`} title={result?.findings.map((f) => f.message).join("\n") || "Offline skill validation; not an agent compatibility guarantee."}>{label}</span>;
}
