// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-content-accessGroups.js";
export default {
  ...en,
  joinTitle: "Richiedi accesso ai gruppi",
  joinDescription:
    "Unisciti a un gruppo per accedere alle categorie riservate autorizzate.",
  noJoinableGroups: "Nessun gruppo accetta richieste",
  joined: "Iscritto",
  pending: "In attesa di revisione",
  apply: "Richiedi accesso",
  approve: "Approva",
  reject: "Rifiuta",
} as const;
