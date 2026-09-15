// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-topic.js";
export default {
  ...en,
  originalImageSize: "Dimensioni originali",
  fitImage: "Adatta allo schermo",
  reply: "Rispondi",
  like: "Mi piace",
  bookmark: "Salva",
  watch: "Segui",
  edit: "Modifica",
  delete: "Elimina",
  cancel: "Annulla",
} as const;
