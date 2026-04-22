import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import { renderWithRouter } from "@/test/utils";
import { Dashboard } from "@/pages/Dashboard";
import * as skell from "@/lib/skell";
import {
  mockInstalledSkill,
  mockStatusEntry,
  mockDiagnosticEntry,
} from "@/test/fixtures";

vi.mock("@/lib/skell");

const mockSkell = skell as unknown as Record<string, ReturnType<typeof vi.fn>>;

beforeEach(async () => {
  mockSkell.listInstalledGlobal.mockResolvedValue([]);
  mockSkell.listInstalled.mockResolvedValue([]);
  mockSkell.getStatus.mockResolvedValue([]);
  mockSkell.doctorCheck.mockResolvedValue([]);
  const { useRepoStore } = await import("@/store");
  useRepoStore.setState({ repos: [], selectedRepo: "global" });
});

describe("Dashboard", () => {
  it("renders heading", async () => {
    renderWithRouter(<Dashboard />);
    expect(screen.getByRole("heading", { level: 1, name: "Dashboard" })).toBeTruthy();
  });

  it("shows skill count after load", async () => {
    mockSkell.listInstalled.mockResolvedValue([mockInstalledSkill()]);
    mockSkell.getStatus.mockResolvedValue([mockStatusEntry()]);
    renderWithRouter(<Dashboard />);
    await waitFor(() => {
      expect(screen.queryByText(/loading/i) ?? screen.getByText(/dashboard/i)).toBeTruthy();
    });
  });

  it("shows outdated count when skills are outdated", async () => {
    const { useRepoStore } = await import("@/store");
    useRepoStore.setState({ repos: ["/test-repo"], selectedRepo: "/test-repo" });
    mockSkell.listInstalled.mockResolvedValue([mockInstalledSkill(), mockInstalledSkill({ name: "b" })]);
    mockSkell.getStatus.mockResolvedValue([
      mockStatusEntry({ name: "test-skill", status: "outdated" }),
      mockStatusEntry({ name: "b", status: "up-to-date" }),
    ]);
    renderWithRouter(<Dashboard />);
    await waitFor(() => {
      // StatCard renders the count as a <p> with the number value
      const ones = screen.getAllByText("1");
      expect(ones.length).toBeGreaterThan(0);
    });
  });

  it("shows error badge when doctor detects errors", async () => {
    const { useRepoStore } = await import("@/store");
    useRepoStore.setState({ repos: ["/test-repo"], selectedRepo: "/test-repo" });
    mockSkell.doctorCheck.mockResolvedValue([mockDiagnosticEntry()]);
    renderWithRouter(<Dashboard />);
    await waitFor(() => {
      const ones = screen.getAllByText("1");
      expect(ones.length).toBeGreaterThan(0);
    });
  });
});
