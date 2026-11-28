import { adminText, type AdminText } from "@gooseforum/i18n/admin";
import type { SimpleKey } from "@gooseforum/i18n";
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

export type AssetTextKey = SimpleKey<"admin.assets">;
export type AssetText = AdminText<"admin.assets">;
export function createAssetText(locale: AuthLocale): AssetText {
  return adminText(locale, "admin.assets");
}
