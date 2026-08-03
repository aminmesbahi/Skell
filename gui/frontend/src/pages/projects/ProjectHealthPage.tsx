import { useEffect, useMemo, useState } from "react";
import { useParams } from "react-router";
import { AlertTriangle, CheckCircle2, ShieldCheck } from "lucide-react";
import { useRepoStore } from "@/store";
import { getProjectDisplayName } from "@/lib/navigation";
import { doctorCheck, validateSkills } from "@/lib/skell";
import type { DiagnosticEntry, SkillValidationResult } from "@/lib/types";

export function ProjectHealthPage() {
  const { projectId: _projectId } = useParams();
  const { repos, selectedRepo } = useRepoStore();

  const projectPath = useMemo(() => {
    if (selectedRepo && selectedRepo !== "global") return selectedRepo;
    return repos[0] ?? "";
  }, [repos, selectedRepo]);

  const [issues, setIssues] = useState<DiagnosticEntry[]>([]);
  const [validations, setValidations] = useState<SkillValidationResult[]>([]);

  useEffect(() => {
    async function loadHealth() {
      if (!projectPath) return;
      const [diagnostics, validationResults] = await Promise.all([
        doctorCheck(projectPath).catch(() => [] as DiagnosticEntry[]),
        validateSkills(projectPath, "", false).catch(() => [] as SkillValidationResult[]),
      ]);
      setIssues(diagnostics);
      setValidations(validationResults);
    }

    void loadHealth();
  }, [projectPath]);

  const errors = issues.filter((issue) => issue.severity === "error").length;
  const warnings = issues.filter((issue) => issue.severity === "warning").length;

  return (
    <div className="mx-auto max-w-6xl px-6 py-8 space-y-6">
      <div className="card">
        <h2 className="text-2xl font-bold text-slate-200">Health for {getProjectDisplayName(projectPath)}</h2>
        <p className="mt-2 text-sm text-slate-400">Validation results and doctor findings for this project.</p>
      </div>

      <div className="card flex items-center gap-3">
        <ShieldCheck size={18} className={errors > 0 ? "text-red-400" : "text-emerald-400"} />
        <div>
          <p className="font-medium text-slate-200">{validations.length} skills checked</p>
          <p className="text-sm text-slate-500">{errors} error{errors !== 1 ? "s" : ""} and {warnings} warning{warnings !== 1 ? "s" : ""}</p>
        </div>
      </div>

      <div className="space-y-3">
        {issues.length === 0 ? (
          <div className="card text-sm text-slate-500">No issues reported.</div>
        ) : (
          issues.map((issue, index) => (
            <div key={`${issue.code}-${index}`} className="card flex items-start gap-3">
              {issue.severity === "error" ? <AlertTriangle size={16} className="text-red-400 mt-0.5" /> : <CheckCircle2 size={16} className="text-amber-400 mt-0.5" />}
              <div>
                <p className="font-medium text-slate-200">{issue.code}</p>
                <p className="mt-1 text-sm text-slate-500">{issue.message}</p>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
