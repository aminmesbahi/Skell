import { beforeEach, describe, expect, it, vi } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import { renderRoute } from "@/test/utils";
import { ProjectSkillsPage } from "@/pages/projects/ProjectSkillsPage";
import { useRepoStore } from "@/store";
import { createProjectId } from "@/lib/navigation";
import * as skell from "@/lib/skell";

vi.mock("@/lib/skell");

const mockSkell = skell as unknown as Record<string, ReturnType<typeof vi.fn>>;

describe("Project skills page", () => {
  beforeEach(() => {
    useRepoStore.setState({ repos: ["/home/user/project"], selectedRepo: "/home/user/project" });
    mockSkell.listInstalled.mockResolvedValue([
      {
        name: "alpha",
        version: "1.0.0",
        registry: "registry",
        source_repo: "source",
        installed_path: ".cursor/skills/alpha",
        installed_at: "",
        pinned: false,
        content_hash: "hash",
      },
    ]);
    mockSkell.getStatus.mockResolvedValue([
      { name: "alpha", installed: "1.0.0", latest: "1.0.0", status: "up-to-date" },
    ]);
    mockSkell.isRepoInitialized.mockResolvedValue(true);
    mockSkell.skellPresent.mockResolvedValue(true);
    mockSkell.initRepo.mockResolvedValue({ success: true, stdout: "ok", stderr: "" });
    mockSkell.listSupportedTargets.mockResolvedValue([
      { id: "claude", displayName: "Anthropic Claude Code", dir: ".claude", detected: false },
      { id: "codex", displayName: "OpenAI Codex", dir: ".codex", detected: false },
      { id: "copilot", displayName: "GitHub Copilot / VS Code", dir: ".github", detected: false },
      { id: "cursor", displayName: "Cursor", dir: ".cursor", detected: false },
      { id: "windsurf", displayName: "Windsurf / Cascade", dir: ".windsurf", detected: false },
      { id: "opencode", displayName: "OpenCode", dir: ".opencode", detected: false },
      { id: "cline", displayName: "Cline", dir: ".cline", detected: false },
      { id: "grok", displayName: "xAI Grok", dir: ".grok", detected: false },
    ]);
    mockSkell.targetFromInstalledPath = vi.fn((path: string) => {
      if (!path) return "";
      const seg = path.split(/[/\\]/)[0];
      return seg?.startsWith(".") ? seg.slice(1) : seg ?? "";
    });
  });

  it("renders installed skills for the selected project", async () => {
    renderRoute(
      "/projects/:projectId/skills",
      <ProjectSkillsPage />,
      `/projects/${createProjectId("/home/user/project")}/skills`
    );

    await waitFor(() => {
      expect(screen.getByText("alpha")).toBeTruthy();
    });
    expect(screen.getByText(/skills for/i)).toBeTruthy();
  });

  it("offers catalog and repository install actions when a project has no skills", async () => {
    mockSkell.listInstalled.mockResolvedValue([]);
    mockSkell.getStatus.mockResolvedValue([]);

    renderRoute(
      "/projects/:projectId/skills",
      <ProjectSkillsPage />,
      `/projects/${createProjectId("/home/user/project")}/skills`
    );

    await waitFor(() => {
      expect(screen.getByText(/no skills installed yet/i)).toBeTruthy();
    });

    expect(screen.getByRole("button", { name: /browse catalog/i })).toBeTruthy();
    expect(screen.getByRole("button", { name: /add from repository/i })).toBeTruthy();
  });
});
