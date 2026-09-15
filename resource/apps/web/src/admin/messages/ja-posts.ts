// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-posts";
export default {
  ...en,
  title: "コンテンツ",
  topics: "トピック",
  replies: "返信",
  search: "コンテンツを検索…",
  save: "保存",
  cancel: "キャンセル",
} as const;
