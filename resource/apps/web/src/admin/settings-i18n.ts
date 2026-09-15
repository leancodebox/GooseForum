import { getAdminDictionary } from "./translation-loader";
type EnglishDictionary = typeof import("./messages/en-settings").default;
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

export type SettingsTextKey = keyof EnglishDictionary;
export function createSettingsText(locale: AuthLocale) {
  return (k: SettingsTextKey) =>
    getAdminDictionary<EnglishDictionary>("settings", locale)[k];
}
