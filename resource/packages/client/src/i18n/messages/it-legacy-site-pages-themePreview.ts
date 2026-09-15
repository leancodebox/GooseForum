// Canonical translation data; compatibility exports and React loaders share this file.
export default {
  presetsTitle: "Preset",
  presetApplied: "Preset {name} applicato. Salva la bozza per mantenerlo.",
  saveDraft: "Salva bozza",
  publishSite: "Pubblica sul sito",
  workflowTitle: "Come funziona:",
  workflowDescription:
    "Stai modificando una bozza del tema. Il salvataggio della bozza non influisce sul sito live; dopo la pubblicazione, il tema personalizzato subentra solo quando l’interruttore di abilitazione a sinistra è attivo.",
  draftSaved: "Bozza del tema salvata. Il sito live è invariato.",
  published: "Tema pubblicato sul sito live.",
  saveFailed: "Salvataggio della bozza del tema non riuscito",
  publishFailed: "Pubblicazione del tema non riuscita",
  restoredToSaved: "Ripristinato alla configurazione salvata",
  themeRestoredDefault: "{name} ripristinato al valore predefinito integrato",
  allRestoredDefault:
    "Ripristinati i temi predefiniti integrati. Salva la bozza di pre-pubblicazione per pubblicare su tutto il sito.",
  cssCopied: "CSS copiato",
  cssCopyFailed: "Copia del CSS non riuscita",
  pageTitle: "Impostazioni anteprima tema",
  resetDefault: "Predefinito",
  restoreToSavedTitle: "Ripristina alla configurazione salvata",
  enableLabel: "Abilita",
  tabLatest: "Recenti",
  tabHot: "Di tendenza",
  tabFeatured: "In evidenza",
  samplePublish: "Pubblica argomento",
  sampleTopic1:
    "Discussione sul refactoring del sistema di temi: colori, raggi e stati dei componenti",
  sampleTopic2: "Leggibilità del testo Markdown in modalità scura",
  sampleTopic3:
    "Controllo visivo dell’onboarding dei nuovi utenti e delle notifiche",
  sampleTopicExcerpt:
    "Osserva come testo attenuato, divisori, tag e sfondi al passaggio del mouse si dispongono a livelli con il tema attuale.",
  sampleTopicTitle: "Il titolo di un post di discussione",
  sampleTopicBody:
    "Questo simula testo, link, citazioni e blocchi di codice. Le variabili del tema dovrebbero mantenere il contenuto leggibile sia in modalità chiara che scura, in particolare testo, bordi e testo attenuato.",
  sampleTopicQuote:
    "La gerarchia dei colori dovrebbe essere discreta, ma mai vaga.",
  sampleMessageIncoming: "Questo sfondo è comodo?",
  sampleMessageOutgoing: "Il contrasto deve rimanere solido.",
  sampleTextarea: "Le variabili del tema coprono input, focus e testo.",
  presets: {
    goose: {
      label: "Goose",
      description: "Predefinito attuale, pulito e stabile",
    },
    clean: {
      label: "Clean",
      description: "Orientato al prodotto, blu e verde acqua tenui",
    },
    warm: {
      label: "Warm",
      description: "Calore più tenue per gli spazi della community",
    },
    deep: {
      label: "Deep",
      description: "Dominato dal nero con contrasto più profondo",
    },
    cupcake: {
      label: "Cupcake",
      description: "Rosa tenue, leggero e dolce",
    },
    retro: {
      label: "Retro",
      description: "Toni vintage caldi per community informali",
    },
    synthwave: {
      label: "Synthwave",
      description: "Viola e ciano neon con presenza più decisa",
    },
    aqua: {
      label: "Aqua",
      description: "Ciano fresco con contrasto brillante",
    },
    forest: {
      label: "Forest",
      description: "Verde a livelli per community di conoscenza",
    },
  },
} as const;
