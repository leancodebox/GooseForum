import zh from "./messages/zh-oidc-consent.js";
import en from "./messages/en-oidc-consent.js";
import ja from "./messages/ja-oidc-consent.js";
import it from "./messages/it-oidc-consent.js";
import type { Locale } from "./auth.js";

export interface OIDCConsentMessages {
  title: string;
  subtitle: string;
  verifying: string;
  expired: string;
  loadFailed: string;
  decisionFailed: string;
  back: string;
  accessAccount: string;
  permissions: string;
  clientId: string;
  trust: string;
  deny: string;
  approve: string;
  loading: string;
  scopes: Record<"openid" | "profile" | "email" | "offline_access", string>;
}

export const oidcConsentResources: Record<Locale, OIDCConsentMessages> = {
  zh,
  en,
  ja,
  it,
} as const;
