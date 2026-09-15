// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-topic.js";
export default {
  ...en,
  originalImageSize: "元のサイズ",
  fitImage: "画面に合わせる",
  reply: "返信",
  like: "いいね",
  bookmark: "保存",
  watch: "フォロー",
  edit: "編集",
  delete: "削除",
  cancel: "キャンセル",
} as const;
