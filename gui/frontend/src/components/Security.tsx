import { memo, useState } from "react";
import { BadgeCheck, FileSignature, KeyRound, ShieldAlert, ShieldCheck, ShieldX } from "lucide-react";
import clsx from "clsx";
import type { FindingSeverity, RiskLevel, ScanReport, SignatureStatus } from "@/lib/types";
import { riskAdvice, signatureLabel, sortedFindings } from "@/lib/security";

const RISK_STYLE: Record<RiskLevel, { label: string; classes: string }> = {
  clean: { label: "Clean", classes: "bg-emerald-500/15 text-emerald-400 border-emerald-500/30" },
  low: { label: "Low risk", classes: "bg-sky-500/15 text-sky-400 border-sky-500/30" },
  medium: { label: "Medium risk", classes: "bg-amber-500/15 text-amber-400 border-amber-500/30" },
  high: { label: "High risk", classes: "bg-orange-500/15 text-orange-400 border-orange-500/30" },
  critical: { label: "Critical", classes: "bg-red-500/15 text-red-400 border-red-500/30" },
};

/** Risk level of a skill from the static security scan. */
export const RiskBadge = memo(function RiskBadge({ level, score, size = "md" }: { level: RiskLevel; score?: number; size?: "sm" | "md" }) {
  const cfg = RISK_STYLE[level] ?? RISK_STYLE.clean;
  const Icon = level === "clean" || level === "low" ? ShieldCheck : ShieldAlert;
  return (
    <span
      title={`${cfg.label}${score ? ` (score ${score}/100)` : ""}. ${riskAdvice(level)}`}
      className={clsx("inline-flex items-center gap-1 whitespace-nowrap rounded-full border font-medium", cfg.classes, size === "sm" ? "px-1.5 py-0.5 text-xs" : "px-2 py-0.5 text-xs")}
    >
      <Icon size={size === "sm" ? 10 : 11} />
      {cfg.label}
      {score ? <span className="opacity-70">· {score}</span> : null}
    </span>
  );
});

const SIGNATURE_STYLE: Record<SignatureStatus, { classes: string; Icon: React.ElementType }> = {
  trusted: { classes: "bg-emerald-500/15 text-emerald-400 border-emerald-500/30", Icon: BadgeCheck },
  "untrusted-key": { classes: "bg-amber-500/15 text-amber-400 border-amber-500/30", Icon: KeyRound },
  tampered: { classes: "bg-red-500/15 text-red-400 border-red-500/30", Icon: ShieldX },
  invalid: { classes: "bg-red-500/15 text-red-400 border-red-500/30", Icon: ShieldX },
  unsigned: { classes: "bg-slate-500/15 text-slate-400 border-slate-500/30", Icon: FileSignature },
};

/** Whether a skill's publisher signature checks out. */
export const SignatureBadge = memo(function SignatureBadge({ status, publisher, size = "md" }: { status: SignatureStatus; publisher?: string; size?: "sm" | "md" }) {
  const cfg = SIGNATURE_STYLE[status] ?? SIGNATURE_STYLE.unsigned;
  const { Icon } = cfg;
  return (
    <span
      title={publisher ? `${signatureLabel(status)} · ${publisher}` : signatureLabel(status)}
      className={clsx("inline-flex items-center gap-1 whitespace-nowrap rounded-full border font-medium", cfg.classes, size === "sm" ? "px-1.5 py-0.5 text-xs" : "px-2 py-0.5 text-xs")}
    >
      <Icon size={size === "sm" ? 10 : 11} />
      {signatureLabel(status)}
    </span>
  );
});

const SEVERITY_CLASS: Record<FindingSeverity, string> = {
  info: "text-slate-500",
  low: "text-sky-400",
  medium: "text-amber-400",
  high: "text-orange-400",
  critical: "text-red-400",
};

/** The findings of one scan: what, where, and the line that triggered it. */
export function FindingsList({ report, initiallyShown = 5, className }: { report: ScanReport; initiallyShown?: number; className?: string }) {
  const [showAll, setShowAll] = useState(false);
  const findings = sortedFindings(report);
  const notes = report.findings.length - findings.length;
  if (findings.length === 0) {
    return <p className={clsx("text-sm text-slate-500", className)}>{notes > 0 ? `No concerns. ${notes} informational note${notes === 1 ? "" : "s"} (such as bundled scripts).` : "No findings."}</p>;
  }
  const shown = showAll ? findings : findings.slice(0, initiallyShown);
  return (
    <div className={className}>
      <ul className="space-y-2">
        {shown.map((f, i) => (
          <li key={`${f.rule}-${f.file}-${f.line ?? 0}-${i}`} className="text-sm">
            <div className="flex flex-wrap items-baseline gap-x-2">
              <span className={clsx("text-xs font-semibold uppercase tracking-wide", SEVERITY_CLASS[f.severity])}>{f.severity}</span>
              <span className="text-slate-200">{f.message}</span>
              <span className="font-mono text-xs text-slate-500">{f.file}{f.line ? `:${f.line}` : ""}</span>
            </div>
            {f.evidence && <code className="mt-0.5 block truncate rounded bg-[var(--palette-0f1225)] px-2 py-1 font-mono text-xs text-slate-400" title={f.evidence}>{f.evidence}</code>}
          </li>
        ))}
      </ul>
      {findings.length > initiallyShown && (
        <button type="button" className="mt-2 text-xs text-brand-400 hover:underline" onClick={() => setShowAll((v) => !v)}>
          {showAll ? "Show fewer" : `Show all ${findings.length} findings`}
        </button>
      )}
    </div>
  );
}
