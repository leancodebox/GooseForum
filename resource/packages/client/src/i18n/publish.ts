import zh from "./messages/zh-publish.js";
import en from "./messages/en-publish.js";
import ja from "./messages/ja-publish.js";
import it from "./messages/it-publish.js";
import type { Locale } from "./auth.js";

type Widen<T> = T extends string ? string : { [K in keyof T]: Widen<T[K]> };
type Messages = Widen<typeof en>;

export const publishResources: Record<Locale, Messages> = {
  zh,
  en,
  ja,
  it,
} as const;
