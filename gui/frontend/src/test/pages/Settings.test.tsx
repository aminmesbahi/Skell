import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor, fireEvent } from "@testing-library/react";
import { renderWithRouter } from "@/test/utils";
import { Settings } from "@/pages/Settings";
import * as skell from "@/lib/skell";
import { mockOkResult, mockErrResult } from "@/test/fixtures";

vi.mock("@/lib/skell");

const mockSkell = skell as unknown as Record<string, ReturnType<typeof vi.fn>>;

beforeEach(() => {
  mockSkell.getSkellVersion.mockResolvedValue("v0.1.0");
  mockSkell.selfUpdateCheck.mockResolvedValue(mockOkResult("Already up to date"));
  mockSkell.selfUpdate.mockResolvedValue(mockOkResult("Updated to v0.2.0"));
});

describe("Settings", () => {
  it("renders heading", () => {
    renderWithRouter(<Settings />);
    expect(screen.getByText(/settings/i)).toBeTruthy();
  });

  it("displays skell version after load", async () => {
    renderWithRouter(<Settings />);
    await waitFor(() => {
      expect(screen.getByText("v0.1.0")).toBeTruthy();
    });
  });

  it("check for updates button triggers selfUpdateCheck", async () => {
    renderWithRouter(<Settings />);
    await waitFor(() => screen.getByText("v0.1.0"));

    const btn = screen.getByRole("button", { name: /check/i });
    fireEvent.click(btn);
    await waitFor(() => {
      expect(mockSkell.selfUpdateCheck).toHaveBeenCalled();
    });
  });

  it("shows update info text after check", async () => {
    mockSkell.selfUpdateCheck.mockResolvedValue(mockOkResult("Already up to date"));
    renderWithRouter(<Settings />);
    await waitFor(() => screen.getByText("v0.1.0"));

    fireEvent.click(screen.getByRole("button", { name: /check/i }));
    await waitFor(() => {
      expect(screen.getByText(/already up to date/i)).toBeTruthy();
    });
  });

  it("shows error info when check fails", async () => {
    mockSkell.selfUpdateCheck.mockResolvedValue(mockErrResult("network timeout"));
    renderWithRouter(<Settings />);
    await waitFor(() => screen.getByText("v0.1.0"));

    fireEvent.click(screen.getByRole("button", { name: /check/i }));
    await waitFor(() => {
      expect(screen.getByText(/network timeout/i)).toBeTruthy();
    });
  });

  it("shows confirmation dialog before updating", async () => {
    mockSkell.selfUpdateCheck.mockResolvedValue(mockOkResult("v0.2.0 available"));
    renderWithRouter(<Settings />);
    await waitFor(() => screen.getByText("v0.1.0"));

    fireEvent.click(screen.getByRole("button", { name: /check/i }));
    await waitFor(() => screen.getByText(/v0.2.0 available/i));

    // Click the "Update Now" button to open the confirmation dialog
    const updateBtn = screen.getByRole("button", { name: /update now/i });
    fireEvent.click(updateBtn);
    await waitFor(() => {
      // ConfirmDialog title appears (ConfirmDialog has no role="dialog")
      expect(screen.getByText("Update Skell CLI?")).toBeTruthy();
    });
  });
});
