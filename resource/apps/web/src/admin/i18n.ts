import { adminText, type Dictionary } from "@gooseforum/i18n/admin";
type Messages = Dictionary<"admin.shell">;
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

export type AdminTextKey = keyof Messages;

export function createAdminText(locale: AuthLocale) {
  return (key: AdminTextKey) =>
    adminText(locale, "admin.shell")(key);
}
