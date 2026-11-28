import { adminText, type Dictionary } from "@gooseforum/i18n/admin";
type Messages = Dictionary<"admin.dashboard">;
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

export type DashboardTextKey = keyof Messages;
export function createDashboardText(l: AuthLocale) {
  return (k: DashboardTextKey) =>
    adminText(l, "admin.dashboard")(k);
}
