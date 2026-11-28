import { adminText } from "@gooseforum/i18n/admin";
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

export type TrustSettingsText = ReturnType<typeof adminText<"admin.trust-settings">>;

export function createTrustSettingsText(locale: AuthLocale): TrustSettingsText {
  return adminText(locale, "admin.trust-settings");
}
