import { lazy, Suspense } from "react";
import { useAsync } from "@/hooks/useAsync";
import { useEditorTheme } from "@/hooks/useEditorTheme";
import { previewRegistrySkill, readFileContent, listDirectory } from "@/lib/skell";
const DiffEditor = lazy(async () => ({ default: (await import("@monaco-editor/react")).DiffEditor }));

export function SkillComparison({ name, registry, url, folder }: { name: string; registry: string; url: string; folder: string }) {
  const theme = useEditorTheme();
  const state = useAsync(`${folder}:${registry}:${url}`, async () => {
    const preview = await previewRegistrySkill(registry, url, name);
    if (!preview.found) throw new Error("Registry copy is not cached. Refresh the registry cache and retry.");
    async function collect(root: string, prefix = ""): Promise<Record<string, string>> {
      const entries = await listDirectory(root);
      const result: Record<string, string> = {};
      for (const entry of entries) {
        if (entry.name === ".git") continue;
        const relative = prefix + entry.name;
        if (entry.is_dir) Object.assign(result, await collect(entry.path, relative + "/"));
        else result[relative] = await readFileContent(entry.path);
      }
      return result;
    }
    const [original, modified] = await Promise.all([collect(preview.source_path), collect(folder)]);
    return [...new Set([...Object.keys(original), ...Object.keys(modified)])].sort().filter((path) => original[path] !== modified[path]).map((path) => ({ path, original: original[path] ?? "", modified: modified[path] ?? "", kind: !(path in original) ? "Local only" : !(path in modified) ? "Registry only" : "Changed" }));
  });
  return <section className="card space-y-3"><h2>Changes against the cached registry version</h2><p className="text-sm">Left: registry · Right: installed files</p>{state.loading && <p role="status">Comparing files…</p>}{state.error && <div role="alert">{state.error}<button className="btn-ghost" onClick={() => void state.refresh()}>Retry</button></div>}{state.data?.length === 0 && <p>No differences found.</p>}{state.data?.map((file) => <details key={file.path} open={file.path.toLowerCase() === "skill.md"}><summary>{file.path} — {file.kind}</summary><Suspense fallback={<p>Loading diff…</p>}><DiffEditor height="380px" original={file.original} modified={file.modified} theme={theme} options={{ readOnly: true, renderSideBySide: true, minimap: { enabled: false } }} /></Suspense></details>)}</section>;
}
