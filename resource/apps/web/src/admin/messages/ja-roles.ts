// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-roles";
export default {
  ...en,
  title: "ロール",
  create: "新規ロール",
  edit: "編集",
  delete: "削除",
  save: "保存",
  cancel: "キャンセル",
} as const;
