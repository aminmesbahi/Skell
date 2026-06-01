import { CheckCircle2, AlertTriangle, XCircle } from "lucide-react";
import type { SkillValidationResult, SkillValidationFinding, SkillAnalysis } from "@/lib/types";
import { SeverityBadge } from "@/components/Badges";

/**
 * Renders a single skill's validation findings and offline analysis metrics.
 * Shared by the per-skill detail tab and the per-project bulk view.
 */
export function ValidationReport({ result }: { result: SkillValidationResult }) {
  const ok = result.errors === 0 && result.warnings === 0;
  return (
    <div className="space-y-4">
      {/* Summary */}
      <div className="flex items-center gap-2">
        {result.errors > 0 ? (
          <XCircle size={16} className="text-red-400" />
        ) : result.warnings > 0 ? (
          <AlertTriangle size={16} className="text-amber-400" />
        ) : (
          <CheckCircle2 size={16} className="text-emerald-400" />
        )}
        <span className="text-sm text-slate-300">
          {ok
            ? "Passed all checks"
            : `${result.errors} error${result.errors !== 1 ? "s" : ""}, ${result.warnings} warning${
                result.warnings !== 1 ? "s" : ""
              }`}
        </span>
      </div>

      {/* Findings */}
      {result.findings.length > 0 && (
        <div className="space-y-1.5">
          {result.findings.map((f, i) => (
            <FindingRow key={i} finding={f} />
          ))}
        </div>
      )}

      {/* Analysis metrics */}
      {result.analysis && <AnalysisGrid analysis={result.analysis} />}
    </div>
  );
}

function FindingRow({ finding }: { finding: SkillValidationFinding }) {
  const loc = finding.file
    ? finding.line && finding.line > 0
      ? `${finding.file}:${finding.line}`
      : finding.file
    : "";
  return (
    <div
      className={`rounded-lg border px-3 py-2 flex items-start gap-2 text-sm ${
        finding.severity === "error"
          ? "bg-red-500/5 border-red-500/20"
          : finding.severity === "warning"
          ? "bg-amber-500/5 border-amber-500/20"
          : "bg-blue-500/5 border-blue-500/20"
      }`}
    >
      <SeverityBadge severity={finding.severity as "error" | "warning" | "info"} />
      <span className="text-xs font-mono text-slate-500 shrink-0">{finding.category}</span>
      <span className="text-slate-300 flex-1">{finding.message}</span>
      {loc && <span className="text-xs font-mono text-slate-600 shrink-0">{loc}</span>}
    </div>
  );
}

function AnalysisGrid({ analysis }: { analysis: SkillAnalysis }) {
  const pct = (v: number) => `${Math.round(v * 100)}%`;
  return (
    <div className="space-y-3">
      {/* Tokens */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
        <Metric label="Total tokens" value={String(analysis.total_tokens)} />
        {analysis.skill_tokens > 0 && (
          <Metric label="SKILL.md tokens" value={String(analysis.skill_tokens)} />
        )}
      </div>

      {analysis.has_content && (
        <div>
          <p className="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1.5">
            Content quality
          </p>
          <div className="grid grid-cols-2 sm:grid-cols-3 gap-2">
            <Metric label="Words" value={String(analysis.word_count)} />
            <Metric label="Info density" value={pct(analysis.information_density)} />
            <Metric label="Imperative" value={pct(analysis.imperative_ratio)} />
            <Metric label="Code ratio" value={pct(analysis.code_block_ratio)} />
            <Metric label="Sections" value={String(analysis.section_count)} />
            <Metric label="List items" value={String(analysis.list_item_count)} />
          </div>
        </div>
      )}

      {analysis.has_contamination && (
        <div>
          <p className="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1.5">
            Contamination
          </p>
          <div className="flex items-center gap-2 flex-wrap">
            <span
              className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium border ${
                analysis.contamination_level === "high"
                  ? "text-red-400 bg-red-500/10 border-red-500/20"
                  : analysis.contamination_level === "medium"
                  ? "text-amber-400 bg-amber-500/10 border-amber-500/20"
                  : "text-emerald-400 bg-emerald-500/10 border-emerald-500/20"
              }`}
            >
              {analysis.contamination_level || "low"} ({analysis.contamination_score.toFixed(2)})
            </span>
            {analysis.code_languages && analysis.code_languages.length > 0 && (
              <span className="text-xs text-slate-500">
                languages: {analysis.code_languages.join(", ")}
              </span>
            )}
            {analysis.mismatched_categories && analysis.mismatched_categories.length > 0 && (
              <span className="text-xs text-amber-400">
                mismatched: {analysis.mismatched_categories.join(", ")}
              </span>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg bg-white/[0.02] border border-[#1e2540] px-3 py-2">
      <p className="text-xs text-slate-600">{label}</p>
      <p className="text-sm font-mono text-slate-300">{value}</p>
    </div>
  );
}
