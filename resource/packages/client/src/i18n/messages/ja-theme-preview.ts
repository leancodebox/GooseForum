// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-theme-preview.js";
export default {
  ...en,
  pageTitle: "テーマプレビュー設定",
  saveDraft: "下書きを保存",
  publishSite: "サイトに公開",
} as const;
