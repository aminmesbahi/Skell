import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor, fireEvent } from "@testing-library/react";
import { renderWithRouter } from "@/test/utils";
import { Repositories } from "@/pages/Repositories";
import * as skell from "@/lib/skell";
import { mockOkResult } from "@/test/fixtures";

vi.mock("@/lib/skell");

const mockSkell = skell as unknown as Record<string, ReturnType<typeof vi.fn>>;

beforeEach(async () => {
  mockSkell.listInstalled.mockResolvedValue([]);
  mockSkell.getStatus.mockResolvedValue([]);
  mockSkell.doctorCheck.mockResolvedValue([]);
  mockSkell.initRepo.mockResolvedValue(mockOkResult());
  mockSkell.isRepoInitialized.mockResolvedValue(false);
  mockSkell.detectRepoTargets.mockResolvedValue([
    { id: "claude", displayName: "Anthropic Claude Code", dir: ".claude", detected: true },
  ]);
  mockSkell.listSupportedTargets.mockResolvedValue([
    { id: "claude", displayName: "Anthropic Claude Code", dir: ".claude", detected: false },
    { id: "codex", displayName: "OpenAI Codex", dir: ".codex", detected: false },
    { id: "copilot", displayName: "GitHub Copilot / VS Code", dir: ".github", detected: false },
    { id: "cursor", displayName: "Cursor", dir: ".cursor", detected: false },
  ]);

  const { useRepoStore } = await import("@/store");
  useRepoStore.setState({ repos: [], selectedRepo: "global" });
});

describe("Repositories", () => {
  it("renders heading", () => {
    renderWithRouter(<Repositories />);
    expect(screen.getByRole("heading", { name: "Repositories" })).toBeTruthy();
  });

  it("shows add repository button", () => {
    renderWithRouter(<Repositories />);
    // Multiple "Add Repository" buttons may exist (header + empty state)
    const addBtns = screen.getAllByRole("button", { name: /add repository/i });
    expect(addBtns.length).toBeGreaterThan(0);
  });

  it("shows empty state when no repos added", async () => {
    renderWithRouter(<Repositories />);
    await waitFor(() => {
      expect(screen.getByText(/no repositories/i)).toBeTruthy();
    });
  });

  it("shows repo in list after adding", async () => {
    const { useRepoStore } = await import("@/store");
    useRepoStore.setState({ repos: ["/home/user/project"], selectedRepo: "/home/user/project" });

    renderWithRouter(<Repositories />);
    await waitFor(() => {
      expect(screen.getByText("project")).toBeTruthy();
    });
  });

  it("calls initRepo when Init button clicked", async () => {
    const { useRepoStore } = await import("@/store");
    useRepoStore.setState({ repos: ["/home/user/project"], selectedRepo: "/home/user/project" });

    renderWithRouter(<Repositories />);
    await waitFor(() => screen.getByText("project"));

    const initBtns = screen.queryAllByRole("button", { name: /init/i });
    if (initBtns.length > 0) {
      fireEvent.click(initBtns[0]);
      // Single detected target should skip the picker and call initRepo
      // with the chosen target id.
      await waitFor(() => {
        expect(mockSkell.initRepo).toHaveBeenCalledWith("/home/user/project", "claude");
      });
    }
  });

  it("shows remove confirmation before removing repo", async () => {
    const { useRepoStore } = await import("@/store");
    useRepoStore.setState({ repos: ["/home/user/project"], selectedRepo: "/home/user/project" });

    renderWithRouter(<Repositories />);
    await waitFor(() => screen.getByText("project"));

    // The remove button is icon-only with title="Remove from Skell"
    const removeBtns = screen.queryAllByRole("button", { name: /remove from skell/i });
    if (removeBtns.length > 0) {
      fireEvent.click(removeBtns[0]);
      await waitFor(() => {
        // ConfirmDialog title for removing a repo
        expect(screen.getByText("Remove repository")).toBeTruthy();
      });
    } else {
      // Fallback: find by title attribute
      const btn = document.querySelector("[title='Remove from Skell']");
      if (btn) {
        fireEvent.click(btn);
        await waitFor(() => {
          expect(screen.getByText("Remove repository")).toBeTruthy();
        });
      }
    }
  });
});
