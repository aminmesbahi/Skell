import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor, fireEvent } from "@testing-library/react";
import { renderWithRouter } from "@/test/utils";
import { Registry } from "@/pages/Registry";
import * as skell from "@/lib/skell";
import { mockRegistrySkill, mockOkResult } from "@/test/fixtures";
import { useUIStore } from "@/store";

vi.mock("@/lib/skell");

const mockSkell = skell as unknown as Record<string, ReturnType<typeof vi.fn>>;

beforeEach(() => {
  mockSkell.searchSkills.mockResolvedValue([]);
  mockSkell.installSkill.mockResolvedValue(mockOkResult());
  useUIStore.setState({ notifications: [] });
});

describe("Registry", () => {
  it("renders search UI heading", () => {
    renderWithRouter(<Registry />);
    expect(screen.getByText(/registry/i)).toBeTruthy();
  });

  it("calls searchSkills on mount", async () => {
    renderWithRouter(<Registry />);
    await waitFor(() => {
      expect(mockSkell.searchSkills).toHaveBeenCalled();
    });
  });

  it("displays skills returned by searchSkills", async () => {
    mockSkell.searchSkills.mockResolvedValue([
      mockRegistrySkill({ name: "my-skill" }),
    ]);
    renderWithRouter(<Registry />);
    await waitFor(() => {
      expect(screen.getByText("my-skill")).toBeTruthy();
    });
  });

  it("re-searches when lifecycle filter changes", async () => {
    renderWithRouter(<Registry />);
    await waitFor(() => expect(mockSkell.searchSkills).toHaveBeenCalledTimes(1));

    const selects = document.querySelectorAll("select");
    expect(selects.length).toBeGreaterThan(0);
    fireEvent.change(selects[0], { target: { value: "stable" } });

    await waitFor(() => {
      expect(mockSkell.searchSkills).toHaveBeenCalledWith(
        expect.objectContaining({ lifecycle: "stable" })
      );
    });
  });

  it("shows 'no skills found' when search returns empty", async () => {
    mockSkell.searchSkills.mockResolvedValue([]);
    renderWithRouter(<Registry />);
    await waitFor(() => {
      expect(screen.getByText(/no skills found/i)).toBeTruthy();
    });
  });

  it("opens install dialog when install button clicked", async () => {
    mockSkell.searchSkills.mockResolvedValue([mockRegistrySkill({ name: "click-skill" })]);
    renderWithRouter(<Registry />);
    await waitFor(() => screen.getByText("click-skill"));

    // The Install button is inside the SkillCard
    const installBtns = screen.getAllByRole("button", { name: /install/i });
    fireEvent.click(installBtns[0]);
    await waitFor(() => {
      // Dialog title: Install "click-skill"
      expect(screen.getByText(/install "click-skill"/i)).toBeTruthy();
    });
  });

  it("shows error notification when search fails", async () => {
    mockSkell.searchSkills.mockRejectedValue(new Error("Network error"));
    renderWithRouter(<Registry />);
    await waitFor(() => {
      const { notifications } = useUIStore.getState();
      expect(notifications.some((n) => /search failed/i.test(n.title))).toBe(true);
    });
  });
});
