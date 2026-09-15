// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-settings.js";
export default {
  ...en,
  tabs: {
    profile: "Profilo",
    account: "Account",
    privacy: "Privacy",
    binding: "Collegamenti",
    applications: "App autorizzate",
  },
  save: "Salva",
  cancel: "Annulla",
  edit: "Modifica",
} as const;
