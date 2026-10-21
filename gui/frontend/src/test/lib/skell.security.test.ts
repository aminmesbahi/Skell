import { beforeEach, describe, expect, it, vi } from "vitest";

type Mock = ReturnType<typeof vi.fn>;
const runSkell = () => window.go.main.App.RunSkell as Mock;
const ok = (stdout: string) => ({ stdout, stderr: "", success: true });

describe("skell.ts — security, bundles, fleet and test commands", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    runSkell().mockResolvedValue(ok("[]"));
  });

  it("scanSkills never fails the command on risky results and can target one skill", async () => {
    const { scanSkills } = await import("@/lib/skell");
    await scanSkills("/repo");
    expect(runSkell()).toHaveBeenCalledWith(["scan", "--repo", "/repo", "--fail-on", "", "--json"]);
    await scanSkills("/repo", "pdf");
    expect(runSkell()).toHaveBeenLastCalledWith(["scan", "pdf", "--repo", "/repo", "--fail-on", "", "--json"]);
    await scanSkills("global");
    expect(runSkell()).toHaveBeenLastCalledWith(["scan", "--global", "--fail-on", "", "--json"]);
  });

  it("verifySkills and trust commands build the right arguments", async () => {
    const lib = await import("@/lib/skell");
    await lib.verifySkills("/repo");
    expect(runSkell()).toHaveBeenLastCalledWith(["verify", "--repo", "/repo", "--json"]);
    await lib.listTrustedKeys();
    expect(runSkell()).toHaveBeenLastCalledWith(["trust", "list", "--json"]);
    await lib.addTrustedKey("acme", "  skell-public-key AAA  ");
    expect(runSkell()).toHaveBeenLastCalledWith(["trust", "add", "acme", "skell-public-key AAA"]);
    await lib.removeTrustedKey("acme");
    expect(runSkell()).toHaveBeenLastCalledWith(["trust", "remove", "acme"]);
  });

  it("returns empty lists when the CLI prints null", async () => {
    runSkell().mockResolvedValue(ok("null"));
    const lib = await import("@/lib/skell");
    expect(await lib.scanSkills("/r")).toEqual([]);
    expect(await lib.verifySkills("/r")).toEqual([]);
    expect(await lib.listTrustedKeys()).toEqual([]);
    expect(await lib.listBundles("/r")).toEqual([]);
  });

  it("bundle commands build the right arguments", async () => {
    const lib = await import("@/lib/skell");
    await lib.listBundles("/repo");
    expect(runSkell()).toHaveBeenLastCalledWith(["bundle", "list", "--repo", "/repo", "--json"]);
    await lib.saveBundle({ repo: "/repo", name: "ui", description: "UI pack", skills: ["a", "b"] });
    expect(runSkell()).toHaveBeenLastCalledWith(["bundle", "create", "ui", "a", "b", "--repo", "/repo", "--description", "UI pack"]);
    await lib.installBundle("/repo", "ui");
    expect(runSkell()).toHaveBeenLastCalledWith(["bundle", "install", "ui", "--repo", "/repo", "--yes"]);
    await lib.deleteBundle("/repo", "ui");
    expect(runSkell()).toHaveBeenLastCalledWith(["bundle", "delete", "ui", "--repo", "/repo"]);
  });

  it("fleet report and export commands build the right arguments", async () => {
    const lib = await import("@/lib/skell");
    runSkell().mockResolvedValue(ok("{}"));
    await lib.fleetReport({ root: "/work", checkUpdates: true });
    expect(runSkell()).toHaveBeenLastCalledWith(["report", "--all-repos", "/work", "--check-updates", "--json"]);
    await lib.fleetReport({ repos: ["/a", "/b"] });
    expect(runSkell()).toHaveBeenLastCalledWith(["report", "--repo", "/a", "--repo", "/b", "--json"]);
    await lib.exportFleetReport({ repos: ["/a"], format: "html", out: "/tmp/r.html" });
    expect(runSkell()).toHaveBeenLastCalledWith(["report", "--format", "html", "--out", "/tmp/r.html", "--repo", "/a"]);
  });

  it("runSkillTests passes the agent only when given, and tolerates a failing exit", async () => {
    const lib = await import("@/lib/skell");
    const payload = { results: [{ skill: "pdf", cases: [], passed: 0, failed: 1, errored: 0, skipped: 0 }], without_evals: ["docx"] };
    // `skell test` exits non-zero on failing cases but still prints JSON.
    runSkell().mockResolvedValue({ stdout: JSON.stringify(payload), stderr: "1 skill(s) have failing evals", success: false });

    const run = await lib.runSkillTests({ repo: "/repo" });
    expect(runSkell()).toHaveBeenLastCalledWith(["test", "--repo", "/repo", "--json"]);
    expect(run.results[0].failed).toBe(1);
    expect(run.without_evals).toEqual(["docx"]);

    await lib.runSkillTests({ repo: "/repo", skillName: "pdf", agent: " claude -p " });
    expect(runSkell()).toHaveBeenLastCalledWith(["test", "pdf", "--repo", "/repo", "--agent", "claude -p", "--json"]);
  });

  it("runSkillTests copes with an empty result", async () => {
    runSkell().mockResolvedValue(ok("[]"));
    const { runSkillTests } = await import("@/lib/skell");
    expect(await runSkillTests({ repo: "/r" })).toEqual({ results: [], without_evals: [] });
  });
});
