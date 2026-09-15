// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-messages.js";
export default {
  ...en,
  title: "Messaggi",
  newMessage: "Nuovo messaggio",
  send: "Invia",
  sending: "Invio…",
  back: "Indietro",
  close: "Chiudi",
} as const;
