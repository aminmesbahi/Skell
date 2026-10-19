import "@testing-library/jest-dom/vitest";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { Hint } from "@/components/Hint";
import { GettingStarted, guideProject } from "@/components/GettingStarted";
import { useRepoStore } from "@/store";
import { mockInstalledSkill, mockRegistrySkill } from "@/test/fixtures";
import * as skell from "@/lib/skell";

vi.mock("@/lib/skell");
const mocks = vi.mocked(skell);

const mount = (ui: React.ReactNode) => render(<MemoryRouter>{ui}</MemoryRouter>);
const reset = (state: Partial<ReturnType<typeof useRepoStore.getState>> = {}) =>
  useRepoStore.setState({ repos: [], selectedRepo: "global", gettingStartedDismissed: false, ...state });

beforeEach(() => {
  vi.resetAllMocks();
  mocks.skellPresent.mockResolvedValue(true);
  mocks.isRepoInitialized.mockResolvedValue(true);
  mocks.listRegistry.mockResolvedValue([]);
  mocks.listInstalled.mockResolvedValue([]);
  reset();
});

describe("Hint", () => {
  it("explains a term on click and closes on Escape", () => {
    mount(<Hint term="source">Source</Hint>);
    expect(screen.queryByRole("tooltip")).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: "What is source?" }));
    expect(screen.getByRole("tooltip")).toHaveTextContent(/git repository or local folder/i);
    fireEvent.keyDown(document, { key: "Escape" });
    expect(screen.queryByRole("tooltip")).toBeNull();
  });

  it("links the button to the tooltip for screen readers", () => {
    mount(<Hint term="pin" />);
    const button = screen.getByRole("button", { name: "What is pin?" });
    fireEvent.click(button);
    expect(button).toHaveAttribute("aria-describedby", screen.getByRole("tooltip").id);
  });
});

describe("guideProject", () => {
  it("prefers the selected project, otherwise the first one, ignoring global", () => {
    expect(guideProject(["/a", "/b"], "/b")).toBe("/b");
    expect(guideProject(["/a", "/b"], "global")).toBe("/a");
    expect(guideProject(["global"], "global")).toBe("");
  });
});

describe("GettingStarted", () => {
  it("asks for a project first when there are none", async () => {
    mount(<GettingStarted />);
    const step = await screen.findByRole("button", { name: "Add Project" });
    expect(step).toBeInTheDocument();
    expect(screen.getByText("1 of 5 steps done")).toBeInTheDocument();
    expect(screen.getByRole("listitem", { current: "step" })).toHaveTextContent("Add a project folder");
  });

  it("points at set-up when the project has no manifest", async () => {
    reset({ repos: ["/work/alpha"], selectedRepo: "/work/alpha" });
    mocks.isRepoInitialized.mockResolvedValue(false);
    mount(<GettingStarted />);
    const link = await screen.findByRole("link", { name: "Set up alpha" });
    expect(link).toHaveAttribute("href", expect.stringMatching(/^\/projects\/project-[a-z0-9]+\/skills$/));
    expect(screen.getByText("2 of 5 steps done")).toBeInTheDocument();
  });

  it("asks for a source once the project is set up but has no skills available", async () => {
    reset({ repos: ["/work/alpha"], selectedRepo: "/work/alpha" });
    mount(<GettingStarted />);
    expect(await screen.findByRole("link", { name: "Choose a source" })).toBeInTheDocument();
    expect(screen.getByText("3 of 5 steps done")).toBeInTheDocument();
  });

  it("asks for a first install once a source provides skills", async () => {
    reset({ repos: ["/work/alpha"], selectedRepo: "/work/alpha" });
    mocks.listRegistry.mockResolvedValue([mockRegistrySkill()]);
    mount(<GettingStarted />);
    expect(await screen.findByRole("link", { name: "Browse skills" })).toBeInTheDocument();
    expect(screen.getByText("4 of 5 steps done")).toBeInTheDocument();
  });

  it("says you're all set when everything is done, and Done hides it", async () => {
    reset({ repos: ["/work/alpha"], selectedRepo: "/work/alpha" });
    mocks.listRegistry.mockResolvedValue([mockRegistrySkill()]);
    mocks.listInstalled.mockResolvedValue([mockInstalledSkill()]);
    mount(<GettingStarted />);
    expect(await screen.findByText("You're all set.")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Hide the getting started guide" }));
    expect(useRepoStore.getState().gettingStartedDismissed).toBe(true);
    await waitFor(() => expect(screen.queryByRole("region", { name: "Getting started" })).toBeNull());
  });

  it("can be dismissed at any step and stays hidden", async () => {
    mount(<GettingStarted />);
    await screen.findByRole("button", { name: "Add Project" });
    fireEvent.click(screen.getByRole("button", { name: "Hide the getting started guide" }));
    expect(screen.queryByRole("heading", { name: "Getting started" })).toBeNull();
  });

  it("treats a failing CLI call as not done instead of crashing", async () => {
    reset({ repos: ["/work/alpha"], selectedRepo: "/work/alpha" });
    mocks.isRepoInitialized.mockRejectedValue(new Error("boom"));
    mount(<GettingStarted />);
    expect(await screen.findByRole("link", { name: "Set up alpha" })).toBeInTheDocument();
  });

  it("renders only the list when embedded, even if dismissed", async () => {
    reset({ gettingStartedDismissed: true });
    mount(<GettingStarted embedded />);
    const list = await screen.findByRole("list", { name: "Getting started" });
    expect(within(list).getByRole("button", { name: "Add Project" })).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "Getting started" })).toBeNull();
  });
});
