// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-moderation.js";
export default {
  ...en,
  title: "モデレーターワークスペース",
  tabs: {
    reports: "報告",
    ban: "ブロック済み",
    logs: "操作ログ",
    guidance: "注意事項",
  },
} as const;
