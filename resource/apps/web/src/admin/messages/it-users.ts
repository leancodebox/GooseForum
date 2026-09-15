// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-users";
export default {
  ...en,
  title: "Utenti",
  search: "Cerca nome utente…",
  edit: "Modifica utente",
  save: "Salva modifiche",
  cancel: "Annulla",
} as const;
