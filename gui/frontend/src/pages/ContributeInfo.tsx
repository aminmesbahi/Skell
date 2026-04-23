import { useNavigate } from "react-router-dom";
import { GitPullRequest, Package, ArrowRight, GitFork, FilePen, GitMerge } from "lucide-react";

const STEPS = [
  {
    icon: Package,
    title: "Open Installed Skills",
    desc: 'Go to "Installed Skills" and find a skill that has missing or incorrect metadata.',
  },
  {
    icon: FilePen,
    title: "Click Fix Metadata",
    desc: 'Click the indigo "Fix Metadata" button in the skill row or on the skill detail page.',
  },
  {
    icon: GitFork,
    title: "Fill in the form",
    desc: "Edit description, tags, lifecycle and owner. Enter the upstream GitHub URL if not pre-filled.",
  },
  {
    icon: GitMerge,
    title: "Open Pull Request",
    desc: "Skell forks the repo if needed, commits the changes, and opens a PR on your behalf.",
  },
];

export function ContributeInfo() {
  const navigate = useNavigate();

  return (
    <div className="p-6 max-w-2xl mx-auto space-y-8">
      {/* Header */}
      <div className="flex items-center gap-3">
        <div className="w-10 h-10 rounded-xl bg-indigo-500/10 flex items-center justify-center shrink-0">
          <GitPullRequest size={20} className="text-indigo-400" />
        </div>
        <div>
          <h1 className="text-xl font-semibold text-slate-100">Contribute Metadata</h1>
          <p className="text-sm text-slate-500 mt-0.5">
            Improve skill metadata and open a PR on the upstream repository.
          </p>
        </div>
      </div>

      {/* Why */}
      <div className="card border border-indigo-500/15 bg-indigo-500/5">
        <p className="text-sm text-slate-300 leading-relaxed">
          Many skills in the registry are missing a description, tags, lifecycle stage, or owner.
          This makes them harder to discover and manage. You can fix that in seconds — Skell handles
          the fork, branch, commit, and PR automatically.
        </p>
      </div>

      {/* Steps */}
      <div className="space-y-3">
        {STEPS.map((step, i) => (
          <div key={i} className="flex items-start gap-4">
            <div className="flex flex-col items-center">
              <div className="w-8 h-8 rounded-lg bg-[#1a1f35] flex items-center justify-center shrink-0">
                <step.icon size={15} className="text-indigo-400" />
              </div>
              {i < STEPS.length - 1 && (
                <div className="w-px h-6 bg-[#1e2540] mt-1" />
              )}
            </div>
            <div className="pt-1">
              <p className="text-sm font-medium text-slate-200">{step.title}</p>
              <p className="text-xs text-slate-500 mt-0.5">{step.desc}</p>
            </div>
          </div>
        ))}
      </div>

      {/* CTA */}
      <button
        onClick={() => navigate("/skills")}
        className="inline-flex items-center gap-2 px-5 py-2.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-medium transition-colors"
      >
        <Package size={15} />
        Browse Installed Skills
        <ArrowRight size={14} />
      </button>
    </div>
  );
}
