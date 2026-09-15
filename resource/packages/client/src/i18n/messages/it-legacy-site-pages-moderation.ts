// Canonical translation data; compatibility exports and React loaders share this file.
export default {
  title: "Spazio di lavoro del moderatore",
  description:
    "Gestisci le attività di gestione della community nel tuo ambito.",
  notice:
    "Da un grande potere derivano grandi responsabilità. Prima di bloccare un post, assicurati che violi le regole della community e, quando opportuno, comunica prima.",
  guidanceTitle: "Promemoria per la moderazione",
  guidanceDescription:
    "Le azioni dei moderatori plasmano la community. Verifica i fatti, le regole e l’impatto prima di agire.",
  guidanceItems: {
    rule: {
      title: "Controlla prima le regole",
      description:
        "Non agire solo per disaccordo. Inizia confermando se il post viola chiaramente le regole della community.",
    },
    context: {
      title: "Leggi tutto il contesto",
      description:
        "Esamina titolo, corpo, categoria e contesto della discussione prima di decidere.",
    },
    restraint: {
      title: "Usa il potere con moderazione",
      description:
        "Preferisci promemoria e dialogo quando possono funzionare. Il blocco è per i casi chiari e necessari.",
    },
  },
  allTitle: "Tutti i post",
  allDescription:
    "Esamina i post nella categoria attuale e agisci in base allo stato.",
  blockedTitle: "Post bloccati",
  blockedDescription:
    "Esamina i post bloccati nella categoria attuale e ripristinali quando serve.",
  allEmptyTitle: "Nessun post",
  blockedEmptyTitle: "Nessun post bloccato",
  banTitle: "Post che puoi bloccare",
  banDescription:
    "Esamina i post normali nelle tue categorie e bloccali quando serve.",
  unbanTitle: "Post bloccati",
  unbanDescription:
    "Esamina i post bloccati nelle tue categorie e ripristinali quando serve.",
  banEmptyTitle: "Nessun post da bloccare",
  unbanEmptyTitle: "Nessun post bloccato",
  banAction: "Blocca",
  unbanAction: "Ripristina",
  total: "{count} elementi",
  knownTotal: "{count} elementi",
  knownTotalMore: "{count}+ elementi",
  blocked: "Bloccato",
  emptyDescription: "Al momento non c’è nulla da gestire nel tuo ambito.",
  managementTabs: {
    reports: "Segnalazioni",
    ban: "Post bloccati",
    logs: "Registro attività",
    guidance: "Promemoria",
    posts: "Moderazione",
    members: "Membri",
    settings: "Regole",
  },
  reports: {
    loading: "Caricamento segnalazioni",
    loadMore: "Carica altro",
    noMore: "Nessun’altra segnalazione",
    emptyTitle: "Nessuna segnalazione in sospeso",
    emptyDescription: "Le segnalazioni degli utenti appariranno qui.",
    statusTabs: {
      open: "In sospeso",
      closed: "Gestite",
    },
    table: {
      report: "Segnalazione",
      reason: "Info segnalazione",
      people: "Persone",
      time: "Ora",
      action: "Azione",
    },
    ban: "Blocca",
    hide: "Blocca",
    reject: "Ignora",
    resolve: "Risolta",
    reasonLabel: "Motivo",
    statusLabel: "Stato",
    submittedAtLabel: "Inviata",
    handledAtLabel: "Gestita",
    reporterLabel: "Segnalante",
    handlerLabel: "Gestore",
    noExcerpt: "Nessun estratto",
    targetTypes: {
      topic: "Topic",
      post: "Post",
    },
    reasons: {
      spam: "Spam",
      abuse: "Molestie",
      illegal: "Contenuto illegale",
      irrelevant: "Fuori tema",
      other: "Altro",
    },
    resolutions: {
      banned: "Bloccato",
      ignored: "Ignorato",
      resolved: "Gestito",
    },
  },
  logs: {
    title: "Registro attività",
    description:
      "Esamina le azioni della console di moderazione in ordine cronologico, così le decisioni possono essere tracciate in seguito.",
    loading: "Caricamento log",
    loadMore: "Carica altro",
    noMore: "Nessun altro log",
    emptyTitle: "Ancora nessuna attività",
    emptyDescription:
      "Blocchi, ripristini e modifiche dell’ambito dei moderatori appariranno qui.",
    table: {
      operation: "Operazione",
      time: "Ora",
    },
    actions: {
      topicBlocked: "ha bloccato",
      topicUnblocked: "ha ripristinato",
      replyBlocked: "ha nascosto",
      replyUnblocked: "ha ripristinato",
      postBlocked: "ha nascosto",
      postUnblocked: "ha ripristinato",
      reportResolved: "ha risolto la segnalazione",
      reportRejected: "ha ignorato la segnalazione",
      categoryModeratorAdded: "ha aggiunto un ambito di moderazione",
      categoryModeratorRemoved: "ha rimosso un ambito di moderazione",
      operation: "ha aggiornato",
    },
  },
  tabs: {
    all: "Tutti",
    blocked: "Bloccati",
    ban: "Tutti",
    unban: "Bloccati",
  },
  table: {
    topic: "Post",
    updatedAt: "Aggiornato",
    author: "Autore",
    activity: "Attività",
    action: "Azione",
  },
  meta: {
    views: "Visualizzazioni",
    replies: "Risposte",
  },
} as const;
