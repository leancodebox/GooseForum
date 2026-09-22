// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-identity-settings";
export default {
  ...en,
  oauth: "Accesso OAuth",
  oidc: "Provider OIDC",
  retry: "Riprova",
  loadFailedHint: "Le impostazioni OAuth non sono temporaneamente disponibili. Controlla la rete o lo stato del server e riprova.",
  providerCheckFailed: "Impossibile verificare la configurazione del provider. Controlla le credenziali, l'URL Discovery e la connessione di rete.",
  callbackHint: "Registra l'URL di callback completo nel provider OAuth. Passa il mouse sui segmenti per vederne l'origine.",
  callbackSiteHint: "Proviene dall'URL configurato nelle informazioni del sito.",
  callbackRouteHint: "Percorso di callback OAuth fisso di GooseForum.",
  callbackProviderHint: "Generato dalla chiave del provider corrente.",
  callbackProviderPlaceholder: "provider-key/callback",
  copyCallback: "Copia l'URL di callback completo",
  callbackCopied: "URL di callback copiato",
  copyFailed: "Copia non riuscita",
  removeProvider: "Rimuovi provider",
  removeProviderTitle: "Rimuovere questo provider OAuth?",
  removeProviderHint: "Il provider verrà rimosso al salvataggio delle impostazioni:",
  addScope: "Aggiungi scope",
  removeScope: "Rimuovi scope",
} as const;
