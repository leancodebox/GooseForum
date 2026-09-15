// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-posts";
export default {
  ...en,
  title: "Contenuti",
  topics: "Discussioni",
  replies: "Risposte",
  search: "Cerca contenuti…",
  save: "Salva",
  cancel: "Annulla",
} as const;
