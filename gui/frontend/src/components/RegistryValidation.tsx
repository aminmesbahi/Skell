import { useAsync } from "@/hooks/useAsync";
import { registryValidation } from "@/lib/registryValidation";
import type { RegistrySkill } from "@/lib/types";
import { ValidationBadge } from "./ValidationBadge";

export function RegistryValidation({ skill }: { skill: RegistrySkill }) {
  const state = useAsync(`${skill.registry_alias}:${skill.registry_url}:${skill.name}:${skill.metadata?.version}`, () => registryValidation(skill));
  return <div><ValidationBadge result={state.data} loading={state.loading} />{state.error && <button className="btn-ghost text-xs" title={state.error} onClick={() => void state.refresh()}>Retry validation</button>}</div>;
}
