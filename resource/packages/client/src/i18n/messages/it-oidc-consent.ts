// Canonical translation data; compatibility exports and React loaders share this file.
export default {
  title: "Accedi con {site}",
  subtitle: "Richiesta di autorizzazione",
  verifying: "Verifica della richiesta…",
  expired: "La richiesta è scaduta. Torna all’app e accedi di nuovo.",
  loadFailed: "Impossibile caricare la richiesta di autorizzazione.",
  decisionFailed: "Impossibile elaborare la richiesta di autorizzazione.",
  back: "Torna a {site}",
  accessAccount: "Vuole accedere al tuo account {site}",
  permissions: "Questa app potrà:",
  clientId: "ID client",
  trust:
    "Assicurati di fidarti di questa app. Puoi revocare l’accesso nelle impostazioni dell’account in qualsiasi momento.",
  deny: "Rifiuta",
  approve: "Consenti e continua",
  loading: "Elaborazione…",
  scopes: {
    openid: "Confermare la tua identità",
    profile: "Leggere nome, nome utente e avatar",
    email: "Leggere email e stato di verifica",
    offline_access: "Mantenere l’accesso quando non sei presente",
  },
} as const;
