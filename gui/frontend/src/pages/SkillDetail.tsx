import { SkillComparison } from "@/components/SkillComparison";
import { LoadState, MissingProject } from "@/components/LoadState";
import { lazy, Suspense, useEffect, useState, useCallback, useRef } from "react";
import { useParams, useLocation, useNavigate, Link } from "react-router";
import {
  ArrowLeft,
  Package,
  ArrowUp,
  Pin,
  PinOff,
  Trash2,
  RefreshCw,
  FileText,
  Code,
  ChevronRight,
  FolderOpen,
  GitPullRequest,
  ShieldCheck,
} from "lucide-react";
import { useRepoStore, useUIStore } from "@/store";
import {
  getInfo, openSkillFolder, targetFromInstalledPath, getStatus,
  upgradeSkill,
  removeSkill,
  pinSkill,
  unpinSkill,
  readFileContent,
  listDirectory,
  validateSkills,
  getGlobalRootDir,
} from "@/lib/skell";
import type { InfoResult, FileEntry, SkillValidationResult } from "@/lib/types";
import { SkillBadge, LifecycleBadge } from "@/components/Badges";
import { MarkdownViewer } from "@/components/MarkdownViewer";
import { ValidationReport } from "@/components/ValidationReport";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import { createProjectId } from "@/lib/navigation";

const CodeViewer = lazy(async () => {
  const mod = await import("@/components/CodeViewer");
  return { default: mod.CodeViewer };
});

type Tab = "info" | "readme" | "files" | "validate";

export function SkillDetail() {
  const location = useLocation();
  return <SkillDetailContent key={location.pathname + location.search + JSON.stringify(location.state)} />;
}

function SkillDetailContent() {
  const { skillName } = useParams<{ skillName: string }>();
  const location = useLocation();
  const navigate = useNavigate();
  const { selectedRepo, repos } = useRepoStore();
  const { notify } = useUIStore();

  const state = location.state as { repo?: string; from?: string; breadcrumb?: string } | undefined;
  const params = new URLSearchParams(location.search);
  const repo = params.get("repo") ?? state?.repo ?? selectedRepo;
  const target = params.get("target") || undefined;
  const invalidRepo = !repo || (params.has("repo") && repo !== "global" && !repos.includes(repo));
  const decoded = skillName ?? "";

  const [info, setInfo] = useState<InfoResult | null>(null);
  const [loading, setLoading] = useState(true);
  const [tab, setTab] = useState<Tab>(params.get("tab") === "validate" ? "validate" : "info");
  const [files, setFiles] = useState<FileEntry[]>([]);
  const [selectedFile, setSelectedFile] = useState<string | null>(null);
  const [fileContent, setFileContent] = useState<string>("");
  const [loadingFile, setLoadingFile] = useState(false);
  const [removing, setRemoving] = useState(false);
  const [acting, setActing] = useState(false);
  const [expandedDirs, setExpandedDirs] = useState<Record<string, FileEntry[]>>({});
  const [loadingDir, setLoadingDir] = useState<string | null>(null);
  const [validation, setValidation] = useState<SkillValidationResult | null>(null);
  const [validating, setValidating] = useState(false);
  const [validationError, setValidationError] = useState<string | null>(null);

  const [folder, setFolder] = useState("");
  const [compare, setCompare] = useState(false);
  const [loadError, setLoadError] = useState<string>();
  const request = useRef(0);
  const loadInfo = useCallback(async () => {
    if (!decoded || invalidRepo) return;
    const generation = ++request.current;
    setLoadError(undefined);
    setLoading(true);
    try {
      const result = await getInfo(decoded, repo, target);
      if (generation !== request.current) return;
      if (!result) { setInfo(null); return; }
      const statuses = await getStatus(repo, target);
      if (generation !== request.current) return;
      const status = statuses.find((entry) => entry.name === decoded);
      if (status) result.status = status.status;
      setInfo(result);

      // Try to list skill files — installed_path may be relative to the repo root.
      if (result?.lock?.installed_path) {
        const installedPath = result.lock.installed_path;
        const isAbsolute = /^[A-Za-z]:[\\/]|^\//.test(installedPath);
        let absPath: string;
        if (isAbsolute || repo === "global") {
          absPath = installedPath;
        } else {
          // Detect the OS path separator from the repo path itself so this
          // works on both Windows (backslash) and Mac/Linux (forward slash).
          const sep = /^[A-Za-z]:/.test(repo) || repo.includes("\\") ? "\\" : "/";
          const rel = installedPath.replace(/[/\\]/g, sep);
          absPath = repo.replace(/[\\/]$/, "") + sep + rel;
        }
        setFolder(absPath);
        const entries = await listDirectory(absPath);
        if (generation === request.current) setFiles(entries);
      }
    } catch (error) { if (generation === request.current) setLoadError(String(error)); } finally {
      if (generation === request.current) setLoading(false);
    }
  }, [decoded, repo, target, invalidRepo]);

  useEffect(() => {
    void loadInfo();
    return () => { request.current++; };
  }, [loadInfo]);
  useEffect(() => { if (tab === "validate" && !invalidRepo) void runValidation(); }, [repo, target]);

  useEffect(() => {
    if (tab === "readme" && !fileContent && !loadingFile) {
      const md = files.find(
        (f) => f.name.toLowerCase() === "skill.md" || f.name.toLowerCase() === "readme.md"
      );
      if (md) void selectFile(md);
    }
    // Only auto-load SKILL.md when tab changes to readme or files list updates.
    // selectFile and fileContent are intentionally excluded from deps.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [files, tab]);

  async function selectFile(entry: FileEntry) {
    if (entry.is_dir) return;
    setSelectedFile(entry.path);
    setLoadingFile(true);
    try {
      const content = await readFileContent(entry.path);
      setFileContent(content);
    } catch {
      setFileContent("// Could not read file");
    } finally {
      setLoadingFile(false);
    }
  }

  async function toggleDir(entry: FileEntry) {
    if (!entry.is_dir) return;
    if (entry.path in expandedDirs) {
      setExpandedDirs((prev) => {
        const next = { ...prev };
        delete next[entry.path];
        return next;
      });
      return;
    }
    setLoadingDir(entry.path);
    try {
      const children = await listDirectory(entry.path).catch(() => [] as FileEntry[]);
      setExpandedDirs((prev) => ({ ...prev, [entry.path]: children }));
    } finally {
      setLoadingDir(null);
    }
  }

  async function runValidation() {
    setValidating(true);
    setValidationError(null);
    try {
      const repoArg = repo === "global" ? await getGlobalRootDir() : repo;
      // full=true → include offline content & contamination analysis.
      const results = await validateSkills(repoArg, decoded, true, target);
      if (!results.length) throw new Error("No validation result was returned.");
      setValidation(results[0]);
    } catch (e) {
      setValidationError(String(e));
    } finally {
      setValidating(false);
    }
  }

  async function handleUpgrade() {
    if (invalidRepo || !info?.lock) return;
    setActing(true);
    try {
      const result = await upgradeSkill({ skillName: decoded, repo, target: target || targetFromInstalledPath(info?.lock?.installed_path ?? "") });
      if (result.success) {
        notify({ kind: "success", title: `Upgraded ${decoded}` });
        void loadInfo();
      } else {
        notify({ kind: "error", title: "Upgrade failed", detail: result.stderr });
      }
    } catch (error) { notify({ kind: "error", title: "Operation failed", detail: String(error) }); } finally {
      setActing(false);
    }
  }

  async function handlePin() {
    if (invalidRepo || !info?.lock) return;
    setActing(true);
    const isPinned = !!info?.lock?.pinned;
    try {
      const result = await (isPinned
        ? unpinSkill({ skillName: decoded, repo, target: target || targetFromInstalledPath(info?.lock?.installed_path ?? "") })
        : pinSkill({ skillName: decoded, repo, target: target || targetFromInstalledPath(info?.lock?.installed_path ?? "") }));
      if (result.success) {
        notify({ kind: "success", title: isPinned ? `Unpinned ${decoded}` : `Pinned ${decoded}` });
        void loadInfo();
      } else {
        notify({ kind: "error", title: "Operation failed", detail: result.stderr });
      }
    } catch (error) { notify({ kind: "error", title: "Operation failed", detail: String(error) }); } finally {
      setActing(false);
    }
  }

  async function handleRemove() {
    if (invalidRepo || !info?.lock) return;
    setRemoving(false);
    setActing(true);
    try {
      const result = await removeSkill({ skillName: decoded, repo, target: target || targetFromInstalledPath(info?.lock?.installed_path ?? "") });
      if (result.success) {
        notify({ kind: "success", title: `Removed ${decoded}` });
        navigate(backTarget);
      } else {
        notify({ kind: "error", title: "Remove failed", detail: result.stderr });
      }
    } catch (error) { notify({ kind: "error", title: "Operation failed", detail: String(error) }); } finally {
      setActing(false);
    }
  }

  const status = info?.lock?.pinned ? "pinned" : info?.status ?? "unknown";
  const isOutdated = info?.status === "outdated";
  const isPinned = !!info?.lock?.pinned;
  const backTarget = state?.from === "catalog"
    ? "/catalog"
    : state?.from === "project" && repo && repo !== "global"
      ? `/projects/${createProjectId(repo)}/skills`
      : state?.from === "health"
        ? "/projects"
        : state?.from === "activity"
          ? "/activity"
          : repo && repo !== "global"
            ? `/projects/${createProjectId(repo)}/skills`
            : "/projects";

  const skillMd = files.find(
    (f) => f.name.toLowerCase() === "skill.md" || f.name.toLowerCase() === "readme.md"
  );

  if (invalidRepo) return <MissingProject />;
  return (
    <div className="p-6 space-y-5 max-w-5xl mx-auto">
      {/* Back + breadcrumb */}
      <div className="flex flex-wrap items-center gap-3">
        <button onClick={() => navigate(backTarget)} className="btn-ghost text-xs">
          <ArrowLeft size={13} />
          Back
        </button>
        {repo && repo !== "global" && (
          <div className="flex items-center gap-2 text-xs text-slate-500">
            <Link to="/projects" className="text-slate-400 hover:text-slate-200 transition-colors">Projects</Link>
            <ChevronRight size={12} />
            <Link to={`/projects/${createProjectId(repo)}`} className="text-slate-300 hover:text-slate-100 transition-colors">
              {repo.split(/[\\/]/).filter(Boolean).pop() ?? repo}
            </Link>
            <ChevronRight size={12} />
            <span className="text-slate-300">{decoded}</span>
          </div>
        )}
      </div>

      <LoadState loading={false} error={loadError} retry={() => void loadInfo()} />
      {loading ? (
        <div className="flex justify-center py-20">
          <div className="spinner w-8 h-8" />
        </div>
      ) : !info && loadError ? null : !info ? (
        <div className="card text-center py-12 text-slate-500">Skill not found.</div>
      ) : (
        <>
          {/* Skill header */}
          <div className="card">
            <div className="flex flex-col gap-4">
              <div className="flex items-start gap-4">
                <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-brand-600/15">
                  <Package size={24} className="text-brand-400" />
                </div>
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-center gap-2.5">
                    <h1 className="text-3xl font-bold leading-tight text-slate-100">{decoded}</h1>
                    <SkillBadge status={status as typeof status} />
                    {info.skill?.metadata?.lifecycle && (
                      <LifecycleBadge lifecycle={info.skill.metadata.lifecycle} />
                    )}
                  </div>
                  {info.skill?.description && (
                    <p className="mt-2 max-w-3xl text-sm leading-6 text-slate-300">{info.skill.description}</p>
                  )}
                  <div className="mt-3 grid gap-1 text-sm text-slate-500 md:grid-cols-2 xl:grid-cols-3">
                    {info.lock?.version && (
                      <p>
                        Version: <span className="font-mono text-slate-300">{info.lock.version}</span>
                      </p>
                    )}
                    {info.skill?.metadata?.version && (
                      <p>
                        Latest: <span className="font-mono text-slate-300">{info.skill.metadata.version}</span>
                      </p>
                    )}
                    {info.lock?.registry && (
                      <p>
                        Registry: <span className="text-slate-300">{info.lock.registry}</span>
                      </p>
                    )}
                    {info.lock?.installed_at && (
                      <p>
                        Installed on: <span className="text-slate-300">{new Date(info.lock.installed_at).toLocaleDateString()}</span>
                      </p>
                    )}
                  </div>
                </div>
              </div>

              <div className="border-t border-[var(--palette-1e2540)] pt-3">
                <div className="flex flex-wrap items-center gap-1.5">
                  {folder && (
                    <>
                      <button className="btn-subtle" onClick={() => void openSkillFolder(folder).catch((e) => notify({ kind: "error", title: "Could not open folder", detail: String(e) }))}>
                        Open folder
                      </button>
                      <button className="btn-subtle" onClick={() => void navigator.clipboard.writeText(folder).catch((e) => notify({ kind: "error", title: "Could not copy path", detail: String(e) }))}>
                        Copy path
                      </button>
                      <button className="btn-subtle" onClick={() => setCompare(!compare)}>
                        Compare with registry
                      </button>
                    </>
                  )}
                  <button aria-label="Refresh skill" title="Refresh skill" onClick={() => void loadInfo()} className="btn-subtle" disabled={acting}>
                    <RefreshCw size={12} /> Refresh
                  </button>
                  <button
                    onClick={() =>
                      navigate(`/contribute/${encodeURIComponent(decoded)}`, {
                        state: {
                          installedPath: info.lock?.installed_path ?? "",
                          sourceRepo: info.lock?.source_repo ?? info.skill?.metadata?.source_repo ?? "",
                          registryAlias: info.lock?.registry ?? "",
                        },
                      })
                    }
                    className="btn-subtle text-indigo-300 hover:bg-indigo-500/10 hover:text-indigo-200"
                    title="Contribute metadata improvement"
                  >
                    <GitPullRequest size={12} />
                    Fix Metadata
                  </button>
                  {isOutdated && (
                    <button onClick={() => void handleUpgrade()} disabled={acting} className="btn-primary h-8 px-3 text-xs">
                      <ArrowUp size={12} />
                      Upgrade
                    </button>
                  )}
                  <button onClick={() => void handlePin()} disabled={acting || invalidRepo || !info?.lock} className="btn-subtle">
                    {isPinned ? <PinOff size={12} /> : <Pin size={12} />}
                    {isPinned ? "Unpin" : "Pin"}
                  </button>
                  <button
                    onClick={() => setRemoving(true)}
                    disabled={acting || invalidRepo || !info?.lock}
                    className="btn-subtle text-red-300 hover:bg-red-500/10 hover:text-red-200"
                  >
                    <Trash2 size={12} />
                    Remove
                  </button>
                </div>
              </div>
            </div>
          </div>

          {compare && folder && <SkillComparison name={decoded} registry={info.lock?.registry ?? ""} url={info.lock?.source_repo ?? ""} folder={folder} />}
          {/* Tabs */}
          <div className="flex items-center gap-1 border-b border-[var(--palette-1e2540)]">
            {(["info", "readme", "files", "validate"] as Tab[]).map((t) => (
              <button
                key={t}
                onClick={() => {
                  setTab(t);
                  if (t === "readme" && skillMd && !fileContent) {
                    void selectFile(skillMd); // stays on readme tab (selectFile no longer switches)
                  }
                  if (t === "validate" && !validation && !validating) {
                    void runValidation();
                  }
                }}
                className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors cursor-pointer ${
                  tab === t
                    ? "border-brand-500 text-brand-400"
                    : "border-transparent text-slate-500 hover:text-slate-300"
                }`}
              >
                {t === "info" && "Metadata"}
                {t === "readme" && "SKILL.md"}
                {t === "files" && "Files"}
                {t === "validate" && "Validate & Analyze"}
              </button>
            ))}
          </div>

          {/* Tab content */}
          {tab === "info" && (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <MetaCard title="Registry Info" data={[
                ["Owner", info.skill?.metadata?.owner],
                ["Scope", info.skill?.metadata?.scope],
                ["Tags", info.skill?.metadata?.tags],
                ["Source", info.skill?.metadata?.source_repo],
                ["License", info.skill?.license],
              ]} />
              <MetaCard title="Install Info" data={[
                ["Version", info.lock?.version],
                ["Registry", info.lock?.registry],
                ["Installed At", info.lock?.installed_at ? new Date(info.lock.installed_at).toLocaleString() : undefined],
                ["Content Hash", info.lock?.content_hash ? info.lock.content_hash.slice(0, 16) + "…" : undefined],
                ["Installed Path", info.lock?.installed_path],
              ]} />
            </div>
          )}

          {tab === "readme" && (
            <div className="card">
              {loadingFile ? (
                <div className="flex justify-center py-8">
                  <div className="spinner w-6 h-6" />
                </div>
              ) : fileContent ? (
                <MarkdownViewer content={fileContent} />
              ) : (
                <div className="text-center py-8 text-slate-500 text-sm">
                  <FileText size={32} className="mx-auto mb-2 text-slate-700" />
                  No SKILL.md found.{" "}
                  {skillMd && (
                    <button
                      className="text-brand-400 underline"
                      onClick={() => void selectFile(skillMd)}
                    >
                      Load it
                    </button>
                  )}
                </div>
              )}
            </div>
          )}

          {tab === "validate" && (
            <div className="card space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-sm font-semibold text-slate-300 flex items-center gap-2">
                    <ShieldCheck size={15} className="text-brand-400" />
                    Validation & Analysis
                  </h3>
                  <p className="text-xs text-slate-600 mt-0.5">
                    Spec conformance plus offline content-quality and contamination analysis
                  </p>
                </div>
                <button
                  onClick={() => void runValidation()}
                  disabled={validating}
                  className="btn-ghost text-xs"
                >
                  <RefreshCw size={12} className={validating ? "animate-spin" : ""} />
                  Re-run
                </button>
              </div>

              {validating ? (
                <div className="flex items-center gap-2 text-sm text-slate-500 py-2">
                  <div className="spinner w-4 h-4" />
                  Validating...
                </div>
              ) : validationError ? (
                <div className="bg-red-500/10 border border-red-500/20 rounded-lg px-4 py-3 text-sm text-red-400">
                  {validationError}
                </div>
              ) : validation ? (
                <ValidationReport result={validation} />
              ) : (
                <p className="text-sm text-slate-600">No results yet.</p>
              )}
            </div>
          )}

          {tab === "files" && (
            <div className="flex gap-4">
              {/* File tree */}
              <div className="w-56 shrink-0">
                <div className="card p-2 space-y-0.5">
                  <p className="text-xs font-semibold text-slate-600 uppercase tracking-wider px-2 py-1">
                    <FolderOpen size={12} className="inline mr-1" />
                    Skill Files
                  </p>
                  {files.length === 0 ? (
                    <p className="text-xs text-slate-700 px-2 py-2">No files found</p>
                  ) : (
                    files.map((f) => (
                      <FileTreeEntry
                        key={f.path}
                        entry={f}
                        depth={0}
                        selectedFile={selectedFile}
                        expandedDirs={expandedDirs}
                        loadingDir={loadingDir}
                        onSelectFile={(e) => { void selectFile(e); }}
                        onToggleDir={(e) => { void toggleDir(e); }}
                      />
                    ))
                  )}
                </div>
              </div>

              {/* File viewer */}
              <div className="flex-1 min-w-0">
                {loadingFile ? (
                  <div className="card flex justify-center py-12">
                    <div className="spinner w-6 h-6" />
                  </div>
                ) : selectedFile ? (
                  selectedFile.toLowerCase().endsWith(".md") ? (
                    <div className="card">
                      <MarkdownViewer content={fileContent} />
                    </div>
                  ) : (
                    <Suspense
                      fallback={
                        <div className="card flex justify-center py-12">
                          <div className="spinner w-6 h-6" />
                        </div>
                      }
                    >
                      <CodeViewer
                        content={fileContent}
                        filename={selectedFile.split(/[/\\]/).at(-1)}
                        height="600px"
                      />
                    </Suspense>
                  )
                ) : (
                  <div className="card flex flex-col items-center py-16 text-center text-slate-600">
                    <Code size={32} className="mb-2 text-slate-700" />
                    <p className="text-sm">Select a file to view</p>
                  </div>
                )}
              </div>
            </div>
          )}
        </>
      )}

      <ConfirmDialog
        open={removing}
        title={`Remove "${decoded}"?`}
        description="This will delete the skill files and remove it from skell.toml and skell.lock."
        confirmLabel="Remove"
        danger
        onConfirm={() => void handleRemove()}
        onCancel={() => setRemoving(false)}
      />
    </div>
  );
}

function MetaCard({
  title,
  data,
}: {
  title: string;
  data: [string, string | undefined][];
}) {
  const filled = data.filter(([, v]) => v);
  return (
    <div className="card">
      <h3 className="text-sm font-semibold text-slate-400 mb-3">{title}</h3>
      <dl className="space-y-2">
        {filled.map(([label, value]) => (
          <div key={label} className="flex gap-3 text-sm">
            <dt className="text-slate-600 w-28 shrink-0">{label}</dt>
            <dd className="text-slate-300 font-mono text-xs break-all">{value}</dd>
          </div>
        ))}
        {filled.length === 0 && (
          <p className="text-slate-700 text-sm">No data available</p>
        )}
      </dl>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Recursive file tree entry — handles both files and expandable folders
// ---------------------------------------------------------------------------

interface FileTreeEntryProps {
  entry: FileEntry;
  depth: number;
  selectedFile: string | null;
  expandedDirs: Record<string, FileEntry[]>;
  loadingDir: string | null;
  onSelectFile: (e: FileEntry) => void;
  onToggleDir: (e: FileEntry) => void;
}

function FileTreeEntry({
  entry,
  depth,
  selectedFile,
  expandedDirs,
  loadingDir,
  onSelectFile,
  onToggleDir,
}: FileTreeEntryProps) {
  const isExpanded = entry.path in expandedDirs;
  const isLoadingThis = loadingDir === entry.path;
  const children = expandedDirs[entry.path] ?? [];
  const indent = depth * 12;

  return (
    <>
      <button
        key={entry.path}
        onClick={() => entry.is_dir ? onToggleDir(entry) : onSelectFile(entry)}
        style={{ paddingLeft: `${8 + indent}px` }}
        className={`w-full flex items-center gap-2 pr-2 py-1.5 rounded text-xs transition-colors text-left ${
          selectedFile === entry.path
            ? "bg-brand-600/20 text-brand-400"
            : entry.is_dir
            ? "text-slate-400 hover:text-slate-200 hover:bg-white/5 cursor-pointer"
            : "text-slate-400 hover:text-slate-200 hover:bg-white/5 cursor-pointer"
        }`}
      >
        {isLoadingThis ? (
          <span className="spinner w-3 h-3 shrink-0" />
        ) : entry.is_dir ? (
          <FolderOpen size={11} className={isExpanded ? "text-brand-400" : ""} />
        ) : entry.name.toLowerCase().endsWith(".md") ? (
          <FileText size={11} />
        ) : (
          <Code size={11} />
        )}
        <span className="truncate flex-1">{entry.name}</span>
        {entry.is_dir && (
          <ChevronRight
            size={10}
            className={`shrink-0 transition-transform ${isExpanded ? "rotate-90" : ""}`}
          />
        )}
      </button>
      {isExpanded && children.map((child) => (
        <FileTreeEntry
          key={child.path}
          entry={child}
          depth={depth + 1}
          selectedFile={selectedFile}
          expandedDirs={expandedDirs}
          loadingDir={loadingDir}
          onSelectFile={onSelectFile}
          onToggleDir={onToggleDir}
        />
      ))}
    </>
  );
}
