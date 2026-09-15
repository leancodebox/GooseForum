// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-content-error.js";
export default {
  ...en,
  notFoundTitle: "Pagina non trovata",
  back: "Indietro",
  home: "Home",
} as const;
