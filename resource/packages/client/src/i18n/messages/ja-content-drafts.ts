// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-content-drafts.js";
export default {
  ...en,
  title: "下書き",
  total: "下書き {count} 件",
  newDraft: "新しい下書き",
  edit: "編集を続ける",
  untitled: "無題の下書き",
  emptyTitle: "下書きはありません",
} as const;
