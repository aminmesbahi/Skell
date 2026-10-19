import type { DiagnosticEntry } from "@/lib/types";

export interface IssueDescription {
  title: string;
  /** What it means and what to do, without CLI jargon. */
  explanation: string;
}

/** Plain-language versions of `skell doctor` codes. Unknown codes fall back to the raw message. */
const DESCRIPTIONS: Record<string, IssueDescription> = {
  "no-manifest": {
    title: "This folder isn't set up for Skell",
    explanation: "There is no skell.toml here yet. Set the folder up from the Skills tab.",
  },
  "no-lockfile": {
    title: "Exact versions aren't recorded yet",
    explanation: "Skell hasn't written skell.lock, so teammates could get different versions. Running sync from the CLI creates it; commit it with skell.toml.",
  },
  "missing-dir": {
    title: "A skill's files are missing",
    explanation: "The lock file lists this skill but its folder is gone. Fix reinstalls it at the locked version.",
  },
  "hash-error": {
    title: "Couldn't check a skill's files",
    explanation: "Skell couldn't read the skill to verify it hasn't changed. Check that the folder is readable, then refresh.",
  },
  "locally-modified": {
    title: "A skill was edited locally",
    explanation: "Skell won't overwrite your edits. Review them with Changes on the Skills tab, or upgrade and choose to discard them.",
  },
  "untracked-skill": {
    title: "A skill isn't tracked by Skell",
    explanation: "It's in the skills folder but not listed in skell.toml. Run sync from the CLI to add it, or remove the folder.",
  },
  "stale-mirror": {
    title: "Another agent's copy is out of date",
    explanation: "A mirrored copy of a skill is missing or old. Fix refreshes every mirror.",
  },
  "validation-warnings": {
    title: "A skill has validation notes",
    explanation: "Open the skill and check its Validate tab for what could be improved.",
  },
};

export function describeIssue(issue: Pick<DiagnosticEntry, "code" | "message">): IssueDescription {
  return DESCRIPTIONS[issue.code] ?? { title: issue.message, explanation: "" };
}
