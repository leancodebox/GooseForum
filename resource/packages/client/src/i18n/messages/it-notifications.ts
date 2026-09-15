// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-notifications.js";
export default {
  ...en,
  title: "Notifiche",
  tabs: { all: "Tutte", unread: "Non lette" },
  markAllRead: "Segna tutte come lette",
  markRead: "Segna come letta",
  loadMore: "Carica altro",
  loadingMore: "Caricamento…",
  noMore: "Nessun'altra notifica",
} as const;
