package oidcprovider

import (
	"context"
	"encoding/json"
	"testing"
)

type claimsRecordingSigner struct {
	testSigner
	claims []IDTokenClaims
}

func (s *claimsRecordingSigner) SignIDToken(ctx context.Context, claims IDTokenClaims) (string, error) {
	s.claims = append(s.claims, claims)
	return s.testSigner.SignIDToken(ctx, claims)
}

func TestCodeAndRefreshKeepProfileClaimsInUserInfo(t *testing.T) {
	p, _ := testProvider(t)
	signer := &claimsRecordingSigner{}
	p.cfg.Signer = signer
	ctx := context.Background()
	verifier := "claims-release-verifier-with-at-least-forty-three-characters"
	code, err := p.Authorize(ctx, AuthorizeRequest{
		ClientID: "client-1", RedirectURI: "https://app.example/callback", ResponseType: "code",
		Scope: []string{"openid", "profile", "email", "offline_access"}, Nonce: "nonce",
		CodeChallenge: base64URLSHA256(verifier), CodeChallengeMethod: "S256", Prompt: []string{"consent"},
	}, testAuthentication(), true)
	if err != nil {
		t.Fatal(err)
	}
	tokens, err := p.ExchangeCode(ctx, TokenRequest{GrantType: "authorization_code", Code: code.Code,
		ClientID: "client-1", ClientSecret: "secret", AuthMethod: ClientSecretPost,
		RedirectURI: "https://app.example/callback", CodeVerifier: verifier})
	if err != nil {
		t.Fatal(err)
	}
	refreshed, err := p.Refresh(ctx, TokenRequest{GrantType: "refresh_token", RefreshToken: tokens.RefreshToken,
		ClientID: "client-1", ClientSecret: "secret", AuthMethod: ClientSecretPost})
	if err != nil {
		t.Fatal(err)
	}
	if len(signer.claims) != 2 {
		t.Fatalf("expected two ID tokens, got %d", len(signer.claims))
	}
	for i, claims := range signer.claims {
		raw, err := json.Marshal(claims)
		if err != nil {
			t.Fatal(err)
		}
		var fields map[string]any
		if err := json.Unmarshal(raw, &fields); err != nil {
			t.Fatal(err)
		}
		for _, key := range []string{"name", "preferred_username", "picture", "email", "email_verified"} {
			if _, present := fields[key]; present {
				t.Errorf("ID token %d exposes %s", i, key)
			}
		}
		if claims.Subject != "u-1" || claims.Audience != "client-1" || claims.AuthTime != testAuthentication().AuthTime.Unix() {
			t.Errorf("ID token %d lost authentication claims: %+v", i, claims)
		}
	}
	if signer.claims[0].Nonce != "nonce" {
		t.Error("initial nonce was lost")
	}
	for _, token := range []string{tokens.AccessToken, refreshed.AccessToken} {
		claims, err := p.UserInfo(ctx, token)
		if err != nil {
			t.Fatal(err)
		}
		if claims.Name != "Alice" || claims.PreferredUsername != "alice" || claims.Email != "alice@example.com" || claims.EmailVerified == nil || !*claims.EmailVerified {
			t.Fatalf("UserInfo lost authorized claims: %+v", claims)
		}
	}
}
