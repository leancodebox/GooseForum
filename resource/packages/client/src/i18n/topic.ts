import zh from "./messages/zh-topic.js";
import en from "./messages/en-topic.js";
import ja from "./messages/ja-topic.js";
import it from "./messages/it-topic.js";
import type { Locale } from "./auth.js";

type Widen<T> = T extends string ? string : { [K in keyof T]: Widen<T[K]> };
type Messages = Widen<typeof en>;

export const topicResources: Record<Locale, Messages> = {
  zh,
  en,
  ja,
  it,
} as const;
