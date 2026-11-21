// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-settings";
export default {
  ...en,
  site: "サイト情報",
  chrome: "サイト表示",
  showInHeader: "ヘッダーに{item}を表示",
  showInSidebar: "サイドバーに{item}を表示",
  more: "もっと見る",
  members: "メンバー",
  accessGroups: "アクセスグループ",
  themePreview: "テーマプレビュー",
} as const;
