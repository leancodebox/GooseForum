import { getAdminDictionary } from "./translation-loader";
type EnglishDictionary =
  typeof import("./messages/en-identity-settings").default;
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

export type IdentityTextKey = keyof EnglishDictionary;
export function createIdentityText(l: AuthLocale) {
  return (k: IdentityTextKey) =>
    getAdminDictionary<EnglishDictionary>("identity-settings", l)[k];
}
