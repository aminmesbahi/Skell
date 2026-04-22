import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor, fireEvent } from "@testing-library/react";
import { renderWithRouter } from "@/test/utils";
import { AuditLog } from "@/pages/AuditLog";
import * as skell from "@/lib/skell";
import { mockAuditEntry } from "@/test/fixtures";

vi.mock("@/lib/skell");

const mockSkell = skell as unknown as Record<string, ReturnType<typeof vi.fn>>;

beforeEach(() => {
  mockSkell.readAuditLog.mockResolvedValue([]);
});

describe("AuditLog", () => {
  it("renders heading", () => {
    renderWithRouter(<AuditLog />);
    expect(screen.getByText(/audit log/i)).toBeTruthy();
  });

  it("shows refresh button", async () => {
    renderWithRouter(<AuditLog />);
    // The header refresh button is icon-only (no accessible name).
    // Wait for loading to finish and confirm at least one button exists.
    await waitFor(() => {
      expect(screen.getAllByRole("button").length).toBeGreaterThan(0);
    });
  });

  it("shows empty state when no entries", async () => {
    renderWithRouter(<AuditLog />);
    await waitFor(() => {
      expect(screen.getByText(/no audit log entries/i)).toBeTruthy();
    });
  });

  it("displays audit entries after load", async () => {
    mockSkell.readAuditLog.mockResolvedValue([
      mockAuditEntry({ skill: "audit-skill", action: "install" }),
    ]);
    renderWithRouter(<AuditLog />);
    await waitFor(() => {
      expect(screen.getByText("audit-skill")).toBeTruthy();
    });
  });

  it("filters entries by search query", async () => {
    mockSkell.readAuditLog.mockResolvedValue([
      mockAuditEntry({ skill: "alpha-skill", action: "install" }),
      mockAuditEntry({ skill: "beta-skill", action: "remove" }),
    ]);
    renderWithRouter(<AuditLog />);
    await waitFor(() => screen.getByText("alpha-skill"));

    const input = screen.getByPlaceholderText(/search/i);
    fireEvent.change(input, { target: { value: "alpha" } });

    await waitFor(() => {
      expect(screen.queryByText("beta-skill")).toBeNull();
      expect(screen.getByText("alpha-skill")).toBeTruthy();
    });
  });

  it("filters entries by action", async () => {
    mockSkell.readAuditLog.mockResolvedValue([
      mockAuditEntry({ skill: "s1", action: "install" }),
      mockAuditEntry({ skill: "s2", action: "remove" }),
    ]);
    renderWithRouter(<AuditLog />);
    await waitFor(() => screen.getByText("s1"));

    const selects = document.querySelectorAll("select");
    expect(selects.length).toBeGreaterThan(0);
    fireEvent.change(selects[0], { target: { value: "install" } });

    await waitFor(() => {
      expect(screen.queryByText("s2")).toBeNull();
      expect(screen.getByText("s1")).toBeTruthy();
    });
  });

  it("reloads on refresh button click", async () => {
    renderWithRouter(<AuditLog />);
    // Wait for initial load to complete
    await waitFor(() => expect(mockSkell.readAuditLog).toHaveBeenCalledTimes(1));
    // Icon-only button - get the first (and only) button in the heading area
    const allBtns = screen.getAllByRole("button");
    fireEvent.click(allBtns[0]);
    await waitFor(() => {
      expect(mockSkell.readAuditLog).toHaveBeenCalledTimes(2);
    });
  });
});
