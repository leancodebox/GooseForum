import zh from "./messages/zh-settings.js";
import en from "./messages/en-settings.js";
import ja from "./messages/ja-settings.js";
import it from "./messages/it-settings.js";
import type { Locale } from "./auth.js";

type Widen<T> = T extends string ? string : { [K in keyof T]: Widen<T[K]> };
type SettingsMessages = Widen<typeof en>;

export const settingsResources: Record<Locale, SettingsMessages> = {
  zh,
  en,
  ja,
  it,
} as const;
