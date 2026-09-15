import zh from "./messages/zh-notifications.js";
import en from "./messages/en-notifications.js";
import ja from "./messages/ja-notifications.js";
import it from "./messages/it-notifications.js";
import type { Locale } from "./auth.js";

type Widen<T> = T extends string ? string : { [K in keyof T]: Widen<T[K]> };
type Messages = Widen<typeof en>;

export const notificationResources: Record<Locale, Messages> = {
  zh,
  en,
  ja,
  it,
} as const;
