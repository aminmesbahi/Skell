import type { SkillTestResult } from "./types";

/** Pass/fail summary of a skill's test result, as a short phrase. */
export function summarizeResult(r: SkillTestResult): string {
  if (r.static_problems?.length) return "The eval file has problems";
  if (r.cases.length > 0 && r.skipped === r.cases.length) return `${r.cases.length} case${r.cases.length === 1 ? "" : "s"}, not run yet`;
  const parts = [`${r.passed} passed`];
  if (r.failed) parts.push(`${r.failed} failed`);
  if (r.errored) parts.push(`${r.errored} could not run`);
  return parts.join(" · ");
}
