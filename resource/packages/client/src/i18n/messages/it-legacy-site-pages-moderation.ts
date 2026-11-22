// Canonical translation data; compatibility exports and React loaders share this file.
export default {
  title: "Spazio di lavoro del moderatore",
  description:
    "Gestisci le attività di gestione della community nel tuo ambito.",
  notice:
    "Da un grande potere derivano grandi responsabilità. Prima di bloccare un post, assicurati che violi le regole della community e, quando opportuno, comunica prima.",
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
  total: "{count} elementi",
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
      time: "Ora",
      action: "Azione",
    },
    ban: "Blocca",
    reject: "Ignora",
    resolve: "Risolta",
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
