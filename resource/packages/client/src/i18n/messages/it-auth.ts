// Canonical translation data; compatibility exports and React loaders share this file.
import { localeLabels } from '../locale.js';
export default {
  locale: localeLabels.it,
  login: "Accedi",
  register: "Registrati",
  loginTitle: "Accedi al tuo account",
  registerTitle: "Crea un account",
  forgotTitle: "Reimposta password",
  loginSubtitle: "Bentornato. Continua le tue discussioni e i tuoi scritti.",
  registerSubtitle:
    "Unisciti a GooseForum e inizia il tuo spazio di discussione.",
  forgotSubtitle:
    "Inserisci la tua email e ti invieremo un messaggio per reimpostare la password.",
  usernameOrEmail: "Nome utente o email",
  username: "Nome utente",
  email: "Email",
  registeredEmail: "Email registrata",
  password: "Password",
  newPassword: "Nuova password",
  confirmPassword: "Conferma password",
  captcha: "Captcha",
  captchaAlt: "Captcha",
  refreshCaptcha: "Aggiorna captcha",
  forgotPassword: "Password dimenticata?",
  agreeTerms: "Ho letto e accetto i termini e l’informativa sulla privacy",
  createAccount: "Crea account",
  sendResetEmail: "Invia email di reimpostazione",
  backToLogin: "Torna all’accesso",
  continueWith: "Oppure continua con",
  resetPasswordTitle: "Reimposta password",
  resetPasswordSubtitle:
    "Imposta una nuova password di accesso. Dopo l’invio, torna ad accedere con essa.",
  resetMissingToken:
    "Nel link di reimpostazione manca un token. Riaprilo dalla tua email.",
  passwordMinLength: "La password deve contenere almeno 6 caratteri",
  saveNewPassword: "Salva nuova password",
  passwordAdviceTitle: "Consigli per la sicurezza della password",
  passwordAdviceDescription:
    "Imposta una nuova password usata solo per GooseForum. I link di reimpostazione hanno validità limitata; richiedi una nuova email dopo la scadenza.",
  passwordAdvice: {
    length: "Almeno 6 caratteri",
    unique: "Evita di riutilizzare password di altri siti",
    loginAfterReset:
      "Dopo la reimpostazione, torna ad accedere con la nuova password",
  },
  validation: {
    loginRequired: "Inserisci account, password e captcha",
    registerRequired: "Completa il modulo di registrazione",
    forgotRequired: "Inserisci email e captcha",
    passwordMismatch: "Le due password non coincidono",
    termsRequired: "Accetta prima i termini e l’informativa sulla privacy",
    captchaLoadFailed: "Caricamento del captcha non riuscito",
    loginFailed: "Accesso non riuscito",
    registerFailed: "Registrazione non riuscita",
    registerSuccess: "Registrazione riuscita",
    resetEmailFailed: "Invio dell’email di reimpostazione non riuscito",
  },
  server: {
    passwordResetMailQueued:
      "Se questa email è registrata, riceverai un’email per reimpostare la password",
    passwordResetSuccess: "Password reimpostata con successo",
    passwordResetFailed: "Reimpostazione della password non riuscita",
  },
} as const;
