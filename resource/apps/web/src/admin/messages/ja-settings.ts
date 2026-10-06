// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-settings";
export default {
  ...en,
  site: "サイト情報",
  chrome: "サイト表示",
  more: "もっと見る",
  members: "メンバー",
  accessGroups: "アクセスグループ",
  themePreview: "テーマプレビュー",
} as const;
