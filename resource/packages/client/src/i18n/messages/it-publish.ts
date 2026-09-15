// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-publish.js";
export default {
  ...en,
  createTitle: "Pubblica argomento",
  editTitle: "Modifica argomento",
  saveDraft: "Salva bozza",
  publishTopic: "Pubblica",
  updateTopic: "Aggiorna",
} as const;
