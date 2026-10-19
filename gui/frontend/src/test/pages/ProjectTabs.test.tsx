import "@testing-library/jest-dom/vitest";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router";
import { ProjectPageHeader } from "@/components/ProjectPageHeader";
import { ProjectHealthPage } from "@/pages/projects/ProjectHealthPage";
import { ProjectSkillsPage } from "@/pages/projects/ProjectSkillsPage";
import { describeIssue } from "@/lib/doctor";
import { useRepoStore } from "@/store";
import { createProjectId } from "@/lib/navigation";
import { mockInstalledSkill } from "@/test/fixtures";
import * as skell from "@/lib/skell";

vi.mock("@/lib/skell");
const mocks = vi.mocked(skell);
const route = (page: string) => `/projects/${createProjectId("/a")}/${page}`;
const mount = (element: React.ReactNode, page: string) =>
  render(<MemoryRouter initialEntries={[route(page)]}><Routes><Route path={`/projects/:projectId/${page}`} element={element} /></Routes></MemoryRouter>);

beforeEach(() => {
  vi.resetAllMocks();
  useRepoStore.setState({ repos: ["/a"], selectedRepo: "/a" });
  mocks.isRepoInitialized.mockResolvedValue(true);
  mocks.listInstalled.mockResolvedValue([]);
  mocks.getStatus.mockResolvedValue([]);
  mocks.listSupportedTargets.mockResolvedValue([{ id: "claude", displayName: "Claude Code", dir: ".claude", detected: true }]);
  mocks.detectRepoTargets.mockResolvedValue([]);
  mocks.targetFromInstalledPath.mockReturnValue("claude");
  mocks.validateSkills.mockResolvedValue([]);
  mocks.doctorCheck.mockResolvedValue([]);
});

describe("project tabs", () => {
  it("lists Skills, Agents, Health and Activity, marking the current one", () => {
    render(<MemoryRouter initialEntries={[route("health")]}><ProjectPageHeader projectPath="/a" title="Health for a" /></MemoryRouter>);
    const tabs = screen.getByRole("navigation", { name: "Project sections" });
    expect(within(tabs).getAllByRole("link").map((link) => link.textContent)).toEqual(["Skills", "Agents", "Health", "Activity"]);
    expect(within(tabs).getByRole("link", { name: "Health" })).toHaveAttribute("aria-current", "page");
    expect(within(tabs).getByRole("link", { name: "Skills" })).not.toHaveAttribute("aria-current");
    expect(within(tabs).getByRole("link", { name: "Agents" })).toHaveAttribute("href", route("agents"));
  });

  it("links the project name in the breadcrumb straight to its skills", () => {
    render(<MemoryRouter><ProjectPageHeader projectPath="/work/alpha" title="x" /></MemoryRouter>);
    expect(within(screen.getByRole("navigation", { name: "Breadcrumb" })).getByRole("link", { name: "alpha" })).toHaveAttribute("href", `/projects/${createProjectId("/work/alpha")}/skills`);
  });
});

describe("skills summary", () => {
  it("summarises the project under the title, replacing the old overview stats", async () => {
    mocks.listInstalled.mockResolvedValue([mockInstalledSkill({ name: "pdf" }), mockInstalledSkill({ name: "docx" })]);
    mocks.getStatus.mockResolvedValue([{ name: "pdf", installed: "", latest: "", status: "outdated", target: "claude" }, { name: "docx", installed: "", latest: "", status: "up-to-date", target: "claude" }]);
    mount(<ProjectSkillsPage />, "skills");
    expect(await screen.findByText("2 skills · 1 outdated")).toBeInTheDocument();
  });

  it("shows no summary for a project with no skills", async () => {
    mount(<ProjectSkillsPage />, "skills");
    await screen.findByRole("heading", { name: "Skills for a" });
    await waitFor(() => expect(screen.queryByText(/\d+ skills?/)).toBeNull());
  });
});

describe("health in plain language", () => {
  it("explains a doctor code and keeps the technical details one click away", async () => {
    mocks.doctorCheck.mockResolvedValue([{ severity: "warning", code: "untracked-skill", message: 'skill "x" is installed but not in manifest', hint: "run 'skell sync' to reconcile" }]);
    mount(<ProjectHealthPage />, "health");
    expect(await screen.findByText("A skill isn't tracked by Skell")).toBeInTheDocument();
    expect(screen.getByText(/not listed in skell\.toml/)).toBeInTheDocument();
    const details = screen.getByText("Technical details").closest("details")!;
    expect(details).not.toHaveAttribute("open");
    expect(within(details).getByText(/untracked-skill: skill "x"/)).toBeInTheDocument();
    expect(within(details).getByText("run 'skell sync' to reconcile")).toBeInTheDocument();
  });

  it("falls back to the raw message for a code it doesn't know", async () => {
    mocks.doctorCheck.mockResolvedValue([{ severity: "error", code: "brand-new", message: "something new broke" }]);
    mount(<ProjectHealthPage />, "health");
    expect(await screen.findByText("something new broke")).toBeInTheDocument();
  });
});

describe("describeIssue", () => {
  it.each(["no-manifest", "no-lockfile", "missing-dir", "hash-error", "locally-modified", "untracked-skill", "stale-mirror", "validation-warnings"])("has a friendly description for %s", (code) => {
    const described = describeIssue({ code, message: "raw" });
    expect(described.title).not.toBe("raw");
    expect(described.explanation.length).toBeGreaterThan(10);
  });
});
