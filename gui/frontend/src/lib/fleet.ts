import type { FleetMatrixRow, FleetRepo, FleetSummary } from "./types";

/** Keep matrix rows whose skill name, or any version, matches the query. */
export function filterMatrix(rows: FleetMatrixRow[], query: string): FleetMatrixRow[] {
  const q = query.trim().toLowerCase();
  if (!q) return rows;
  return rows.filter((r) => r.name.toLowerCase().includes(q) || r.versions.some((v) => v.version.toLowerCase().includes(q)));
}

/** Keep repositories whose name, path or any skill matches the query. */
export function filterRepos(repos: FleetRepo[], query: string): FleetRepo[] {
  const q = query.trim().toLowerCase();
  if (!q) return repos;
  return repos.filter((r) => r.name.toLowerCase().includes(q) || r.path.toLowerCase().includes(q) || r.skills.some((s) => s.name.toLowerCase().includes(q)));
}

export interface FleetTile {
  key: string;
  label: string;
  value: number;
  tone: "neutral" | "ok" | "warn" | "bad";
  hint?: string;
}

/** The headline numbers, each with a tone: green when nothing needs doing. */
export function fleetTiles(s: FleetSummary): FleetTile[] {
  const tone = (n: number, bad = false): FleetTile["tone"] => (n === 0 ? "ok" : bad ? "bad" : "warn");
  return [
    { key: "repos", label: "Repositories", value: s.repos, tone: "neutral" },
    { key: "skills", label: "Different skills", value: s.unique_skills, tone: "neutral", hint: `${s.installs} installs in total` },
    { key: "outdated", label: "Outdated", value: s.outdated, tone: tone(s.outdated), hint: "Run with “check for updates” to refresh" },
    { key: "modified", label: "Edited by hand", value: s.modified, tone: tone(s.modified) },
    { key: "risk", label: "High-risk installs", value: s.high_risk, tone: tone(s.high_risk, true) },
    { key: "spread", label: "On mixed versions", value: s.version_spread, tone: tone(s.version_spread) },
    { key: "attention", label: "Repos needing attention", value: s.unhealthy_repos, tone: tone(s.unhealthy_repos) },
  ];
}

/** "report 2026-10-21" style file stem used for exports. */
export function reportFileName(format: "html" | "md" | "csv", now = new Date()): string {
  return `skell-fleet-report-${now.toISOString().slice(0, 10)}.${format}`;
}

/** Join a folder and file name using the separator the folder already uses. */
export function joinPath(dir: string, file: string): string {
  const sep = dir.includes("\\") && !dir.includes("/") ? "\\" : "/";
  return dir.endsWith("/") || dir.endsWith("\\") ? dir + file : dir + sep + file;
}
