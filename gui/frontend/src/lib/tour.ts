/** One stop on the product tour. Steps whose target isn't on screen are skipped. */
export interface TourStep {
  id: string;
  /** CSS selector of the element to highlight (a data-tour attribute). */
  target: string;
  title: string;
  text: string;
}

export const TOUR_STEPS: TourStep[] = [
  {
    id: "nav",
    target: '[data-tour="nav"]',
    title: "Find your way around",
    text: "Home shows what needs attention. Projects lists your folders, Catalog is where you find skills, and Activity logs every change Skell makes.",
  },
  {
    id: "add-project",
    target: '[data-tour="add-project"]',
    title: "Add a project",
    text: "A project is a folder with your code. Skell installs skills into it for the agent you use there, such as Claude Code or Cursor.",
  },
  {
    id: "catalog",
    target: '[data-tour="nav-catalog"]',
    title: "Browse the Catalog",
    text: "Add a source (a git repository or folder of skills), then install skills from it into any project. Skills that ship scripts ask for your review first.",
  },
  {
    id: "guide",
    target: '[data-tour="guide"]',
    title: "Follow the checklist",
    text: "This guide tracks what's done and shows the next step. Hide it any time and bring it back from Home.",
  },
  {
    id: "attention",
    target: '[data-tour="attention"]',
    title: "See what needs you",
    text: "Once skills are installed, projects with updates or local edits are listed here, with a link straight to them.",
  },
  {
    id: "search",
    target: '[data-tour="search"]',
    title: "Search from anywhere",
    text: "Press Ctrl+K (or /) on any page to search skills.",
  },
  {
    id: "replay",
    target: '[data-tour="replay"]',
    title: "That's the tour",
    text: "Hover the ? icons next to unfamiliar words for a one-line explanation. Replay this tour whenever you like from here.",
  },
];
