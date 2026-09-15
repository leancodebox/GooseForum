import { getAdminDictionary } from "./translation-loader";
type EnglishDictionary = typeof import("./messages/en-users").default;
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

export type UserTextKey = keyof EnglishDictionary;
export function createUserText(locale: AuthLocale) {
  return (key: UserTextKey) =>
    getAdminDictionary<EnglishDictionary>("users", locale)[key];
}
