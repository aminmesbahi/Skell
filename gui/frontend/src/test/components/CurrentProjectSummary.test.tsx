import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import { CurrentProjectSummary } from "@/components/CurrentProjectSummary";

describe("CurrentProjectSummary", () => {
  it("opens the project when any part of the card is clicked", () => {
    const onOpen = vi.fn();
    render(
      <CurrentProjectSummary
        projectPath="D:\\amin\\Desktop\\New folder (2)"
        onOpen={onOpen}
      />
    );

    fireEvent.click(screen.getByRole("button", { name: "Open current project New folder (2)" }));

    expect(onOpen).toHaveBeenCalledOnce();
  });
});
