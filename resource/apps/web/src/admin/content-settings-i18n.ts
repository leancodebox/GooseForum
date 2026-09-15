import { getAdminDictionary } from "./translation-loader";
type EnglishDictionary =
  typeof import("./messages/en-content-settings").default;
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

export type ContentSettingsTextKey = keyof EnglishDictionary;
export function createContentSettingsText(l: AuthLocale) {
  return (k: ContentSettingsTextKey) =>
    getAdminDictionary<EnglishDictionary>("content-settings", l)[k];
}
