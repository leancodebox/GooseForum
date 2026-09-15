import { getAdminDictionary } from "./translation-loader";
type EnglishDictionary =
  typeof import("./messages/en-moderation-settings").default;
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

export type ModerationSettingsTextKey = keyof EnglishDictionary;
export function createModerationSettingsText(l: AuthLocale) {
  return (k: ModerationSettingsTextKey) =>
    getAdminDictionary<EnglishDictionary>("moderation-settings", l)[k];
}
