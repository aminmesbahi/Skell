import type { FindingSeverity, RiskLevel, ScanReport, SignatureStatus, VerifyEntry } from "./types";

const RISK_ORDER: Record<RiskLevel, number> = { clean: 0, low: 1, medium: 2, high: 3, critical: 4 };
const SEVERITY_ORDER: Record<FindingSeverity, number> = { info: 0, low: 1, medium: 2, high: 3, critical: 4 };

export function riskRank(level: RiskLevel | string): number {
  return RISK_ORDER[level as RiskLevel] ?? 0;
}

export function severityRank(severity: FindingSeverity | string): number {
  return SEVERITY_ORDER[severity as FindingSeverity] ?? 0;
}

/** True when a skill's risk deserves attention before it is trusted. */
export function isHighRisk(level: RiskLevel | string): boolean {
  return riskRank(level) >= RISK_ORDER.high;
}

/** One sentence a person can act on, per risk level. */
export function riskAdvice(level: RiskLevel): string {
  switch (level) {
    case "clean":
      return "Nothing suspicious found.";
    case "low":
      return "Minor notes only. Usually fine.";
    case "medium":
      return "Worth reading the flagged lines before you rely on this skill.";
    case "high":
      return "Review the findings before using this skill. Remove it if you do not recognise them.";
    case "critical":
      return "Do not use this skill unless you wrote it and understand every finding.";
  }
}

export function signatureLabel(status: SignatureStatus): string {
  switch (status) {
    case "trusted":
      return "Signed · trusted";
    case "untrusted-key":
      return "Signed · untrusted";
    case "tampered":
      return "Tampered";
    case "invalid":
      return "Invalid signature";
    case "unsigned":
      return "Unsigned";
  }
}

export function signatureAdvice(entry: Pick<VerifyEntry, "status" | "publisher" | "detail">): string {
  switch (entry.status) {
    case "trusted":
      return `Signed by ${entry.publisher || "a publisher you trust"} and unchanged since.`;
    case "untrusted-key":
      return `Validly signed${entry.publisher ? ` by ${entry.publisher}` : ""}, but you have not chosen to trust that publisher yet. Add them in Settings.`;
    case "tampered":
      return "Files changed after the skill was signed. Reinstall it from its source.";
    case "invalid":
      return entry.detail || "The signature file is damaged or does not verify.";
    case "unsigned":
      return "This skill carries no signature, like most skills today.";
  }
}

export interface SecurityRow {
  name: string;
  scan?: ScanReport;
  signature?: VerifyEntry;
}

/** Join scan and verify results by skill name, riskiest first. */
export function joinSecurity(scans: ScanReport[], sigs: VerifyEntry[]): SecurityRow[] {
  const rows = new Map<string, SecurityRow>();
  for (const s of scans) rows.set(s.skill, { name: s.skill, scan: s });
  for (const v of sigs) rows.set(v.skill, { ...(rows.get(v.skill) ?? { name: v.skill }), signature: v });
  return [...rows.values()].sort((a, b) => {
    const bad = (r: SecurityRow) => (r.signature && ["tampered", "invalid"].includes(r.signature.status) ? 10 : 0);
    const ra = riskRank(a.scan?.level ?? "clean") + bad(a);
    const rb = riskRank(b.scan?.level ?? "clean") + bad(b);
    return rb - ra || (b.scan?.score ?? 0) - (a.scan?.score ?? 0) || a.name.localeCompare(b.name);
  });
}

/** Findings that matter, riskiest first; notes ("info") only when asked. */
export function sortedFindings(report: ScanReport, includeNotes = false) {
  return [...report.findings]
    .filter((f) => includeNotes || f.severity !== "info")
    .sort((a, b) => severityRank(b.severity) - severityRank(a.severity) || a.file.localeCompare(b.file) || (a.line ?? 0) - (b.line ?? 0));
}
