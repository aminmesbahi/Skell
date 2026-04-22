import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor, fireEvent } from "@testing-library/react";
import { renderWithRouter } from "@/test/utils";
import { Doctor } from "@/pages/Doctor";
import * as skell from "@/lib/skell";
import { mockDiagnosticEntry } from "@/test/fixtures";

vi.mock("@/lib/skell");

const mockSkell = skell as unknown as Record<string, ReturnType<typeof vi.fn>>;

beforeEach(async () => {
  mockSkell.doctorCheck.mockResolvedValue([]);
  const { useRepoStore } = await import("@/store");
  // Use selectedRepo="global" so targets=repos (stable reference, avoids infinite re-render)
  useRepoStore.setState({ repos: ["/repo"], selectedRepo: "global" });
});

describe("Doctor", () => {
  it("renders heading", () => {
    renderWithRouter(<Doctor />);
    expect(screen.getByRole("heading", { name: /doctor/i })).toBeTruthy();
  });

  it("shows run button", () => {
    renderWithRouter(<Doctor />);
    expect(screen.getByRole("button", { name: /run/i })).toBeTruthy();
  });

  it("calls doctorCheck on run all click", async () => {
    renderWithRouter(<Doctor />);
    const runBtn = screen.getByRole("button", { name: /run/i });
    fireEvent.click(runBtn);
    await waitFor(() => {
      expect(mockSkell.doctorCheck).toHaveBeenCalledWith("/repo");
    });
  });

  it("shows issues when doctorCheck returns entries", async () => {
    mockSkell.doctorCheck.mockResolvedValue([
      mockDiagnosticEntry({ severity: "error", message: "Missing version" }),
    ]);
    renderWithRouter(<Doctor />);
    const runBtn = screen.getByRole("button", { name: /run/i });
    fireEvent.click(runBtn);
    await waitFor(() => {
      expect(screen.getByText(/missing version/i)).toBeTruthy();
    });
  });

  it("shows warning entries correctly", async () => {
    mockSkell.doctorCheck.mockResolvedValue([
      mockDiagnosticEntry({ severity: "warning", message: "Deprecated lifecycle" }),
    ]);
    renderWithRouter(<Doctor />);
    const runBtn = screen.getByRole("button", { name: /run/i });
    fireEvent.click(runBtn);
    await waitFor(() => {
      expect(screen.getByText(/deprecated lifecycle/i)).toBeTruthy();
    });
  });

  it("shows 'no issues' after clean run", async () => {
    mockSkell.doctorCheck.mockResolvedValue([]);
    renderWithRouter(<Doctor />);
    const runBtn = screen.getByRole("button", { name: /run/i });
    fireEvent.click(runBtn);
    await waitFor(() => {
      expect(screen.getByText(/no issues/i)).toBeTruthy();
    });
  });
});
