// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-content-accessGroups.js";
export default {
  ...en,
  joinTitle: "アクセスグループへの参加申請",
  joinDescription:
    "グループに参加すると、許可された制限付きカテゴリにアクセスできます。",
  noJoinableGroups: "申請可能なグループはありません",
  joined: "参加済み",
  pending: "審査待ち",
  apply: "参加を申請",
  approve: "承認",
  reject: "却下",
} as const;
