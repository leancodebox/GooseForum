package oauthservice

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
)

func TestBuildProvidersBuiltIns(t *testing.T) {
	config := pageConfig.OAuthSettingsConfig{
		GitHub:  pageConfig.OAuthProviderConfig{Enabled: true, ClientID: "github-id", ClientSecret: "github-secret"},
		Google:  pageConfig.OAuthProviderConfig{Enabled: true, ClientID: "google-id", ClientSecret: "google-secret"},
		Discord: pageConfig.OAuthProviderConfig{Enabled: true, ClientID: "discord-id", ClientSecret: "discord-secret"},
	}
	providers, enabled, err := buildProviders(config, "https://forum.example.com/")
	if err != nil {
		t.Fatal(err)
	}
	if len(providers) != 3 || len(enabled) != 3 {
		t.Fatalf("providers = %d, enabled = %d", len(providers), len(enabled))
	}

	wantHosts := map[string]string{
		ProviderGitHub: "github.com", ProviderGoogle: "accounts.google.com", ProviderDiscord: "discord.com",
	}
	for _, provider := range providers {
		session, err := provider.BeginAuth("state-value")
		if err != nil {
			t.Fatalf("begin %s auth: %v", provider.Name(), err)
		}
		authURL, err := session.GetAuthURL()
		if err != nil {
			t.Fatalf("get %s auth URL: %v", provider.Name(), err)
		}
		parsed, err := url.Parse(authURL)
		if err != nil {
			t.Fatal(err)
		}
		if parsed.Host != wantHosts[provider.Name()] {
			t.Errorf("%s auth host = %q", provider.Name(), parsed.Host)
		}
		if parsed.Query().Get("state") != "state-value" {
			t.Errorf("%s state was not preserved", provider.Name())
		}
		wantCallback := "https://forum.example.com/api/auth/" + provider.Name() + "/callback"
		if parsed.Query().Get("redirect_uri") != wantCallback {
			t.Errorf("%s callback = %q, want %q", provider.Name(), parsed.Query().Get("redirect_uri"), wantCallback)
		}
	}
}

func TestBuildProvidersCustomOIDC(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/openid-configuration" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{
			"issuer":                 server.URL,
			"authorization_endpoint": server.URL + "/authorize",
			"token_endpoint":         server.URL + "/token",
			"userinfo_endpoint":      server.URL + "/userinfo",
		})
	}))
	defer server.Close()

	config := pageConfig.OAuthSettingsConfig{Custom: []pageConfig.OIDCProviderConfig{{
		Key: "company-sso", DisplayName: "Company SSO", Enabled: true,
		ClientID: "client-id", ClientSecret: "client-secret",
		DiscoveryURL: server.URL + "/.well-known/openid-configuration", Scopes: []string{"profile", "email"},
	}}}
	providers, enabled, err := buildProviders(config, "https://forum.example.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(providers) != 1 || enabled["company-sso"].DisplayName != "Company SSO" {
		t.Fatalf("unexpected custom providers: %#v", enabled)
	}
	session, err := providers[0].BeginAuth("oidc-state")
	if err != nil {
		t.Fatal(err)
	}
	authURL, _ := session.GetAuthURL()
	parsed, _ := url.Parse(authURL)
	if parsed.Path != "/authorize" || parsed.Query().Get("state") != "oidc-state" {
		t.Fatalf("custom auth URL = %q", authURL)
	}
	if scopes := strings.Fields(parsed.Query().Get("scope")); !slices.Contains(scopes, "openid") || !slices.Contains(scopes, "email") {
		t.Fatalf("custom scopes = %#v", scopes)
	}
}

func TestBuildProvidersRejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		config pageConfig.OAuthSettingsConfig
		part   string
	}{
		{
			name:   "incomplete built in",
			config: pageConfig.OAuthSettingsConfig{Google: pageConfig.OAuthProviderConfig{Enabled: true, ClientID: "id"}},
			part:   "client ID and client secret",
		},
		{
			name:   "reserved custom key",
			config: pageConfig.OAuthSettingsConfig{Custom: []pageConfig.OIDCProviderConfig{{Key: "github", DisplayName: "Other"}}},
			part:   "duplicate",
		},
		{
			name:   "invalid custom key",
			config: pageConfig.OAuthSettingsConfig{Custom: []pageConfig.OIDCProviderConfig{{Key: "Bad Key", DisplayName: "Other"}}},
			part:   "must match",
		},
		{
			name: "insecure remote discovery",
			config: pageConfig.OAuthSettingsConfig{Custom: []pageConfig.OIDCProviderConfig{{
				Key: "other", DisplayName: "Other", Enabled: true, ClientID: "id", ClientSecret: "secret", DiscoveryURL: "http://id.example.com/.well-known/openid-configuration",
			}}},
			part: "must use HTTPS",
		},
		{
			name: "invalid scope token",
			config: pageConfig.OAuthSettingsConfig{Custom: []pageConfig.OIDCProviderConfig{{
				Key: "other", DisplayName: "Other", Enabled: true, ClientID: "id", ClientSecret: "secret",
				DiscoveryURL: "https://id.example.com/.well-known/openid-configuration", Scopes: []string{"openid", "bad scope"},
			}}},
			part: "invalid scope",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, err := buildProviders(test.config, "https://forum.example.com")
			if err == nil || !strings.Contains(err.Error(), test.part) {
				t.Fatalf("error = %v, want containing %q", err, test.part)
			}
		})
	}
}

func TestNormalizeOIDCConfig(t *testing.T) {
	config := normalizeOIDCConfig(pageConfig.OIDCProviderConfig{
		Key: " Company-SSO ", DisplayName: " Company SSO ", Scopes: []string{"email", "email", " profile "},
	})
	if config.Key != "company-sso" || config.DisplayName != "Company SSO" {
		t.Fatalf("normalized identity = %#v", config)
	}
	if want := []string{"openid", "email", "profile"}; !slices.Equal(config.Scopes, want) {
		t.Fatalf("scopes = %#v, want %#v", config.Scopes, want)
	}
}

func TestMergedSecret(t *testing.T) {
	if got := mergedSecret("old", ProviderUpdate{}); got != "old" {
		t.Fatalf("blank update replaced secret: %q", got)
	}
	if got := mergedSecret("old", ProviderUpdate{ClientSecret: " new "}); got != "new" {
		t.Fatalf("new secret = %q", got)
	}
	if got := mergedSecret("old", ProviderUpdate{ClientSecret: "new", ClearClientSecret: true}); got != "" {
		t.Fatalf("cleared secret = %q", got)
	}
}

func TestOAuthUsername(t *testing.T) {
	validUsername := regexp.MustCompile(`^[a-zA-Z0-9_-]{6,32}$`)
	tests := []struct {
		name string
		user OAuthUserInfo
	}{
		{name: "github login", user: OAuthUserInfo{Provider: "github", ID: "1", Login: "octocat"}},
		{name: "google display name", user: OAuthUserInfo{Provider: "google", ID: "2", Name: "Ada Lovelace"}},
		{name: "discord missing nickname", user: OAuthUserInfo{Provider: "discord", ID: "3"}},
		{name: "long login", user: OAuthUserInfo{Provider: "github", ID: "4", Login: strings.Repeat("a", 80)}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			username := oauthUsername(test.user)
			if !validUsername.MatchString(username) {
				t.Fatalf("invalid OAuth username %q", username)
			}
		})
	}
}

func TestValidateOAuthUserInfo(t *testing.T) {
	for _, user := range []OAuthUserInfo{
		{Provider: "", ID: "123"},
		{Provider: "github", ID: ""},
		{Provider: "github", ID: "   "},
	} {
		if err := validateOAuthUserInfo(user); err == nil {
			t.Fatalf("expected invalid OAuth identity %#v to fail", user)
		}
	}
	if err := validateOAuthUserInfo(OAuthUserInfo{Provider: "github", ID: "123"}); err != nil {
		t.Fatalf("valid OAuth identity failed: %v", err)
	}
}
