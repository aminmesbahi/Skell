import { describe, expect, it } from "vitest";
import { versionLabel } from "@/lib/format";

describe("versionLabel", () => {
  it("prefers the version", () => expect(versionLabel("1.2.0", "aeac41d4d00")).toBe("1.2.0"));
  it("falls back to the short source commit", () => expect(versionLabel("", "aeac41d4d005999432bb")).toBe("@aeac41d"));
  it("shows a dash when there is neither", () => {
    expect(versionLabel("", "")).toBe("—");
    expect(versionLabel(undefined, undefined)).toBe("—");
  });
});

import { plural } from "@/lib/format";

describe("plural", () => {
  it("only pluralises when needed", () => {
    expect(plural(0, "skill")).toBe("0 skills");
    expect(plural(1, "skill")).toBe("1 skill");
    expect(plural(3, "skill")).toBe("3 skills");
  });
});
