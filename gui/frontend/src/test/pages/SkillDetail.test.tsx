import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor, fireEvent } from "@testing-library/react";
import { renderRoute } from "@/test/utils";
import { SkillDetail } from "@/pages/SkillDetail";
import * as skell from "@/lib/skell";
import {
  mockInfoResult,
  mockOkResult,
  mockFileEntry,
  mockRegistrySkill,
  mockInstalledSkill,
} from "@/test/fixtures";

vi.mock("@/lib/skell");

const mockSkell = skell as unknown as Record<string, ReturnType<typeof vi.fn>>;

beforeEach(() => {
  mockSkell.getInfo.mockResolvedValue(mockInfoResult());
  mockSkell.listDirectory.mockResolvedValue([]);
  mockSkell.readFileContent.mockResolvedValue("# SKILL.md\n\nTest content");
  mockSkell.upgradeSkill.mockResolvedValue(mockOkResult());
  mockSkell.removeSkill.mockResolvedValue(mockOkResult());
  mockSkell.pinSkill.mockResolvedValue(mockOkResult());
  mockSkell.unpinSkill.mockResolvedValue(mockOkResult());
});

describe("SkillDetail", () => {
  it("renders skill name after loading", async () => {
    mockSkell.getInfo.mockResolvedValue(mockInfoResult({ name: "my-skill" }));
    renderRoute("/skills/:skillName", <SkillDetail />, "/skills/my-skill");
    await waitFor(() => {
      expect(screen.getByText("my-skill")).toBeTruthy();
    });
  });

  it("renders 'Skill not found' when info returns null", async () => {
    mockSkell.getInfo.mockResolvedValue(null as never);
    renderRoute("/skills/:skillName", <SkillDetail />, "/skills/unknown");
    await waitFor(() => {
      expect(screen.getByText(/skill not found/i)).toBeTruthy();
    });
  });

  it("does not crash when skill.metadata is undefined (regression test)", async () => {
    const info = mockInfoResult();
    (info.skill as { metadata?: unknown }).metadata = undefined;
    mockSkell.getInfo.mockResolvedValue(info);
    renderRoute("/skills/:skillName", <SkillDetail />, "/skills/test-skill");
    await waitFor(() => {
      expect(screen.getByText("test-skill")).toBeTruthy();
    });
  });

  it("renders lifecycle badge when metadata has lifecycle", async () => {
    mockSkell.getInfo.mockResolvedValue(
      mockInfoResult({
        skill: mockRegistrySkill({ metadata: { version: "1", owner: "o", lifecycle: "stable", scope: "s", tags: "", source_repo: "" } }),
      })
    );
    renderRoute("/skills/:skillName", <SkillDetail />, "/skills/test-skill");
    await waitFor(() => {
      expect(screen.getByText(/stable/i)).toBeTruthy();
    });
  });

  it("shows metadata tab content", async () => {
    renderRoute("/skills/:skillName", <SkillDetail />, "/skills/test-skill");
    await waitFor(() => screen.getByText("Metadata"));
    fireEvent.click(screen.getByText("Metadata"));
    await waitFor(() => {
      expect(screen.getByText("Registry Info")).toBeTruthy();
    });
  });

  it("switches to files tab and lists files", async () => {
    mockSkell.listDirectory.mockResolvedValue([
      mockFileEntry({ name: "SKILL.md" }),
    ]);
    renderRoute("/skills/:skillName", <SkillDetail />, "/skills/test-skill");
    await waitFor(() => screen.getByText("Files"));
    fireEvent.click(screen.getByText("Files"));
    await waitFor(() => {
      expect(screen.queryAllByText("SKILL.md").length).toBeGreaterThan(0);
    });
  });

  it("shows pinned badge for pinned skill", async () => {
    mockSkell.getInfo.mockResolvedValue(
      mockInfoResult({ lock: mockInstalledSkill({ pinned: true }) })
    );
    renderRoute("/skills/:skillName", <SkillDetail />, "/skills/test-skill");
    await waitFor(() => {
      expect(screen.getByText(/pinned/i)).toBeTruthy();
    });
  });

  it("handles remove with confirmation", async () => {
    const { useRepoStore } = await import("@/store");
    useRepoStore.setState({ selectedRepo: "/repo", repos: ["/repo"] });
    renderRoute("/skills/:skillName", <SkillDetail />, "/skills/test-skill");
    await waitFor(() => screen.getByRole("button", { name: /remove/i }));
    fireEvent.click(screen.getByRole("button", { name: /remove/i }));
    await waitFor(() => {
      expect(screen.getByText(/remove.*test-skill/i)).toBeTruthy();
    });
  });
});
