// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-page-config";
export default {
  ...en,
  linksTitle: "リンク",
  sponsorsTitle: "スポンサー",
  save: "保存",
  cancel: "キャンセル",
} as const;
