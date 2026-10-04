export default {
  title: 'Autenticazione a due fattori', enabled: 'Attiva', disabled: 'Disattiva',
  cancel: 'Annulla configurazione',
  unavailable: "L'autenticazione a due fattori non è disponibile. Contatta l'amministratore.",
  code: 'Codice di autenticazione o recupero', verify: 'Verifica', enable: 'Configura autenticatore', confirm: 'Attiva autenticazione a due fattori',
  disable: 'Disattiva autenticazione a due fattori', regenerate: 'Rigenera codici di recupero', password: "Password dell'account",
  failed: 'Verifica non riuscita o scaduta. Controlla password e codice.',
  statusFailed: 'Impossibile caricare le impostazioni di autenticazione a due fattori.', retry: 'Riprova',
  secret: 'Chiave di configurazione', qr: "Codice QR per l'autenticatore", recoveryTitle: 'Codici di recupero',
  recoveryNotice: 'Salva i codici ora. Ogni codice è utilizzabile una sola volta. Tutti i dispositivi sono stati disconnessi.',
  download: 'Scarica codici di recupero', back: 'Torna al login', remaining: '{count} codici di recupero rimanenti',
  logoutNotice: 'La modifica disconnette tutti i dispositivi.',
  oauthPassword: "Se accedi tramite un provider, aggiungi e verifica un'email nel profilo, poi reimposta la password dell'account.",
} as const
