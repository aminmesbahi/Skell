import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { ProjectsAttention } from "@/components/ProjectsAttention";
import { Sidebar } from "@/components/Sidebar";
import { ProjectPageHeader } from "@/components/ProjectPageHeader";
import { ValidationBadge } from "@/components/ValidationBadge";
import { useRepoStore } from "@/store";
import * as skell from "@/lib/skell";
import { mockInstalledSkill } from "@/test/fixtures";

vi.mock("@/lib/skell");
const mocks = vi.mocked(skell);

function mockMatchMedia(matches: boolean) {
  Object.defineProperty(window, "matchMedia", {
    configurable: true,
    writable: true,
    value: (query: string) => ({ matches, media: query, addEventListener: () => undefined, removeEventListener: () => undefined }),
  });
}

describe("ProjectsAttention", () => {
  beforeEach(() => { vi.resetAllMocks(); mocks.isRepoInitialized.mockResolvedValue(true); });

  it("lists only the projects that need something done, with counts", async () => {
    mocks.listInstalled.mockResolvedValue([mockInstalledSkill()]);
    mocks.getStatus.mockImplementation(async (repo: string) =>
      repo === "/work/alpha"
        ? [{ name: "pdf", installed: "", latest: "", status: "outdated" }, { name: "docx", installed: "", latest: "", status: "locally-modified" }]
        : [{ name: "pdf", installed: "", latest: "", status: "up-to-date" }]
    );
    render(<MemoryRouter><ProjectsAttention repos={["/work/alpha", "/work/beta"]} /></MemoryRouter>);

    const link = await screen.findByRole("link", { name: /alpha/ });
    expect(within(link).getByText("1 outdated")).toBeTruthy();
    expect(within(link).getByText("1 edited locally")).toBeTruthy();
    expect(screen.queryByRole("link", { name: /beta/ })).toBeNull();
  });

  it("says so when everything is up to date", async () => {
    mocks.listInstalled.mockResolvedValue([mockInstalledSkill()]);
    mocks.getStatus.mockResolvedValue([{ name: "pdf", installed: "", latest: "", status: "up-to-date" }]);
    render(<MemoryRouter><ProjectsAttention repos={["/work/alpha"]} /></MemoryRouter>);
    expect(await screen.findByText(/project is up to date/)).toBeTruthy();
  });

  it("flags a project that could not be checked instead of hiding it", async () => {
    mocks.listInstalled.mockRejectedValue(new Error("boom"));
    mocks.getStatus.mockRejectedValue(new Error("boom"));
    render(<MemoryRouter><ProjectsAttention repos={["/work/alpha"]} /></MemoryRouter>);
    expect(await screen.findByText("Could not check")).toBeTruthy();
  });

  it("says a folder is not set up yet rather than calling it an error", async () => {
    mocks.isRepoInitialized.mockResolvedValue(false);
    render(<MemoryRouter><ProjectsAttention repos={["/work/alpha"]} /></MemoryRouter>);
    expect(await screen.findByText("Not set up yet")).toBeTruthy();
    expect(screen.queryByText("Could not check")).toBeNull();
    expect(mocks.listInstalled).not.toHaveBeenCalled();
  });
});

describe("Sidebar", () => {
  beforeEach(() => {
    useRepoStore.setState({ repos: ["/work/alpha"], selectedRepo: "/work/alpha", sidebarCollapsed: false, recentProjects: [{ id: "p1", path: "/work/alpha", displayName: "alpha", lastOpenedAt: "2026-01-01T00:00:00Z" }] });
  });

  it("offers a single Add Project action", () => {
    mockMatchMedia(false);
    render(<MemoryRouter><Sidebar /></MemoryRouter>);
    expect(screen.getAllByRole("button", { name: /add project/i })).toHaveLength(1);
  });

  it("collapses to the icon rail in a narrow window", () => {
    mockMatchMedia(true);
    render(<MemoryRouter><Sidebar /></MemoryRouter>);
    expect(screen.queryByText("Catalog")).toBeNull();
    expect(screen.queryByTitle("Collapse sidebar")).toBeNull();
  });
});

describe("ProjectPageHeader", () => {
  it("navigates with the breadcrumb only (no duplicate back button)", () => {
    render(<MemoryRouter><ProjectPageHeader projectPath="/work/alpha" title="Skills for alpha" /></MemoryRouter>);
    expect(screen.queryByText(/back to projects/i)).toBeNull();
    const crumbs = screen.getByRole("navigation", { name: "Breadcrumb" });
    expect(within(crumbs).getByRole("link", { name: "Projects" })).toBeTruthy();
  });
});

describe("ValidationBadge", () => {
  it("explains the not-yet-validated state", async () => {
    render(<ValidationBadge />);
    const badge = await waitFor(() => screen.getByText("Not validated yet"));
    expect(badge.getAttribute("title")).toContain("Preview");
  });
});
