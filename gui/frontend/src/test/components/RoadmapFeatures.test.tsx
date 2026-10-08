import { beforeEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router";
import * as skell from "@/lib/skell";
import { InstallReviewDialog } from "@/components/InstallReviewDialog";
import { FeaturedSources } from "@/components/FeaturedSources";
import { Catalog } from "@/pages/Catalog";
import { ProjectAgentsPage } from "@/pages/projects/ProjectAgentsPage";
import { ProjectHealthPage } from "@/pages/projects/ProjectHealthPage";
import { renderWithRouter } from "@/test/utils";
import { mockOkResult, mockRegistrySkill } from "@/test/fixtures";
import { useRepoStore, useUIStore } from "@/store";
import { createProjectId } from "@/lib/navigation";
import type { SkillReview } from "@/lib/types";

vi.mock("@/lib/skell");
const mockSkell = skell as unknown as Record<string, ReturnType<typeof vi.fn>>;

const review: SkillReview = {
  skill: mockRegistrySkill({ name: "pdf" }),
  registry: "anthropic",
  registry_url: "https://github.com/anthropics/skills",
  commit: "683bc88e56f3",
  files: ["SKILL.md", "scripts/fill.py"],
  scripts: ["scripts/fill.py"],
  allowed_tools: ["Bash"],
  links: ["https://example.com"],
  total_bytes: 2048,
  warnings: ["pre-approves unrestricted shell access (Bash)", "ships 1 script/executable file(s) the agent may run"],
};

beforeEach(() => {
  vi.resetAllMocks();
  useUIStore.setState({ notifications: [] });
});

describe("InstallReviewDialog", () => {
  it("shows warnings, tools and scripts and confirms", () => {
    const onConfirm = vi.fn();
    render(<InstallReviewDialog review={review} onConfirm={onConfirm} onCancel={vi.fn()} />);
    const dialog = screen.getByRole("dialog");
    expect(dialog.textContent).toContain("pre-approves unrestricted shell access");
    expect(dialog.textContent).toContain("scripts/fill.py");
    expect(dialog.textContent).toContain("683bc88");
    fireEvent.click(within(dialog).getByRole("button", { name: /install anyway/i }));
    expect(onConfirm).toHaveBeenCalled();
  });

  it("renders nothing without a review", () => {
    const { container } = render(<InstallReviewDialog review={null} onConfirm={vi.fn()} onCancel={vi.fn()} />);
    expect(container.innerHTML).toBe("");
  });
});

describe("FeaturedSources", () => {
  it("adds a catalog source to the project and marks configured ones", async () => {
    mockSkell.listCatalog.mockResolvedValue([
      { id: "anthropic", name: "Anthropic Agent Skills", url: "https://github.com/anthropics/skills", description: "" },
      { id: "openai", name: "OpenAI Codex Skills", url: "https://github.com/openai/skills", description: "" },
    ]);
    mockSkell.addSource.mockResolvedValue([]);
    const onAdded = vi.fn();
    render(<FeaturedSources repo="/proj" configuredUrls={["https://github.com/openai/skills.git"]} onAdded={onAdded} />);

    await screen.findByText("Anthropic Agent Skills");
    expect((screen.getByRole("button", { name: "OpenAI Codex Skills added" }) as HTMLButtonElement).disabled).toBe(true);
    fireEvent.click(screen.getByRole("button", { name: "Add Anthropic Agent Skills" }));
    await waitFor(() => expect(mockSkell.addSource).toHaveBeenCalledWith({ source: "anthropic", repo: "/proj", alias: "anthropic" }));
    expect(onAdded).toHaveBeenCalled();
  });

  it("adds a shared source when no project is chosen", async () => {
    mockSkell.listCatalog.mockResolvedValue([{ id: "anthropic", name: "Anthropic", url: "https://github.com/anthropics/skills", description: "" }]);
    mockSkell.addSkillSource.mockResolvedValue(undefined);
    render(<FeaturedSources />);
    fireEvent.click(await screen.findByRole("button", { name: "Add Anthropic" }));
    await waitFor(() => expect(mockSkell.addSkillSource).toHaveBeenCalledWith("anthropic", "https://github.com/anthropics/skills"));
  });
});

describe("Catalog install review", () => {
  beforeEach(() => {
    mockSkell.listRegistry.mockResolvedValue([mockRegistrySkill({ name: "pdf", registry_alias: "anthropic" })]);
    mockSkell.listInstalled.mockResolvedValue([]);
    mockSkell.isRepoInitialized.mockResolvedValue(true);
    mockSkell.listSupportedTargets.mockResolvedValue([{ id: "claude", displayName: "Claude", dir: ".claude", detected: true }]);
    mockSkell.activeRepoTarget.mockResolvedValue("claude");
    mockSkell.installSkill.mockResolvedValue(mockOkResult());
    mockSkell.listCatalog.mockResolvedValue([]);
    useRepoStore.setState({ repos: ["/proj"], selectedRepo: "/proj", recentProjects: [], sidebarCollapsed: false });
  });

  it("asks before installing a skill with scripts, and installs on confirm", async () => {
    mockSkell.reviewSkill.mockResolvedValue(review);
    mockSkell.reviewNeedsConfirmation.mockReturnValue(true);
    renderWithRouter(<Catalog />);
    const [install] = await screen.findAllByRole("button", { name: /install/i });
    fireEvent.click(install);

    const dialog = await screen.findByRole("dialog", { name: /review before installing pdf/i });
    expect(mockSkell.installSkill).not.toHaveBeenCalled();
    fireEvent.click(within(dialog).getByRole("button", { name: /install anyway/i }));
    await waitFor(() => expect(mockSkell.installSkill).toHaveBeenCalled());
  });

  it("cancelling the review installs nothing", async () => {
    mockSkell.reviewSkill.mockResolvedValue(review);
    mockSkell.reviewNeedsConfirmation.mockReturnValue(true);
    renderWithRouter(<Catalog />);
    const buttons = await screen.findAllByRole("button", { name: /install/i });
    fireEvent.click(buttons[0]);
    const dialog = await screen.findByRole("dialog", { name: /review before installing pdf/i });
    fireEvent.click(within(dialog).getByRole("button", { name: "Cancel" }));
    await waitFor(() => expect(screen.queryByRole("dialog", { name: /review before installing/i })).toBeNull());
    expect(mockSkell.installSkill).not.toHaveBeenCalled();
  });
});

function renderProjectPage(element: React.ReactElement, sub: string) {
  const path = createProjectId("/proj");
  return render(
    <MemoryRouter initialEntries={[`/projects/${path}/${sub}`]}>
      <Routes>
        <Route path="/projects/:projectId/*" element={element} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("Project agents page", () => {
  it("toggles mirroring for a non-primary agent", async () => {
    useRepoStore.setState({ repos: ["/proj"], selectedRepo: "/proj", recentProjects: [], sidebarCollapsed: false });
    mockSkell.detectRepoTargets.mockResolvedValue([{ id: "claude", displayName: "Claude Code", dir: ".claude", detected: true }]);
    mockSkell.listSupportedTargets.mockResolvedValue([
      { id: "claude", displayName: "Claude Code", dir: ".claude", detected: false },
      { id: "cursor", displayName: "Cursor", dir: ".cursor", detected: false },
    ]);
    mockSkell.activeRepoTarget.mockResolvedValue("claude");
    mockSkell.listMirrors.mockResolvedValue([]);
    mockSkell.setMirror.mockResolvedValue({ targets: ["cursor"], copied: [], removed: [] });
    renderProjectPage(<ProjectAgentsPage />, "agents");

    const toggle = await screen.findByRole("checkbox", { name: "Provide skills to Cursor" });
    expect(screen.getByText("Primary")).toBeTruthy();
    fireEvent.click(toggle);
    await waitFor(() => expect(mockSkell.setMirror).toHaveBeenCalledWith("/proj", "cursor", true));
  });
});

describe("Project health page", () => {
  it("offers an automatic fix for repairable issues", async () => {
    useRepoStore.setState({ repos: ["/proj"], selectedRepo: "/proj", recentProjects: [], sidebarCollapsed: false });
    mockSkell.doctorCheck.mockResolvedValue([{ severity: "warning", code: "stale-mirror", message: "mirror copy cursor/pdf is out of date" }]);
    mockSkell.validateSkills.mockResolvedValue([]);
    mockSkell.doctorFix.mockResolvedValue([]);
    renderProjectPage(<ProjectHealthPage />, "health");

    fireEvent.click(await screen.findByRole("button", { name: /fix 1 issue/i }));
    await waitFor(() => expect(mockSkell.doctorFix).toHaveBeenCalledWith("/proj"));
  });
});
