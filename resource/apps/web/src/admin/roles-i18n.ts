import { adminText, type Dictionary } from "@gooseforum/i18n/admin";
type Messages = Dictionary<"admin.roles">;
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

export type RoleTextKey = keyof Messages;
export function createRoleText(locale: AuthLocale) {
  return (key: RoleTextKey) =>
    adminText(locale, "admin.roles")(key);
}
