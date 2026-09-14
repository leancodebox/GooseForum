# Admin OAuth and OIDC migration record

- React routes: `/admin/settings/oauth`, `/admin/settings/oidc-provider`
- Vue references: `OAuthSettingsPage.vue`, `OIDCProviderSettingsPage.vue`
- OAuth preserves built-in/custom providers, enablement, client ID, secret keep/clear semantics, discovery, scopes, callback display, and custom-provider removal.
- OIDC preserves provider status, enablement, signing-key rotation, client create/edit/enablement, public/confidential invariants, PKCE, scopes/grants, client-secret rotation, and one-time secret disclosure/copy.
- Provider status and client lists load independently. Locale changes do not reset unsaved forms or revealed credentials.
- Client tests cover OAuth save, provider enablement, and client-secret rotation bodies. No live identity configuration was changed.
