package oidcprovider

import (
	"testing"
	"time"
)

func TestIDTokenHintValidation(t *testing.T) {
	p, _ := testProvider(t)
	keys, err := GenerateRSAKeySet("hint-key", 2048)
	if err != nil {
		t.Fatal(err)
	}
	p.cfg.Signer = keys
	now := p.cfg.Now()
	req := AuthorizeRequest{ClientID: "client-1", RedirectURI: "https://app.example/callback", ResponseType: "code", Scope: []string{"openid"}, CodeChallenge: "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM", CodeChallengeMethod: "S256"}
	auth := Authentication{UserID: "u-1", AuthTime: now}
	user, err := p.resolveUser(t.Context(), auth.UserID)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, issuer, subject, audience, want string }{
		{"matching", p.cfg.Issuer, user.Subject, req.ClientID, ""},
		{"different user", p.cfg.Issuer, "other-user", req.ClientID, "login_required"},
		{"different issuer", "https://other.example", user.Subject, req.ClientID, "invalid_request"},
		{"different client", p.cfg.Issuer, user.Subject, "other-client", "invalid_request"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// An expired, correctly signed token remains useful as a past-session hint.
			raw, err := keys.SignIDToken(t.Context(), IDTokenClaims{Issuer: tc.issuer, Subject: tc.subject, Audience: tc.audience, IssuedAt: now.Add(-time.Hour).Unix(), ExpiresAt: now.Add(-time.Minute).Unix()})
			if err != nil {
				t.Fatal(err)
			}
			req.IDTokenHint = raw
			for _, complete := range []bool{false, true} {
				var err error
				if complete {
					_, err = p.Authorize(t.Context(), req, auth, true)
				} else {
					_, err = p.BeginAuthorization(t.Context(), req, auth)
				}
				if tc.want == "" {
					if err != nil {
						t.Fatal(err)
					}
				} else if e, ok := err.(*OAuthError); !ok || e.Code != tc.want {
					t.Fatalf("got %v, want %s", err, tc.want)
				}
			}
		})
	}
	otherKeys, err := GenerateRSAKeySet("hint-key", 2048)
	if err != nil {
		t.Fatal(err)
	}
	req.IDTokenHint, err = otherKeys.SignIDToken(t.Context(), IDTokenClaims{Issuer: p.cfg.Issuer, Subject: user.Subject, Audience: req.ClientID, IssuedAt: now.Unix(), ExpiresAt: now.Add(time.Minute).Unix()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.BeginAuthorization(t.Context(), req, auth); err == nil {
		t.Fatal("forged signature accepted")
	}
	req.IDTokenHint = "eyJhbGciOiJub25lIn0.e30."
	if _, err := p.BeginAuthorization(t.Context(), req, auth); err == nil {
		t.Fatal("unsigned hint accepted")
	}
}

func TestMaxAgeZeroRequiresFreshAuthenticationAtSameInstant(t *testing.T) {
	p, _ := testProvider(t)
	zero := time.Duration(0)
	req := AuthorizeRequest{ClientID: "client-1", RedirectURI: "https://app.example/callback", ResponseType: "code", Scope: []string{"openid"}, CodeChallenge: "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM", CodeChallengeMethod: "S256", MaxAge: &zero}
	decision, err := p.BeginAuthorization(t.Context(), req, Authentication{UserID: "u-1", AuthTime: p.cfg.Now()})
	if err != nil || decision.Action != AuthorizationNeedLogin {
		t.Fatalf("decision=%+v err=%v", decision, err)
	}
}
