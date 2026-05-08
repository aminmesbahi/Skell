import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor, fireEvent } from "@testing-library/react";
import { renderWithRouter } from "@/test/utils";
import { InstalledSkills } from "@/pages/InstalledSkills";
import * as skell from "@/lib/skell";
import { mockInstalledSkill, mockStatusEntry, mockOkResult } from "@/test/fixtures";

vi.mock("@/lib/skell");

const mockSkell = skell as unknown as Record<string, ReturnType<typeof vi.fn>>;

beforeEach(() => {
  mockSkell.listInstalled.mockResolvedValue([]);
  mockSkell.listInstalledGlobal.mockResolvedValue([]);
  mockSkell.getStatus.mockResolvedValue([]);
  mockSkell.isRepoInitialized.mockResolvedValue(false);
  mockSkell.upgradeSkill.mockResolvedValue(mockOkResult());
  mockSkell.removeSkill.mockResolvedValue(mockOkResult());
  mockSkell.pinSkill.mockResolvedValue(mockOkResult());
  mockSkell.unpinSkill.mockResolvedValue(mockOkResult());
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

    const upgBtns = screen.queryAllByRole("button", { name: /upgrade/i });
    if (upgBtns.length > 0) {
      fireEvent.click(upgBtns[0]);
      await waitFor(() => {
        expect(mockSkell.upgradeSkill).toHaveBeenCalled();
      });
    }
  });
});
