// Canonical translation data; compatibility exports and React loaders share this file.
export default {
  title: "Sign in with {site}",
  subtitle: "Authorization request",
  verifying: "Verifying authorization request…",
  expired: "This request has expired. Return to the app and sign in again.",
  loadFailed: "Unable to load the authorization request.",
  decisionFailed: "Unable to process the authorization request.",
  back: "Return to {site}",
  accessAccount: "Wants to access your {site} account",
  permissions: "This app will be allowed to:",
  clientId: "Client ID",
  trust:
    "Make sure you trust this app. You can revoke access in your account settings at any time.",
  deny: "Deny",
  approve: "Allow and continue",
  loading: "Working…",
  scopes: {
    openid: "Confirm your identity",
    profile: "Read your name, username, and avatar",
    email: "Read your email and verification status",
    offline_access: "Keep you signed in while you are away",
  },
} as const;
