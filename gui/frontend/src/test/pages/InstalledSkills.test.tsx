import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor, fireEvent, within } from "@testing-library/react";
import { renderWithRouter } from "@/test/utils";
import { InstalledSkills } from "@/pages/InstalledSkills";
import * as skell from "@/lib/skell";
import { mockInstalledSkill, mockStatusEntry, mockOkResult } from "@/test/fixtures";

vi.mock("@/lib/skell");

const mockSkell = skell as unknown as Record<string, ReturnType<typeof vi.fn>>;

beforeEach(() => {
  mockSkell.listInstalled.mockResolvedValue([]);
  mockSkell.skellPresent = vi.fn().mockResolvedValue(true);
  mockSkell.listInstalledGlobal.mockResolvedValue([]);
  mockSkell.getStatus.mockResolvedValue([]);
  mockSkell.isRepoInitialized.mockResolvedValue(false);
  mockSkell.upgradeSkill.mockResolvedValue(mockOkResult());
  mockSkell.removeSkill.mockResolvedValue(mockOkResult());
  mockSkell.pinSkill.mockResolvedValue(mockOkResult());
  mockSkell.unpinSkill.mockResolvedValue(mockOkResult());
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

describe("InstalledSkills", () => {
  it("renders heading", () => {
    renderWithRouter(<InstalledSkills />);
    expect(screen.getByText(/my skills/i)).toBeTruthy();
  });

  it("renders empty state when no skills", async () => {
    renderWithRouter(<InstalledSkills />);
    await waitFor(() => {
      expect(screen.getByText(/no skills/i)).toBeTruthy();
    });
  });

  it("renders skill rows from installed list", async () => {
    mockSkell.listInstalledGlobal.mockResolvedValue([
      mockInstalledSkill({ name: "pdf-skill" }),
    ]);
    mockSkell.getStatus.mockResolvedValue([mockStatusEntry({ name: "pdf-skill" })]);
    renderWithRouter(<InstalledSkills />);
    await waitFor(() => {
      expect(screen.getByText("pdf-skill")).toBeTruthy();
    });
  });

  it("filters skills by search query", async () => {
    mockSkell.listInstalledGlobal.mockResolvedValue([
      mockInstalledSkill({ name: "alpha-skill" }),
      mockInstalledSkill({ name: "beta-skill" }),
    ]);
    renderWithRouter(<InstalledSkills />);
    await waitFor(() => screen.getByText("alpha-skill"));

    const input = screen.getByPlaceholderText(/search/i);
    fireEvent.change(input, { target: { value: "alpha" } });

    await waitFor(() => {
      expect(screen.queryByText("beta-skill")).toBeNull();
      expect(screen.getByText("alpha-skill")).toBeTruthy();
    });
  });

  it("shows outdated badge for outdated skills", async () => {
    mockSkell.listInstalledGlobal.mockResolvedValue([mockInstalledSkill({ name: "old-skill" })]);
    mockSkell.getStatus.mockResolvedValue([
      mockStatusEntry({ name: "old-skill", status: "outdated" }),
    ]);
    renderWithRouter(<InstalledSkills />);
    await waitFor(() => {
      expect(screen.getByText(/outdated/i)).toBeTruthy();
    });
  });

  it("upgrade button calls upgradeSkill", async () => {
    mockSkell.listInstalled.mockResolvedValue([mockInstalledSkill({ name: "upg-skill" })]);
    mockSkell.getStatus.mockResolvedValue([
      mockStatusEntry({ name: "upg-skill", status: "outdated" }),
    ]);
    // pre-set selectedRepo to something non-global so upgrade button is visible
    const { useRepoStore } = await import("@/store");
    useRepoStore.setState({ selectedRepo: "/repo", repos: ["/repo"] });

    renderWithRouter(<InstalledSkills />);
    await waitFor(() => screen.getByText("upg-skill"));

    mockSkell.diffSkill.mockResolvedValue({
      name: "upg-skill", registry: "default", patch: "--- installed/SKILL.md\n+++ latest/SKILL.md\n@@ -1 +1 @@\n-old\n+new\n", locally_modified: false,
    });
    fireEvent.click(screen.getByRole("button", { name: /review changes and upgrade/i }));

    // The diff is shown before anything changes.
    const dialog = await screen.findByRole("dialog");
    await waitFor(() => expect(dialog.textContent).toContain("+new"));
    expect(mockSkell.upgradeSkill).not.toHaveBeenCalled();

    fireEvent.click(within(dialog).getByRole("button", { name: "Upgrade" }));
    await waitFor(() => {
      expect(mockSkell.upgradeSkill).toHaveBeenCalledWith({ skillName: "upg-skill", repo: "/repo", force: false });
    });
  });

  it("requires opting in before discarding local edits", async () => {
    mockSkell.listInstalled.mockResolvedValue([mockInstalledSkill({ name: "mod-skill" })]);
    mockSkell.getStatus.mockResolvedValue([mockStatusEntry({ name: "mod-skill", status: "outdated" })]);
    mockSkell.diffSkill.mockResolvedValue({ name: "mod-skill", registry: "default", patch: "-mine\n+theirs\n", locally_modified: true });
    const { useRepoStore } = await import("@/store");
    useRepoStore.setState({ selectedRepo: "/repo", repos: ["/repo"] });

    renderWithRouter(<InstalledSkills />);
    await waitFor(() => screen.getByText("mod-skill"));
    fireEvent.click(screen.getByRole("button", { name: /review changes and upgrade/i }));
    const dialog = await screen.findByRole("dialog");
    const upgrade = await within(dialog).findByRole("button", { name: "Upgrade" });
    await waitFor(() => expect((upgrade as HTMLButtonElement).disabled).toBe(true));
    fireEvent.click(within(dialog).getByRole("checkbox"));
    expect((upgrade as HTMLButtonElement).disabled).toBe(false);
    fireEvent.click(upgrade);
    await waitFor(() => expect(mockSkell.upgradeSkill).toHaveBeenCalledWith({ skillName: "mod-skill", repo: "/repo", force: true }));
  });
});
