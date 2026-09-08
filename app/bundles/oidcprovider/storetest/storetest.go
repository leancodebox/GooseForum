// Package storetest provides a reusable contract suite for oidcprovider Store
// implementations.
package storetest

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"sync"
	"testing"
	"time"

	provider "github.com/leancodebox/GooseForum/app/bundles/oidcprovider"
)

type Factory func(*testing.T) (provider.Store, func(*provider.Client))

func Run(t *testing.T, factory Factory) {
	t.Helper()
	t.Run("client and consent round trip", func(t *testing.T) {
		store, seed := factory(t)
		client := &provider.Client{ID: "client", RedirectURIs: []string{"https://client.example/callback"}, Scopes: []string{"openid"}, GrantTypes: []string{"authorization_code"}, TokenEndpointAuthMethod: provider.ClientAuthNone, Public: true, Enabled: true}
		seed(client)
		loaded, err := store.GetClient(context.Background(), client.ID)
		if err != nil || loaded == nil || loaded.ID != client.ID {
			t.Fatalf("client round trip: %+v, %v", loaded, err)
		}
		consent := &provider.Consent{UserID: "user", ClientID: client.ID, Scopes: []string{"openid"}}
		if err := store.SaveConsent(context.Background(), consent); err != nil {
			t.Fatal(err)
		}
		loadedConsent, err := store.GetConsent(context.Background(), "user", client.ID)
		if err != nil || loadedConsent == nil || len(loadedConsent.Scopes) != 1 {
			t.Fatalf("consent round trip: %+v, %v", loadedConsent, err)
		}
	})

	t.Run("authorization code has one atomic winner", func(t *testing.T) {
		store, _ := factory(t)
		now := time.Unix(1_700_000_000, 0)
		verifier := "store-contract-verifier-with-at-least-43-characters"
		code := &provider.AuthorizationCode{Hash: "code", GrantID: "grant", UserID: "user", ClientID: "client", RedirectURI: "https://client.example/callback", Scopes: []string{"openid"}, CodeChallenge: challenge(verifier), CodeChallengeMethod: "S256", ExpiresAt: now.Add(time.Minute)}
		if err := store.SaveAuthorizationCode(context.Background(), code); err != nil {
			t.Fatal(err)
		}
		invalidAccess := &provider.Token{Hash: "invalid-access", UserID: "user", ClientID: "client", ExpiresAt: now.Add(time.Hour)}
		err := store.ExchangeAuthorizationCode(context.Background(), provider.AuthorizationCodeExchange{CodeHash: code.Hash, ClientID: code.ClientID, RedirectURI: code.RedirectURI, CodeVerifier: "wrong-verifier-with-at-least-forty-three-characters", Now: now, AccessToken: invalidAccess})
		if !errors.Is(err, provider.ErrInvalidGrant) {
			t.Fatalf("invalid PKCE accepted: %v", err)
		}
		if token, _ := store.GetAccessToken(context.Background(), invalidAccess.Hash); token != nil {
			t.Fatal("failed PKCE persisted an access token")
		}
		const workers = 12
		var wg sync.WaitGroup
		results := make(chan error, workers)
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				access := &provider.Token{Hash: "access-" + string(rune('a'+i)), GrantID: "grant", UserID: "user", ClientID: "client", Scopes: []string{"openid"}, ExpiresAt: now.Add(time.Hour)}
				refresh := &provider.Token{Hash: "refresh-" + string(rune('a'+i)), GrantID: "grant", FamilyID: "family", UserID: "user", ClientID: "client", Scopes: []string{"openid"}, ExpiresAt: now.Add(time.Hour), FamilyExpiresAt: now.Add(24 * time.Hour)}
				results <- store.ExchangeAuthorizationCode(context.Background(), provider.AuthorizationCodeExchange{CodeHash: code.Hash, ClientID: code.ClientID, RedirectURI: code.RedirectURI, CodeVerifier: verifier, Now: now, AccessToken: access, RefreshToken: refresh})
			}(i)
		}
		wg.Wait()
		close(results)
		winners := 0
		for err := range results {
			if err == nil {
				winners++
			} else if !errors.Is(err, provider.ErrInvalidGrant) {
				t.Fatalf("unexpected loser error: %v", err)
			}
		}
		if winners != 1 {
			t.Fatalf("atomic exchange had %d winners", winners)
		}
		if err := store.RevokeAuthorizationCodeGrant(context.Background(), code.Hash, now); err != nil {
			t.Fatal(err)
		}
		revoked := 0
		for i := 0; i < workers; i++ {
			token, _ := store.GetAccessToken(context.Background(), "access-"+string(rune('a'+i)))
			if token != nil && token.RevokedAt != nil {
				revoked++
			}
		}
		if revoked != 1 {
			t.Fatalf("code replay grant revocation affected %d winning tokens", revoked)
		}
	})

	t.Run("authorization code expires exclusively without writes", func(t *testing.T) {
		store, _ := factory(t)
		now := time.Unix(1_700_000_000, 0)
		verifier := "store-expiry-verifier-with-at-least-43-characters"
		code := &provider.AuthorizationCode{Hash: "expired-code", UserID: "user", ClientID: "client", RedirectURI: "https://client.example/callback", CodeChallenge: challenge(verifier), ExpiresAt: now}
		if err := store.SaveAuthorizationCode(context.Background(), code); err != nil {
			t.Fatal(err)
		}
		access := &provider.Token{Hash: "must-not-exist", UserID: "user", ClientID: "client", ExpiresAt: now.Add(time.Hour)}
		err := store.ExchangeAuthorizationCode(context.Background(), provider.AuthorizationCodeExchange{CodeHash: code.Hash, ClientID: code.ClientID, RedirectURI: code.RedirectURI, CodeVerifier: verifier, Now: now, AccessToken: access})
		if !errors.Is(err, provider.ErrInvalidGrant) {
			t.Fatalf("code valid at expiry boundary: %v", err)
		}
		storedCode, _ := store.GetAuthorizationCode(context.Background(), code.Hash)
		storedAccess, _ := store.GetAccessToken(context.Background(), access.Hash)
		if storedCode == nil || storedCode.Used || storedAccess != nil {
			t.Fatalf("failed expiry exchange mutated state: code=%+v access=%+v", storedCode, storedAccess)
		}
	})

	t.Run("refresh reuse revokes derived access", func(t *testing.T) {
		store, _ := factory(t)
		now := time.Unix(1_700_000_000, 0)
		verifier := "store-refresh-verifier-with-at-least-43-characters"
		code := &provider.AuthorizationCode{Hash: "code", UserID: "user", ClientID: "client", RedirectURI: "https://client.example/callback", Scopes: []string{"openid"}, CodeChallenge: challenge(verifier), ExpiresAt: now.Add(time.Minute)}
		if err := store.SaveAuthorizationCode(context.Background(), code); err != nil {
			t.Fatal(err)
		}
		old := &provider.Token{Hash: "old", FamilyID: "family", UserID: "user", ClientID: "client", Scopes: []string{"openid"}, ExpiresAt: now.Add(time.Hour), FamilyExpiresAt: now.Add(24 * time.Hour)}
		if err := store.ExchangeAuthorizationCode(context.Background(), provider.AuthorizationCodeExchange{CodeHash: code.Hash, ClientID: code.ClientID, RedirectURI: code.RedirectURI, CodeVerifier: verifier, Now: now, AccessToken: &provider.Token{Hash: "initial-access", FamilyID: "family", UserID: "user", ClientID: "client", ExpiresAt: now.Add(time.Hour)}, RefreshToken: old}); err != nil {
			t.Fatal(err)
		}
		rotation := provider.RefreshTokenRotation{OldHash: old.Hash, ClientID: "client", Now: now, AccessToken: &provider.Token{Hash: "new-access", FamilyID: "family", UserID: "user", ClientID: "client", ExpiresAt: now.Add(time.Hour)}, RefreshToken: &provider.Token{Hash: "new-refresh", FamilyID: "family", UserID: "user", ClientID: "client", ExpiresAt: now.Add(time.Hour), FamilyExpiresAt: now.Add(24 * time.Hour)}}
		if err := store.RotateRefreshToken(context.Background(), rotation); err != nil {
			t.Fatal(err)
		}
		if err := store.RotateRefreshToken(context.Background(), rotation); !errors.Is(err, provider.ErrInvalidGrant) {
			t.Fatalf("refresh reuse accepted: %v", err)
		}
		access, _ := store.GetAccessToken(context.Background(), "new-access")
		if access == nil || access.RevokedAt == nil {
			t.Fatal("refresh reuse did not revoke derived access")
		}
	})

	t.Run("refresh family expiry is exclusive without writes", func(t *testing.T) {
		store, _ := factory(t)
		now := time.Unix(1_700_000_000, 0)
		verifier := "store-family-expiry-verifier-with-at-least-43-characters"
		code := &provider.AuthorizationCode{Hash: "family-code", UserID: "user", ClientID: "client", RedirectURI: "https://client.example/callback", CodeChallenge: challenge(verifier), ExpiresAt: now.Add(time.Minute)}
		if err := store.SaveAuthorizationCode(context.Background(), code); err != nil {
			t.Fatal(err)
		}
		old := &provider.Token{Hash: "family-old", FamilyID: "family", UserID: "user", ClientID: "client", ExpiresAt: now.Add(time.Hour), FamilyExpiresAt: now}
		if err := store.ExchangeAuthorizationCode(context.Background(), provider.AuthorizationCodeExchange{CodeHash: code.Hash, ClientID: code.ClientID, RedirectURI: code.RedirectURI, CodeVerifier: verifier, Now: now.Add(-time.Second), AccessToken: &provider.Token{Hash: "family-initial-access", FamilyID: "family", UserID: "user", ClientID: "client", ExpiresAt: now.Add(time.Hour)}, RefreshToken: old}); err != nil {
			t.Fatal(err)
		}
		rotation := provider.RefreshTokenRotation{OldHash: old.Hash, ClientID: "client", Now: now, AccessToken: &provider.Token{Hash: "family-forbidden-access", FamilyID: "family", UserID: "user", ClientID: "client", ExpiresAt: now.Add(time.Hour)}, RefreshToken: &provider.Token{Hash: "family-forbidden-refresh", FamilyID: "family", UserID: "user", ClientID: "client", ExpiresAt: now.Add(time.Hour), FamilyExpiresAt: now}}
		if err := store.RotateRefreshToken(context.Background(), rotation); !errors.Is(err, provider.ErrInvalidGrant) {
			t.Fatalf("family valid at expiry boundary: %v", err)
		}
		access, _ := store.GetAccessToken(context.Background(), rotation.AccessToken.Hash)
		refresh, _ := store.GetRefreshToken(context.Background(), rotation.RefreshToken.Hash)
		if access != nil || refresh != nil {
			t.Fatalf("failed family rotation persisted tokens: access=%+v refresh=%+v", access, refresh)
		}
	})

	t.Run("grant revocation covers consent codes and tokens", func(t *testing.T) {
		store, _ := factory(t)
		now := time.Unix(1_700_000_000, 0)
		verifier := "store-revoke-verifier-with-at-least-43-characters"
		if err := store.SaveConsent(context.Background(), &provider.Consent{UserID: "user", ClientID: "client", Scopes: []string{"openid"}}); err != nil {
			t.Fatal(err)
		}
		usedCode := &provider.AuthorizationCode{Hash: "used-code", UserID: "user", ClientID: "client", RedirectURI: "https://client.example/callback", Scopes: []string{"openid"}, CodeChallenge: challenge(verifier), ExpiresAt: now.Add(time.Minute)}
		unusedCode := &provider.AuthorizationCode{Hash: "unused-code", UserID: "user", ClientID: "client", RedirectURI: usedCode.RedirectURI, Scopes: []string{"openid"}, CodeChallenge: challenge(verifier), ExpiresAt: now.Add(time.Minute)}
		if err := store.SaveAuthorizationCode(context.Background(), usedCode); err != nil {
			t.Fatal(err)
		}
		if err := store.SaveAuthorizationCode(context.Background(), unusedCode); err != nil {
			t.Fatal(err)
		}
		access := &provider.Token{Hash: "access", FamilyID: "family", UserID: "user", ClientID: "client", ExpiresAt: now.Add(time.Hour)}
		refresh := &provider.Token{Hash: "refresh", FamilyID: "family", UserID: "user", ClientID: "client", ExpiresAt: now.Add(time.Hour), FamilyExpiresAt: now.Add(24 * time.Hour)}
		if err := store.ExchangeAuthorizationCode(context.Background(), provider.AuthorizationCodeExchange{CodeHash: usedCode.Hash, ClientID: "client", RedirectURI: usedCode.RedirectURI, CodeVerifier: verifier, Now: now, AccessToken: access, RefreshToken: refresh}); err != nil {
			t.Fatal(err)
		}
		if err := store.RevokeGrant(context.Background(), "user", "client", now); err != nil {
			t.Fatal(err)
		}
		consent, _ := store.GetConsent(context.Background(), "user", "client")
		code, _ := store.GetAuthorizationCode(context.Background(), unusedCode.Hash)
		storedAccess, _ := store.GetAccessToken(context.Background(), access.Hash)
		storedRefresh, _ := store.GetRefreshToken(context.Background(), refresh.Hash)
		if consent != nil || code == nil || !code.Used || storedAccess == nil || storedAccess.RevokedAt == nil || storedRefresh == nil || storedRefresh.RevokedAt == nil {
			t.Fatalf("incomplete grant revocation: consent=%+v code=%+v access=%+v refresh=%+v", consent, code, storedAccess, storedRefresh)
		}
	})
}

func challenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
