import { adminText, type Dictionary } from "@gooseforum/i18n/admin";
type Messages = Dictionary<"admin.page-config">;
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

export type PageConfigTextKey = keyof Messages;
export function createPageConfigText(locale: AuthLocale) {
  return (key: PageConfigTextKey) =>
    adminText(locale, "admin.page-config")(key);
}
