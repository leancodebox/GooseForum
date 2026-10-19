// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-notifications.js";
export default {
  ...en,
  trustLevel: "信頼レベル",
  templates: { ...en.templates, trustPromotion: "「{level}」に昇格しました。権限と上限を確認" },
  title: "通知",
  tabs: { all: "すべて", unread: "未読" },
  markAllRead: "すべて既読",
  markRead: "既読にする",
  loadMore: "さらに読み込む",
  loadingMore: "読み込み中…",
  noMore: "通知は以上です",
} as const;
