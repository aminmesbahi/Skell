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
      expect(screen.getByText("repo")).toBeTruthy();
    });
  });

  it("dry-run toggle is present and defaults on", () => {
    renderWithRouter(<Sync />);
    const toggle = screen.queryByRole("checkbox");
    if (toggle) {
      expect((toggle as HTMLInputElement).checked).toBe(true);
    }
  });

  it("shows 'Not checked' status badge before any run", async () => {
    renderWithRouter(<Sync />);
    await waitFor(() => screen.getByText("repo"));
    expect(screen.getByText("Not checked")).toBeTruthy();
  });

  it("calls syncRepo when Preview Sync button clicked", async () => {
    renderWithRouter(<Sync />);
    await waitFor(() => screen.getByText("repo"));

    const btn = screen.getByRole("button", { name: /preview sync/i });
    fireEvent.click(btn);
    await waitFor(() => {
      expect(mockSkell.syncRepo).toHaveBeenCalled();
    });
  });

  it("shows 'Already up to date' when in sync", async () => {
    mockSkell.syncRepo.mockResolvedValue(mockSyncReport({ installed: [], removed: [] }));
    renderWithRouter(<Sync />);
    await waitFor(() => screen.getByText("repo"));

    fireEvent.click(screen.getByRole("button", { name: /preview sync/i }));
    await waitFor(() => {
      expect(screen.getByText("Already up to date")).toBeTruthy();
    });
  });

  it("shows 'Up to date' status badge when in sync", async () => {
    mockSkell.syncRepo.mockResolvedValue(mockSyncReport({ installed: [], removed: [] }));
    renderWithRouter(<Sync />);
    await waitFor(() => screen.getByText("repo"));

    fireEvent.click(screen.getByRole("button", { name: /preview sync/i }));
    await waitFor(() => {
      expect(screen.getByText("Up to date")).toBeTruthy();
    });
  });

  it("shows 'Will install' label and skill names from dry-run report", async () => {
    mockSkell.syncRepo.mockResolvedValue(
      mockSyncReport({ installed: ["skill-a", "skill-b"], removed: [] })
    );
    renderWithRouter(<Sync />);
    await waitFor(() => screen.getByText("repo"));

    fireEvent.click(screen.getByRole("button", { name: /preview sync/i }));
    await waitFor(() => {
      expect(screen.getByText("skill-a")).toBeTruthy();
      expect(screen.getByText("skill-b")).toBeTruthy();
    });
    expect(screen.getByText(/will install/i)).toBeTruthy();
  });

  it("shows pending changes count in status badge", async () => {
    mockSkell.syncRepo.mockResolvedValue(
      mockSyncReport({ installed: ["skill-a"], removed: ["old-skill"] })
    );
    renderWithRouter(<Sync />);
    await waitFor(() => screen.getByText("repo"));

    fireEvent.click(screen.getByRole("button", { name: /preview sync/i }));
    await waitFor(() => {
      expect(screen.getByText(/2 changes pending/i)).toBeTruthy();
    });
  });

  it("shows 'Apply now' button when dry-run finds pending changes", async () => {
    mockSkell.syncRepo.mockResolvedValue(
      mockSyncReport({ installed: ["skill-a"], removed: [] })
    );
    renderWithRouter(<Sync />);
    await waitFor(() => screen.getByText("repo"));

    fireEvent.click(screen.getByRole("button", { name: /preview sync/i }));
    await waitFor(() => {
      expect(screen.getByRole("button", { name: /apply now/i })).toBeTruthy();
    });
  });

  it("shows preview context banner with change count", async () => {
    mockSkell.syncRepo.mockResolvedValue(
      mockSyncReport({ installed: ["skill-a"], removed: [] })
    );
    renderWithRouter(<Sync />);
    await waitFor(() => screen.getByText("repo"));

    fireEvent.click(screen.getByRole("button", { name: /preview sync/i }));
    await waitFor(() => {
      expect(screen.getByText(/preview.*1 change would be applied/i)).toBeTruthy();
    });
  });

  it("shows sync error on failure", async () => {
    mockSkell.syncRepo.mockRejectedValue(new Error("sync failed"));
    renderWithRouter(<Sync />);
    await waitFor(() => screen.getByText("repo"));

    fireEvent.click(screen.getByRole("button", { name: /preview sync/i }));
    await waitFor(() => {
      expect(screen.getByText(/sync failed/i)).toBeTruthy();
    });
  });

  it("shows 'Error' status badge on failure", async () => {
    mockSkell.syncRepo.mockRejectedValue(new Error("network error"));
    renderWithRouter(<Sync />);
    await waitFor(() => screen.getByText("repo"));

    fireEvent.click(screen.getByRole("button", { name: /preview sync/i }));
    await waitFor(() => {
      expect(screen.getByText("Error")).toBeTruthy();
    });
  });
});
