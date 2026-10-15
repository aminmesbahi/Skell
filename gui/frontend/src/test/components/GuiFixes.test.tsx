import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { RouterProvider, createMemoryRouter } from "react-router";
import { RouteError } from "@/components/RouteError";
import { SkillCard } from "@/components/SkillCard";

vi.mock("@/lib/skell");

function Boom(): never {
  throw new Error("kaboom");
}

describe("RouteError", () => {
  it("replaces a crashed page with a recoverable message", () => {
    const router = createMemoryRouter(
      [{ path: "/", errorElement: <RouteError />, element: <Boom /> }],
      { initialEntries: ["/"] },
    );
    const spy = vi.spyOn(console, "error").mockImplementation(() => {});
    render(<RouterProvider router={router} />);
    spy.mockRestore();
    expect(screen.getByRole("alert")).toBeTruthy();
    expect(screen.getByText(/something went wrong/i)).toBeTruthy();
    expect(screen.getByText("kaboom")).toBeTruthy();
    expect(screen.getByRole("button", { name: /reload/i })).toBeTruthy();
  });
});

describe("SkillCard lifecycle badge", () => {
  const base = { name: "demo", description: "A demo skill", license: "", metadata: { version: "", owner: "", scope: "", tags: "", source_repo: "", paths: "", disable_model_invocation: false, compatibility: "", license: "" } };

  it("does not claim 'Stable' for a skill that declares no lifecycle", () => {
    render(<SkillCard skill={{ ...base, metadata: { ...base.metadata, lifecycle: "" } } as never} installing={false} installed={false} canInstall onPreview={() => {}} onInstall={() => {}} />);
    expect(screen.queryByText("Stable")).toBeNull();
  });

  it("shows the declared lifecycle", () => {
    render(<SkillCard skill={{ ...base, metadata: { ...base.metadata, lifecycle: "experimental" } } as never} installing={false} installed={false} canInstall onPreview={() => {}} onInstall={() => {}} />);
    expect(screen.getByText("Experimental")).toBeTruthy();
  });
});
