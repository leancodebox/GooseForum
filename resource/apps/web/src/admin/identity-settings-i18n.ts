import { adminText, type Dictionary } from "@gooseforum/i18n/admin";
type Messages = Dictionary<"admin.identity-settings">;
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

export type IdentityTextKey = keyof Messages;
export function createIdentityText(l: AuthLocale) {
  return (k: IdentityTextKey) =>
    adminText(l, "admin.identity-settings")(k);
}
