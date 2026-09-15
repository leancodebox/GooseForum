import { getAdminDictionary } from "./translation-loader";
type EnglishDictionary = typeof import("./messages/en-dashboard").default;
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

export type DashboardTextKey = keyof EnglishDictionary;
export function createDashboardText(l: AuthLocale) {
  return (k: DashboardTextKey) =>
    getAdminDictionary<EnglishDictionary>("dashboard", l)[k];
}
