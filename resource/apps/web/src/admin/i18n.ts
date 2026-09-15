import { getAdminDictionary } from "./translation-loader";
type EnglishDictionary = typeof import("./messages/en-shell").default;
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

export type AdminTextKey = keyof EnglishDictionary;

export function createAdminText(locale: AuthLocale) {
  return (key: AdminTextKey) =>
    getAdminDictionary<EnglishDictionary>("shell", locale)[key];
}
