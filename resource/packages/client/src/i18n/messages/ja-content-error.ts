// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-content-error.js";
export default {
  ...en,
  notFoundTitle: "ページが見つかりません",
  back: "戻る",
  home: "ホーム",
} as const;
