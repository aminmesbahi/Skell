import { Link } from "react-router";

export function MissingProject() {
  return <div className="card m-6"><h1>Project not found</h1><p>This link does not match a saved project.</p><Link className="btn-primary mt-4" to="/projects">Choose a project</Link></div>;
}

export function LoadState({ loading, error, retry }: { loading: boolean; error?: string; retry: () => void }) {
  return <>
    {loading && <p role="status" className="text-sm text-slate-400">Loading…</p>}
    {error && <div role="alert" className="card border-red-500/40"><p>Could not load the latest data.</p><details className="text-sm"><summary>Details</summary>{error}</details><button className="btn-ghost mt-2" onClick={retry}>Retry</button></div>}
  </>;
}
