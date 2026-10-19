import "@testing-library/jest-dom/vitest";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";
import { ProductTour } from "@/components/ProductTour";
import { useRepoStore, useUIStore } from "@/store";

function Page({ targets }: { targets: string[] }) {
  return (
    <div>
      {targets.map((name) => <button key={name} data-tour={name}>{name}</button>)}
      <ProductTour />
    </div>
  );
}

const start = () => act(() => useUIStore.getState().startTour());
const dialog = () => screen.getByRole("dialog", { name: "Product tour" });

beforeEach(() => {
  useUIStore.setState({ tourOpen: false });
  useRepoStore.setState({ tourSeen: false });
});

describe("ProductTour", () => {
  it("renders nothing until started", () => {
    render(<Page targets={["nav"]} />);
    expect(screen.queryByRole("dialog")).toBeNull();
  });

  it("walks through only the steps whose targets are on screen", () => {
    render(<Page targets={["nav", "search", "replay"]} />);
    start();
    expect(dialog()).toHaveTextContent("Step 1 of 3");
    expect(dialog()).toHaveTextContent("Find your way around");
    fireEvent.click(screen.getByRole("button", { name: "Next" }));
    expect(dialog()).toHaveTextContent("Step 2 of 3");
    expect(dialog()).toHaveTextContent("Search from anywhere");
    fireEvent.click(screen.getByRole("button", { name: "Back" }));
    expect(dialog()).toHaveTextContent("Step 1 of 3");
  });

  it("finishing on the last step closes the tour and remembers it", () => {
    render(<Page targets={["nav", "replay"]} />);
    start();
    fireEvent.click(screen.getByRole("button", { name: "Next" }));
    expect(screen.queryByRole("button", { name: "Skip tour" })).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: "Finish" }));
    expect(screen.queryByRole("dialog")).toBeNull();
    expect(useUIStore.getState().tourOpen).toBe(false);
    expect(useRepoStore.getState().tourSeen).toBe(true);
  });

  it("Skip tour and Escape both close it and mark it seen", () => {
    const { unmount } = render(<Page targets={["nav", "search"]} />);
    start();
    fireEvent.click(screen.getByRole("button", { name: "Skip tour" }));
    expect(screen.queryByRole("dialog")).toBeNull();
    expect(useRepoStore.getState().tourSeen).toBe(true);
    unmount();

    useRepoStore.setState({ tourSeen: false });
    render(<Page targets={["nav", "search"]} />);
    start();
    fireEvent.keyDown(dialog(), { key: "Escape" });
    expect(screen.queryByRole("dialog")).toBeNull();
    expect(useRepoStore.getState().tourSeen).toBe(true);
  });

  it("supports the arrow keys", () => {
    render(<Page targets={["nav", "search"]} />);
    start();
    fireEvent.keyDown(dialog(), { key: "ArrowRight" });
    expect(dialog()).toHaveTextContent("Step 2 of 2");
    fireEvent.keyDown(dialog(), { key: "ArrowLeft" });
    expect(dialog()).toHaveTextContent("Step 1 of 2");
  });

  it("gives up quietly when nothing to point at is on the page", () => {
    render(<Page targets={[]} />);
    start();
    expect(screen.queryByRole("dialog")).toBeNull();
    expect(useUIStore.getState().tourOpen).toBe(false);
  });

  it("moves focus to Next so the keyboard works straight away", () => {
    render(<Page targets={["nav", "search"]} />);
    start();
    expect(screen.getByRole("button", { name: "Next" })).toHaveFocus();
  });
});
