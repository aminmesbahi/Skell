import Editor from "@monaco-editor/react";

interface CodeViewerProps {
  content: string;
  language?: string;
  height?: string;
}

function detectLanguage(filename: string): string {
  const ext = filename.split(".").at(-1)?.toLowerCase() ?? "";
  const map: Record<string, string> = {
    ts: "typescript",
    tsx: "typescript",
    js: "javascript",
    jsx: "javascript",
    go: "go",
    py: "python",
    rs: "rust",
    toml: "toml",
    yaml: "yaml",
    yml: "yaml",
    json: "json",
    md: "markdown",
    sh: "shell",
    bash: "shell",
    ps1: "powershell",
    rb: "ruby",
    lock: "json",
  };
  return map[ext] ?? "plaintext";
}

interface CodeViewerProps {
  content: string;
  language?: string;
  filename?: string;
  height?: string;
}

export function CodeViewer({
  content,
  language,
  filename,
  height = "400px",
}: CodeViewerProps) {
  const lang = language ?? (filename ? detectLanguage(filename) : "plaintext");

  return (
    <div className="monaco-editor-container border border-[#1e2540] rounded-xl overflow-hidden">
      <Editor
        height={height}
        language={lang}
        value={content}
        theme="vs-dark"
        options={{
          readOnly: true,
          minimap: { enabled: false },
          scrollBeyondLastLine: false,
          fontSize: 13,
          fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
          lineNumbers: "on",
          wordWrap: "on",
          padding: { top: 12, bottom: 12 },
          scrollbar: { verticalScrollbarSize: 6 },
        }}
      />
    </div>
  );
}

export { detectLanguage };
