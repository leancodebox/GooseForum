// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-access-groups";
export default {
  ...en,
  memberAndAbove: "Membro o superiore",
  regularAndAbove: "Membro abituale o superiore",
  communityLeaderAndAbove: "Leader della comunità o superiore",
  systemEveryone: "Tutti",
  systemRegistered: "Utenti registrati",
  title: "Gruppi di accesso",
  create: "Nuovo gruppo",
  groups: "Gruppi",
  members: "Membri",
  categoryPermissions: "Permessi categorie",
  save: "Salva",
  cancel: "Annulla",
} as const;
