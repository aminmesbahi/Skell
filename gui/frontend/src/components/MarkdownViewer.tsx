import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import rehypeHighlight from "rehype-highlight";
import "highlight.js/styles/github-dark.css";

interface MarkdownViewerProps {
  content: string;
  className?: string;
}

// A SKILL.md starts with YAML front matter; as markdown it renders as a stray
// heading made of "name: … description: …", so it is dropped (the metadata is
// shown separately by the pages that embed this viewer).
const FRONT_MATTER = /^﻿?---[ \t]*\r?\n[\s\S]*?\r?\n---[ \t]*(?:\r?\n|$)/;

export function stripFrontMatter(content: string): string {
  return content.replace(FRONT_MATTER, "").replace(/^\s+/, "");
}

export function MarkdownViewer({ content, className }: MarkdownViewerProps) {
  return (
    <div className={`md-content ${className ?? ""}`}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[rehypeHighlight]}
      >
        {stripFrontMatter(content)}
      </ReactMarkdown>
    </div>
  );
}
