// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-publish.js";
export default {
  ...en,
  createTitle: "トピックを投稿",
  editTitle: "トピックを編集",
  saveDraft: "下書きを保存",
  publishTopic: "トピックを投稿",
  updateTopic: "トピックを更新",
} as const;
