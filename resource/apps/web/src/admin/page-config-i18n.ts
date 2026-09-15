import { getAdminDictionary } from "./translation-loader";
type EnglishDictionary = typeof import("./messages/en-page-config").default;
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

export type PageConfigTextKey = keyof EnglishDictionary;
export function createPageConfigText(locale: AuthLocale) {
  return (key: PageConfigTextKey) =>
    getAdminDictionary<EnglishDictionary>("page-config", locale)[key];
}
