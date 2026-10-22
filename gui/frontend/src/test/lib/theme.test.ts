import { describe, expect, it } from "vitest";
// theme.css is read from disk: the bundler hands tests an empty string for CSS.
// The module name is built at run time because the node typings are not installed.
const fsModule = "node:fs";
const { readFileSync } = await import(/* @vite-ignore */ fsModule);
const theme: string = readFileSync("src/theme.css", "utf8") // tests run from gui/frontend;

// Every component source (the tests themselves are left out).
const sources = import.meta.glob(["../../**/*.{ts,tsx}", "!../../test/**"], {
  query: "?raw",
  import: "default",
  eager: true,
}) as Record<string, string>;

// Colours whose pale shades are written for the dark theme. On the light theme
// they must be given a dark shade in theme.css, or the text disappears on the
// white / pale-tinted panels (the "skell CLI was not found" banner did).
const PALE = /\btext-(amber|yellow|red|emerald|sky|cyan|blue|indigo|violet|purple|orange|teal)-(100|200|300|400)(?:\/\d+)?\b/g;
// Medium shades of the common status colours are mapped one by one in theme.css.
const MAPPED_400 = new Set(["red", "emerald", "amber", "yellow", "indigo", "blue", "violet", "purple"]);

describe("light theme", () => {
  it("finds the component sources", () => {
    expect(Object.keys(sources).length).toBeGreaterThan(20);
  });

  it("gives every pale text colour used in the app a readable light-theme shade", () => {
    const missing = new Set<string>();
    for (const source of Object.values(sources)) {
      for (const [, colour, shade] of source.matchAll(PALE)) {
        if (shade === "400" && MAPPED_400.has(colour)) continue;
        if (!theme.includes(`text-${colour}-${shade}`)) missing.add(`text-${colour}-${shade}`);
      }
    }
    expect([...missing].sort()).toEqual([]);
  });

  it("covers opacity variants such as text-amber-200/90", () => {
    expect(theme).toContain('[class*="text-amber-200"]');
  });
});
