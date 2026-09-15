// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-content-drafts.js";
export default {
  ...en,
  title: "Bozze",
  total: "{count} bozze",
  newDraft: "Nuova bozza",
  edit: "Continua a modificare",
  untitled: "Bozza senza titolo",
  emptyTitle: "Nessuna bozza",
} as const;
