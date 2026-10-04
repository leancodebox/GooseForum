# Admin OAuth and OIDC migration record

- React routes: `/admin/settings/oauth`, `/admin/settings/oidc-provider`
- Vue references: `OAuthSettingsPage.vue`, `OIDCProviderSettingsPage.vue`
- OAuth preserves built-in/custom providers, enablement, client ID, secret keep/clear semantics, discovery, scopes, callback display, and custom-provider removal.
- OIDC preserves provider status, enablement, signing-key rotation, client create/edit/enablement, public/confidential invariants, PKCE, scopes/grants, client-secret rotation, and one-time secret disclosure/copy.
- Provider status and client lists load independently. Locale changes do not reset unsaved forms or revealed credentials.
- Client tests cover OAuth save, provider enablement, and client-secret rotation bodies. No live identity configuration was changed.

## OIDC management

- The provider page displays issuer, discovery, authorization, token, UserInfo, JWKS and revocation addresses with copy actions.
- Client grants list usernames and IDs, scopes, creation and update times. Filtering is by exact user ID. Pagination uses `(client_id, user_id)` and reads 21 rows for a 20-row page, without count or offset.
- Administrators can revoke one user's grant or all grants for a client. Authorization codes and opaque access/refresh tokens are invalidated, and consent is removed. These actions cannot terminate sessions already created by another application or recall issued ID tokens.
- Client deletion disables the registration first, revokes its credentials, then deletes it. Failed cleanup can be retried. Disabled/deleted clients cannot use existing access tokens.
- An enabled provider with an undecryptable stored private key can reset it after confirmation. Reset revokes all OIDC credentials and grants, replaces the key and reloads the provider. Recover the original system secret instead when available.
- `app.secretKey` is the preferred configuration name; absent that field, the old `app.signingKey` is read. Renaming must retain the exact value. Signing-key rotation changes the independently generated OIDC RSA key, not the system secret.
- Administrator revocations, client deletions and signing-key recovery are saved in operation records.
