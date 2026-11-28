import { adminText, type Dictionary } from "@gooseforum/i18n/admin";
type Messages = Dictionary<"admin.content-settings">;
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

export type ContentSettingsTextKey = keyof Messages;
export function createContentSettingsText(l: AuthLocale) {
  return (k: ContentSettingsTextKey) =>
    adminText(l, "admin.content-settings")(k);
}
