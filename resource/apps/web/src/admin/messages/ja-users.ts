// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-users";
export default {
  ...en,
  title: "ユーザー",
  search: "ユーザー名を検索…",
  edit: "ユーザーを編集",
  save: "変更を保存",
  cancel: "キャンセル",
} as const;
