import {
  CheckCircle2,
  ArrowUp,
  Pin,
  AlertTriangle,
  Archive,
  GitBranch,
  HelpCircle,
  AlertCircle,
  Hash,
} from "lucide-react";
import type { SkillStatus, Lifecycle } from "@/lib/types";
import clsx from "clsx";

// ---------------------------------------------------------------------------
// Skill status badge
// ---------------------------------------------------------------------------

const STATUS_CONFIG: Record<
  SkillStatus,
  { label: string; classes: string; Icon: React.ElementType }
> = {
  "up-to-date": {
    label: "Up to date",
    classes: "bg-emerald-500/15 text-emerald-400 border-emerald-500/30",
    Icon: CheckCircle2,
  },
  outdated: {
    label: "Outdated",
    classes: "bg-amber-500/15 text-amber-400 border-amber-500/30",
    Icon: ArrowUp,
  },
  pinned: {
    label: "Pinned",
    classes: "bg-blue-500/15 text-blue-400 border-blue-500/30",
    Icon: Pin,
  },
  deprecated: {
    label: "Deprecated",
    classes: "bg-orange-500/15 text-orange-400 border-orange-500/30",
    Icon: AlertTriangle,
  },
  archived: {
    label: "Archived",
    classes: "bg-slate-500/15 text-slate-400 border-slate-500/30",
    Icon: Archive,
  },
  "locally-modified": {
    label: "Modified",
    classes: "bg-purple-500/15 text-purple-400 border-purple-500/30",
    Icon: GitBranch,
  },
  unknown: {
    label: "Unknown",
    classes: "bg-slate-600/15 text-slate-500 border-slate-600/30",
    Icon: HelpCircle,
  },
  "missing-metadata": {
    label: "No metadata",
    classes: "bg-yellow-500/15 text-yellow-400 border-yellow-500/30",
    Icon: AlertCircle,
  },
  unversioned: {
    label: "Unversioned",
    classes: "bg-yellow-600/15 text-yellow-500 border-yellow-600/30",
    Icon: Hash,
  },
};

interface SkillBadgeProps {
  status: SkillStatus;
  size?: "sm" | "md";
}

export function SkillBadge({ status, size = "md" }: SkillBadgeProps) {
  const cfg = STATUS_CONFIG[status] ?? STATUS_CONFIG["unknown"];
  const { Icon } = cfg;
  return (
    <span
      className={clsx(
        "inline-flex items-center gap-1 rounded-full border font-medium",
        cfg.classes,
        size === "sm" ? "px-1.5 py-0.5 text-xs" : "px-2 py-0.5 text-xs"
      )}
    >
      <Icon size={size === "sm" ? 10 : 11} />
      {cfg.label}
    </span>
  );
}

// ---------------------------------------------------------------------------
// Lifecycle badge
// ---------------------------------------------------------------------------

const LIFECYCLE_CONFIG: Record<
  Lifecycle,
  { label: string; classes: string }
> = {
  stable: { label: "Stable", classes: "bg-emerald-500/15 text-emerald-400 border-emerald-500/30" },
  experimental: { label: "Experimental", classes: "bg-yellow-500/15 text-yellow-400 border-yellow-500/30" },
  draft: { label: "Draft", classes: "bg-slate-500/15 text-slate-400 border-slate-500/30" },
  deprecated: { label: "Deprecated", classes: "bg-orange-500/15 text-orange-400 border-orange-500/30" },
  archived: { label: "Archived", classes: "bg-slate-600/15 text-slate-500 border-slate-600/30" },
};

export function LifecycleBadge({ lifecycle }: { lifecycle: Lifecycle }) {
  const cfg = LIFECYCLE_CONFIG[lifecycle] ?? LIFECYCLE_CONFIG["stable"];
  return (
    <span
      className={clsx(
        "inline-flex items-center rounded-full border px-2 py-0.5 text-xs font-medium",
        cfg.classes
      )}
    >
      {cfg.label}
    </span>
  );
}

// ---------------------------------------------------------------------------
// Scope badge (Global / Local)
// ---------------------------------------------------------------------------

export function ScopeBadge({ scope }: { scope: "global" | "local" }) {
  return scope === "global" ? (
    <span className="inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-xs font-medium bg-indigo-500/15 text-indigo-400 border-indigo-500/30">
      🌐 Global
    </span>
  ) : (
    <span className="inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-xs font-medium bg-teal-500/15 text-teal-400 border-teal-500/30">
      📁 Local
    </span>
  );
}

// ---------------------------------------------------------------------------
// Severity badge (Doctor diagnostics)
// ---------------------------------------------------------------------------

export function SeverityBadge({
  severity,
}: {
  severity: "error" | "warning" | "info";
}) {
  const cfg = {
    error: { label: "Error", classes: "bg-red-500/15 text-red-400 border-red-500/30", Icon: AlertCircle },
    warning: { label: "Warning", classes: "bg-amber-500/15 text-amber-400 border-amber-500/30", Icon: AlertTriangle },
    info: { label: "Info", classes: "bg-blue-500/15 text-blue-400 border-blue-500/30", Icon: CheckCircle2 },
  }[severity];
  const { Icon } = cfg;
  return (
    <span
      className={clsx(
        "inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-xs font-medium",
        cfg.classes
      )}
    >
      <Icon size={10} />
      {cfg.label}
    </span>
  );
}
