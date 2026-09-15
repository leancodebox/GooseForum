import { getAdminDictionary } from "./translation-loader";
type EnglishDictionary = typeof import("./messages/en-roles").default;
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

export type RoleTextKey = keyof EnglishDictionary;
export function createRoleText(locale: AuthLocale) {
  return (key: RoleTextKey) =>
    getAdminDictionary<EnglishDictionary>("roles", locale)[key];
}
