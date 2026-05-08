import { lazy, Suspense, useEffect, useState, useCallback } from "react";
import { useParams, useLocation, useNavigate } from "react-router-dom";
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
} from "lucide-react";
import { useRepoStore, useUIStore } from "@/store";
import {
  getInfo,
  upgradeSkill,
  removeSkill,
  pinSkill,
  unpinSkill,
  readFileContent,
  listDirectory,
} from "@/lib/skell";
import type { InfoResult, FileEntry } from "@/lib/types";
import { SkillBadge, LifecycleBadge } from "@/components/Badges";
import { MarkdownViewer } from "@/components/MarkdownViewer";
import { ConfirmDialog } from "@/components/ConfirmDialog";

const CodeViewer = lazy(async () => {
  const mod = await import("@/components/CodeViewer");
  return { default: mod.CodeViewer };
});

type Tab = "info" | "readme" | "files";

export function SkillDetail() {
  const { skillName } = useParams<{ skillName: string }>();
  const location = useLocation();
  const navigate = useNavigate();
  const { selectedRepo } = useRepoStore();
  const { notify } = useUIStore();

  const repo = (location.state as { repo?: string })?.repo ?? selectedRepo;
  const decoded = decodeURIComponent(skillName ?? "");

  const [info, setInfo] = useState<InfoResult | null>(null);
  const [loading, setLoading] = useState(true);
  const [tab, setTab] = useState<Tab>("info");
  const [files, setFiles] = useState<FileEntry[]>([]);
  const [selectedFile, setSelectedFile] = useState<string | null>(null);
  const [fileContent, setFileContent] = useState<string>("");
  const [loadingFile, setLoadingFile] = useState(false);
  const [removing, setRemoving] = useState(false);
  const [acting, setActing] = useState(false);
  // Folder expansion state for the file tree
  const [expandedDirs, setExpandedDirs] = useState<Record<string, FileEntry[]>>({});
  const [loadingDir, setLoadingDir] = useState<string | null>(null);

  const loadInfo = useCallback(async () => {
    if (!decoded) return;
    setLoading(true);
    try {
      const result = await getInfo(decoded, repo === "global" ? undefined : repo);
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
        const entries = await listDirectory(absPath).catch(() => [] as FileEntry[]);
        setFiles(entries);
      }
    } finally {
      setLoading(false);
    }
  }, [decoded, repo]);

  useEffect(() => {
    void loadInfo();
  }, [loadInfo]);

  // Auto-load SKILL.md content once files are available and the readme tab is active.
  useEffect(() => {
    if (tab === "readme" && !fileContent && !loadingFile) {
      const md = files.find(
        (f) => f.name.toLowerCase() === "skill.md" || f.name.toLowerCase() === "readme.md"
      );
      if (md) void selectFile(md);
    }
    // selectFile is stable (defined in the same component scope), files and tab change trigger this
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
    // NOTE: do NOT switch tabs here — caller owns tab state
  }

  async function toggleDir(entry: FileEntry) {
    if (!entry.is_dir) return;
    // Collapse if already expanded
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

  async function handleUpgrade() {
    if (!repo || repo === "global") return;
    setActing(true);
    try {
      const result = await upgradeSkill({ skillName: decoded, repo });
      if (result.success) {
        notify({ kind: "success", title: `Upgraded ${decoded}` });
        void loadInfo();
      } else {
        notify({ kind: "error", title: "Upgrade failed", detail: result.stderr });
      }
    } finally {
      setActing(false);
    }
  }

  async function handlePin() {
    if (!repo || repo === "global") return;
    setActing(true);
    const isPinned = !!info?.lock?.pinned;
    try {
      const result = await (isPinned
        ? unpinSkill({ skillName: decoded, repo })
        : pinSkill({ skillName: decoded, repo }));
      if (result.success) {
        notify({ kind: "success", title: isPinned ? `Unpinned ${decoded}` : `Pinned ${decoded}` });
        void loadInfo();
      } else {
        notify({ kind: "error", title: "Operation failed", detail: result.stderr });
      }
    } finally {
      setActing(false);
    }
  }

  async function handleRemove() {
    if (!repo || repo === "global") return;
    setRemoving(false);
    setActing(true);
    try {
      const result = await removeSkill({ skillName: decoded, repo });
      if (result.success) {
        notify({ kind: "success", title: `Removed ${decoded}` });
        navigate(-1);
      } else {
        notify({ kind: "error", title: "Remove failed", detail: result.stderr });
      }
    } finally {
      setActing(false);
    }
  }

  const status = info?.lock?.pinned ? "pinned" : info?.status ?? "unknown";
  const isOutdated = info?.status === "outdated";
  const isPinned = !!info?.lock?.pinned;

  // Find SKILL.md in file list
  const skillMd = files.find(
    (f) => f.name.toLowerCase() === "skill.md" || f.name.toLowerCase() === "readme.md"
  );

  return (
    <div className="p-6 space-y-5 max-w-5xl mx-auto">
      {/* Back */}
      <button onClick={() => navigate(-1)} className="btn-ghost text-xs">
        <ArrowLeft size={13} />
        Back
      </button>

      {loading ? (
        <div className="flex justify-center py-20">
          <div className="spinner w-8 h-8" />
        </div>
      ) : !info ? (
        <div className="card text-center py-12 text-slate-500">Skill not found.</div>
      ) : (
        <>
          {/* Skill header */}
          <div className="card">
            <div className="flex items-start gap-4">
              <div className="w-12 h-12 rounded-xl bg-brand-600/15 flex items-center justify-center shrink-0">
                <Package size={24} className="text-brand-400" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-3 flex-wrap">
                  <h1 className="text-xl font-bold text-slate-200">{decoded}</h1>
                  <SkillBadge status={status as typeof status} />
                  {info.skill?.metadata?.lifecycle && (
                    <LifecycleBadge lifecycle={info.skill.metadata.lifecycle} />
                  )}
                </div>
                {info.skill?.description && (
                  <p className="text-sm text-slate-400 mt-1">{info.skill.description}</p>
                )}
                <div className="flex items-center gap-4 mt-2 text-xs text-slate-600 flex-wrap">
                  {info.lock?.version && (
                    <span>
                      Installed: <span className="font-mono text-slate-400">{info.lock.version}</span>
                    </span>
                  )}
                  {info.skill?.metadata?.version && (
                    <span>
                      Latest: <span className="font-mono text-slate-400">{info.skill.metadata.version}</span>
                    </span>
                  )}
                  {info.lock?.registry && (
                    <span>
                      Registry: <span className="text-slate-400">{info.lock.registry}</span>
                    </span>
                  )}
                  {info.lock?.installed_at && (
                    <span>
                      Installed: <span className="text-slate-400">{new Date(info.lock.installed_at).toLocaleDateString()}</span>
                    </span>
                  )}
                </div>
              </div>

              {/* Actions */}
              <div className="flex items-center gap-2 shrink-0">
                <button onClick={() => void loadInfo()} className="btn-ghost" disabled={acting}>
                  <RefreshCw size={13} />
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
                  className="btn-ghost text-xs text-indigo-400 hover:bg-indigo-500/10"
                  title="Contribute metadata improvement"
                >
                  <GitPullRequest size={13} />
                  Fix Metadata
                </button>
                {isOutdated && (
                  <button onClick={() => void handleUpgrade()} disabled={acting} className="btn-primary py-1.5 text-xs">
                    <ArrowUp size={13} />
                    Upgrade
                  </button>
                )}
                <button onClick={() => void handlePin()} disabled={acting || !repo || repo === "global"} className="btn-ghost text-xs">
                  {isPinned ? <PinOff size={13} /> : <Pin size={13} />}
                  {isPinned ? "Unpin" : "Pin"}
                </button>
                <button
                  onClick={() => setRemoving(true)}
                  disabled={acting || !repo || repo === "global"}
                  className="btn-ghost text-xs text-red-400 hover:bg-red-500/10"
                >
                  <Trash2 size={13} />
                  Remove
                </button>
              </div>
            </div>
          </div>

          {/* Tabs */}
          <div className="flex items-center gap-1 border-b border-[#1e2540]">
            {(["info", "readme", "files"] as Tab[]).map((t) => (
              <button
                key={t}
                onClick={() => {
                  setTab(t);
                  if (t === "readme" && skillMd && !fileContent) {
                    void selectFile(skillMd); // stays on readme tab (selectFile no longer switches)
                  }
                }}
                className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
                  tab === t
                    ? "border-brand-500 text-brand-400"
                    : "border-transparent text-slate-500 hover:text-slate-300"
                }`}
              >
                {t === "info" && "Metadata"}
                {t === "readme" && "SKILL.md"}
                {t === "files" && "Files"}
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
            : "text-slate-400 hover:text-slate-200 hover:bg-white/5"
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
