// These endpoints use their protocol-native JSON responses instead of the
// GooseForum API envelope, so they are intentionally kept out of api/routes.ts.
export const protocolRoutes = {
  oidcConsentDetails: ['GET', '/oauth2/consent/details'],
  oidcConsentDecision: ['POST', '/oauth2/consent'],
} as const
