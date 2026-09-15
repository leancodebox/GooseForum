// Canonical translation data; compatibility exports and React loaders share this file.
import en from "./en-server-messages.js";

export default {
  ...en,
  "page.notFound": "La pagina non esiste o è stata eliminata.",
  "route.notFound":
    "Route non trovata. Controlla l’URL e il metodo della richiesta.",
  "admin.moderator.userRequired": "Inserisci un utente moderatore.",
  "admin.moderator.userNotFound": "Utente moderatore non trovato.",
  "admin.moderator.notFound": "Record del moderatore non trovato.",
  "report.targetInvalid":
    "Il contenuto segnalato non esiste o non può essere segnalato.",
  "report.ownContent": "Non puoi segnalare i tuoi contenuti.",
  "report.duplicate": "Già segnalato e in attesa di revisione.",
  "report.createFailed":
    "Invio della segnalazione non riuscito. Riprova più tardi.",
  "report.notFound": "Segnalazione non trovata.",
  "upload.dailyLimit":
    "Hai caricato {count} file oggi e hai raggiunto il limite giornaliero",
  "upload.dailyLimit.avatar":
    "Hai caricato {count} file oggi. Caricare un avatar richiede {fileCount} slot e supererebbe il limite giornaliero",
  "auth.required": "Accedi per continuare.",
  "auth.signupDisabled": "La registrazione è attualmente disabilitata.",
  "auth.emailDomain.invalid": "L’indirizzo email non è valido.",
  "auth.emailDomain.notAllowed": "Questo dominio email non è consentito.",
  "auth.username.invalid": "Il formato del nome utente non è valido.",
  "auth.username.exists": "Questo nome utente è già in uso.",
  "auth.email.exists": "Questo indirizzo email è già in uso.",
  "auth.password.tooShort": "La password deve contenere almeno {minLength} caratteri.",
  "auth.password.tooLong": "La password è troppo lunga.",
  "auth.password.needsLetterNumber": "La password deve contenere lettere e numeri.",
  "auth.captcha.invalid": "Il codice di verifica è errato o scaduto.",
  "auth.register.failed": "Registrazione non riuscita. Riprova più tardi.",
  "auth.register.retryLogin": "Registrazione completata, ma l’accesso automatico non è riuscito. Accedi manualmente.",
  "auth.register.emailVerify": "Registrazione completata. Verifica il tuo indirizzo email.",
  "auth.login.success": "Accesso eseguito.",
  "auth.login.invalidRequest": "La richiesta di accesso è scaduta. Aggiorna la pagina.",
  "auth.password.invalidFormat": "Il formato della password non è valido.",
  "auth.credentials.invalid": "Nome utente, email o password non corretti.",
  "auth.account.frozen": "Questo account è sospeso.",
  "auth.email.unverified": "Verifica prima il tuo indirizzo email.",
  "auth.login.failed": "Accesso non riuscito.",
} as const;
