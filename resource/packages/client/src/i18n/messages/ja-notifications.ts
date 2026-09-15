// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-notifications.js";
export default {
  ...en,
  title: "通知",
  tabs: { all: "すべて", unread: "未読" },
  markAllRead: "すべて既読",
  markRead: "既読にする",
  loadMore: "さらに読み込む",
  loadingMore: "読み込み中…",
  noMore: "通知は以上です",
} as const;
