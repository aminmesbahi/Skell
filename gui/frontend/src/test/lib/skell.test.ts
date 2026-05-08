import { describe, it, expect, vi, beforeEach } from "vitest";

describe("skell.ts — runJSON / run helpers", () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it("runJSON parses stdout as JSON", async () => {
    const payload = [{ name: "s", version: "1", registry: "r", source_repo: "", installed_path: "", installed_at: "", pinned: false, content_hash: "" }];
    (window.go.main.App.RunSkell as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      stdout: JSON.stringify(payload),
      stderr: "",
      success: true,
    });
    const { listInstalled } = await import("@/lib/skell");
    const result = await listInstalled("/repo");
    expect(result).toEqual(payload);
    expect(window.go.main.App.RunSkell).toHaveBeenCalledWith(["list", "--repo", "/repo", "--json"]);
  });

  it("runJSON throws on failure", async () => {
    (window.go.main.App.RunSkell as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      stdout: "",
      stderr: "something went wrong",
      success: false,
    });
    const { listInstalled } = await import("@/lib/skell");
    await expect(listInstalled("/repo")).rejects.toThrow("something went wrong");
  });
});

describe("skell.ts — command argument construction", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    (window.go.main.App.RunSkell as ReturnType<typeof vi.fn>).mockResolvedValue({
      stdout: "[]",
      stderr: "",
      success: true,
    });
  });

  it("listInstalledGlobal sends --global flag", async () => {
    const { listInstalledGlobal } = await import("@/lib/skell");
    await listInstalledGlobal();
    expect(window.go.main.App.RunSkell).toHaveBeenCalledWith(["list", "--global", "--json"]);
  });

  it("listRegistry sends --source registry flag", async () => {
    const { listRegistry } = await import("@/lib/skell");
    await listRegistry();
    expect(window.go.main.App.RunSkell).toHaveBeenCalledWith([
      "list",
      "--source",
      "registry",
      "--json",
    ]);
  });

  it("getStatus sends --repo", async () => {
    const { getStatus } = await import("@/lib/skell");
    await getStatus("/repo");
    expect(window.go.main.App.RunSkell).toHaveBeenCalledWith(["status", "--repo", "/repo", "--json"]);
  });

  it("getInfo without repo omits --repo flag", async () => {
    (window.go.main.App.RunSkell as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      stdout: JSON.stringify({ name: "x", status: "up-to-date" }),
      stderr: "",
      success: true,
    });
    const { getInfo } = await import("@/lib/skell");
    await getInfo("my-skill");
    expect(window.go.main.App.RunSkell).toHaveBeenCalledWith(["info", "my-skill", "--json"]);
  });

  it("getInfo with repo appends --repo", async () => {
    (window.go.main.App.RunSkell as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      stdout: JSON.stringify({ name: "x", status: "up-to-date" }),
      stderr: "",
      success: true,
    });
    const { getInfo } = await import("@/lib/skell");
    await getInfo("my-skill", "/repo");
    expect(window.go.main.App.RunSkell).toHaveBeenCalledWith([
      "info",
      "my-skill",
      "--repo",
      "/repo",
      "--json",
    ]);
  });

  it("installSkill builds correct args", async () => {
    (window.go.main.App.RunSkell as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      stdout: "",
      stderr: "",
      success: true,
    });
    const { installSkill } = await import("@/lib/skell");
    await installSkill({ skillName: "s", repo: "/repo", registry: "r", dryRun: true });
    expect(window.go.main.App.RunSkell).toHaveBeenCalledWith([
      "install",
      "s",
      "--repo",
      "/repo",
      "--registry",
      "r",
      "--dry-run",
    ]);
  });

  it("upgradeSkill without skillName upgrades all", async () => {
    (window.go.main.App.RunSkell as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      stdout: "",
      stderr: "",
      success: true,
    });
    const { upgradeSkill } = await import("@/lib/skell");
    await upgradeSkill({ repo: "/repo" });
    expect(window.go.main.App.RunSkell).toHaveBeenCalledWith(["upgrade", "--repo", "/repo"]);
  });

  it("removeSkill builds correct args", async () => {
    (window.go.main.App.RunSkell as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      stdout: "",
      stderr: "",
      success: true,
    });
    const { removeSkill } = await import("@/lib/skell");
    await removeSkill({ skillName: "s", repo: "/repo" });
    expect(window.go.main.App.RunSkell).toHaveBeenCalledWith(["remove", "s", "--repo", "/repo"]);
  });

  it("pinSkill with version appends --version", async () => {
    (window.go.main.App.RunSkell as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      stdout: "",
      stderr: "",
      success: true,
    });
    const { pinSkill } = await import("@/lib/skell");
    await pinSkill({ skillName: "s", repo: "/repo", version: "1.2.3" });
    expect(window.go.main.App.RunSkell).toHaveBeenCalledWith([
      "pin",
      "s",
      "--repo",
      "/repo",
      "--version",
      "1.2.3",
    ]);
  });

  it("unpinSkill builds correct args", async () => {
    (window.go.main.App.RunSkell as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      stdout: "",
      stderr: "",
      success: true,
    });
    const { unpinSkill } = await import("@/lib/skell");
    await unpinSkill({ skillName: "s", repo: "/repo" });
    expect(window.go.main.App.RunSkell).toHaveBeenCalledWith(["unpin", "s", "--repo", "/repo"]);
  });

  it("syncRepo with dryRun and check appends flags", async () => {
    (window.go.main.App.RunSkell as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      stdout: JSON.stringify({ installed: [], removed: [] }),
      stderr: "",
      success: true,
    });
    const { syncRepo } = await import("@/lib/skell");
    await syncRepo({ repo: "/repo", dryRun: true, check: true });
    expect(window.go.main.App.RunSkell).toHaveBeenCalledWith([
      "sync",
      "--repo",
      "/repo",
      "--check",
      "--dry-run",
      "--json",
    ]);
  });

  it("searchSkills without filters sends only 'search'", async () => {
    const { searchSkills } = await import("@/lib/skell");
    await searchSkills({});
    expect(window.go.main.App.RunSkell).toHaveBeenCalledWith(["search", "--json"]);
  });

  it("searchSkills passes all optional filters", async () => {
    const { searchSkills } = await import("@/lib/skell");
    await searchSkills({ query: "pdf", lifecycle: "stable", owner: "org", tag: "ai", repo: "/r" });
    expect(window.go.main.App.RunSkell).toHaveBeenCalledWith([
      "search",
      "pdf",
      "--lifecycle",
      "stable",
      "--owner",
      "org",
      "--tag",
      "ai",
      "--repo",
      "/r",
      "--json",
    ]);
  });

  it("doctorCheck sends correct args", async () => {
    const { doctorCheck } = await import("@/lib/skell");
    await doctorCheck("/repo");
    expect(window.go.main.App.RunSkell).toHaveBeenCalledWith(["doctor", "--repo", "/repo", "--json"]);
  });

  it("cacheStatus calls cache status", async () => {
    (window.go.main.App.RunSkell as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      stdout: "Cache: OK",
      stderr: "",
      success: true,
    });
    const { cacheStatus } = await import("@/lib/skell");
    const r = await cacheStatus();
    expect(r.stdout).toBe("Cache: OK");
    expect(window.go.main.App.RunSkell).toHaveBeenCalledWith(["cache", "status"]);
  });

  it("cacheRefresh calls cache refresh", async () => {
    (window.go.main.App.RunSkell as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      stdout: "Refreshed",
      stderr: "",
      success: true,
    });
    const { cacheRefresh } = await import("@/lib/skell");
    await cacheRefresh();
    expect(window.go.main.App.RunSkell).toHaveBeenCalledWith(["cache", "refresh"]);
  });

  it("cacheClear calls cache clear", async () => {
    (window.go.main.App.RunSkell as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      stdout: "",
      stderr: "",
      success: true,
    });
    const { cacheClear } = await import("@/lib/skell");
    await cacheClear();
    expect(window.go.main.App.RunSkell).toHaveBeenCalledWith(["cache", "clear"]);
  });

  it("selfUpdateCheck sends --check flag", async () => {
    (window.go.main.App.RunSkell as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      stdout: "up to date",
      stderr: "",
      success: true,
    });
    const { selfUpdateCheck } = await import("@/lib/skell");
    await selfUpdateCheck();
    expect(window.go.main.App.RunSkell).toHaveBeenCalledWith(["selfupdate", "--check"]);
  });

  it("selfUpdate sends selfupdate without flags", async () => {
    (window.go.main.App.RunSkell as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      stdout: "Updated",
      stderr: "",
      success: true,
    });
    const { selfUpdate } = await import("@/lib/skell");
    await selfUpdate();
    expect(window.go.main.App.RunSkell).toHaveBeenCalledWith(["selfupdate"]);
  });
});

describe("skell.ts — readAuditLog", () => {
  it("returns empty array when path is empty", async () => {
    (window.go.main.App.AuditLogPath as ReturnType<typeof vi.fn>).mockResolvedValueOnce("");
    const { readAuditLog } = await import("@/lib/skell");
    const result = await readAuditLog();
    expect(result).toEqual([]);
  });

  it("parses NDJSON lines and reverses order", async () => {
    (window.go.main.App.AuditLogPath as ReturnType<typeof vi.fn>).mockResolvedValueOnce("/path/audit.log");
    const line1 = JSON.stringify({ timestamp: "2024-01-01T00:00:00Z", action: "install", skill: "a" });
    const line2 = JSON.stringify({ timestamp: "2024-01-02T00:00:00Z", action: "remove", skill: "b" });
    (window.go.main.App.ReadFileContent as ReturnType<typeof vi.fn>).mockResolvedValueOnce(
      `${line1}\n${line2}\n`
    );
    const { readAuditLog } = await import("@/lib/skell");
    const result = await readAuditLog();
    expect(result).toHaveLength(2);
    // newest first
    expect(result[0].skill).toBe("b");
    expect(result[1].skill).toBe("a");
  });

  it("skips malformed lines", async () => {
    (window.go.main.App.AuditLogPath as ReturnType<typeof vi.fn>).mockResolvedValueOnce("/path/audit.log");
    (window.go.main.App.ReadFileContent as ReturnType<typeof vi.fn>).mockResolvedValueOnce(
      `not-json\n${JSON.stringify({ timestamp: "t", action: "install", skill: "s" })}\n`
    );
    const { readAuditLog } = await import("@/lib/skell");
    const result = await readAuditLog();
    expect(result).toHaveLength(1);
    expect(result[0].skill).toBe("s");
  });
});

describe("skell.ts — getSkellVersion / listDirectory / readFileContent", () => {
  it("getSkellVersion delegates to SkellVersion()", async () => {
    (window.go.main.App.SkellVersion as ReturnType<typeof vi.fn>).mockResolvedValueOnce("v0.2.0");
    const { getSkellVersion } = await import("@/lib/skell");
    const v = await getSkellVersion();
    expect(v).toBe("v0.2.0");
  });

  it("listDirectory delegates to ListDirectory()", async () => {
    const entries = [{ name: "SKILL.md", path: "/p/SKILL.md", is_dir: false }];
    (window.go.main.App.ListDirectory as ReturnType<typeof vi.fn>).mockResolvedValueOnce(entries);
    const { listDirectory } = await import("@/lib/skell");
    const r = await listDirectory("/p");
    expect(r).toEqual(entries);
  });
});
