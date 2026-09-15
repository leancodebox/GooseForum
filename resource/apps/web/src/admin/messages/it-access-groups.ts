// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-access-groups";
export default {
  ...en,
  title: "Gruppi di accesso",
  create: "Nuovo gruppo",
  groups: "Gruppi",
  members: "Membri",
  categoryPermissions: "Permessi categorie",
  save: "Salva",
  cancel: "Annulla",
} as const;
