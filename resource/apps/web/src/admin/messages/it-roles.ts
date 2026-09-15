// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-roles";
export default {
  ...en,
  title: "Ruoli",
  create: "Nuovo ruolo",
  edit: "Modifica",
  delete: "Elimina",
  save: "Salva",
  cancel: "Annulla",
} as const;
