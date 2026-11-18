// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-topic.js";
export default {
  ...en,
  positionHelp: "Trascina in verticale e rilascia per spostarti, oppure inserisci il numero del messaggio.",
  jumpToPost: "Vai al messaggio",
  jumpGo: "Vai",
  originalImageSize: "Dimensioni originali",
  fitImage: "Adatta allo schermo",
  reply: "Rispondi",
  replyToTopic: "Rispondi alla discussione",
  like: "Mi piace",
  bookmark: "Salva",
  watch: "Segui",
  edit: "Modifica",
  delete: "Elimina",
  cancel: "Annulla",
} as const;
