import { getAdminDictionary } from "./translation-loader";
type EnglishDictionary = typeof import("./messages/en-assets").default;
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

export type AssetTextKey = keyof EnglishDictionary;
export function createAssetText(locale: AuthLocale) {
  return (key: AssetTextKey) =>
    getAdminDictionary<EnglishDictionary>("assets", locale)[key];
}
