import { useState } from "react";
import { BadgeCheck, KeyRound, Trash2 } from "lucide-react";
import { useAsync } from "@/hooks/useAsync";
import { addTrustedKey, listTrustedKeys, removeTrustedKey } from "@/lib/skell";
import { useUIStore } from "@/store";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import { LoadState } from "@/components/LoadState";

/** Publishers whose skill signatures Skell accepts (`skell trust`). */
export function TrustedPublishers() {
  const { notify } = useUIStore();
  const state = useAsync("trusted-keys", listTrustedKeys);
  const [name, setName] = useState("");
  const [key, setKey] = useState("");
  const [busy, setBusy] = useState(false);
  const [confirmRemove, setConfirmRemove] = useState<string | null>(null);
  const keys = state.data ?? [];

  async function add(e: React.FormEvent) {
    e.preventDefault();
    if (!name.trim() || !key.trim()) return;
    setBusy(true);
    try {
      const res = await addTrustedKey(name.trim(), key);
      if (res.success) {
        notify({ kind: "success", title: `Now trusting ${name.trim()}` });
        setName("");
        setKey("");
        await state.refresh();
      } else {
        notify({ kind: "error", title: "That key was not accepted", detail: (res.stderr || res.stdout).trim() });
      }
    } finally {
      setBusy(false);
    }
  }

  async function remove(id: string) {
    setConfirmRemove(null);
    const res = await removeTrustedKey(id);
    notify(res.success ? { kind: "success", title: "Publisher removed" } : { kind: "error", title: "Could not remove the publisher", detail: res.stderr });
    await state.refresh();
  }

  return (
    <div className="card">
      <h2 className="section-title flex items-center gap-2 text-base"><BadgeCheck size={16} className="text-brand-400" />Trusted publishers</h2>
      <p className="mb-4 mt-0.5 text-sm text-slate-500">
        Publishers can sign their skills so you know who made them and that nobody changed them afterwards. Skills signed by someone on this list show as “Signed · trusted”.
      </p>

      <LoadState loading={state.loading && !state.data} error={state.error} retry={() => void state.refresh()} />
      {state.data && keys.length === 0 && (
        <p className="mb-4 rounded-lg border border-dashed border-[var(--palette-1e2640)] p-4 text-center text-sm text-slate-500">No trusted publishers yet.</p>
      )}
      {keys.length > 0 && (
        <ul className="mb-4 space-y-2">
          {keys.map((k) => (
            <li key={k.key_id} className="flex items-center justify-between gap-3 rounded-lg border border-[var(--palette-1e2640)] bg-[var(--palette-0f1225)] px-4 py-3">
              <div className="flex min-w-0 items-center gap-3">
                <KeyRound size={16} className="shrink-0 text-emerald-400" />
                <div className="min-w-0">
                  <div className="truncate text-sm font-medium text-slate-200">{k.publisher}</div>
                  <div className="font-mono text-xs text-slate-500">key {k.key_id}</div>
                </div>
              </div>
              <button className="rounded p-1.5 text-slate-500 transition hover:bg-red-500/10 hover:text-red-400" title="Stop trusting this publisher" aria-label={`Stop trusting ${k.publisher}`} onClick={() => setConfirmRemove(k.key_id)}>
                <Trash2 size={15} />
              </button>
            </li>
          ))}
        </ul>
      )}

      <form onSubmit={(e) => void add(e)} className="space-y-3 border-t border-[var(--palette-1e2640)] pt-4">
        <p className="text-sm font-medium text-slate-300">Trust a new publisher</p>
        <input className="input" placeholder="Publisher name, e.g. acme" value={name} onChange={(e) => setName(e.target.value)} aria-label="Publisher name" />
        <textarea className="input !h-20 py-2 font-mono text-xs" placeholder="Paste their public key (the contents of their .pub file)" value={key} onChange={(e) => setKey(e.target.value)} aria-label="Public key" />
        <button className="btn-primary" type="submit" disabled={busy || !name.trim() || !key.trim()}>{busy ? "Adding…" : "Trust publisher"}</button>
      </form>

      <ConfirmDialog
        open={confirmRemove !== null}
        title="Stop trusting this publisher?"
        description="Their skills stay installed, but will show as signed by an untrusted publisher."
        confirmLabel="Stop trusting"
        danger
        onConfirm={() => confirmRemove && void remove(confirmRemove)}
        onCancel={() => setConfirmRemove(null)}
      />
    </div>
  );
}
