// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-notifications.js";
export default {
  ...en,
  trustLevel: "Livello di fiducia",
  templates: { ...en.templates, trustPromotion: "Promozione a {level}. Visualizza permessi e limiti" },
  title: "Notifiche",
  tabs: { all: "Tutte", unread: "Non lette" },
  markAllRead: "Segna tutte come lette",
  markRead: "Segna come letta",
  loadMore: "Carica altro",
  loadingMore: "Caricamento…",
  noMore: "Nessun'altra notifica",
} as const;
