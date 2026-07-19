import { describe, expect, it } from "vitest";
import { migrateRepoStoreState } from "@/store";

describe("repo store migration", () => {
  it("migrates legacy repo arrays without losing projects", () => {
    const migrated = migrateRepoStoreState({
      repos: ["/tmp/Project A", "C:/tmp/project-a", "global", "/tmp/Project A"],
      selectedRepo: "global",
    } as any);

    expect(migrated.repos).toEqual(["/tmp/Project A", "C:/tmp/project-a", "global"]);
    expect(migrated.recentProjects).toEqual([
      expect.objectContaining({ path: "/tmp/Project A", displayName: "Project A" }),
      expect.objectContaining({ path: "C:/tmp/project-a", displayName: "project-a" }),
    ]);
  });

  it("keeps a stable selected project after migration", () => {
    const migrated = migrateRepoStoreState({
      repos: ["/tmp/Alpha", "/tmp/Beta"],
      selectedRepo: "/tmp/Beta",
    } as any);

    expect(migrated.selectedRepo).toBe("/tmp/Beta");
  });
});
