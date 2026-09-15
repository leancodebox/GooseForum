// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-content-settings";
export default {
  ...en,
  posting: "Pubblicazione",
  announcement: "Annuncio",
  exampleContent:
    "## Avviso di manutenzione\n\nStanotte il sito sarà in manutenzione dalle 22:00 alle 24:00. Potrebbe essere temporaneamente non disponibile.",
} as const;
