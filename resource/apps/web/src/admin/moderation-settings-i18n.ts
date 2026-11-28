import { adminText, type Dictionary } from "@gooseforum/i18n/admin";
type Messages = Dictionary<"admin.moderation-settings">;
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

export type ModerationSettingsTextKey = keyof Messages;
export function createModerationSettingsText(l: AuthLocale) {
  return (k: ModerationSettingsTextKey) =>
    adminText(l, "admin.moderation-settings")(k);
}
