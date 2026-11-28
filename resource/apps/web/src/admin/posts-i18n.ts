import { adminText, type Dictionary } from "@gooseforum/i18n/admin";
type Messages = Dictionary<"admin.posts">;
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

export type PostTextKey = keyof Messages;
export function createPostText(locale: AuthLocale) {
  return (key: PostTextKey) =>
    adminText(locale, "admin.posts")(key);
}
