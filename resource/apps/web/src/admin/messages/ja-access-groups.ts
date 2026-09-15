// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-access-groups";
export default {
  ...en,
  title: "アクセスグループ",
  create: "新規グループ",
  groups: "グループ",
  members: "メンバー",
  categoryPermissions: "カテゴリー権限",
  save: "保存",
  cancel: "キャンセル",
} as const;
