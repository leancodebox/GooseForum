import zh from "./messages/zh-auth.js";
import en from "./messages/en-auth.js";
import ja from "./messages/ja-auth.js";
import it from "./messages/it-auth.js";
export { supportedLocales, normalizeLocale } from "./locale.js";
import type { Locale } from "./locale.js";
export type { Locale } from "./locale.js";

export interface AuthMessages {
  locale: { label: string; short: string };
  login: string;
  register: string;
  loginTitle: string;
  registerTitle: string;
  forgotTitle: string;
  loginSubtitle: string;
  registerSubtitle: string;
  forgotSubtitle: string;
  usernameOrEmail: string;
  username: string;
  email: string;
  registeredEmail: string;
  password: string;
  newPassword: string;
  confirmPassword: string;
  captcha: string;
  captchaAlt: string;
  refreshCaptcha: string;
  forgotPassword: string;
  agreeTerms: string;
  createAccount: string;
  sendResetEmail: string;
  backToLogin: string;
  continueWith: string;
  resetPasswordTitle: string;
  resetPasswordSubtitle: string;
  resetMissingToken: string;
  passwordMinLength: string;
  saveNewPassword: string;
  passwordAdviceTitle: string;
  passwordAdviceDescription: string;
  passwordAdvice: { length: string; unique: string; loginAfterReset: string };
  validation: {
    loginRequired: string;
    registerRequired: string;
    forgotRequired: string;
    passwordMismatch: string;
    termsRequired: string;
    captchaLoadFailed: string;
    loginFailed: string;
    registerFailed: string;
    registerSuccess: string;
    resetEmailFailed: string;
  };
  server: {
    passwordResetMailQueued: string;
    passwordResetSuccess: string;
    passwordResetFailed: string;
  };
}

export const authResources = { zh, en, ja, it } as const;
