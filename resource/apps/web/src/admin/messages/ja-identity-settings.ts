// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-identity-settings";
export default {
  ...en,
  oauth: "OAuth ログイン",
  oidc: "OIDC プロバイダー",
} as const;
