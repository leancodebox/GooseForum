// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-topic.js";
export default {
  ...en,
  positionHelp: "上下にドラッグし、指を離すと移動します。投稿番号でも移動できます。",
  jumpToPost: "投稿番号に移動",
  jumpGo: "移動",
  originalImageSize: "元のサイズ",
  fitImage: "画面に合わせる",
  reply: "返信",
  replyToTopic: "トピックに返信",
  like: "いいね",
  bookmark: "保存",
  watch: "フォロー",
  edit: "編集",
  delete: "削除",
  cancel: "キャンセル",
} as const;
