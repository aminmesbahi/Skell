import { useState, useEffect, FormEvent } from "react";
import { useParams, useLocation, useNavigate } from "react-router-dom";
import {
  GitPullRequest,
  ArrowLeft,
  CheckCircle,
  AlertCircle,
  Loader,
  ExternalLink,
  Info,
} from "lucide-react";
import { ReadSkillMetadata, ContributeMetadata } from "../../wailsjs/go/main/App";
import { main } from "../../wailsjs/go/models";

const LIFECYCLE_OPTIONS = [
  "draft",
  "experimental",
  "stable",
  "deprecated",
  "archived",
] as const;

interface RouteState {
  installedPath?: string;
  sourceRepo?: string;
}

export function ContributeMetadataPage() {
  const { skillName } = useParams<{ skillName: string }>();
  const location = useLocation();
  const navigate = useNavigate();

  const state = (location.state ?? {}) as RouteState;
  const installedPath = state.installedPath ?? "";
  const sourceRepo = state.sourceRepo ?? "";

  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [prURL, setPrURL] = useState<string | null>(null);
  const [submitError, setSubmitError] = useState<string | null>(null);

  const [description, setDescription] = useState("");
  const [tags, setTags] = useState("");
  const [lifecycle, setLifecycle] = useState<string>("stable");
  const [owner, setOwner] = useState("");
  const [githubToken, setGithubToken] = useState("");

  useEffect(() => {
    if (!installedPath) {
      setLoading(false);
      return;
    }
    ReadSkillMetadata(installedPath)
      .then((fields) => {
        setDescription(fields.description ?? "");
        setTags(fields.tags ?? "");
        setLifecycle(fields.lifecycle || "stable");
        setOwner(fields.owner ?? "");
      })
      .catch(() => {
        // no SKILL.md found — start with blank form
      })
      .finally(() => setLoading(false));
  }, [installedPath]);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    if (!sourceRepo) {
      setSubmitError("This skill has no source repository URL — cannot create a PR.");
      return;
    }
    setSubmitting(true);
    setSubmitError(null);
    try {
      const result = await ContributeMetadata(
        main.ContributeParams.createFrom({
          installedPath,
          sourceRepo,
          skillName: skillName ?? "",
          metadata: main.SkillMetadataFields.createFrom({ description, tags, lifecycle, owner }),
          githubToken,
        })
      );
      if (result.success) {
        setPrURL(result.prUrl);
      } else {
        setSubmitError(result.error ?? "Unknown error");
      }
    } finally {
      setSubmitting(false);
    }
  }

  const decodedName = skillName ? decodeURIComponent(skillName) : "";

  return (
    <div className="p-6 max-w-2xl mx-auto">
      {/* Header */}
      <div className="flex items-center gap-3 mb-6">
        <button
          onClick={() => navigate(-1)}
          className="p-1.5 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-white/5 transition-colors"
        >
          <ArrowLeft size={18} />
        </button>
        <div className="flex items-center gap-2">
          <GitPullRequest size={22} className="text-indigo-400" />
          <h1 className="text-xl font-semibold text-slate-100">
            Contribute Metadata
          </h1>
        </div>
      </div>

      {/* Context banner */}
      <div className="card mb-6 border border-indigo-500/20 bg-indigo-500/5">
        <div className="flex gap-3">
          <Info size={16} className="text-indigo-400 mt-0.5 shrink-0" />
          <div className="text-sm text-slate-300 space-y-1">
            <p>
              You are about to improve the metadata of{" "}
              <span className="font-semibold text-slate-100">{decodedName}</span>{" "}
              and open a Pull Request on the skill's upstream repository.
            </p>
            <p className="text-slate-400 text-xs">
              If you don't own the repository, Skell will automatically fork it
              first. A GitHub token with <code className="text-indigo-300">public_repo</code> scope is required.
            </p>
          </div>
        </div>
      </div>

      {/* Success state */}
      {prURL && (
        <div className="card border border-emerald-500/30 bg-emerald-500/5 mb-6">
          <div className="flex items-start gap-3">
            <CheckCircle size={20} className="text-emerald-400 shrink-0 mt-0.5" />
            <div>
              <p className="font-semibold text-emerald-300 mb-1">Pull Request opened!</p>
              <a
                href={prURL}
                target="_blank"
                rel="noreferrer"
                className="inline-flex items-center gap-1.5 text-sm text-indigo-400 hover:text-indigo-300 transition-colors"
              >
                {prURL}
                <ExternalLink size={12} />
              </a>
            </div>
          </div>
        </div>
      )}

      {/* Form */}
      {!prURL && (
        <form onSubmit={(e) => void handleSubmit(e)} className="space-y-5">
          {loading ? (
            <div className="flex justify-center py-10">
              <Loader size={24} className="animate-spin text-indigo-400" />
            </div>
          ) : (
            <>
              {/* Description */}
              <div>
                <label className="block text-sm font-medium text-slate-300 mb-1.5">
                  Description
                  <span className="ml-1 text-slate-500 font-normal">(top-level field)</span>
                </label>
                <textarea
                  rows={3}
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="A clear, one-sentence description of what this skill does."
                  className="w-full px-3 py-2 rounded-lg bg-[#0f1221] border border-[#1e2640] text-slate-200 placeholder:text-slate-600 text-sm focus:outline-none focus:border-indigo-500/60 resize-none"
                />
              </div>

              {/* Tags */}
              <div>
                <label className="block text-sm font-medium text-slate-300 mb-1.5">
                  Tags
                  <span className="ml-1 text-slate-500 font-normal">(comma-separated)</span>
                </label>
                <input
                  type="text"
                  value={tags}
                  onChange={(e) => setTags(e.target.value)}
                  placeholder="e.g. testing, react, typescript"
                  className="w-full px-3 py-2 rounded-lg bg-[#0f1221] border border-[#1e2640] text-slate-200 placeholder:text-slate-600 text-sm focus:outline-none focus:border-indigo-500/60"
                />
              </div>

              {/* Lifecycle */}
              <div>
                <label className="block text-sm font-medium text-slate-300 mb-1.5">
                  Lifecycle
                </label>
                <select
                  value={lifecycle}
                  onChange={(e) => setLifecycle(e.target.value)}
                  className="w-full px-3 py-2 rounded-lg bg-[#0f1221] border border-[#1e2640] text-slate-200 text-sm focus:outline-none focus:border-indigo-500/60"
                >
                  {LIFECYCLE_OPTIONS.map((lc) => (
                    <option key={lc} value={lc}>
                      {lc}
                    </option>
                  ))}
                </select>
              </div>

              {/* Owner */}
              <div>
                <label className="block text-sm font-medium text-slate-300 mb-1.5">
                  Owner
                  <span className="ml-1 text-slate-500 font-normal">(GitHub username or org)</span>
                </label>
                <input
                  type="text"
                  value={owner}
                  onChange={(e) => setOwner(e.target.value)}
                  placeholder="e.g. aminmesbahi"
                  className="w-full px-3 py-2 rounded-lg bg-[#0f1221] border border-[#1e2640] text-slate-200 placeholder:text-slate-600 text-sm focus:outline-none focus:border-indigo-500/60"
                />
              </div>

              {/* Divider */}
              <div className="border-t border-[#1a1f35] pt-5">
                <label className="block text-sm font-medium text-slate-300 mb-1.5">
                  GitHub Token
                  <span className="ml-1 text-slate-500 font-normal">(optional if git credentials are configured)</span>
                </label>
                <input
                  type="password"
                  value={githubToken}
                  onChange={(e) => setGithubToken(e.target.value)}
                  placeholder="ghp_xxxx  (needs public_repo scope)"
                  className="w-full px-3 py-2 rounded-lg bg-[#0f1221] border border-[#1e2640] text-slate-200 placeholder:text-slate-600 text-sm font-mono focus:outline-none focus:border-indigo-500/60"
                />
                <p className="mt-1.5 text-xs text-slate-500">
                  Leave blank to use the token stored in your git credential manager.
                </p>
              </div>

              {/* Source repo info */}
              {sourceRepo && (
                <p className="text-xs text-slate-500">
                  Target repository:{" "}
                  <span className="text-slate-400 font-mono">{sourceRepo}</span>
                </p>
              )}

              {/* Error */}
              {submitError && (
                <div className="flex items-start gap-2 text-sm text-red-400 bg-red-500/10 border border-red-500/20 rounded-lg px-3 py-2.5">
                  <AlertCircle size={15} className="shrink-0 mt-0.5" />
                  <span>{submitError}</span>
                </div>
              )}

              {/* Submit */}
              <div className="flex justify-end gap-3 pt-1">
                <button
                  type="button"
                  onClick={() => navigate(-1)}
                  className="px-4 py-2 text-sm rounded-lg text-slate-400 hover:text-slate-200 hover:bg-white/5 transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={submitting}
                  className="inline-flex items-center gap-2 px-5 py-2 text-sm font-medium rounded-lg bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed text-white transition-colors"
                >
                  {submitting ? (
                    <>
                      <Loader size={14} className="animate-spin" />
                      Opening PR…
                    </>
                  ) : (
                    <>
                      <GitPullRequest size={14} />
                      Open Pull Request
                    </>
                  )}
                </button>
              </div>
            </>
          )}
        </form>
      )}
    </div>
  );
}
