import { beforeEach, describe, expect, it, vi } from "vitest";
import { act, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { useState } from "react";
import { MemoryRouter, Route, Routes, useNavigate } from "react-router";
import { ProjectSkillsPage } from "@/pages/projects/ProjectSkillsPage";
import { ProjectHealthPage } from "@/pages/projects/ProjectHealthPage";
import { ProjectAgentsPage } from "@/pages/projects/ProjectAgentsPage";
import { ProjectActivityPage } from "@/pages/projects/ProjectActivityPage";
import { ProjectOverview } from "@/pages/projects/ProjectOverview";
import { SkillDetail } from "@/pages/SkillDetail";
import { Catalog } from "@/pages/Catalog";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import { useRepoStore } from "@/store";
import { createProjectId, skillRoute } from "@/lib/navigation";
import { mockInstalledSkill, mockInfoResult, mockRegistrySkill, mockOkResult } from "@/test/fixtures";
import * as skell from "@/lib/skell";

vi.mock("@/lib/skell");
const mocks = vi.mocked(skell);
const route = (project: string, page = "skills") => `/projects/${createProjectId(project)}/${page}`;
function mount(element: React.ReactNode, page = "skills", url = route("/a", page)) {
  return render(<MemoryRouter initialEntries={[url]}><Routes><Route path={`/projects/:projectId/${page}`} element={element} /></Routes></MemoryRouter>);
}
beforeEach(() => {
  vi.resetAllMocks();
  useRepoStore.setState({ repos: ["/a", "/b"], selectedRepo: "/b" });
  mocks.listInstalled.mockResolvedValue([]);
  mocks.getStatus.mockResolvedValue([]);
  mocks.listSupportedTargets.mockResolvedValue([{ id: "claude", displayName: "Claude", dir: ".claude", detected: true }, { id: "cursor", displayName: "Cursor", dir: ".cursor", detected: true }]);
  mocks.detectRepoTargets.mockResolvedValue([]);
  mocks.activeRepoTarget.mockResolvedValue("claude");
  mocks.isRepoInitialized.mockResolvedValue(true);
  mocks.doctorCheck.mockResolvedValue([]);
  mocks.validateSkills.mockResolvedValue([]);
  mocks.readAuditLog.mockResolvedValue([]);
  mocks.listRegistry.mockResolvedValue([]);
  mocks.previewRegistrySkill.mockResolvedValue({ found: false, readme_content: "", source_path: "", source_url: "", source_type: "git" });
  mocks.targetFromInstalledPath.mockImplementation((path) => path.includes(".cursor") ? "cursor" : "claude");
  mocks.upgradeSkill.mockResolvedValue(mockOkResult());
  mocks.removeSkill.mockResolvedValue(mockOkResult());
  mocks.installSkill.mockResolvedValue(mockOkResult());
  mocks.listDirectory.mockResolvedValue([]);
  mocks.getInfo.mockResolvedValue(mockInfoResult());
});

describe("project identity and truthful states", () => {
  it.each([["skills", ProjectSkillsPage, "listInstalled"], ["health", ProjectHealthPage, "doctorCheck"], ["agents", ProjectAgentsPage, "detectRepoTargets"]] as const)("uses the URL project for %s", async (page, Component, method) => {
    mount(<Component />, page);
    await waitFor(() => expect(mocks[method]).toHaveBeenCalledWith("/a"));
    expect(mocks[method]).not.toHaveBeenCalledWith("/b");
  });
  it("resolves overview independently of selection", async () => {
    render(<MemoryRouter initialEntries={[`/projects/${createProjectId("/a")}`]}><Routes><Route path="/projects/:projectId" element={<ProjectOverview />} /></Routes></MemoryRouter>);
    await waitFor(() => expect(mocks.listInstalled).toHaveBeenCalledWith("/a"));
  });
  it("rejects an unknown project without loading another", () => {
    mount(<ProjectSkillsPage />, "skills", "/projects/unknown/skills");
    expect(screen.getByText("Project not found")).toBeTruthy();
    expect(mocks.listInstalled).not.toHaveBeenCalled();
  });
  it("shows validation findings and incomplete checks instead of healthy", async () => {
    mocks.doctorCheck.mockRejectedValue(new Error("doctor unavailable"));
    mocks.validateSkills.mockResolvedValue([{ name: "broken", target: "cursor", errors: 1, warnings: 1, findings: [{ severity: "error", category: "metadata", message: "Missing description" }] }]);
    mount(<ProjectHealthPage />, "health");
    expect(screen.queryByText("No issues reported.")).toBeNull();
    expect(await screen.findByText("Missing description")).toBeTruthy();
    expect(screen.getByText(/Checks incomplete/)).toBeTruthy();
    expect(screen.getByText(/1 errors · 1 warnings/)).toBeTruthy();
  });
  it("distinguishes failed skills loads from empty projects", async () => {
    mocks.listInstalled.mockRejectedValue(new Error("CLI unavailable"));
    mount(<ProjectSkillsPage />);
    expect((await screen.findByRole("alert")).textContent).toContain("Could not load");
    expect(screen.queryByText("No skills installed yet")).toBeNull();
  });
  it("opens repository import from the empty state", async () => {
    mount(<ProjectSkillsPage />);
    fireEvent.click(await screen.findByRole("button", { name: "Add from Repository" }));
    expect(screen.getByRole("dialog").textContent).toContain("Add Skill from Repository");
  });
  it("shows only actual audit events for this project", async () => {
    mocks.readAuditLog.mockResolvedValue([{ timestamp: "2026-09-01T12:00:00Z", action: "install", skill: "alpha", repo: "/a" }, { timestamp: "2026-09-02T12:00:00Z", action: "remove", skill: "other", repo: "/b" }]);
    mount(<ProjectActivityPage />, "activity");
    expect(await screen.findByText("alpha")).toBeTruthy();
    expect(screen.queryByText("other")).toBeNull();
  });
  it("ignores a slow response after navigating to a different project", async () => {
    let finish!: (value: Awaited<ReturnType<typeof skell.detectRepoTargets>>) => void;
    mocks.detectRepoTargets.mockImplementation((project) => project === "/a" ? new Promise((resolve) => { finish = resolve; }) : Promise.resolve([{ id: "cursor", displayName: "B agent", dir: ".cursor", detected: true }]));
    function Switch() { const navigate = useNavigate(); return <><button onClick={() => navigate(route("/b", "agents"))}>Switch</button><ProjectAgentsPage /></>; }
    mount(<Switch />, "agents");
    fireEvent.click(screen.getByText("Switch"));
    await screen.findByText("B agent");
    await act(async () => finish([{ id: "claude", displayName: "A agent", dir: ".claude", detected: true }]));
    expect(screen.queryByText("A agent")).toBeNull();
    expect(screen.getByText("B agent")).toBeTruthy();
  });
});

describe("installation actions", () => {
  it("uses target-specific detail links and excludes pinned skills from bulk upgrade", async () => {
    mocks.listInstalled.mockResolvedValue([mockInstalledSkill({ name: "alpha", installed_path: ".cursor/skills/alpha" }), mockInstalledSkill({ name: "pinned", pinned: true, installed_path: ".claude/skills/pinned" })]);
    mocks.getStatus.mockResolvedValue([{ name: "alpha", installed: "1", latest: "2", status: "outdated" }, { name: "pinned", installed: "1", latest: "2", status: "outdated" }]);
    mount(<ProjectSkillsPage />);
    expect((await screen.findByRole("link", { name: "alpha" })).getAttribute("href")).toBe(skillRoute("/a", "alpha", "cursor"));
    fireEvent.click(screen.getByRole("button", { name: "Upgrade all outdated (1)" }));
    fireEvent.click(within(screen.getByRole("dialog")).getByRole("button", { name: "Upgrade" }));
    await waitFor(() => expect(mocks.upgradeSkill).toHaveBeenCalledWith({ repo: "/a", skillName: "alpha", target: "cursor" }));
    expect(mocks.upgradeSkill).toHaveBeenCalledTimes(1);
  });
  it("keeps detail actions on the project and agent in the URL", async () => {
    mocks.getInfo.mockResolvedValue(mockInfoResult({ lock: mockInstalledSkill({ installed_path: ".cursor/skills/alpha" }) }));
    render(<MemoryRouter initialEntries={[skillRoute("/a", "alpha", "cursor")]}><Routes><Route path="/skills/:skillName" element={<SkillDetail />} /></Routes></MemoryRouter>);
    fireEvent.click(await screen.findByRole("button", { name: "Remove" }));
    fireEvent.click(within(screen.getByRole("dialog")).getByRole("button", { name: "Remove" }));
    await waitFor(() => expect(mocks.removeSkill).toHaveBeenCalledWith({ repo: "/a", skillName: "alpha", target: "cursor" }));
    expect(mocks.getInfo).toHaveBeenCalledWith("alpha", "/a", "cursor");
  });
  it("supports multiple agent installs and reports partial failure", async () => {
    mocks.listRegistry.mockResolvedValue([mockRegistrySkill({ name: "alpha" })]);
    mocks.installSkill.mockImplementation(async (opts) => opts.target === "cursor" ? { success: false, stdout: "", stderr: "disk full" } : mockOkResult());
    render(<MemoryRouter><Catalog /></MemoryRouter>);
    await screen.findByRole("button", { name: "Install" });
    fireEvent.click(screen.getByRole("checkbox", { name: "Cursor" }));
    await waitFor(() => expect((screen.getByRole("button", { name: "Install" }) as HTMLButtonElement).disabled).toBe(false));
    fireEvent.click(screen.getByRole("button", { name: "Install" }));
    expect(await screen.findByText("Cursor: Failed — disk full")).toBeTruthy();
    expect(screen.getByText("Claude: Installed")).toBeTruthy();
    expect(mocks.installSkill).toHaveBeenCalledTimes(2);
  });
});

it("traps dialog focus, closes on Escape, and restores the trigger", async () => {
  function Example() { const [open, setOpen] = useState(false); return <><button onClick={() => setOpen(true)}>Open</button><ConfirmDialog open={open} title="Remove?" description="Review removal" onCancel={() => setOpen(false)} onConfirm={() => {}} /></>; }
  render(<Example />);
  const trigger = screen.getByText("Open"); trigger.focus(); fireEvent.click(trigger);
  const dialog = screen.getByRole("dialog");
  const buttons = within(dialog).getAllByRole("button");
  buttons.at(-1)!.focus(); fireEvent.keyDown(buttons.at(-1)!, { key: "Tab" });
  expect(document.activeElement).toBe(buttons[0]);
  fireEvent.keyDown(buttons[0], { key: "Escape" });
  expect(screen.queryByRole("dialog")).toBeNull();
  expect(document.activeElement).toBe(trigger);
});
