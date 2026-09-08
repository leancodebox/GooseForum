package oidcProviderStore

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	core "github.com/leancodebox/GooseForum/app/bundles/oidcprovider"
	"github.com/leancodebox/GooseForum/app/bundles/oidcprovider/storetest"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestInteractionIsUserBoundSingleUseAndExpires(t *testing.T) {
	store := newTestStore(t)
	now := time.Unix(1_700_000_000, 0)
	interaction := &InteractionEntity{
		Hash: string(make([]byte, 64)), RequestJSON: []byte(`{"ClientID":"client"}`),
		Purpose: InteractionPurposeConsent, UserID: "user-1", AuthTime: now.Add(-time.Minute),
		StartedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Minute),
	}
	if err := store.CreateInteraction(t.Context(), interaction); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetInteraction(t.Context(), interaction.Hash, "user-2", now); !errors.Is(err, ErrInteractionInvalid) {
		t.Fatalf("cross-user read error = %v", err)
	}

	const workers = 12
	var wg sync.WaitGroup
	results := make(chan error, workers)
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := store.ConsumeInteraction(context.Background(), interaction.Hash, "user-1", now)
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		} else if !errors.Is(err, ErrInteractionInvalid) {
			t.Fatalf("unexpected consume error: %v", err)
		}
	}
	if successes != 1 {
		t.Fatalf("successful consumers = %d, want 1", successes)
	}

	expired := *interaction
	expired.Hash = string(make([]byte, 63)) + "x"
	expired.ExpiresAt = now
	if err := store.CreateInteraction(t.Context(), &expired); err != nil {
		t.Fatal(err)
	}
	if _, err := store.ConsumeInteraction(t.Context(), expired.Hash, expired.UserID, now); !errors.Is(err, ErrInteractionInvalid) {
		t.Fatalf("expiry-boundary consume error = %v", err)
	}
	deleted, err := store.PurgeInteractions(t.Context(), now)
	if err != nil || deleted != 1 {
		t.Fatalf("PurgeInteractions deleted=%d err=%v", deleted, err)
	}
}

func TestLoginInteractionRequiresNewAuthenticationAndIsSingleUse(t *testing.T) {
	store := newTestStore(t)
	now := time.Unix(1_700_000_000, 0)
	interaction := &InteractionEntity{
		Hash: string(make([]byte, 63)) + "r", RequestJSON: []byte(`{"ClientID":"client"}`),
		Purpose: InteractionPurposeLogin, PreviousAuthID: "old-auth", StartedAt: now,
		ExpiresAt: now.Add(time.Minute),
	}
	if err := store.CreateInteraction(t.Context(), interaction); err != nil {
		t.Fatal(err)
	}
	for name, attempt := range map[string]struct {
		authID   string
		authTime time.Time
	}{
		"same authentication": {"old-auth", now.Add(time.Second)},
		"too early":           {"new-auth", now.Add(-time.Second)},
		"missing auth id":     {"", now.Add(time.Second)},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := store.ConsumeLoginInteraction(t.Context(), interaction.Hash, attempt.authID, attempt.authTime, now); !errors.Is(err, ErrInteractionInvalid) {
				t.Fatalf("error = %v", err)
			}
		})
	}
	consumed, err := store.ConsumeLoginInteraction(t.Context(), interaction.Hash, "new-auth", now, now)
	if err != nil || consumed.Purpose != InteractionPurposeLogin {
		t.Fatalf("consume result=%#v error=%v", consumed, err)
	}
	if _, err := store.ConsumeLoginInteraction(t.Context(), interaction.Hash, "another-auth", now.Add(time.Second), now); !errors.Is(err, ErrInteractionInvalid) {
		t.Fatalf("replay error = %v", err)
	}

	preLogin := *interaction
	preLogin.Hash = string(make([]byte, 63)) + "p"
	preLogin.PreviousAuthID = ""
	if err := store.CreateInteraction(t.Context(), &preLogin); err != nil {
		t.Fatalf("create pre-login interaction: %v", err)
	}
	if _, err := store.ConsumeLoginInteraction(t.Context(), preLogin.Hash, "first-auth", now, now); err != nil {
		t.Fatalf("consume pre-login interaction: %v", err)
	}
}

func TestStoreContract(t *testing.T) {
	storetest.Run(t, func(t *testing.T) (core.Store, func(*core.Client)) {
		t.Helper()
		dsn := filepath.Join(t.TempDir(), "oidc-store.sqlite") + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
		if err != nil {
			t.Fatal(err)
		}
		sqlDB, err := db.DB()
		if err != nil {
			t.Fatal(err)
		}
		// A single SQLite writer makes transaction scheduling deterministic;
		// MySQL deployments retain their configured connection pool.
		sqlDB.SetMaxOpenConns(1)
		t.Cleanup(func() { _ = sqlDB.Close() })

		store, err := New(db)
		if err != nil {
			t.Fatal(err)
		}
		if err := store.Migrate(t.Context()); err != nil {
			t.Fatal(err)
		}
		return store, func(client *core.Client) {
			if err := store.SaveClient(t.Context(), client); err != nil {
				t.Fatal(err)
			}
		}
	})
}

func TestNewRejectsNilDatabase(t *testing.T) {
	if _, err := New(nil); err == nil {
		t.Fatal("New(nil) succeeded")
	}
}

func TestCreateClientNeverOverwritesExistingRegistration(t *testing.T) {
	store := newTestStore(t)
	client := &core.Client{
		ID: "client", Name: "Original", SecretHash: core.HashClientSecret("secret"),
		RedirectURIs: []string{"https://client.example/callback"}, Scopes: []string{"openid"},
		GrantTypes: []string{"authorization_code"}, TokenEndpointAuthMethod: core.ClientSecretBasic, Enabled: true,
	}
	if err := store.CreateClient(t.Context(), client); err != nil {
		t.Fatal(err)
	}
	replacement := *client
	replacement.Name = "Replacement"
	if err := store.CreateClient(t.Context(), &replacement); err == nil {
		t.Fatal("duplicate CreateClient unexpectedly succeeded")
	}
	stored, err := store.GetClient(t.Context(), client.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored == nil || stored.Name != "Original" {
		t.Fatalf("existing client was overwritten: %+v", stored)
	}
}

func TestListUserGrantsIsUserBound(t *testing.T) {
	store := newTestStore(t)
	client := &core.Client{
		ID: "client", Name: "Wiki", SecretHash: core.HashClientSecret("secret"),
		RedirectURIs: []string{"https://client.example/callback"}, Scopes: []string{"openid", "profile"},
		GrantTypes: []string{"authorization_code"}, TokenEndpointAuthMethod: core.ClientSecretBasic, Enabled: true,
	}
	if err := store.CreateClient(t.Context(), client); err != nil {
		t.Fatal(err)
	}
	for _, userID := range []string{"user-1", "user-2"} {
		if err := store.SaveConsent(t.Context(), &core.Consent{UserID: userID, ClientID: client.ID, Scopes: []string{"openid"}}); err != nil {
			t.Fatal(err)
		}
	}
	grants, err := store.ListUserGrants(t.Context(), "user-1")
	if err != nil || len(grants) != 1 || grants[0].ClientID != client.ID || grants[0].Name != "Wiki" || len(grants[0].Scopes) != 1 {
		t.Fatalf("grants = %+v, %v", grants, err)
	}
}

func TestExchangeRollsBackCodeConsumptionWhenTokenInsertFails(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	now := time.Unix(1_700_000_000, 0)
	verifier := "transaction-rollback-verifier-with-at-least-43-characters"
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	redirectURI := "https://client.example/callback"

	firstCode := &core.AuthorizationCode{Hash: "first-code", UserID: "user", ClientID: "client", RedirectURI: redirectURI, CodeChallenge: challenge, ExpiresAt: now.Add(time.Minute)}
	if err := store.SaveAuthorizationCode(ctx, firstCode); err != nil {
		t.Fatal(err)
	}
	collidingToken := &core.Token{Hash: "same-token", UserID: "user", ClientID: "client", ExpiresAt: now.Add(time.Hour)}
	if err := store.ExchangeAuthorizationCode(ctx, core.AuthorizationCodeExchange{CodeHash: firstCode.Hash, ClientID: "client", RedirectURI: redirectURI, CodeVerifier: verifier, Now: now, AccessToken: collidingToken}); err != nil {
		t.Fatal(err)
	}

	secondCode := &core.AuthorizationCode{Hash: "second-code", UserID: "user", ClientID: "client", RedirectURI: redirectURI, CodeChallenge: challenge, ExpiresAt: now.Add(time.Minute)}
	if err := store.SaveAuthorizationCode(ctx, secondCode); err != nil {
		t.Fatal(err)
	}
	if err := store.ExchangeAuthorizationCode(ctx, core.AuthorizationCodeExchange{CodeHash: secondCode.Hash, ClientID: "client", RedirectURI: redirectURI, CodeVerifier: verifier, Now: now, AccessToken: collidingToken}); err == nil {
		t.Fatal("duplicate token insert unexpectedly succeeded")
	}
	storedCode, err := store.GetAuthorizationCode(ctx, secondCode.Hash)
	if err != nil {
		t.Fatal(err)
	}
	if storedCode == nil || storedCode.Used {
		t.Fatalf("failed exchange consumed authorization code: %+v", storedCode)
	}
}

func TestSchemaHasOperationalIndexes(t *testing.T) {
	store := newTestStore(t)
	migrator := store.db.Migrator()
	checks := []struct {
		model any
		index string
	}{
		{&AuthorizationCodeEntity{}, "idx_oidc_codes_expiry"},
		{&AuthorizationCodeEntity{}, "idx_oidc_codes_grant_owner"},
		{&TokenEntity{}, "idx_oidc_tokens_family"},
		{&TokenEntity{}, "idx_oidc_tokens_owner"},
		{&TokenEntity{}, "idx_oidc_tokens_type_expiry"},
		{&TokenEntity{}, "idx_oidc_tokens_revoked"},
	}
	for _, check := range checks {
		if !migrator.HasIndex(check.model, check.index) {
			t.Errorf("missing index %s", check.index)
		}
	}
}

func TestPurgeRetainsReplayTombstonesUntilCredentialsAreTerminal(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	now := time.Unix(1_700_000_000, 0)
	verifier := "purge-tombstone-verifier-with-at-least-43-characters"
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	code := &core.AuthorizationCode{Hash: "used-code", GrantID: "grant", UserID: "user", ClientID: "client", RedirectURI: "https://client.example/callback", CodeChallenge: challenge, ExpiresAt: now.Add(time.Minute)}
	if err := store.SaveAuthorizationCode(ctx, code); err != nil {
		t.Fatal(err)
	}
	access := &core.Token{Hash: "access", GrantID: "grant", FamilyID: "family", UserID: "user", ClientID: "client", ExpiresAt: now.Add(time.Hour)}
	refresh := &core.Token{Hash: "refresh", GrantID: "grant", FamilyID: "family", UserID: "user", ClientID: "client", ExpiresAt: now.Add(time.Minute), FamilyExpiresAt: now.Add(2 * time.Hour)}
	if err := store.ExchangeAuthorizationCode(ctx, core.AuthorizationCodeExchange{CodeHash: code.Hash, ClientID: code.ClientID, RedirectURI: code.RedirectURI, CodeVerifier: verifier, Now: now, AccessToken: access, RefreshToken: refresh}); err != nil {
		t.Fatal(err)
	}
	if err := store.RevokeToken(ctx, refresh.Hash, code.ClientID, now); err != nil {
		t.Fatal(err)
	}

	result, err := store.Purge(ctx, core.PurgeRequest{Now: now.Add(90 * time.Minute), RevokedBefore: now.Add(90 * time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	if result.AuthorizationCodes != 0 || result.RefreshTokens != 0 {
		t.Fatalf("replay tombstones purged before family expiry: %+v", result)
	}
	storedCode, _ := store.GetAuthorizationCode(ctx, code.Hash)
	storedRefresh, _ := store.GetRefreshToken(ctx, refresh.Hash)
	if storedCode == nil || storedRefresh == nil {
		t.Fatalf("missing replay tombstone: code=%+v refresh=%+v", storedCode, storedRefresh)
	}

	result, err = store.Purge(ctx, core.PurgeRequest{Now: now.Add(2 * time.Hour), RevokedBefore: now.Add(2 * time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	if result.AuthorizationCodes != 1 || result.RefreshTokens != 1 {
		t.Fatalf("terminal replay tombstones retained: %+v", result)
	}
}

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dsn := filepath.Join(t.TempDir(), "oidc-store.sqlite") + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	store, err := New(db)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Migrate(t.Context()); err != nil {
		t.Fatal(err)
	}
	return store
}
