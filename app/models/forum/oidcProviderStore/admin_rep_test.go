package oidcProviderStore

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"errors"
	core "github.com/leancodebox/GooseForum/app/bundles/oidcprovider"
	"gorm.io/gorm"
)

func TestAdminGrantsPaginationAndRevocation(t *testing.T) {
	s := newTestStore(t)
	ctx := t.Context()
	for _, id := range []string{"managed", "other"} {
		client := &core.Client{ID: id, Name: id, RedirectURIs: []string{"https://example.com/callback"}, Scopes: []string{"openid"}, GrantTypes: []string{"authorization_code"}, TokenEndpointAuthMethod: core.ClientAuthNone, Public: true, RequirePKCE: true, Enabled: true}
		if err := s.CreateClient(ctx, client); err != nil {
			t.Fatal(err)
		}
		for i := 1; i <= 22; i++ {
			if err := s.db.Create(&ConsentEntity{UserID: fmt.Sprintf("%03d", i), ClientID: id, Scopes: []string{"openid"}}).Error; err != nil {
				t.Fatal(err)
			}
		}
		if err := s.db.Create(&TokenEntity{Hash: id, Type: tokenTypeAccess, UserID: "001", ClientID: id, Scopes: []string{"openid"}, ExpiresAt: time.Now().Add(time.Hour)}).Error; err != nil {
			t.Fatal(err)
		}
		if err := s.db.Create(&AuthorizationCodeEntity{Hash: id, UserID: "001", ClientID: id, Scopes: []string{"openid"}, ExpiresAt: time.Now().Add(time.Hour)}).Error; err != nil {
			t.Fatal(err)
		}
	}
	rows, more, err := s.ListClientGrants(ctx, "managed", "", "")
	if err != nil || !more || len(rows) != 20 {
		t.Fatalf("page: %d %v %v", len(rows), more, err)
	}
	updated, err := s.GetClient(ctx, "other")
	if err != nil {
		t.Fatal(err)
	}
	updated.Enabled = false
	updated.Name = "Updated client"
	if err := s.UpdateClient(ctx, updated); err != nil {
		t.Fatal(err)
	}
	actual, err := s.GetClient(ctx, "other")
	if err != nil || actual.Enabled || actual.Name != updated.Name || len(actual.Scopes) != 1 || actual.Scopes[0] != "openid" {
		t.Fatalf("client update did not persist zero values/serialized fields: %v %v", actual, err)
	}
	rows, more, err = s.ListClientGrants(ctx, "managed", rows[19].UserID, "")
	if err != nil || more || len(rows) != 2 || rows[0].UserID != "021" {
		t.Fatalf("next page: %v %v %v", rows, more, err)
	}
	rows, _, err = s.ListClientGrants(ctx, "managed", "", "001")
	if err != nil || len(rows) != 1 {
		t.Fatalf("filter: %v %v", rows, err)
	}
	var plan []struct{ Detail string }
	if err := s.db.Raw("EXPLAIN QUERY PLAN SELECT user_id, scopes, created_at, updated_at FROM oidc_consents WHERE client_id = ? AND user_id > ? ORDER BY user_id LIMIT 21", "managed", "").Scan(&plan).Error; err != nil {
		t.Fatal(err)
	}
	if len(plan) != 1 || !strings.Contains(plan[0].Detail, "idx_oidc_consents_client_user") || strings.Contains(plan[0].Detail, "TEMP") {
		t.Fatalf("plan=%v", plan)
	}
	if err := s.DeleteClient(ctx, "managed", time.Now()); err != nil {
		t.Fatal(err)
	}
	stale := &core.Client{ID: "managed", Name: "stale edit", RedirectURIs: []string{"https://example.com/callback"}, Scopes: []string{"openid"}, GrantTypes: []string{"authorization_code"}, TokenEndpointAuthMethod: core.ClientAuthNone, Public: true, RequirePKCE: true, Enabled: true}
	if err := s.UpdateClient(ctx, stale); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("stale update recreated deleted client: %v", err)
	}
	if err := s.DeleteClient(ctx, "managed", time.Now()); err != nil {
		t.Fatal(err)
	}
	client, err := s.GetClient(ctx, "managed")
	if err != nil || client != nil {
		t.Fatalf("deleted client=%v %v", client, err)
	}
	for _, id := range []string{"managed", "other"} {
		var code AuthorizationCodeEntity
		var token TokenEntity
		if err := s.db.Where("code_hash = ?", id).Take(&code).Error; err != nil {
			t.Fatal(err)
		}
		if err := s.db.Where("token_hash = ?", id).Take(&token).Error; err != nil {
			t.Fatal(err)
		}
		if (id == "managed") != code.Used || (id == "managed") != (token.RevokedAt != nil) {
			t.Fatalf("invalidated unrelated or retained managed credentials: %s", id)
		}
	}
	if err := s.ResetSigningKey(ctx, &SigningKeyEntity{KID: "replacement", EncryptedPrivateKey: []byte("encrypted")}, time.Now()); err != nil {
		t.Fatal(err)
	}
	rows, _, err = s.ListClientGrants(ctx, "other", "", "")
	if err != nil || len(rows) != 0 {
		t.Fatalf("reset retained consents: %v %v", rows, err)
	}
	var token TokenEntity
	if err := s.db.Where("token_hash = ?", "other").Take(&token).Error; err != nil || token.RevokedAt == nil {
		t.Fatalf("reset retained token: %v", err)
	}
}
