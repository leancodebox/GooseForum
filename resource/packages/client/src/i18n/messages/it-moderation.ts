// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-moderation.js";
export default {
  ...en,
  title: "Area moderazione",
  tabs: {
    reports: "Segnalazioni",
    ban: "Contenuti bloccati",
    logs: "Registro attività",
    guidance: "Promemoria",
  },
} as const;
