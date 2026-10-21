/** Skill names typed into a free-text box: comma or whitespace separated, de-duplicated. */
export function parseSkillList(text: string): string[] {
  return [...new Set(text.split(/[\s,]+/).map((s) => s.trim()).filter(Boolean))];
}

/** Bundle names follow the CLI rule: lowercase letters, digits and hyphens. */
export const BUNDLE_NAME_RX = /^[a-z0-9][a-z0-9-]*$/;
