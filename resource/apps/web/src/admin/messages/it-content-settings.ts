// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-content-settings";
export default {
  ...en,
  announcementEditorVisual: "Visuale",
  announcementEditorVisualPlaceholder: "Modifica direttamente l’annuncio",
  announcementList: "Annunci",
  announcementAdd: "Aggiungi annuncio",
  announcementTitle: "Titolo dell’annuncio",
  announcementUntitled: "Annuncio senza titolo",
  announcementMoveUp: "Sposta su",
  announcementMoveDown: "Sposta giù",
  announcementReorder: "Trascina per riordinare",
  announcementEmpty: "Nessun annuncio",
  announcementEmptyHint: "Aggiungi un annuncio per impostare titolo, contenuto e ordine.",
  announcementContentRequired: "Gli annunci attivi richiedono un contenuto.",
  announcementOrderHint: "Mostra un annuncio alla volta nell’ordine dell’elenco.",

  externalLinks: "Link esterni nei post",
  externalLinksEnabled: "Mostra un avviso prima di lasciare il sito",
  externalLinksHint: "Si applica solo ai contenuti e alle risposte, non agli annunci. Se disattivato, o per i domini consentiti, il reindirizzamento è immediato.",
  externalLinksWhitelist: "Domini consentiti",
  externalLinksWhitelistHint: "Un nome host per riga, ad esempio example.com. Corrispondenza esatta; elenca separatamente i sottodomini, senza protocollo, percorso o porta.",
  posting: "Pubblicazione",
  announcement: "Annuncio",
  announcementContentHint: "Modifica in modalità visuale o Markdown; l’anteprima usa il layout pubblico.",
  announcementEditorPreview: "Anteprima",
  announcementEditorEmptyPreview: "Non c’è ancora alcun contenuto da visualizzare.",
  announcementEditorUploadImage: "Carica immagine",
  announcementEditorUploadFailed: "Caricamento immagine non riuscito",
  announcementEditorPlaceholder: "Scrivi l’annuncio in Markdown",
  exampleContent:
    "## Avviso di manutenzione\n\nStanotte il sito sarà in manutenzione dalle 22:00 alle 24:00. Potrebbe essere temporaneamente non disponibile.",
} as const;
