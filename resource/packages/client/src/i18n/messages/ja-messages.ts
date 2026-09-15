// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-messages.js";
export default {
  ...en,
  title: "メッセージ",
  newMessage: "新規メッセージ",
  send: "送信",
  sending: "送信中…",
  back: "戻る",
  close: "閉じる",
} as const;
