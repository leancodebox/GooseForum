import type { LucideIconData } from "lucide-react";

export async function fetchCatalog(): Promise<Record<string, LucideIconData>> {
  const response = await fetch(new URL("../../../../icons/lucide.json", import.meta.url));
  if (!response.ok) throw new Error("Could not load the navigation icon catalog");
  const data = await response.json() as { icons: Record<string, LucideIconData>; aliases?: Record<string, string> };
  for (const [alias, name] of Object.entries(data.aliases ?? {})) data.icons[alias] = data.icons[name];
  return data.icons;
}
