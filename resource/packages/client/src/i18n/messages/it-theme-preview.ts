// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-theme-preview.js";
export default {
  ...en,
  pageTitle: "Anteprima tema",
  saveDraft: "Salva bozza",
  publishSite: "Pubblica sul sito",
} as const;
