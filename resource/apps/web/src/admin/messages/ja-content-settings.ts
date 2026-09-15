// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-content-settings";
export default {
  ...en,
  posting: "投稿設定",
  announcement: "お知らせ",
} as const;
