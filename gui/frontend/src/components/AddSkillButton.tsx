import { useMemo, useState } from "react";
import { Plus, Library, GitBranchPlus } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { useRepoStore, useUIStore } from "@/store";
import { AddFromURLDialog } from "./AddFromURLDialog";

interface AddSkillButtonProps {
  projectPath: string;
  onRefresh?: () => void;
  label?: string;
}

export function AddSkillButton({ projectPath, onRefresh, label = "Add Skill" }: AddSkillButtonProps) {
  const navigate = useNavigate();
  const { selectedRepo } = useRepoStore();
  const { notify } = useUIStore();
  const [menuOpen, setMenuOpen] = useState(false);
  const [addUrlOpen, setAddUrlOpen] = useState(false);

  const destination = useMemo(() => projectPath || selectedRepo || "", [projectPath, selectedRepo]);

  function handleBrowseCatalog() {
    setMenuOpen(false);
    if (!destination) {
      notify({ kind: "info", title: "Select a project first", detail: "Choose a project before installing a catalog skill." });
      return;
    }
    navigate("/catalog", { state: { installDestination: destination } });
  }

  function handleAddFromRepository() {
    setMenuOpen(false);
    if (!destination) {
      notify({ kind: "info", title: "Select a project first", detail: "Choose a project before adding a skill from a repository." });
      return;
    }
    setAddUrlOpen(true);
  }

  return (
    <>
      <div className="relative">
        <button onClick={() => setMenuOpen((open) => !open)} className="btn-primary inline-flex items-center gap-2 text-sm">
          <Plus size={14} />
          {label}
        </button>
        {menuOpen && (
          <div className="absolute right-0 z-20 mt-2 w-56 rounded-xl border border-[#26314d] bg-[#11162a] p-2 shadow-2xl">
            <button onClick={handleBrowseCatalog} className="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left text-sm text-slate-300 hover:bg-white/5">
              <Library size={14} className="text-brand-400" />
              Browse Catalog
            </button>
            <button onClick={handleAddFromRepository} className="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left text-sm text-slate-300 hover:bg-white/5">
              <GitBranchPlus size={14} className="text-brand-400" />
              Add from Repository
            </button>
            <button onClick={() => { setMenuOpen(false); handleAddFromRepository(); }} className="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left text-sm text-slate-300 hover:bg-white/5">
              <GitBranchPlus size={14} className="text-brand-400" />
              Add from Local Path
            </button>
          </div>
        )}
      </div>
      <AddFromURLDialog
        open={addUrlOpen}
        onClose={() => setAddUrlOpen(false)}
        initialRepo={destination}
        onSuccess={() => onRefresh?.()}
      />
    </>
  );
}
