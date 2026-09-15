import { getAdminDictionary } from "./translation-loader";
type EnglishDictionary = typeof import("./messages/en-posts").default;
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

export type PostTextKey = keyof EnglishDictionary;
export function createPostText(locale: AuthLocale) {
  return (key: PostTextKey) =>
    getAdminDictionary<EnglishDictionary>("posts", locale)[key];
}
