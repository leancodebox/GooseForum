package oidcprovider

import (
	"context"
	"errors"
	"testing"
)

func TestForumCodeExchangeUsesOneAccountSnapshot(t *testing.T) {
	for _, changed := range []bool{false, true} {
		t.Run(map[bool]string{false: "single_snapshot", true: "credentials_changed"}[changed], func(t *testing.T) {
			p, store := testProvider(t)
			ctx := context.Background()
			client, err := store.GetClient(ctx, "client-1")
			if err != nil {
				t.Fatal(err)
			}
			client.Scopes = append(client.Scopes, ScopeForumRead)
			store.PutClient(client)
			p.cfg.SupportedScopes = append(p.cfg.SupportedScopes, ScopeForumRead)
			verifier := "forum-authorization-verifier-with-43-characters-minimum"
			code, err := p.Authorize(ctx, AuthorizeRequest{ClientID: client.ID, RedirectURI: client.RedirectURIs[0], ResponseType: "code", Scope: []string{"openid", ScopeForumRead}, CodeChallenge: base64URLSHA256(verifier), CodeChallengeMethod: "S256"}, testAuthentication(), true)
			if err != nil {
				t.Fatal(err)
			}
			calls := 0
			p.cfg.Users = userResolverFunc(func(context.Context, string) (*User, error) {
				calls++
				user, _ := testUsers{}.ResolveUser(ctx, "user-1")
				if changed || calls > 1 {
					user.Version = 1
				}
				return user, nil
			})
			tokens, err := p.ExchangeCode(ctx, TokenRequest{GrantType: "authorization_code", ClientID: client.ID, ClientSecret: "secret", AuthMethod: ClientSecretPost, RedirectURI: client.RedirectURIs[0], Code: code.Code, CodeVerifier: verifier})
			if calls != 1 {
				t.Fatalf("exchange resolved account %d times", calls)
			}
			if changed {
				if !errors.Is(err, ErrInvalidGrant) {
					t.Fatalf("changed credentials: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			stored, err := p.ValidateAccessToken(ctx, tokens.AccessToken)
			if err != nil || stored.UserVersion != 0 {
				t.Fatalf("token did not preserve validated version: %#v %v", stored, err)
			}
		})
	}
}
