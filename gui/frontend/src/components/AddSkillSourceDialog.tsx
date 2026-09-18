import { useEffect, useState } from "react";
import { FolderOpen, Globe, Loader2, X } from "lucide-react";
import { SelectDirectory } from "../../bindings/skell-gui/app";
import { addSkillSource } from "@/lib/skell";
import { useUIStore } from "@/store";

interface AddSkillSourceDialogProps {
	open: boolean;
	onClose: () => void;
	onSuccess?: () => void;
}

function guessAlias(source: string): string {
	const trimmed = source.trim().replace(/\/+$/, "");
	if (!trimmed) return "";
	const normalized = trimmed.replace(/^[a-z]+:\/\//i, "");
	const segments = normalized.split(/[\\/:]/).filter(Boolean);
	return segments.at(-1)?.replace(/\.git$/i, "") ?? "";
}

export function AddSkillSourceDialog({ open, onClose, onSuccess }: AddSkillSourceDialogProps) {
	const { notify } = useUIStore();
	const [alias, setAlias] = useState("");
	const [source, setSource] = useState("");
	const [loading, setLoading] = useState(false);
	const [aliasTouched, setAliasTouched] = useState(false);

	const title = "Add Source";
	const helper = "Paste a repo URL, local path, or file URI. Skell will detect the source type.";

	useEffect(() => {
		if (!open) return;
		setAlias("");
		setSource("");
		setLoading(false);
		setAliasTouched(false);
	}, [open]);

	if (!open) return null;

	function updateSource(nextSource: string) {
		setSource(nextSource);
		if (!aliasTouched) {
			setAlias(guessAlias(nextSource));
		}
	}

	async function chooseFolder() {
		const selected = await SelectDirectory();
		if (!selected) return;
		updateSource(selected);
	}

	async function handleSubmit(e: React.FormEvent) {
		e.preventDefault();
		const trimmedAlias = alias.trim();
		const trimmedSource = source.trim();
		if (!trimmedAlias || !trimmedSource) return;

		setLoading(true);
		try {
			await addSkillSource(trimmedAlias, trimmedSource);
			notify({
				kind: "success",
				title: "Source added",
				detail: trimmedAlias,
			});
			onSuccess?.();
			onClose();
		} catch (error) {
			notify({
				kind: "error",
				title: "Failed to add source",
				detail: error instanceof Error ? error.message : String(error),
			});
		} finally {
			setLoading(false);
		}
	}

	return (
		<div className="fixed inset-0 z-50 flex items-center justify-center">
			<div className="absolute inset-0 bg-black/60 backdrop-blur-sm" onClick={onClose} />
			<form onSubmit={handleSubmit} className="relative z-10 w-full max-w-lg mx-4 rounded-2xl border border-[#2d3348] bg-[#13162a] shadow-2xl">
				<div className="flex items-center justify-between border-b border-[#2d3348] px-6 py-4">
					<div className="flex items-center gap-2 text-slate-200 font-semibold">
						<Globe size={18} className="text-blue-400" />
						{title}
					</div>
					<button type="button" onClick={onClose} className="rounded-lg p-1.5 text-slate-500 hover:bg-white/5 hover:text-slate-300">
						<X size={16} />
					</button>
				</div>

				<div className="space-y-4 p-6">
					<p className="text-sm text-slate-400">{helper}</p>

					<div className="space-y-1.5">
						<label className="text-xs font-medium uppercase tracking-wider text-slate-400">Alias</label>
						<input
							autoFocus
							type="text"
							value={alias}
							onChange={(e) => {
								setAliasTouched(true);
								setAlias(e.target.value);
							}}
							placeholder="company-skills"
							className="w-full rounded-lg border border-[#2d3348] bg-[#0e1120] px-3 py-2 text-sm text-slate-200 placeholder-slate-600 focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500/40"
						/>
					</div>

					<div className="space-y-1.5">
						<label className="text-xs font-medium uppercase tracking-wider text-slate-400">Source URL, path, or URI</label>
						<div className="flex gap-2">
							<input
								type="text"
								value={source}
								onChange={(e) => updateSource(e.target.value)}
								placeholder="https://github.com/owner/repo.git or C:\\skills\\design"
								className="min-w-0 flex-1 rounded-lg border border-[#2d3348] bg-[#0e1120] px-3 py-2 text-sm text-slate-200 placeholder-slate-600 focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500/40"
							/>
							<button type="button" onClick={() => void chooseFolder()} className="btn-ghost shrink-0">
								<FolderOpen size={14} />
								Folder
							</button>
						</div>
					</div>

					<div className="flex justify-end gap-3 pt-2">
						<button type="button" onClick={onClose} className="btn-ghost" disabled={loading}>
							Cancel
						</button>
						<button type="submit" className="btn-primary flex items-center gap-2" disabled={loading || !alias.trim() || !source.trim()}>
							{loading && <Loader2 size={14} className="animate-spin" />}
							Add Source
						</button>
					</div>
				</div>
			</form>
		</div>
	);
}
