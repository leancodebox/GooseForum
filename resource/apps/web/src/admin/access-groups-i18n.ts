import { getAdminDictionary } from "./translation-loader";
type EnglishDictionary = typeof import("./messages/en-access-groups").default;
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

export type AccessGroupTextKey = keyof EnglishDictionary;

export function createAccessGroupText(locale: AuthLocale) {
  return (key: AccessGroupTextKey) =>
    getAdminDictionary<EnglishDictionary>("access-groups", locale)[key];
}
