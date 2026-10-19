import "@testing-library/jest-dom/vitest";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import { Catalog } from "@/pages/Catalog";
import { renderWithRouter } from "@/test/utils";
import { mockRegistrySkill } from "@/test/fixtures";
import { useRepoStore } from "@/store";
import * as skell from "@/lib/skell";

vi.mock("@/lib/skell");
const mocks = vi.mocked(skell);

const setRepos = (repos: string[], selectedRepo = "global") =>
  useRepoStore.setState({ repos, selectedRepo, recentProjects: [], sidebarCollapsed: false });

beforeEach(() => {
  vi.resetAllMocks();
  mocks.listRegistry.mockResolvedValue([]);
  mocks.listInstalled.mockResolvedValue([]);
  mocks.isRepoInitialized.mockResolvedValue(true);
  mocks.validateDirectory.mockResolvedValue([]);
  mocks.previewRegistrySkill.mockResolvedValue({ found: false, source_type: "git", source_path: "", source_url: "", readme_content: "" });
  mocks.listSupportedTargets.mockResolvedValue([
    { id: "claude", displayName: "Claude Code", dir: ".claude", detected: true },
    { id: "cursor", displayName: "Cursor", dir: ".cursor", detected: false },
  ]);
  mocks.activeRepoTarget.mockResolvedValue("claude");
  mocks.listCatalog.mockResolvedValue([{ id: "anthropic", name: "Anthropic", description: "Official skills", url: "https://github.com/anthropics/skills" }]);
  setRepos(["/work/alpha"], "/work/alpha");
});

describe("Catalog guidance", () => {
  it("leads with featured sources instead of an empty search when there are no skills", async () => {
    renderWithRouter(<Catalog />, { initialEntries: ["/catalog"] });
    expect(await screen.findByText("Start with a source")).toBeInTheDocument();
    expect(await screen.findByRole("button", { name: "Add Anthropic" })).toBeInTheDocument();
    expect(screen.queryByRole("textbox", { name: "Search skills" })).toBeNull();
  });

  it("shows search and filters once skills exist, with featured sources demoted", async () => {
    mocks.listRegistry.mockResolvedValue([mockRegistrySkill({ name: "pdf" })]);
    renderWithRouter(<Catalog />, { initialEntries: ["/catalog"] });
    expect(await screen.findByRole("textbox", { name: "Search skills" })).toBeInTheDocument();
    expect(screen.queryByText("Start with a source")).toBeNull();
  });

  it("preselects the only project", async () => {
    setRepos(["/work/alpha"], "global");
    renderWithRouter(<Catalog />, { initialEntries: ["/catalog"] });
    const select = await screen.findByLabelText("Install into");
    await waitFor(() => expect(select).toHaveValue("/work/alpha"));
  });

  it("says once, near the selector, why Install is unavailable", async () => {
    setRepos(["/work/alpha", "/work/beta"], "global");
    mocks.listRegistry.mockResolvedValue([mockRegistrySkill({ name: "pdf" }), mockRegistrySkill({ name: "docx" })]);
    renderWithRouter(<Catalog />, { initialEntries: ["/catalog"] });
    await screen.findByText("Choose a project to enable Install.");
    expect(screen.getAllByText("Choose a project to enable Install.")).toHaveLength(1);
  });

  it("asks for an agent when none is selected", async () => {
    mocks.activeRepoTarget.mockResolvedValue("");
    mocks.listRegistry.mockResolvedValue([mockRegistrySkill({ name: "pdf" })]);
    renderWithRouter(<Catalog />, { initialEntries: ["/catalog"] });
    expect(await screen.findByText("Choose at least one agent to enable Install.")).toBeInTheDocument();
  });

  it("offers Add Project when there are no projects at all", async () => {
    setRepos([]);
    renderWithRouter(<Catalog />, { initialEntries: ["/catalog"] });
    expect(await screen.findByText(/Add one to get started/)).toBeInTheDocument();
    expect(screen.getAllByRole("button", { name: "Add Project" }).length).toBeGreaterThan(0);
  });
});
