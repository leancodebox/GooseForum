import zh from "./messages/zh-moderation.js";
import en from "./messages/en-moderation.js";
import ja from "./messages/ja-moderation.js";
import it from "./messages/it-moderation.js";
import type { Locale } from "./auth.js";

type Widen<T> = T extends string ? string : { [K in keyof T]: Widen<T[K]> };
type Messages = Widen<typeof en>;

export const moderationResources: Record<Locale, Messages> = {
  zh,
  en,
  ja,
  it,
} as const;
