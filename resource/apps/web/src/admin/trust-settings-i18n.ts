import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";
import { getAdminDictionary } from "./translation-loader";

type Dictionary = typeof import("./messages/en-trust-settings").default;
export type TrustSettingsText = (key: keyof Dictionary, params?: Record<string, string>) => string;

export function createTrustSettingsText(locale: AuthLocale): TrustSettingsText {
  return (key, params = {}) => getAdminDictionary<Dictionary>("trust-settings", locale)[key]
    .replace(/\{(\w+)\}/g, (match, name: string) => params[name] ?? match);
}
