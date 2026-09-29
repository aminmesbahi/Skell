import { previewRegistrySkill, validateDirectory } from "./skell";
import type { RegistrySkill, SkillValidationResult } from "./types";

// Bound native CLI work when a catalog contains hundreds of cards. A short cache
// also avoids repeating identical checks when changing search/source filters.
let active = 0;
const queue: Array<() => void> = [];
const cache = new Map<string, { expires: number; result: Promise<SkillValidationResult | undefined> }>();
async function limited<T>(work: () => Promise<T>): Promise<T> {
  await new Promise<void>((resolve) => {
    const start = () => { active++; resolve(); };
    if (active < 3) start(); else queue.push(start);
  });
  try { return await work(); }
  finally { active--; queue.shift()?.(); }
}

export function registryValidation(skill: RegistrySkill): Promise<SkillValidationResult | undefined> {
  const key = JSON.stringify([skill.registry_alias, skill.registry_url, skill.name, skill.metadata?.version]);
  const existing = cache.get(key);
  if (existing && existing.expires > Date.now()) return existing.result;
  const result = limited(async () => {
    const preview = await previewRegistrySkill(skill.registry_alias ?? "", skill.registry_url ?? "", skill.name);
    if (!preview.found) return undefined;
    return (await validateDirectory(preview.source_path))[0];
  });
  for (const [id, entry] of cache) if (entry.expires <= Date.now()) cache.delete(id);
  cache.set(key, { expires: Date.now() + 60_000, result });
  void result.then((value) => { if (!value) cache.delete(key); }, () => cache.delete(key));
  return result;
}
