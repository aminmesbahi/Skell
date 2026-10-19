import "@testing-library/jest-dom/vitest";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { Outlet, RouterProvider, createMemoryRouter } from "react-router";
import { Sidebar } from "@/components/Sidebar";
import { Settings } from "@/pages/Settings";
import { useRepoStore, useUIStore } from "@/store";
import { renderWithRouter } from "@/test/utils";
import * as skell from "@/lib/skell";

vi.mock("@/lib/skell");
const mocks = vi.mocked(skell);

beforeEach(() => {
  vi.resetAllMocks();
  useRepoStore.setState({ repos: [], selectedRepo: "global", gettingStartedDismissed: true });
  useUIStore.setState({ tourOpen: false });
});

describe("Sidebar", () => {
  it("highlights the clicked item immediately, while its page is still loading", async () => {
    let release: () => void = () => undefined;
    const gate = new Promise<void>((resolve) => { release = resolve; });
    const router = createMemoryRouter([
      {
        path: "/",
        element: <><Sidebar /><Outlet /></>,
        children: [
          { index: true, element: <p>home page</p> },
          { path: "catalog", lazy: async () => { await gate; return { Component: () => <p>catalog page</p> }; } },
        ],
      },
    ], { initialEntries: ["/"] });
    render(<RouterProvider router={router} />);

    const catalog = screen.getByRole("link", { name: "Catalog" });
    expect(catalog.className).not.toContain("bg-brand-600/20");
    fireEvent.click(catalog);
    // Still on Home until the lazy route resolves, but the click already registers.
    expect(await screen.findByText("home page")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Catalog" }).className).toContain("bg-brand-600/20");

    await act(async () => { release(); await gate; });
    expect(await screen.findByText("catalog page")).toBeInTheDocument();
  });
});

describe("Settings", () => {
  it("shows placeholders instead of blanks while loading", async () => {
    mocks.getSkellVersion.mockReturnValue(new Promise(() => undefined));
    mocks.listSkillSources.mockReturnValue(new Promise(() => undefined));
    renderWithRouter(<Settings />);
    expect(screen.getByRole("status", { name: "Loading sources" })).toBeInTheDocument();
    expect(screen.getByRole("status", { name: "Checking version" })).toBeInTheDocument();
  });

  it("can restart the tour and bring back the getting started guide", async () => {
    mocks.getSkellVersion.mockResolvedValue("v1.0.0");
    mocks.listSkillSources.mockResolvedValue([]);
    vi.useFakeTimers();
    try {
      renderWithRouter(<Settings />);
      fireEvent.click(screen.getByRole("button", { name: "Show the getting started guide" }));
      expect(useRepoStore.getState().gettingStartedDismissed).toBe(false);

      fireEvent.click(screen.getByRole("button", { name: "Take the tour" }));
      await act(async () => { await vi.runAllTimersAsync(); });
      expect(useUIStore.getState().tourOpen).toBe(true);
    } finally {
      vi.useRealTimers();
    }
  });
});
