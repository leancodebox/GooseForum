import zh from "./messages/zh-content.js";
import en from "./messages/en-content.js";
import ja from "./messages/ja-content.js";
import it from "./messages/it-content.js";
import type { Locale } from "./auth.js";

type Widen<T> = T extends string ? string : { [K in keyof T]: Widen<T[K]> };
type ContentMessages = Widen<typeof en>;

export const contentResources: Record<Locale, ContentMessages> = {
  zh,
  en,
  ja,
  it,
} as const;
