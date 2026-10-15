/** Shows a skill's version, or the short source commit for the many skills that have none. */
export function versionLabel(version?: string | null, commit?: string | null): string {
  if (version) return version;
  return commit ? `@${commit.slice(0, 7)}` : "—";
}
