// Canonical translation data; compatibility exports and React loaders share this file.
export default {
  moderationPending: "Contenuto salvato. Sarà pubblico dopo l’approvazione.",
  moderationRejected:
    "Il contenuto è stato salvato ma non ha superato la moderazione e non è pubblico. Modificalo e riprova oppure contatta un amministratore.",
  createTitle: "Pubblica argomento",
  editTitle: "Modifica argomento",
  subtitle:
    "Scrivi un titolo chiaro e scegli categorie adatte per rendere la discussione più facile da trovare.",
  titlePlaceholder: "Inserisci il titolo dell’argomento",
  maxCategories: "Fino a 3",
  mainCategoryHint:
    "La prima categoria, “{category}”, decide chi può leggere questa discussione",
  mainCategoryBadge: "Principale",
  mainCategoryChangeTitle: "Cambiare la categoria principale?",
  mainCategoryChangeDescription:
    "Rimuovendo “{current}”, “{next}” diventerà la categoria principale e cambierà chi può leggere questa discussione.",
  confirmMainCategoryChange: "Cambia categoria principale",
  restrictedCategoryHint:
    "Visibile solo ai membri con accesso a questa categoria",
  restrictedCategorySingleHint:
    "Una categoria riservata deve essere selezionata da sola; selezionandola sostituisci le categorie correnti",
  noCreatePermission:
    "Non puoi più pubblicare nelle categorie selezionate, ma puoi ancora salvare una bozza.",
  markdownMode: "Markdown",
  visualMode: "Editor",
  preview: "Anteprima",
  pastePlainText: "Incolla come testo semplice",
  clipboardReadFailed:
    "Impossibile leggere gli appunti. Controlla le autorizzazioni del browser.",
  visualUnsupported:
    "Il contenuto include elenchi di attività. Continua a modificarlo in modalità Markdown.",
  processingImage: "Elaborazione immagine...",
  processingImages: "Elaborazione immagini {done}/{total}",
  imageInserted: "Immagine inserita.",
  imagesInserted: "{count} immagini inserite.",
  moreImageFailures: "; altre {count} non riuscite",
  noUploadableImages: "Nessun file immagine caricabile",
  topicUpdated: "Argomento aggiornato.",
  topicPublished: "Argomento pubblicato.",
  saveFailed: "Salvataggio non riuscito",
  draftSaveFailed: "Salvataggio della bozza non riuscito",
  saveDraft: "Salva bozza",
  updateTopic: "Aggiorna argomento",
  publishTopic: "Pubblica argomento",
  uploadImageTitle:
    "Carica immagini; selezione multipla, incolla e trascinamento supportati",
  bodyPlaceholder:
    "Inserisci il testo, Markdown supportato; incolla o trascina le immagini qui",
  visualPlaceholder: "Scrivi e formatta direttamente il contenuto",
  dropToUpload: "Rilascia per caricare e inserire le immagini",
  emptyPreview: "Non c’è ancora contenuto da visualizzare in anteprima.",
  selectedCategories: "Categorie selezionate",
  leaveTitle: "Salvare le modifiche non finite?",
  leaveDescription:
    "Il contenuto attuale non è stato salvato. Salvalo come bozza prima di uscire, così potrai continuare più tardi.",
  draftRequirement:
    "Una bozza richiede un titolo, un corpo e almeno una categoria prima di poter essere salvata.",
  continueEditing: "Continua a modificare",
  leaveWithoutSaving: "Esci senza salvare",
  fields: {
    title: "Titolo",
    category: "Categoria",
    body: "Corpo",
  },
  validation: {
    requiredFields:
      "Aggiungi un titolo, seleziona almeno una categoria e completa il contenuto.",
    categoryRequired: "Seleziona almeno una categoria prima di pubblicare.",
  },
  toolbar: {
    bold: "Grassetto",
    italic: "Corsivo",
    strike: "Barrato",
    inlineCode: "Codice inline",
    link: "Link",
    linkUrl: "Inserisci URL",
    applyLink: "Applica",
    quote: "Citazione",
    code: "Blocco di codice",
    bulletList: "Elenco puntato",
    orderedList: "Elenco numerato",
    horizontalRule: "Linea divisoria",
    hardBreak: "Interruzione forzata",
    table: "Inserisci tabella",
    tableSize: "{rows} righe × {columns} colonne",
    blockType: "Formato paragrafo",
    paragraph: "Paragrafo",
    heading: "Titolo {level}",
    codeBlock: "Blocco di codice",
  },
  placeholder: {
    bold: "testo in grassetto",
    italic: "testo in corsivo",
    strike: "testo barrato",
    link: "testo del link",
    quote: "contenuto citato",
    listItem: "voce elenco",
  },
  checklist: {
    title: "Checklist di pubblicazione",
    done: "Compilato",
    pending: "In sospeso",
    characters: "{count} caratteri",
  },
} as const;
