/** One-sentence explanations of Skell's vocabulary, shown by <Hint term="…" />. */
export const GLOSSARY = {
  skill: {
    label: "Skill",
    text: "A folder with a SKILL.md file that teaches an AI agent how to do one task, such as working with PDFs.",
  },
  source: {
    label: "Source",
    text: "A git repository or local folder that contains skills. Add a source first, then install skills from it.",
  },
  agent: {
    label: "Agent",
    text: "The AI tool that reads your skills, such as Claude Code, Cursor or Copilot. Each one keeps skills in its own folder.",
  },
  setup: {
    label: "Set up",
    text: "Creates a small skell.toml file in the project so Skell can track which skills it installed. Commit it to share the setup with your team.",
  },
  pin: {
    label: "Pin",
    text: "Freezes a skill at its current version, so upgrades skip it until you unpin it.",
  },
  mirror: {
    label: "Also give skills to",
    text: "Gives another agent in this project a synced copy of every skill, so you can use Claude, Cursor and Copilot with the same set.",
  },
  outdated: {
    label: "Outdated",
    text: "The source has a newer version of this skill. Review the changes, then upgrade.",
  },
  "locally-modified": {
    label: "Edited locally",
    text: "You changed the installed files. Skell won't overwrite your edits unless you choose to.",
  },
} as const;

export type GlossaryTerm = keyof typeof GLOSSARY;
