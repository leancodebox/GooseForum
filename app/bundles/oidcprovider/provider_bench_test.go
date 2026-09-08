package oidcprovider

import (
	"context"
	"testing"
	"time"
)

func BenchmarkAuthorizeAndExchangeCode(b *testing.B) {
	p, store := testProvider(b)
	ctx := context.Background()
	verifier := "benchmark-verifier-with-at-least-forty-three-characters"
	challenge := base64URLSHA256(verifier)
	req := AuthorizeRequest{ClientID: "client-1", RedirectURI: "https://app.example/callback", ResponseType: "code", Scope: []string{"openid", "profile"}, CodeChallenge: challenge, CodeChallengeMethod: "S256"}
	tokenReq := TokenRequest{GrantType: "authorization_code", ClientID: "client-1", ClientSecret: "secret", AuthMethod: ClientSecretPost, RedirectURI: req.RedirectURI, CodeVerifier: verifier}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result, err := p.Authorize(ctx, req, testAuthentication(), true)
		if err != nil {
			b.Fatal(err)
		}
		tokenReq.Code = result.Code
		if _, err := p.ExchangeCode(ctx, tokenReq); err != nil {
			b.Fatal(err)
		}
		if i%1024 == 1023 {
			if _, err := store.Purge(ctx, PurgeRequest{Now: time.Unix(1_800_000_000, 0), RevokedBefore: time.Unix(1_800_000_000, 0)}); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkRS256SignIDToken(b *testing.B) {
	keys, err := GenerateRSAKeySet("benchmark", 2048)
	if err != nil {
		b.Fatal(err)
	}
	claims := IDTokenClaims{Issuer: "https://forum.example.com/oauth2", Subject: "user-1", Audience: "client-1", IssuedAt: 1_700_000_000, ExpiresAt: 1_700_000_600}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := keys.SignIDToken(context.Background(), claims); err != nil {
			b.Fatal(err)
		}
	}
}
