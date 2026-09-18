import type { RegistrySkill } from "./types";

export type RegistrySourceFilter = "all" | "global" | "local";
export type NormalizedRegistrySource = Exclude<RegistrySourceFilter, "all"> | "unknown";

const GLOBAL_SOURCE_VALUES = new Set([
	"global",
	"shared",
	"shared-source",
	"shared_source",
	"registry",
	"remote",
]);

const LOCAL_SOURCE_VALUES = new Set([
	"local",
	"project",
	"repo",
	"folder",
	"project-local",
	"project_local",
]);

function isLocalPath(value: string): boolean {
	return /^[a-zA-Z]:[\\/]/.test(value) || value.startsWith("/") || value.startsWith("\\\\");
}

export function normalizeRegistrySource(source?: string | null): NormalizedRegistrySource {
	const value = source?.trim().toLowerCase();
	if (!value) return "unknown";
	if (GLOBAL_SOURCE_VALUES.has(value)) return "global";
	if (LOCAL_SOURCE_VALUES.has(value)) return "local";
	return "unknown";
}

export function inferRegistrySource(skill: Pick<RegistrySkill, "registry_source" | "registry_url">): NormalizedRegistrySource {
	const normalized = normalizeRegistrySource(skill.registry_source);
	if (normalized !== "unknown") return normalized;
	if (skill.registry_url && isLocalPath(skill.registry_url.trim())) return "local";
	return "global";
}

export function matchesRegistrySource(skill: Pick<RegistrySkill, "registry_source" | "registry_url">, filter: RegistrySourceFilter): boolean {
	if (filter === "all") return true;
	return inferRegistrySource(skill) === filter;
}
