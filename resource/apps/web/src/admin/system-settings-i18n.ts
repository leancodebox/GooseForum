import { getAdminDictionary } from "./translation-loader";
type EnglishDictionary = typeof import("./messages/en-system-settings").default;
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

export type SystemSettingsTextKey = keyof EnglishDictionary;
export function createSystemSettingsText(l: AuthLocale) {
  return (k: SystemSettingsTextKey) =>
    getAdminDictionary<EnglishDictionary>("system-settings", l)[k];
}
