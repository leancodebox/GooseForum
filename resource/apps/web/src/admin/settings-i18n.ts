import { adminText, type AdminText } from "@gooseforum/i18n/admin";
import type { SimpleKey } from "@gooseforum/i18n";
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

export type SettingsTextKey = SimpleKey<"admin.settings">;
export type SettingsText = AdminText<"admin.settings">;
export function createSettingsText(locale: AuthLocale): SettingsText {
  return adminText(locale, "admin.settings");
}
