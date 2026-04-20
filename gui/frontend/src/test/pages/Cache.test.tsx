import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor, fireEvent } from "@testing-library/react";
import { renderWithRouter } from "@/test/utils";
import { Cache } from "@/pages/Cache";
import * as skell from "@/lib/skell";
import { mockOkResult, mockErrResult } from "@/test/fixtures";
import { useUIStore } from "@/store";

vi.mock("@/lib/skell");

const mockSkell = skell as unknown as Record<string, ReturnType<typeof vi.fn>>;

beforeEach(() => {
  mockSkell.cacheStatus.mockResolvedValue(mockOkResult("Cache size: 42 MB\n3 registries"));
  mockSkell.cacheRefresh.mockResolvedValue(mockOkResult("Cache refreshed"));
  mockSkell.cacheClear.mockResolvedValue(mockOkResult());
  // Reset notification store between tests
  useUIStore.setState({ notifications: [] });
});

describe("Cache", () => {
  it("renders heading", () => {
    renderWithRouter(<Cache />);
    expect(screen.getByRole("heading", { level: 1, name: "Cache" })).toBeTruthy();
  });

  it("loads and displays cache status on mount", async () => {
    renderWithRouter(<Cache />);
    await waitFor(() => {
      expect(screen.getByText(/cache size/i)).toBeTruthy();
    });
  });

  it("calls cacheRefresh when Refresh button clicked", async () => {
    renderWithRouter(<Cache />);
    await waitFor(() => screen.getByText(/cache size/i));

    const btn = screen.getByRole("button", { name: /refresh/i });
    fireEvent.click(btn);
    await waitFor(() => {
      expect(mockSkell.cacheRefresh).toHaveBeenCalled();
    });
  });

  it("shows error notification when cacheRefresh fails", async () => {
    mockSkell.cacheRefresh.mockResolvedValue(mockErrResult("network error"));
    renderWithRouter(<Cache />);
    await waitFor(() => screen.getByText(/cache size/i));

    fireEvent.click(screen.getByRole("button", { name: /^Refresh$/i }));
    await waitFor(() => {
      const { notifications } = useUIStore.getState();
      expect(notifications.some((n) => /refresh failed/i.test(n.title))).toBe(true);
    });
  });

  it("shows confirmation dialog before clearing cache", async () => {
    renderWithRouter(<Cache />);
    await waitFor(() => screen.getByText(/cache size/i));

    const clearBtn = screen.getByRole("button", { name: /^Clear$/i });
    fireEvent.click(clearBtn);
    await waitFor(() => {
      // ConfirmDialog title appears
      expect(screen.getByText("Clear cache?")).toBeTruthy();
    });
  });

  it("calls cacheClear after confirmation", async () => {
    renderWithRouter(<Cache />);
    await waitFor(() => screen.getByText(/cache size/i));

    fireEvent.click(screen.getByRole("button", { name: /^Clear$/i }));
    await waitFor(() => screen.getByText("Clear cache?"));

    // Click the confirm button inside the dialog ("Clear Cache")
    const confirmBtn = screen.getByRole("button", { name: "Clear Cache" });
    fireEvent.click(confirmBtn);
    await waitFor(() => {
      expect(mockSkell.cacheClear).toHaveBeenCalled();
    });
  });
});
