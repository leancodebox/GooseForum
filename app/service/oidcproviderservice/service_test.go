package oidcproviderservice

import (
	"bytes"
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	core "github.com/leancodebox/GooseForum/app/bundles/oidcprovider"
	"github.com/leancodebox/GooseForum/app/models/forum/oidcProviderStore"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestNewPersistsAndReusesEncryptedSigningKey(t *testing.T) {
	db := testDB(t)
	secret := []byte("01234567890123456789012345678901")
	opts := Options{DB: db, SiteURL: "https://forum.example/base/", KeyEncryptionSecret: secret}
	first, err := New(t.Context(), opts)
	if err != nil {
		t.Fatal(err)
	}
	firstKeys, err := first.Provider().PublicJWKS(t.Context())
	if err != nil || len(firstKeys.Keys) != 1 {
		t.Fatalf("first JWKS = %+v, err = %v", firstKeys, err)
	}
	second, err := New(t.Context(), opts)
	if err != nil {
		t.Fatal(err)
	}
	secondKeys, err := second.Provider().PublicJWKS(t.Context())
	if err != nil || len(secondKeys.Keys) != 1 || secondKeys.Keys[0] != firstKeys.Keys[0] {
		t.Fatalf("signing key changed across initialization: first=%+v second=%+v err=%v", firstKeys, secondKeys, err)
	}
	if got := first.Provider().Metadata().Issuer; got != "https://forum.example/base/oauth2" {
		t.Fatalf("issuer = %q", got)
	}
	var stored oidcProviderStore.SigningKeyEntity
	if err := db.First(&stored).Error; err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(stored.EncryptedPrivateKey, []byte("PRIVATE KEY")) {
		t.Fatal("database contains plaintext private key")
	}
}

func TestSigningKeyRotationRetainsOldJWKSUntilIDTokensExpire(t *testing.T) {
	db := testDB(t)
	now := time.Unix(1_700_000_000, 0)
	secret := []byte("01234567890123456789012345678901")
	opts := Options{
		DB: db, SiteURL: "https://forum.example", KeyEncryptionSecret: secret,
		Now: func() time.Time { return now }, IDTokenTTL: 10 * time.Minute,
	}
	service, err := New(t.Context(), opts)
	if err != nil {
		t.Fatal(err)
	}
	before, err := service.Provider().PublicJWKS(t.Context())
	if err != nil || len(before.Keys) != 1 {
		t.Fatalf("initial JWKS = %+v, %v", before, err)
	}
	if err := service.RotateSigningKey(t.Context(), secret); err != nil {
		t.Fatal(err)
	}
	reloaded, err := New(t.Context(), opts)
	if err != nil {
		t.Fatal(err)
	}
	after, err := reloaded.Provider().PublicJWKS(t.Context())
	if err != nil || len(after.Keys) != 2 || after.Keys[0].Kid == after.Keys[1].Kid {
		t.Fatalf("rotated JWKS = %+v, %v", after, err)
	}
	now = now.Add(10*time.Minute - time.Nanosecond)
	if _, err := reloaded.PurgeDefault(t.Context()); err != nil {
		t.Fatal(err)
	}
	beforeExpiry, _ := reloaded.Provider().PublicJWKS(t.Context())
	if len(beforeExpiry.Keys) != 2 {
		t.Fatalf("old key removed before expiry: %+v", beforeExpiry)
	}
	now = now.Add(time.Nanosecond)
	if _, err := reloaded.PurgeDefault(t.Context()); err != nil {
		t.Fatal(err)
	}
	afterExpiry, _ := reloaded.Provider().PublicJWKS(t.Context())
	if len(afterExpiry.Keys) != 1 || afterExpiry.Keys[0].Kid == before.Keys[0].Kid {
		t.Fatalf("old key retained after expiry: %+v", afterExpiry)
	}
}

func TestNewFailsExplicitlyForWrongEncryptionSecret(t *testing.T) {
	db := testDB(t)
	base := Options{DB: db, SiteURL: "https://forum.example", KeyEncryptionSecret: []byte("01234567890123456789012345678901")}
	if _, err := New(t.Context(), base); err != nil {
		t.Fatal(err)
	}
	base.KeyEncryptionSecret = []byte("abcdefghijklmnopqrstuvwxyzABCDEF")
	if _, err := New(t.Context(), base); err == nil || !strings.Contains(err.Error(), "decrypt OIDC signing key") {
		t.Fatalf("wrong secret error = %v", err)
	}
}

func TestNewValidation(t *testing.T) {
	db := testDB(t)
	secret := []byte("01234567890123456789012345678901")
	tests := []struct {
		name string
		opts Options
	}{
		{name: "nil database", opts: Options{SiteURL: "https://forum.example", KeyEncryptionSecret: secret}},
		{name: "empty site URL", opts: Options{DB: db, KeyEncryptionSecret: secret}},
		{name: "short encryption secret", opts: Options{DB: db, SiteURL: "https://forum.example", KeyEncryptionSecret: []byte("short")}},
		{name: "production HTTP issuer", opts: Options{DB: db, SiteURL: "http://forum.example", KeyEncryptionSecret: secret}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := New(t.Context(), tt.opts); err == nil {
				t.Fatal("New succeeded")
			}
		})
	}
}

func TestNewAllowsHTTPOnlyWhenExplicitlyConfigured(t *testing.T) {
	db := testDB(t)
	service, err := New(t.Context(), Options{
		DB: db, SiteURL: "http://localhost:5234", KeyEncryptionSecret: []byte("01234567890123456789012345678901"), AllowInsecureIssuer: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := service.Provider().Metadata().Issuer; got != "http://localhost:5234/oauth2" {
		t.Fatalf("issuer = %q", got)
	}
}

func TestUserResolverMapsClaimsAndAvailability(t *testing.T) {
	db := testDB(t)
	if err := db.Create(&users.EntityComplete{Id: 7, Username: "ada", Nickname: "Ada", Email: "ada@example.com", AvatarUrl: "/avatar.png", IsActivated: users.ActivationSuccess}).Error; err != nil {
		t.Fatal(err)
	}
	resolver := &userResolver{db: db, siteBase: mustURL(t, "https://forum.example/community")}
	user, err := resolver.ResolveUser(t.Context(), "7")
	if err != nil {
		t.Fatal(err)
	}
	if user.Subject != "7" || user.Name != "Ada" || user.PreferredUsername != "ada" || !user.EmailVerified || user.Picture != "https://forum.example/file/img/avatar.png" {
		t.Fatalf("resolved user = %+v", user)
	}
	if err := db.Model(&users.EntityComplete{}).Where("id = ?", 7).Update("is_frozen", users.StatusFrozen).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.ResolveUser(t.Context(), "7"); !errors.Is(err, core.ErrUserUnavailable) {
		t.Fatalf("frozen user error = %v", err)
	}
	if _, err := resolver.ResolveUser(t.Context(), "invalid"); !errors.Is(err, core.ErrUserUnavailable) {
		t.Fatalf("invalid subject error = %v", err)
	}
}

func TestPurgeHasNoBackgroundLifecycle(t *testing.T) {
	db := testDB(t)
	service, err := New(t.Context(), Options{
		DB: db, SiteURL: "https://forum.example", KeyEncryptionSecret: []byte("01234567890123456789012345678901"), Now: func() time.Time { return time.Unix(1_700_000_000, 0) },
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.PurgeDefault(context.Background())
	if err != nil || result != (core.PurgeResult{}) {
		t.Fatalf("PurgeDefault = %+v, %v", result, err)
	}
	if _, err := service.Purge(t.Context(), -time.Second); err == nil {
		t.Fatal("negative retention accepted")
	}
}

func TestRuntimeStateCanRecoverAfterInitializationFailure(t *testing.T) {
	var state runtimeState
	want := &Service{}
	initialErr := errors.New("invalid configuration")
	state.replace(nil, initialErr)
	if service, err := state.get(); service != nil || !errors.Is(err, initialErr) {
		t.Fatalf("failed state = %#v, %v", service, err)
	}
	state.replace(want, nil)
	if service, err := state.get(); service != want || err != nil {
		t.Fatalf("recovered state = %#v, %v", service, err)
	}
	state.replace(nil, ErrUnavailable)
	if service, err := state.get(); service != nil || !errors.Is(err, ErrUnavailable) {
		t.Fatalf("disabled state = %#v, %v", service, err)
	}
}

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(append(oidcProviderStore.Models(), &users.EntityComplete{})...); err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}

func mustURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}
