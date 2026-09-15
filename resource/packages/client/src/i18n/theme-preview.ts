import zh from "./messages/zh-theme-preview.js";
import en from "./messages/en-theme-preview.js";
import ja from "./messages/ja-theme-preview.js";
import it from "./messages/it-theme-preview.js";
import type { Locale } from "./auth.js";

type Widen<T> = T extends string ? string : { [K in keyof T]: Widen<T[K]> };
type Messages = Widen<typeof en>;

export const themePreviewResources: Record<Locale, Messages> = {
  zh,
  en,
  ja,
  it,
} as const;
