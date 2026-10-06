// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-settings";
export default {
  ...en,
  site: "Informazioni sito",
  chrome: "Aspetto sito",
  more: "Altro",
  members: "Membri",
  accessGroups: "Gruppi di accesso",
  themePreview: "Anteprima tema",
} as const;
