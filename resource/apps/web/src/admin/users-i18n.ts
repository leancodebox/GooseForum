import { adminText, type Dictionary } from "@gooseforum/i18n/admin";
type Messages = Dictionary<"admin.users">;
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

export type UserTextKey = keyof Messages;
export function createUserText(locale: AuthLocale) {
  return (key: UserTextKey) =>
    adminText(locale, "admin.users")(key);
}
