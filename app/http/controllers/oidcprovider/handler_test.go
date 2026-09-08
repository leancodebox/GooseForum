package oidcprovider_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	core "github.com/leancodebox/GooseForum/app/bundles/oidcprovider"
	oidchttp "github.com/leancodebox/GooseForum/app/http/controllers/oidcprovider"
	"github.com/leancodebox/GooseForum/app/models/forum/oidcProviderStore"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newInteractionStore(t *testing.T) *oidcProviderStore.Store {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "interactions.sqlite")), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	store, err := oidcProviderStore.New(db)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Migrate(t.Context()); err != nil {
		t.Fatal(err)
	}
	return store
}

type users struct{}

func (users) ResolveUser(context.Context, string) (*core.User, error) {
	return &core.User{Subject: "user-1", Name: "Alice", Email: "alice@example.com", EmailVerified: true}, nil
}

type signer struct{}

func (signer) SignIDToken(context.Context, core.IDTokenClaims) (string, error) {
	return "header.payload.signature", nil
}

func (signer) PublicKeys(context.Context) ([]core.JWK, error) {
	return []core.JWK{{KTY: "RSA", Kid: "test", Use: "sig", Alg: "RS256", N: "n", E: "AQAB"}}, nil
}

func newFixture(t *testing.T, method core.ClientAuthenticationMethod) (*core.Provider, *core.MemoryStore, time.Time) {
	t.Helper()
	store := core.NewMemoryStore()
	client := &core.Client{
		ID: "client", Name: "Client", RedirectURIs: []string{"https://client.example/callback"},
		Scopes: []string{"openid", "profile"}, GrantTypes: []string{"authorization_code"},
		TokenEndpointAuthMethod: method, Enabled: true,
	}
	if method != core.ClientAuthNone {
		client.SecretHash = core.HashClientSecret("secret")
	} else {
		client.Public = true
	}
	store.PutClient(client)
	now := time.Unix(1_700_000_000, 0)
	provider, err := core.New(core.Config{
		Issuer: "https://forum.example/oauth2", Store: store, Users: users{}, Signer: signer{},
		Now: func() time.Time { return now },
	})
	if err != nil {
		t.Fatal(err)
	}
	return provider, store, now
}

func newEngine(t *testing.T, provider *core.Provider, resolver oidchttp.AuthenticationResolver) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	handler, err := oidchttp.New(func() (*core.Provider, error) { return provider, nil }, resolver)
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.GET("/.well-known/openid-configuration", handler.Discovery)
	router.GET("/authorize", handler.Authorize)
	router.POST("/token", handler.Token)
	router.GET("/userinfo", handler.UserInfo)
	router.POST("/revoke", handler.Revoke)
	router.GET("/jwks.json", handler.JWKS)
	return router
}

func TestDiscoveryAndJWKSUseStandardDocuments(t *testing.T) {
	provider, _, _ := newFixture(t, core.ClientSecretPost)
	router := newEngine(t, provider, nil)

	for _, path := range []string{"/.well-known/openid-configuration", "/jwks.json"} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK || !strings.HasPrefix(response.Header().Get("Content-Type"), "application/json") {
			t.Fatalf("%s returned status=%d content-type=%q", path, response.Code, response.Header().Get("Content-Type"))
		}
		if strings.Contains(response.Body.String(), `"data"`) || strings.Contains(response.Body.String(), `"code":200`) {
			t.Fatalf("%s unexpectedly used API envelope: %s", path, response.Body.String())
		}
	}
}

func TestTokenUserInfoAndRevocationFlow(t *testing.T) {
	provider, _, now := newFixture(t, core.ClientSecretPost)
	router := newEngine(t, provider, nil)
	authorization, err := provider.Authorize(context.Background(), core.AuthorizeRequest{
		ClientID: "client", RedirectURI: "https://client.example/callback", ResponseType: "code",
		Scope: []string{"openid", "profile"},
	}, core.Authentication{UserID: "1", AuthTime: now}, true)
	if err != nil {
		t.Fatal(err)
	}

	form := url.Values{
		"grant_type": {"authorization_code"}, "code": {authorization.Code},
		"redirect_uri": {"https://client.example/callback"}, "client_id": {"client"}, "client_secret": {"secret"},
	}
	response := performForm(router, "/token", form)
	if response.Code != http.StatusOK {
		t.Fatalf("token status=%d body=%s", response.Code, response.Body.String())
	}
	if response.Header().Get("Cache-Control") != "no-store" || response.Header().Get("Pragma") != "no-cache" {
		t.Fatal("token response is cacheable")
	}
	var token struct {
		AccessToken string `json:"access_token"`
		IDToken     string `json:"id_token"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &token); err != nil || token.AccessToken == "" || token.IDToken == "" {
		t.Fatalf("invalid token response: %s (%v)", response.Body.String(), err)
	}

	userinfoRequest := httptest.NewRequest(http.MethodGet, "/userinfo", nil)
	userinfoRequest.Header.Set("Authorization", "Bearer "+token.AccessToken)
	userinfo := httptest.NewRecorder()
	router.ServeHTTP(userinfo, userinfoRequest)
	if userinfo.Code != http.StatusOK || !strings.Contains(userinfo.Body.String(), `"sub":"user-1"`) {
		t.Fatalf("userinfo status=%d body=%s", userinfo.Code, userinfo.Body.String())
	}

	revocation := performForm(router, "/revoke", url.Values{
		"token": {token.AccessToken}, "client_id": {"client"}, "client_secret": {"secret"},
	})
	if revocation.Code != http.StatusOK || revocation.Body.Len() != 0 {
		t.Fatalf("revoke status=%d body=%s", revocation.Code, revocation.Body.String())
	}
	userinfo = httptest.NewRecorder()
	router.ServeHTTP(userinfo, userinfoRequest)
	if userinfo.Code != http.StatusUnauthorized || !strings.HasPrefix(userinfo.Header().Get("WWW-Authenticate"), "Bearer ") {
		t.Fatalf("revoked userinfo status=%d authenticate=%q", userinfo.Code, userinfo.Header().Get("WWW-Authenticate"))
	}
}

func TestTokenRequiresFormAndRejectsDuplicateParameters(t *testing.T) {
	provider, _, _ := newFixture(t, core.ClientSecretPost)
	router := newEngine(t, provider, nil)

	request := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(`{"grant_type":"authorization_code"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assertOAuthError(t, response, http.StatusBadRequest, "invalid_request")
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("error response is cacheable")
	}

	response = performForm(router, "/token", url.Values{
		"grant_type": {"authorization_code", "refresh_token"}, "client_id": {"client"}, "client_secret": {"secret"},
	})
	assertOAuthError(t, response, http.StatusBadRequest, "invalid_request")
}

func TestBasicAuthenticationAndMixedMethods(t *testing.T) {
	provider, _, now := newFixture(t, core.ClientSecretBasic)
	router := newEngine(t, provider, nil)
	authorization, err := provider.Authorize(context.Background(), core.AuthorizeRequest{
		ClientID: "client", RedirectURI: "https://client.example/callback", ResponseType: "code", Scope: []string{"openid"},
	}, core.Authentication{UserID: "1", AuthTime: now}, true)
	if err != nil {
		t.Fatal(err)
	}
	form := url.Values{"grant_type": {"authorization_code"}, "code": {authorization.Code}, "redirect_uri": {"https://client.example/callback"}}
	request := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("client", "secret")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("basic token status=%d body=%s", response.Code, response.Body.String())
	}

	mixed := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(url.Values{"grant_type": {"authorization_code"}, "client_id": {"client"}}.Encode()))
	mixed.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mixed.SetBasicAuth("client", "secret")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, mixed)
	assertOAuthError(t, response, http.StatusBadRequest, "invalid_request")
}

func TestAuthorizeReauthenticatesLegacySessionWithoutAuthenticationTime(t *testing.T) {
	provider, _, now := newFixture(t, core.ClientSecretPost)
	handler, err := oidchttp.New(func() (*core.Provider, error) { return provider, nil }, func(*gin.Context) core.Authentication {
		return core.Authentication{UserID: "1", AuthID: "legacy-auth"}
	})
	if err != nil {
		t.Fatal(err)
	}
	handler.WithInteractions(newInteractionStore(t), func() time.Time { return now }, nil)
	router := gin.New()
	router.GET("/authorize", handler.Authorize)
	requestTarget := "/authorize?client_id=client&redirect_uri=https%3A%2F%2Fclient.example%2Fcallback&response_type=code&scope=openid&state=s"
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, requestTarget, nil))
	if response.Code != http.StatusFound || !strings.HasPrefix(response.Header().Get("Location"), "/login?force=true&redirect=") {
		t.Fatalf("authorize status=%d location=%q", response.Code, response.Header().Get("Location"))
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("authorization response is cacheable")
	}
	location, err := url.Parse(response.Header().Get("Location"))
	resume, resumeErr := url.Parse(location.Query().Get("redirect"))
	if err != nil || resumeErr != nil || resume.Path != "/oauth2/authorize/resume" || resume.Query().Get("interaction") == "" || resume.Query().Get("client_id") != "" {
		t.Fatalf("unsafe or incomplete login return URL: %q (%v)", response.Header().Get("Location"), err)
	}
}

func TestPromptLoginAndMaxAgeRequireTrustedReauthentication(t *testing.T) {
	for _, tc := range []struct {
		name, query string
		initial     bool
	}{
		{"reauth prompt login", "&prompt=login", false},
		{"reauth max age", "&max_age=0", false},
		{"reauth max age one second", "&max_age=1", false},
		{"initial prompt login", "&prompt=login", true},
		{"initial max age", "&max_age=0", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			provider, protocolStore, now := newFixture(t, core.ClientSecretPost)
			if err := protocolStore.SaveConsent(t.Context(), &core.Consent{UserID: "1", ClientID: "client", Scopes: []string{"openid"}}); err != nil {
				t.Fatal(err)
			}
			interactions := newInteractionStore(t)
			auth := core.Authentication{UserID: "1", AuthTime: now.Add(-time.Hour), AuthID: "old-auth"}
			if tc.initial {
				auth = core.Authentication{}
			}
			handler, err := oidchttp.New(func() (*core.Provider, error) { return provider, nil }, func(*gin.Context) core.Authentication { return auth })
			if err != nil {
				t.Fatal(err)
			}
			handler.WithInteractions(interactions, func() time.Time { return now }, nil)
			router := gin.New()
			router.GET("/authorize", handler.Authorize)
			router.GET("/oauth2/authorize/resume", handler.ResumeAuthorization)

			requestTarget := "/authorize?client_id=client&redirect_uri=https%3A%2F%2Fclient.example%2Fcallback&response_type=code&scope=openid&state=s" + tc.query
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, requestTarget, nil))
			location, parseErr := url.Parse(response.Header().Get("Location"))
			if response.Code != http.StatusFound || parseErr != nil || location.Path != "/login" || location.Query().Get("force") != "true" {
				t.Fatalf("reauth redirect status=%d location=%q err=%v", response.Code, response.Header().Get("Location"), parseErr)
			}
			resume, parseErr := url.Parse(location.Query().Get("redirect"))
			if parseErr != nil || resume.Path != "/oauth2/authorize/resume" || resume.Query().Get("interaction") == "" || strings.Contains(resume.RawQuery, "client_id") {
				t.Fatalf("resume URL=%q err=%v", location.Query().Get("redirect"), parseErr)
			}

			oldAttempt := httptest.NewRecorder()
			router.ServeHTTP(oldAttempt, httptest.NewRequest(http.MethodGet, resume.String(), nil))
			if tc.initial {
				assertOAuthError(t, oldAttempt, http.StatusUnauthorized, "login_required")
			} else {
				assertOAuthError(t, oldAttempt, http.StatusBadRequest, "invalid_request")
			}

			// An OAuth callback can establish a new session but cannot prove reauthentication.
			auth = core.Authentication{UserID: "1", AuthTime: now}
			oauthAttempt := httptest.NewRecorder()
			router.ServeHTTP(oauthAttempt, httptest.NewRequest(http.MethodGet, resume.String(), nil))
			assertOAuthError(t, oauthAttempt, http.StatusUnauthorized, "login_required")

			auth = core.Authentication{UserID: "1", AuthTime: now, AuthID: "new-auth"}
			completed := httptest.NewRecorder()
			router.ServeHTTP(completed, httptest.NewRequest(http.MethodGet, resume.String(), nil))
			callback, parseErr := url.Parse(completed.Header().Get("Location"))
			if completed.Code != http.StatusFound || parseErr != nil || callback.Host != "client.example" || callback.Query().Get("code") == "" || callback.Query().Get("state") != "s" {
				t.Fatalf("completed status=%d location=%q err=%v", completed.Code, completed.Header().Get("Location"), parseErr)
			}

			replay := httptest.NewRecorder()
			router.ServeHTTP(replay, httptest.NewRequest(http.MethodGet, resume.String(), nil))
			assertOAuthError(t, replay, http.StatusBadRequest, "invalid_request")
		})
	}
}

func TestConsentInteractionApproveDenyAndReplay(t *testing.T) {
	provider, _, now := newFixture(t, core.ClientSecretPost)
	store := newInteractionStore(t)
	auth := core.Authentication{UserID: "1", AuthTime: now.Add(-time.Minute), Fresh: true}
	handler, err := oidchttp.New(func() (*core.Provider, error) { return provider, nil }, func(*gin.Context) core.Authentication { return auth })
	if err != nil {
		t.Fatal(err)
	}
	handler.WithInteractions(store, func() time.Time { return now }, nil)
	router := gin.New()
	router.GET("/authorize", handler.Authorize)
	router.GET("/oauth2/consent/details", handler.ConsentDetails)
	router.POST("/oauth2/consent", handler.ConsentDecision)

	start := func(prompt string) string {
		t.Helper()
		query := url.Values{
			"client_id": {"client"}, "redirect_uri": {"https://client.example/callback"},
			"response_type": {"code"}, "scope": {"openid profile"}, "state": {"original-state"},
		}
		if prompt != "" {
			query.Set("prompt", prompt)
		}
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/authorize?"+query.Encode(), nil))
		if response.Code != http.StatusFound {
			t.Fatalf("authorize status=%d body=%s", response.Code, response.Body.String())
		}
		location, err := url.Parse(response.Header().Get("Location"))
		if err != nil || location.Path != "/oauth2/consent" || location.Query().Get("interaction") == "" {
			t.Fatalf("consent location=%q err=%v", response.Header().Get("Location"), err)
		}
		return location.Query().Get("interaction")
	}

	interaction := start("")
	details := httptest.NewRecorder()
	router.ServeHTTP(details, httptest.NewRequest(http.MethodGet, "/oauth2/consent/details?interaction="+url.QueryEscape(interaction), nil))
	if details.Code != http.StatusOK || !strings.Contains(details.Body.String(), `"name":"Client"`) || !strings.Contains(details.Body.String(), `"scopes":["openid","profile"]`) || strings.Contains(details.Body.String(), "redirect_uri") || strings.Contains(details.Body.String(), "original-state") {
		t.Fatalf("details status=%d body=%s", details.Code, details.Body.String())
	}

	// Unknown authorization fields are rejected before consumption. They are
	// never merged into the persisted request.
	forged := performJSON(router, "/oauth2/consent", `{"interaction":"`+interaction+`","decision":"approve","redirect_uri":"https://evil.example"}`)
	assertOAuthError(t, forged, http.StatusBadRequest, "invalid_request")

	approved := performJSON(router, "/oauth2/consent", `{"interaction":"`+interaction+`","decision":"approve"}`)
	if approved.Code != http.StatusOK {
		t.Fatalf("approve status=%d body=%s", approved.Code, approved.Body.String())
	}
	var result struct {
		RedirectURL string `json:"redirect_url"`
	}
	if err := json.Unmarshal(approved.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	redirect, err := url.Parse(result.RedirectURL)
	if err != nil || redirect.Host != "client.example" || redirect.Query().Get("code") == "" || redirect.Query().Get("state") != "original-state" {
		t.Fatalf("approve redirect=%q err=%v", result.RedirectURL, err)
	}
	replay := performJSON(router, "/oauth2/consent", `{"interaction":"`+interaction+`","decision":"approve"}`)
	assertOAuthError(t, replay, http.StatusBadRequest, "invalid_request")

	deniedInteraction := start("consent")
	denied := performJSON(router, "/oauth2/consent", `{"interaction":"`+deniedInteraction+`","decision":"deny"}`)
	if denied.Code != http.StatusOK {
		t.Fatalf("deny status=%d body=%s", denied.Code, denied.Body.String())
	}
	if err := json.Unmarshal(denied.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	redirect, err = url.Parse(result.RedirectURL)
	if err != nil || redirect.Query().Get("error") != "access_denied" || redirect.Query().Get("state") != "original-state" {
		t.Fatalf("deny redirect=%q err=%v", result.RedirectURL, err)
	}
}

func TestUnavailableProviderReturnsServiceUnavailable(t *testing.T) {
	handler, err := oidchttp.New(func() (*core.Provider, error) { return nil, errors.New("not initialized") }, nil)
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.GET("/discovery", handler.Discovery)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/discovery", nil))
	assertOAuthError(t, response, http.StatusServiceUnavailable, "temporarily_unavailable")
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("unavailable response is cacheable")
	}
}

func performForm(router http.Handler, path string, form url.Values) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func performJSON(router http.Handler, path, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "http://example.com")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func assertOAuthError(t *testing.T, response *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if response.Code != status || !strings.Contains(response.Body.String(), `"error":"`+code+`"`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestUnsupportedRequestObjectsRedirectSafely(t *testing.T) {
	provider, _, _ := newFixture(t, core.ClientSecretPost)
	router := newEngine(t, provider, nil)
	router.POST("/authorize", func(c *gin.Context) {
		h, _ := oidchttp.New(func() (*core.Provider, error) { return provider, nil }, nil)
		h.Authorize(c)
	})
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		for parameter, expected := range map[string]string{"request": "request_not_supported", "request_uri": "request_uri_not_supported"} {
			for _, redirect := range []string{"https://client.example/callback", "https://attacker.example/callback"} {
				values := url.Values{"client_id": {"client"}, "redirect_uri": {redirect}, "response_type": {"code"}, "scope": {"openid"}, "state": {"original state"}, parameter: {"unsupported"}}
				target := "/authorize"
				body := ""
				if method == http.MethodGet {
					target += "?" + values.Encode()
				} else {
					body = values.Encode()
				}
				req := httptest.NewRequest(method, target, strings.NewReader(body))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				response := httptest.NewRecorder()
				router.ServeHTTP(response, req)
				if strings.Contains(redirect, "attacker") {
					assertOAuthError(t, response, 400, "invalid_request")
					continue
				}
				location, err := url.Parse(response.Header().Get("Location"))
				if err != nil || response.Code != 302 || location.Query().Get("error") != expected || location.Query().Get("state") != "original state" {
					t.Fatalf("%s %s: %d %s", method, parameter, response.Code, response.Header().Get("Location"))
				}
			}
		}
	}
}
