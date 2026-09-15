// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-page-config";
export default {
  ...en,
  linksTitle: "Link",
  sponsorsTitle: "Sponsor",
  save: "Salva",
  cancel: "Annulla",
} as const;
