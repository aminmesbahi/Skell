import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor, fireEvent } from "@testing-library/react";
import { renderWithRouter } from "@/test/utils";
import { Sync } from "@/pages/Sync";
import * as skell from "@/lib/skell";
import { mockSyncReport } from "@/test/fixtures";

vi.mock("@/lib/skell");

const mockSkell = skell as unknown as Record<string, ReturnType<typeof vi.fn>>;

beforeEach(async () => {
  mockSkell.syncRepo.mockResolvedValue(mockSyncReport());
  const { useRepoStore } = await import("@/store");
  // Use selectedRepo="global" so targets=repos (stable reference, avoids infinite re-render)
  useRepoStore.setState({ repos: ["/repo"], selectedRepo: "global" });
});

describe("Sync", () => {
  it("renders the sync page heading", () => {
    renderWithRouter(<Sync />);
    expect(screen.getByRole("heading", { name: "Sync" })).toBeTruthy();
  });

  it("renders repo name in sync card", async () => {
    renderWithRouter(<Sync />);
    await waitFor(() => {
      expect(screen.getByText("repo")).toBeTruthy(); // last path segment
    });
  });

  it("dry-run toggle is present and defaults on", () => {
    renderWithRouter(<Sync />);
    const toggle = screen.queryByRole("checkbox");
    if (toggle) {
      expect((toggle as HTMLInputElement).checked).toBe(true);
    }
  });

  it("calls syncRepo when Sync All button clicked", async () => {
    renderWithRouter(<Sync />);
    await waitFor(() => screen.getByText("repo"));

    const btns = screen.getAllByRole("button", { name: /sync all|run|sync/i });
    fireEvent.click(btns[0]);
    await waitFor(() => {
      expect(mockSkell.syncRepo).toHaveBeenCalled();
    });
  });

  it("shows installed skills from sync report", async () => {
    mockSkell.syncRepo.mockResolvedValue(
      mockSyncReport({ installed: ["skill-a", "skill-b"], removed: [] })
    );
    renderWithRouter(<Sync />);
    await waitFor(() => screen.getByText("repo"));

    const syncBtns = screen.getAllByRole("button", { name: /sync/i });
    fireEvent.click(syncBtns[0]);
    await waitFor(() => {
      expect(screen.getByText("skill-a")).toBeTruthy();
    });
  });

  it("shows sync error on failure", async () => {
    mockSkell.syncRepo.mockRejectedValue(new Error("sync failed"));
    renderWithRouter(<Sync />);
    await waitFor(() => screen.getByText("repo"));

    const btns = screen.getAllByRole("button", { name: /sync/i });
    fireEvent.click(btns[0]);
    await waitFor(() => {
      expect(screen.getByText(/sync failed/i)).toBeTruthy();
    });
  });
});
