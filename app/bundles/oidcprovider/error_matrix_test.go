package oidcprovider

import (
	"context"
	"errors"
	"testing"
)

func TestAuthorizationErrorMatrix(t *testing.T) {
	base := AuthorizeRequest{
		ClientID:            "client-1",
		RedirectURI:         "https://app.example/callback",
		ResponseType:        "code",
		ResponseMode:        "query",
		Scope:               []string{"openid"},
		State:               "state-that-must-round-trip",
		CodeChallenge:       base64URLSHA256("authorization-error-matrix-verifier-123456789"),
		CodeChallengeMethod: "S256",
	}

	tests := []struct {
		name            string
		mutate          func(*AuthorizeRequest, *MemoryStore)
		code            string
		cause           error
		redirectAllowed bool
		responseMode    string
	}{
		{
			name: "unknown client is not redirectable",
			mutate: func(req *AuthorizeRequest, _ *MemoryStore) {
				req.ClientID = "unknown"
			},
			code: "unauthorized_client", cause: ErrInvalidClient,
		},
		{
			name: "disabled client is not redirectable",
			mutate: func(_ *AuthorizeRequest, store *MemoryStore) {
				store.PutClient(&Client{ID: "client-1", Enabled: false})
			},
			code: "unauthorized_client", cause: ErrInvalidClient,
		},
		{
			name: "unregistered redirect URI is not redirectable",
			mutate: func(req *AuthorizeRequest, _ *MemoryStore) {
				req.RedirectURI = "https://attacker.example/callback"
			},
			code: "invalid_request", cause: ErrInvalidRequest,
		},
		{
			name: "unsupported response type uses validated redirect",
			mutate: func(req *AuthorizeRequest, _ *MemoryStore) {
				req.ResponseType = "token"
			},
			code: "unsupported_response_type", cause: ErrUnsupportedResponse,
			redirectAllowed: true, responseMode: "query",
		},
		{
			name: "unsupported response mode safely falls back to query",
			mutate: func(req *AuthorizeRequest, _ *MemoryStore) {
				req.ResponseMode = "fragment"
			},
			code: "unsupported_response_mode", cause: ErrUnsupportedResponseMode,
			redirectAllowed: true, responseMode: "query",
		},
		{
			name: "invalid scope uses validated redirect",
			mutate: func(req *AuthorizeRequest, _ *MemoryStore) {
				req.Scope = []string{"profile"}
			},
			code: "invalid_scope", cause: ErrInvalidScope,
			redirectAllowed: true, responseMode: "query",
		},
		{
			name: "invalid PKCE uses validated redirect",
			mutate: func(req *AuthorizeRequest, _ *MemoryStore) {
				req.CodeChallenge = "short"
			},
			code: "invalid_request", cause: ErrInvalidRequest,
			redirectAllowed: true, responseMode: "query",
		},
		{
			name: "login required uses validated redirect",
			mutate: func(req *AuthorizeRequest, _ *MemoryStore) {
				req.Prompt = []string{"none"}
			},
			code: "login_required", cause: ErrLoginRequired,
			redirectAllowed: true, responseMode: "query",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, store := testProvider(t)
			req := base
			req.Scope = append([]string(nil), base.Scope...)
			tt.mutate(&req, store)
			auth := testAuthentication()
			if tt.cause == ErrLoginRequired {
				auth = Authentication{}
			}

			_, err := provider.BeginAuthorization(context.Background(), req, auth)
			assertOAuthError(t, err, tt.code, tt.cause)

			var oauthErr *OAuthError
			if !errors.As(err, &oauthErr) {
				t.Fatalf("error is not OAuthError: %T", err)
			}
			if oauthErr.RedirectAllowed != tt.redirectAllowed {
				t.Fatalf("RedirectAllowed = %v, want %v", oauthErr.RedirectAllowed, tt.redirectAllowed)
			}
			if !tt.redirectAllowed {
				if oauthErr.RedirectURI != "" || oauthErr.State != "" || oauthErr.ResponseMode != "" {
					t.Fatalf("unsafe redirect data leaked: %+v", oauthErr)
				}
				return
			}
			if oauthErr.RedirectURI != base.RedirectURI || oauthErr.State != base.State || oauthErr.ResponseMode != tt.responseMode {
				t.Fatalf("redirect context = (%q, %q, %q), want (%q, %q, %q)", oauthErr.RedirectURI, oauthErr.State, oauthErr.ResponseMode, base.RedirectURI, base.State, tt.responseMode)
			}
		})
	}
}

func TestAuthorizationDenialErrorIsSafelyRedirectable(t *testing.T) {
	provider, _ := testProvider(t)
	req := validErrorMatrixAuthorizeRequest([]string{"openid"})
	req.State = "denial-state"

	_, err := provider.Authorize(context.Background(), req, testAuthentication(), false)
	assertOAuthError(t, err, "access_denied", ErrAccessDenied)

	var oauthErr *OAuthError
	if !errors.As(err, &oauthErr) || !oauthErr.RedirectAllowed {
		t.Fatalf("validated authorization denial must be redirectable: %#v", err)
	}
	if oauthErr.RedirectURI != req.RedirectURI || oauthErr.State != req.State || oauthErr.ResponseMode != "query" {
		t.Fatalf("authorization denial lost redirect context: %+v", oauthErr)
	}
}

func TestAuthorizationCodeTokenErrorMatrix(t *testing.T) {
	tests := []struct {
		name    string
		request func(*testing.T, *Provider) TokenRequest
		code    string
		cause   error
	}{
		{
			name: "unsupported grant type",
			request: func(_ *testing.T, _ *Provider) TokenRequest {
				return TokenRequest{GrantType: "password"}
			},
			code: "unsupported_grant_type", cause: ErrUnsupportedGrant,
		},
		{
			name: "unknown client",
			request: func(_ *testing.T, _ *Provider) TokenRequest {
				return TokenRequest{GrantType: "authorization_code", ClientID: "unknown", AuthMethod: ClientSecretPost}
			},
			code: "invalid_client", cause: ErrInvalidClient,
		},
		{
			name: "wrong client secret",
			request: func(_ *testing.T, _ *Provider) TokenRequest {
				return TokenRequest{GrantType: "authorization_code", ClientID: "client-1", ClientSecret: "wrong", AuthMethod: ClientSecretPost}
			},
			code: "invalid_client", cause: ErrInvalidClient,
		},
		{
			name: "unknown authorization code",
			request: func(_ *testing.T, _ *Provider) TokenRequest {
				return TokenRequest{GrantType: "authorization_code", Code: "unknown", ClientID: "client-1", ClientSecret: "secret", AuthMethod: ClientSecretPost, RedirectURI: "https://app.example/callback"}
			},
			code: "invalid_grant", cause: ErrInvalidGrant,
		},
		{
			name: "PKCE mismatch",
			request: func(t *testing.T, provider *Provider) TokenRequest {
				request := issueErrorMatrixCode(t, provider, []string{"openid"})
				request.CodeVerifier = "wrong-verifier-with-at-least-forty-three-characters"
				return request
			},
			code: "invalid_grant", cause: ErrInvalidGrant,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, _ := testProvider(t)
			_, err := provider.ExchangeCode(context.Background(), tt.request(t, provider))
			assertOAuthError(t, err, tt.code, tt.cause)
			assertTokenErrorHasNoRedirect(t, err)
		})
	}
}

func TestRefreshTokenErrorMatrix(t *testing.T) {
	tests := []struct {
		name    string
		request func(*testing.T, *Provider) TokenRequest
		code    string
		cause   error
	}{
		{
			name: "unsupported grant type",
			request: func(_ *testing.T, _ *Provider) TokenRequest {
				return TokenRequest{GrantType: "authorization_code"}
			},
			code: "unsupported_grant_type", cause: ErrUnsupportedGrant,
		},
		{
			name: "wrong client authentication method",
			request: func(_ *testing.T, _ *Provider) TokenRequest {
				return TokenRequest{GrantType: "refresh_token", ClientID: "client-1", ClientSecret: "secret", AuthMethod: ClientSecretBasic}
			},
			code: "invalid_client", cause: ErrInvalidClient,
		},
		{
			name: "unknown refresh token",
			request: func(_ *testing.T, _ *Provider) TokenRequest {
				return TokenRequest{GrantType: "refresh_token", RefreshToken: "unknown", ClientID: "client-1", ClientSecret: "secret", AuthMethod: ClientSecretPost}
			},
			code: "invalid_grant", cause: ErrInvalidGrant,
		},
		{
			name: "scope escalation",
			request: func(t *testing.T, provider *Provider) TokenRequest {
				tokens := issueErrorMatrixTokens(t, provider)
				return TokenRequest{GrantType: "refresh_token", RefreshToken: tokens.RefreshToken, ClientID: "client-1", ClientSecret: "secret", AuthMethod: ClientSecretPost, Scope: []string{"openid", "profile"}}
			},
			code: "invalid_scope", cause: ErrInvalidScope,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, _ := testProvider(t)
			_, err := provider.Refresh(context.Background(), tt.request(t, provider))
			assertOAuthError(t, err, tt.code, tt.cause)
			assertTokenErrorHasNoRedirect(t, err)
		})
	}
}

func assertOAuthError(t *testing.T, err error, code string, cause error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected OAuth error %q", code)
	}
	var oauthErr *OAuthError
	if !errors.As(err, &oauthErr) {
		t.Fatalf("error %T is not OAuthError: %v", err, err)
	}
	if oauthErr.Code != code {
		t.Fatalf("OAuth error code = %q, want %q", oauthErr.Code, code)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("errors.Is(%v) = false for cause %v", err, cause)
	}
}

func assertTokenErrorHasNoRedirect(t *testing.T, err error) {
	t.Helper()
	var oauthErr *OAuthError
	if !errors.As(err, &oauthErr) {
		t.Fatalf("error %T is not OAuthError", err)
	}
	if oauthErr.RedirectAllowed || oauthErr.RedirectURI != "" || oauthErr.State != "" || oauthErr.ResponseMode != "" {
		t.Fatalf("token endpoint error contains redirect context: %+v", oauthErr)
	}
}

func validErrorMatrixAuthorizeRequest(scopes []string) AuthorizeRequest {
	return AuthorizeRequest{
		ClientID:            "client-1",
		RedirectURI:         "https://app.example/callback",
		ResponseType:        "code",
		Scope:               scopes,
		CodeChallenge:       base64URLSHA256("error-matrix-code-verifier-with-43-characters"),
		CodeChallengeMethod: "S256",
		Prompt:              []string{"consent"},
	}
}

func issueErrorMatrixCode(t *testing.T, provider *Provider, scopes []string) TokenRequest {
	t.Helper()
	verifier := "error-matrix-code-verifier-with-43-characters"
	req := validErrorMatrixAuthorizeRequest(scopes)
	result, err := provider.Authorize(context.Background(), req, testAuthentication(), true)
	if err != nil {
		t.Fatalf("issue authorization code: %v", err)
	}
	return TokenRequest{
		GrantType:    "authorization_code",
		Code:         result.Code,
		ClientID:     "client-1",
		ClientSecret: "secret",
		AuthMethod:   ClientSecretPost,
		RedirectURI:  req.RedirectURI,
		CodeVerifier: verifier,
	}
}

func issueErrorMatrixTokens(t *testing.T, provider *Provider) *TokenResponse {
	t.Helper()
	request := issueErrorMatrixCode(t, provider, []string{"openid", "email", "offline_access"})
	tokens, err := provider.ExchangeCode(context.Background(), request)
	if err != nil {
		t.Fatalf("exchange authorization code: %v", err)
	}
	if tokens.RefreshToken == "" {
		t.Fatal("expected refresh token")
	}
	return tokens
}
