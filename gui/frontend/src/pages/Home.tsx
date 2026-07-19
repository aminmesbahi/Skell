import { Link } from "react-router-dom";
import { Plus, Compass, Sparkles } from "lucide-react";
import { useRepoStore } from "@/store";

export function Home() {
  const { repos } = useRepoStore();

  if (repos.length === 0) {
    return (
      <div className="mx-auto flex min-h-full max-w-3xl flex-col justify-center px-6 py-16">
        <div className="rounded-2xl border border-[#1a1f35] bg-[#0f1324] p-8 shadow-2xl shadow-black/20">
          <div className="mb-6 flex h-12 w-12 items-center justify-center rounded-full bg-brand-600/20 text-brand-400">
            <Sparkles size={24} />
          </div>
          <h1 className="text-2xl font-bold text-slate-200">Manage agent skills across your projects</h1>
          <p className="mt-3 text-sm leading-6 text-slate-400">
            Skell helps you discover, install, validate, and keep agent skills organised for the coding tools you use.
          </p>
          <div className="mt-8 flex flex-wrap gap-3">
            <Link to="/projects" className="inline-flex items-center gap-2 rounded-lg bg-brand-600 px-4 py-2 text-sm font-medium text-white hover:bg-brand-500">
              <Plus size={16} />
              Add Project
            </Link>
            <Link to="/catalog" className="inline-flex items-center gap-2 rounded-lg border border-[#2a3353] px-4 py-2 text-sm font-medium text-slate-300 hover:bg-white/5">
              <Compass size={16} />
              Browse Catalog
            </Link>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-6xl px-6 py-8">
      <h1 className="text-2xl font-bold text-slate-200">Home</h1>
      <p className="mt-2 text-sm text-slate-400">Actionable summaries for the projects you manage.</p>
      <div className="mt-8 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        <Link to="/projects" className="rounded-2xl border border-[#1a1f35] bg-[#0f1324] p-5 hover:bg-white/5">
          <p className="text-sm font-medium text-slate-200">Projects requiring setup</p>
          <p className="mt-2 text-2xl font-semibold text-white">{repos.length}</p>
        </Link>
        <Link to="/catalog" className="rounded-2xl border border-[#1a1f35] bg-[#0f1324] p-5 hover:bg-white/5">
          <p className="text-sm font-medium text-slate-200">Browse the catalog</p>
          <p className="mt-2 text-sm text-slate-400">Discover skills for the tools you use.</p>
        </Link>
        <Link to="/activity" className="rounded-2xl border border-[#1a1f35] bg-[#0f1324] p-5 hover:bg-white/5">
          <p className="text-sm font-medium text-slate-200">Recent activity</p>
          <p className="mt-2 text-sm text-slate-400">Open the activity feed for recent changes.</p>
        </Link>
      </div>
    </div>
  );
}
