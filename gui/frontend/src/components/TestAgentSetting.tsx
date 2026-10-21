import { useState } from "react";
import { FlaskConical } from "lucide-react";
import { useRepoStore } from "@/store";

const EXAMPLES = ["claude -p", "codex exec"];

/** The command Skell uses to answer test prompts. Empty means “only check the eval files”. */
export function TestAgentSetting() {
  const testAgent = useRepoStore((s) => s.testAgent);
  const setTestAgent = useRepoStore((s) => s.setTestAgent);
  const [draft, setDraft] = useState(testAgent);
  const dirty = draft.trim() !== testAgent;

  return (
    <div className="card">
      <h2 className="section-title flex items-center gap-2 text-base"><FlaskConical size={16} className="text-brand-400" />Agent for skill tests</h2>
      <p className="mb-4 mt-0.5 text-sm text-slate-500">
        Skills can ship example prompts (evals). To run them, Skell needs a command that reads a prompt and prints your agent’s answer. Each case is one real call to that agent, so nothing runs until you set this.
      </p>
      <form className="flex flex-wrap items-center gap-2" onSubmit={(e) => { e.preventDefault(); setTestAgent(draft.trim()); }}>
        <input className="input max-w-sm font-mono" placeholder="claude -p" value={draft} onChange={(e) => setDraft(e.target.value)} aria-label="Agent command" />
        <button className="btn-primary" type="submit" disabled={!dirty}>Save</button>
        {testAgent && <button type="button" className="btn-ghost" onClick={() => { setDraft(""); setTestAgent(""); }}>Clear</button>}
      </form>
      <p className="mt-2 text-xs text-slate-600">
        Examples:{" "}
        {EXAMPLES.map((example) => (
          <button key={example} type="button" className="mr-2 font-mono text-brand-400 hover:underline" onClick={() => setDraft(example)}>{example}</button>
        ))}
      </p>
    </div>
  );
}
