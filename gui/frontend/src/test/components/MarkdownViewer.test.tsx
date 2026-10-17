import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import { MarkdownViewer, stripFrontMatter } from "@/components/MarkdownViewer";

describe("MarkdownViewer", () => {
  it("drops the YAML front matter of a SKILL.md", () => {
    const md = "---\nname: pdf\ndescription: Work with PDFs.\n---\n\n# PDF\n\nBody text.\n";
    expect(stripFrontMatter(md)).toBe("# PDF\n\nBody text.\n");
    expect(stripFrontMatter(md.replace(/\n/g, "\r\n"))).toBe("# PDF\r\n\r\nBody text.\r\n");
  });

  it("leaves documents without front matter, and later horizontal rules, alone", () => {
    expect(stripFrontMatter("# Title\n\n---\n\ntext\n")).toBe("# Title\n\n---\n\ntext\n");
    expect(stripFrontMatter("plain")).toBe("plain");
  });

  it("does not render front matter as a heading", () => {
    render(<MarkdownViewer content={"---\nname: pdf\ndescription: Work with PDFs.\n---\n# PDF\n"} />);
    expect(screen.getByRole("heading", { name: "PDF" })).toBeTruthy();
    expect(screen.queryByText(/name: pdf/)).toBeNull();
  });
});
