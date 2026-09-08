# OIDC Provider Core

`oidcprovider` is GooseForum's transport- and persistence-independent OAuth
2.0/OpenID Connect provider core. It does not import `net/http`, Gin, GORM, or
GooseForum user models.

## Supported profile

- OpenID Connect Authorization Code Flow
- PKCE using `S256` only; mandatory for public clients and configurable for confidential clients
- public clients using `none`
- confidential clients using `client_secret_basic` or `client_secret_post`
- opaque access tokens
- rotating opaque refresh tokens with family revocation on reuse
- fixed refresh-token family maximum lifetime that rotation cannot extend
- `openid`, `profile`, `email`, and `offline_access` scopes
- UserInfo claims (profile/email scope claims are returned here, not implicitly in ID Tokens)
- RS256 ID Tokens and JWKS key rotation
- Discovery metadata
- RFC 7009 token revocation
- `prompt=none`, `prompt=login`, `prompt=consent`, and `max_age`
- RFC 9207 authorization response issuer

Implicit, hybrid, password, client credentials, dynamic client registration,
introspection, device authorization, logout protocols, pairwise subjects, and
encrypted ID Tokens are intentionally not advertised.

## Host responsibilities

The host adapter must:

1. Parse HTTP requests without accepting duplicate security-sensitive fields.
2. Distinguish Basic, form-post, and unauthenticated client authentication and
   populate `ClientAuthenticationMethod` accurately.
3. Never redirect an `OAuthError` unless `RedirectAllowed` is true.
4. Preserve `state` exactly and include `AuthorizeResult.Issuer` as `iss`.
5. Authenticate the resource owner and provide trusted `Authentication` state.
6. Require TLS in production and set appropriate no-store/cache-control headers
   on authorization and token responses.
7. Encode scopes as one space-delimited string at the HTTP boundary.
8. Encrypt private signing keys at rest and retain old public keys until every
   ID Token they signed has expired.

## Store transaction contract

`ExchangeAuthorizationCode` must atomically validate and consume the code and
persist its access/refresh tokens. `RotateRefreshToken` must atomically consume
the old refresh token and persist both replacement tokens. If an already-used
refresh token is observed, the complete family must be revoked in the same
transaction. It must also reject both `ExpiresAt` and `FamilyExpiresAt` at the
exact expiry boundary and preserve the original family expiry on every
replacement token. Family revocation covers both refresh tokens and every
access token derived from that family.

`RevokeGrant` is also a transaction boundary. It must remove consent and revoke
all unused authorization codes, access tokens, and refresh tokens belonging to
the user/client pair before returning.

Authorization codes and their derived tokens share a `GrantID`. When a used
code is presented again, `RevokeAuthorizationCodeGrant` must revoke every token
derived from that code before the core returns `invalid_grant`.

`RevokeToken` must not reveal whether a token exists. Revoking an access token
affects that token only. Revoking a refresh token revokes its complete family
and all access tokens derived from that family.

Authorization codes, access tokens, refresh tokens, and client secrets are
passed to stores only as SHA-256 digests. Raw bearer credentials must never be
persisted or logged.

`MemoryStore` is the reference implementation for these semantics. It is meant
for tests and local development, not production persistence.

The core never starts cleanup goroutines. The host must schedule `Provider.Purge`
at a bounded interval. MemoryStore rebuilds its credential maps during purge so
expired entries and oversized map buckets become garbage-collectable instead of
causing unbounded process growth.

GooseForum's browser consent adapter persists a validated authorization
request in `oidc_interactions`. Only a SHA-256 digest of its 256-bit random
browser handle is stored. Interactions are user-bound, expire after ten
minutes, and use a conditional transactional consume so concurrent approve and
deny calls cannot both succeed. The host service purge also removes expired
interactions; it adds no background goroutine or unbounded in-memory cache.

All injected implementations (`Store`, `UserResolver`, `SigningKeyProvider`,
`Config.Random`, and `Config.Now`) may be called concurrently and therefore
must be concurrency-safe. The core itself starts no goroutines, timers, files,
or network connections.

After key rotation, call `RSAKeySet.Retire` with a time no earlier than the
latest ID Token expiry for the old key, then call `Prune` from host maintenance.
This keeps the JWKS set bounded without invalidating outstanding tokens.

Every production Store implementation must run `storetest.Run` with a factory
that creates an isolated database and a client seeder. The suite verifies
atomic authorization-code consumption under concurrency, refresh reuse
detection, derived access-token revocation, and complete grant revocation.

`UserResolver` must return `ErrUserUnavailable` when the subject has been
deleted, disabled, frozen, or otherwise may no longer receive tokens. Storage
and network failures must be returned as their original errors. The core uses
that distinction to invalidate grants without mistaking an outage for a user
status change.

## Verification

The package includes unit, concurrency, race, fuzz, benchmark, real RS256/JWKS,
and `golang.org/x/oauth2` interoperability tests. The Gin adapter lives in `app/http/controllers/oidcprovider` and has separate
HTTP and persisted-interaction tests. These local tests do not replace the
OpenID Foundation Provider Conformance Suite. Run Basic OP and Config OP against
a deployed HTTPS instance; official suite results are not yet recorded.
