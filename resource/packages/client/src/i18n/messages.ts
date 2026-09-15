import zh from "./messages/zh-messages.js";
import en from "./messages/en-messages.js";
import ja from "./messages/ja-messages.js";
import it from "./messages/it-messages.js";
import type { Locale } from "./auth.js";

type Widen<T> = T extends string ? string : { [K in keyof T]: Widen<T[K]> };
type Messages = Widen<typeof en>;

export const messageResources: Record<Locale, Messages> = {
  zh,
  en,
  ja,
  it,
} as const;
