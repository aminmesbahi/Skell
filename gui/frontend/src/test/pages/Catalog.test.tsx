import { beforeEach, describe, expect, it, vi } from "vitest";
import { fireEvent, screen, waitFor } from "@testing-library/react";
import { Catalog } from "@/pages/Catalog";
import { renderWithRouter } from "@/test/utils";
import { mockOkResult, mockRegistrySkill } from "@/test/fixtures";
import { useRepoStore, useUIStore } from "@/store";
import * as skell from "@/lib/skell";

vi.mock("@/lib/skell");

const mockSkell = skell as unknown as Record<string, ReturnType<typeof vi.fn>>;
const projectPath = "D:\\amin\\Desktop\\New folder (2)";

beforeEach(() => {
  vi.resetAllMocks();
  mockSkell.listRegistry.mockResolvedValue([]);
  mockSkell.listInstalled.mockResolvedValue([]);
  mockSkell.listInstalledGlobal.mockResolvedValue([]);
  mockSkell.installSkill.mockResolvedValue(mockOkResult());
  mockSkell.previewRegistrySkill.mockResolvedValue({
    found: true,
    source_type: "git",
    source_path: "",
    source_url: "https://github.com/WordPress/agent-skills",
    readme_content: "# Blueprint",
  });
  useRepoStore.setState({
    repos: [projectPath],
    selectedRepo: projectPath,
    recentProjects: [],
    sidebarCollapsed: false,
  });
  useUIStore.setState({ notifications: [] });
});

describe("Catalog", () => {
  it("opens the real skill preview without running install", async () => {
    mockSkell.listRegistry.mockResolvedValue([
      mockRegistrySkill({
        name: "blueprint",
        registry_alias: "wordpress-skills",
        registry_url: "https://github.com/WordPress/agent-skills",
      }),
    ]);

    renderWithRouter(<Catalog />, { initialEntries: ["/catalog"] });

    fireEvent.click(await screen.findByRole("button", { name: "Preview" }));

    await waitFor(() => {
      expect(mockSkell.previewRegistrySkill).toHaveBeenCalledWith(
        "wordpress-skills",
        "https://github.com/WordPress/agent-skills",
        "blueprint"
      );
      expect(screen.getByRole("dialog")).toBeTruthy();
      expect(mockSkell.installSkill).not.toHaveBeenCalled();
      expect(useUIStore.getState().notifications).toHaveLength(0);
    });
  });

  it("passes the catalog skill registry source to the real install", async () => {
    mockSkell.listRegistry.mockResolvedValue([
      mockRegistrySkill({
        name: "blueprint",
        registry_alias: "wordpress-skills",
        registry_url: "https://github.com/WordPress/agent-skills",
      }),
    ]);

    renderWithRouter(<Catalog />, { initialEntries: ["/catalog"] });
    fireEvent.click(await screen.findByRole("button", { name: "Install" }));

    await waitFor(() => {
      expect(mockSkell.installSkill).toHaveBeenCalledWith({
        skillName: "blueprint",
        repo: projectPath,
        registry: "wordpress-skills",
        registryURL: "https://github.com/WordPress/agent-skills",
      });
    });
  });

  it("hides Cobra usage text from install errors", async () => {
    mockSkell.listRegistry.mockResolvedValue([mockRegistrySkill({ name: "blueprint" })]);
    mockSkell.installSkill.mockResolvedValue({
      success: false,
      stdout: "",
      stderr: "registry not configured\nUsage: skell install <skill-name> [flags]",
    });

    renderWithRouter(<Catalog />, { initialEntries: ["/catalog"] });
    fireEvent.click(await screen.findByRole("button", { name: "Install" }));

    await waitFor(() => {
      const notification = useUIStore.getState().notifications.at(-1);
      expect(notification?.detail).toBe("registry not configured");
    });
  });
});
