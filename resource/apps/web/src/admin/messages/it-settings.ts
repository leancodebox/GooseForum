// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-settings";
export default {
  ...en,
  site: "Informazioni sito",
  chrome: "Aspetto sito",
} as const;
