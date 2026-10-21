import { describe, expect, it } from "vitest";
import { isHighRisk, joinSecurity, riskAdvice, riskRank, signatureAdvice, signatureLabel, sortedFindings } from "@/lib/security";
import { filterMatrix, filterRepos, fleetTiles, joinPath, reportFileName } from "@/lib/fleet";
import { BUNDLE_NAME_RX, parseSkillList } from "@/lib/bundles";
import { summarizeResult } from "@/lib/skilltests";
import type { FleetMatrixRow, FleetRepo, FleetSummary, ScanReport, SkillTestResult, VerifyEntry } from "@/lib/types";

const scan = (skill: string, level: ScanReport["level"], score = 0, findings: ScanReport["findings"] = []): ScanReport => ({ skill, level, score, files: 1, findings });
const sig = (skill: string, status: VerifyEntry["status"]): VerifyEntry => ({ skill, status });

describe("security helpers", () => {
  it("ranks risk levels and flags high risk", () => {
    expect(riskRank("critical")).toBeGreaterThan(riskRank("high"));
    expect(riskRank("clean")).toBe(0);
    expect(riskRank("nonsense")).toBe(0);
    expect(isHighRisk("high")).toBe(true);
    expect(isHighRisk("critical")).toBe(true);
    expect(isHighRisk("medium")).toBe(false);
  });

  it("explains every level and signature state in plain words", () => {
    for (const level of ["clean", "low", "medium", "high", "critical"] as const) expect(riskAdvice(level).length).toBeGreaterThan(10);
    expect(signatureLabel("trusted")).toBe("Signed · trusted");
    expect(signatureLabel("unsigned")).toBe("Unsigned");
    expect(signatureAdvice({ status: "trusted", publisher: "acme" })).toContain("acme");
    expect(signatureAdvice({ status: "untrusted-key", publisher: "acme" })).toContain("Settings");
    expect(signatureAdvice({ status: "invalid", detail: "bad" })).toBe("bad");
    expect(signatureAdvice({ status: "tampered" })).toContain("Reinstall");
  });

  it("joins scans and signatures, riskiest and tampered first", () => {
    const rows = joinSecurity(
      [scan("calm", "clean"), scan("shady", "critical", 75), scan("meh", "medium", 20)],
      [sig("calm", "tampered"), sig("meh", "trusted"), sig("only-signed", "unsigned")],
    );
    // A tampered signature outranks a critical scan: it means the files are not what was signed.
    expect(rows[0].name).toBe("calm");
    expect(rows.map((r) => r.name)).toEqual(["calm", "shady", "meh", "only-signed"]);
    expect(rows.find((r) => r.name === "only-signed")?.scan).toBeUndefined();
  });

  it("sorts findings by severity and hides notes unless asked", () => {
    const report = scan("x", "high", 40, [
      { rule: "script", severity: "info", message: "script", file: "a.sh" },
      { rule: "b", severity: "medium", message: "m", file: "b.md", line: 2 },
      { rule: "c", severity: "high", message: "h", file: "a.md", line: 9 },
    ]);
    expect(sortedFindings(report).map((f) => f.rule)).toEqual(["c", "b"]);
    expect(sortedFindings(report, true).map((f) => f.rule)).toEqual(["c", "b", "script"]);
  });
});

const row = (name: string, versions: string[]): FleetMatrixRow => ({
  name,
  repos: versions.length,
  versions: versions.map((v) => ({ version: v, repos: 1 })),
  spread: new Set(versions).size > 1,
});

const repo = (name: string, skills: string[]): FleetRepo => ({
  path: `/work/${name}`,
  name,
  skills: skills.map((s) => ({ name: s, version: "1", registry: "r", status: "installed", risk: "clean", risk_score: 0, signature: "unsigned" })),
  errors: 0,
  warnings: 0,
  health: "healthy",
});

describe("fleet helpers", () => {
  it("filters the matrix by skill name or version", () => {
    const rows = [row("pdf", ["1.0.0", "1.2.0"]), row("docx", ["2.0.0"])];
    expect(filterMatrix(rows, "")).toHaveLength(2);
    expect(filterMatrix(rows, "PD").map((r) => r.name)).toEqual(["pdf"]);
    expect(filterMatrix(rows, "2.0").map((r) => r.name)).toEqual(["pdf", "docx"]);
    expect(filterMatrix(rows, "zzz")).toEqual([]);
  });

  it("filters repositories by name, path or skill", () => {
    const repos = [repo("api", ["pdf"]), repo("web", ["css"])];
    expect(filterRepos(repos, "api").map((r) => r.name)).toEqual(["api"]);
    expect(filterRepos(repos, "css").map((r) => r.name)).toEqual(["web"]);
    expect(filterRepos(repos, "/work").length).toBe(2);
  });

  it("tones the headline tiles: green when nothing needs doing", () => {
    const summary: FleetSummary = { repos: 3, installs: 9, unique_skills: 4, outdated: 0, modified: 2, pinned: 1, high_risk: 1, unsigned: 9, version_spread: 0, unhealthy_repos: 0 };
    const tiles = Object.fromEntries(fleetTiles(summary).map((t) => [t.key, t]));
    expect(tiles.outdated.tone).toBe("ok");
    expect(tiles.modified.tone).toBe("warn");
    expect(tiles.risk.tone).toBe("bad");
    expect(tiles.repos.tone).toBe("neutral");
  });

  it("names export files and joins paths with the folder's own separator", () => {
    expect(reportFileName("html", new Date("2026-10-21T10:00:00Z"))).toBe("skell-fleet-report-2026-10-21.html");
    expect(joinPath("C:\\Users\\me", "r.html")).toBe("C:\\Users\\me\\r.html");
    expect(joinPath("/home/me", "r.csv")).toBe("/home/me/r.csv");
    expect(joinPath("/home/me/", "r.csv")).toBe("/home/me/r.csv");
  });
});

describe("bundle and test helpers", () => {
  it("parses typed skill lists", () => {
    expect(parseSkillList("pdf, docx  xlsx,,pdf")).toEqual(["pdf", "docx", "xlsx"]);
    expect(parseSkillList("  ")).toEqual([]);
    expect(BUNDLE_NAME_RX.test("front-end")).toBe(true);
    expect(BUNDLE_NAME_RX.test("Front End")).toBe(false);
  });

  it("summarises test results", () => {
    const base: SkillTestResult = { skill: "s", cases: [], passed: 0, failed: 0, errored: 0, skipped: 0 };
    expect(summarizeResult({ ...base, static_problems: ["x"] })).toContain("problems");
    expect(summarizeResult({ ...base, cases: [{ name: "a", status: "skipped", duration_ms: 0 }], skipped: 1 })).toBe("1 case, not run yet");
    expect(summarizeResult({ ...base, passed: 2, failed: 1, errored: 1 })).toBe("2 passed · 1 failed · 1 could not run");
  });
});
