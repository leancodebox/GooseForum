import { adminText, type Dictionary } from "@gooseforum/i18n/admin";
type Messages = Dictionary<"admin.system-settings">;
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

export type SystemSettingsTextKey = keyof Messages;
export function createSystemSettingsText(l: AuthLocale) {
  return (k: SystemSettingsTextKey) =>
    adminText(l, "admin.system-settings")(k);
}
