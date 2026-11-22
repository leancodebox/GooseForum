// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-settings";
export default {
  ...en,
  assignAvatar: "Assegna un avatar predefinito ai nuovi utenti",
  assignAvatarHint: "Si applica ai nuovi utenti senza avatar. Senza immagine viene assegnato un avatar casuale della raccolta. Disattiva per mostrare l’iniziale del nome visualizzato.",
  defaultAvatar: "Immagine avatar predefinita",
  removeAvatar: "Rimuovi immagine",
  site: "Informazioni sito",
  chrome: "Aspetto sito",
  showInHeader: "Mostra {item} nella navigazione in alto",
  showInSidebar: "Mostra {item} nella barra laterale",
  more: "Altro",
  members: "Membri",
  accessGroups: "Gruppi di accesso",
  themePreview: "Anteprima tema",
} as const;
