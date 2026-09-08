package oidcprovider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

// TestOAuth2ClientInterop keeps HTTP outside the library while proving that a
// conventional OAuth2 client can consume results produced by the core.
func TestOAuth2ClientInterop(t *testing.T) {
	store := NewMemoryStore()
	secret := "interop-client-secret"
	store.PutClient(&Client{ID: "interop", SecretHash: HashClientSecret(secret), RedirectURIs: []string{"https://client.example/callback"}, Scopes: []string{"openid", "profile"}, GrantTypes: []string{"authorization_code"}, TokenEndpointAuthMethod: ClientSecretPost, Enabled: true})
	store.PutClient(&Client{ID: "interop-basic", SecretHash: HashClientSecret(secret), RedirectURIs: []string{"https://client.example/callback"}, Scopes: []string{"openid"}, GrantTypes: []string{"authorization_code"}, TokenEndpointAuthMethod: ClientSecretBasic, Enabled: true})
	signer, err := GenerateRSAKeySet("interop-key", 2048)
	if err != nil {
		t.Fatal(err)
	}
	var provider *Provider
	mux := http.NewServeMux()
	server := httptest.NewTLSServer(mux)
	defer server.Close()
	now := time.Unix(1_700_000_100, 0)
	provider, err = New(Config{Issuer: server.URL, Store: store, Users: testUsers{}, Signer: signer, Now: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}

	mux.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		result, err := provider.Authorize(r.Context(), AuthorizeRequest{ClientID: q.Get("client_id"), RedirectURI: q.Get("redirect_uri"), ResponseType: q.Get("response_type"), Scope: strings.Fields(q.Get("scope")), State: q.Get("state"), CodeChallenge: q.Get("code_challenge"), CodeChallengeMethod: q.Get("code_challenge_method")}, Authentication{UserID: "user-1", AuthTime: now}, true)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		location, _ := url.Parse(result.RedirectURI)
		values := location.Query()
		values.Set("code", result.Code)
		values.Set("state", result.State)
		values.Set("iss", result.Issuer)
		location.RawQuery = values.Encode()
		http.Redirect(w, r, location.String(), http.StatusFound)
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		clientID, clientSecret, method := r.Form.Get("client_id"), r.Form.Get("client_secret"), ClientSecretPost
		if basicID, basicSecret, ok := r.BasicAuth(); ok {
			clientID, clientSecret, method = basicID, basicSecret, ClientSecretBasic
		}
		result, err := provider.ExchangeCode(r.Context(), TokenRequest{GrantType: r.Form.Get("grant_type"), Code: r.Form.Get("code"), ClientID: clientID, ClientSecret: clientSecret, AuthMethod: method, RedirectURI: r.Form.Get("redirect_uri"), CodeVerifier: r.Form.Get("code_verifier")})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": result.AccessToken, "token_type": result.TokenType, "expires_in": result.ExpiresIn, "id_token": result.IDToken, "scope": strings.Join(result.Scope, " ")})
	})
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(provider.Metadata())
	})
	mux.HandleFunc("/jwks.json", func(w http.ResponseWriter, r *http.Request) {
		keys, err := provider.PublicJWKS(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(keys)
	})
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		raw := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		claims, err := provider.UserInfo(r.Context(), raw)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(claims)
	})

	config := oauth2.Config{ClientID: "interop", ClientSecret: secret, RedirectURL: "https://client.example/callback", Scopes: []string{"openid", "profile"}, Endpoint: oauth2.Endpoint{AuthURL: server.URL + "/authorize", TokenURL: server.URL + "/token", AuthStyle: oauth2.AuthStyleInParams}}
	verifier := oauth2.GenerateVerifier()
	authURL := config.AuthCodeURL("state-1", oauth2.S256ChallengeOption(verifier))
	client := server.Client()
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	response, err := client.Get(authURL)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	redirect, err := url.Parse(response.Header.Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	if redirect.Query().Get("state") != "state-1" {
		t.Fatal("state was not preserved")
	}
	if redirect.Query().Get("iss") != server.URL {
		t.Fatal("authorization response issuer was not preserved")
	}
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, server.Client())
	token, err := config.Exchange(ctx, redirect.Query().Get("code"), oauth2.VerifierOption(verifier))
	if err != nil {
		t.Fatal(err)
	}
	if !token.Valid() || token.Extra("id_token") == nil {
		t.Fatalf("standard client received invalid token: %+v", token)
	}
	discoveryResponse, err := server.Client().Get(server.URL + "/.well-known/openid-configuration")
	if err != nil {
		t.Fatal(err)
	}
	var metadata DiscoveryMetadata
	if err := json.NewDecoder(discoveryResponse.Body).Decode(&metadata); err != nil {
		t.Fatal(err)
	}
	discoveryResponse.Body.Close()
	if metadata.Issuer != server.URL || metadata.JWKSEndpoint != server.URL+"/jwks.json" {
		t.Fatalf("invalid discovery metadata: %+v", metadata)
	}
	jwksResponse, err := server.Client().Get(metadata.JWKSEndpoint)
	if err != nil {
		t.Fatal(err)
	}
	var jwks JWKSet
	if err := json.NewDecoder(jwksResponse.Body).Decode(&jwks); err != nil {
		t.Fatal(err)
	}
	jwksResponse.Body.Close()
	if len(jwks.Keys) != 1 || jwks.Keys[0].Kid != "interop-key" {
		t.Fatalf("invalid HTTP JWKS: %+v", jwks)
	}
	userinfoRequest, _ := http.NewRequestWithContext(ctx, http.MethodGet, metadata.UserinfoEndpoint, nil)
	userinfoRequest.Header.Set("Authorization", "Bearer "+token.AccessToken)
	userinfoResponse, err := server.Client().Do(userinfoRequest)
	if err != nil {
		t.Fatal(err)
	}
	var claims UserInfoClaims
	if err := json.NewDecoder(userinfoResponse.Body).Decode(&claims); err != nil {
		t.Fatal(err)
	}
	userinfoResponse.Body.Close()
	if claims.Subject != "u-1" || claims.Name != "Alice" {
		t.Fatalf("invalid UserInfo response: %+v", claims)
	}

	basicConfig := oauth2.Config{ClientID: "interop-basic", ClientSecret: secret, RedirectURL: "https://client.example/callback", Scopes: []string{"openid"}, Endpoint: oauth2.Endpoint{AuthURL: server.URL + "/authorize", TokenURL: server.URL + "/token", AuthStyle: oauth2.AuthStyleInHeader}}
	basicVerifier := oauth2.GenerateVerifier()
	basicResponse, err := client.Get(basicConfig.AuthCodeURL("basic-state", oauth2.S256ChallengeOption(basicVerifier)))
	if err != nil {
		t.Fatal(err)
	}
	basicResponse.Body.Close()
	basicRedirect, err := url.Parse(basicResponse.Header.Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	basicToken, err := basicConfig.Exchange(ctx, basicRedirect.Query().Get("code"), oauth2.VerifierOption(basicVerifier))
	if err != nil {
		t.Fatal(err)
	}
	if !basicToken.Valid() {
		t.Fatal("client_secret_basic token is invalid")
	}
}
