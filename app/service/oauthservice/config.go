package oauthservice

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/userOAuth"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/discord"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
	"github.com/markbates/goth/providers/openidConnect"
)

const (
	ProviderGitHub  = "github"
	ProviderGoogle  = "google"
	ProviderDiscord = "discord"
)

var providerKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{1,31}$`)
var scopeTokenPattern = regexp.MustCompile(`^[\x21\x23-\x5B\x5D-\x7E]+$`)

const oidcDiscoveryMaxBytes = 1024 * 1024

type PublicProvider struct {
	Key         string `json:"key"`
	DisplayName string `json:"displayName"`
}

type AdminProviderView struct {
	Key                    string   `json:"key"`
	DisplayName            string   `json:"displayName"`
	Kind                   string   `json:"kind"`
	Enabled                bool     `json:"enabled"`
	ClientID               string   `json:"clientId"`
	ClientSecretConfigured bool     `json:"clientSecretConfigured"`
	CallbackURL            string   `json:"callbackUrl"`
	DiscoveryURL           string   `json:"discoveryUrl,omitempty"`
	Scopes                 []string `json:"scopes,omitempty"`
}

type AdminSettingsView struct {
	Providers []AdminProviderView `json:"providers"`
}

type BindingProvider struct {
	Key         string     `json:"key"`
	DisplayName string     `json:"displayName"`
	Enabled     bool       `json:"enabled"`
	Bound       bool       `json:"bound"`
	BoundAt     *time.Time `json:"boundAt,omitempty"`
}

type ProviderUpdate struct {
	Key               string   `json:"key"`
	DisplayName       string   `json:"displayName"`
	Kind              string   `json:"kind"`
	Enabled           bool     `json:"enabled"`
	ClientID          string   `json:"clientId"`
	ClientSecret      string   `json:"clientSecret"`
	ClearClientSecret bool     `json:"clearClientSecret"`
	DiscoveryURL      string   `json:"discoveryUrl"`
	Scopes            []string `json:"scopes"`
}

type SettingsUpdate struct {
	Providers []ProviderUpdate `json:"providers" validate:"required"`
}

type runtimeProvider struct {
	PublicProvider
	Kind string
}

var providerRuntime = struct {
	sync.RWMutex
	enabled map[string]runtimeProvider
}{enabled: map[string]runtimeProvider{}}

func callbackURL(siteURL, provider string) string {
	return strings.TrimRight(strings.TrimSpace(siteURL), "/") + "/api/auth/" + url.PathEscape(provider) + "/callback"
}

func IsProviderEnabled(key string) bool {
	providerRuntime.RLock()
	defer providerRuntime.RUnlock()
	_, ok := providerRuntime.enabled[key]
	return ok
}

func EnabledProviders() []PublicProvider {
	providerRuntime.RLock()
	defer providerRuntime.RUnlock()
	providers := make([]PublicProvider, 0, len(providerRuntime.enabled))
	for _, provider := range providerRuntime.enabled {
		providers = append(providers, provider.PublicProvider)
	}
	slices.SortFunc(providers, func(a, b PublicProvider) int {
		return strings.Compare(a.DisplayName, b.DisplayName)
	})
	return providers
}

func BindingProviders(userID uint64) []BindingProvider {
	config := loadSettings()
	names := map[string]string{
		ProviderGitHub: "GitHub", ProviderGoogle: "Google", ProviderDiscord: "Discord",
	}
	for _, custom := range config.Custom {
		names[custom.Key] = custom.DisplayName
	}
	bindings := make(map[string]userOAuth.Entity)
	for _, binding := range userOAuth.ListByUserID(userID) {
		bindings[binding.Provider] = binding
	}
	candidates := make(map[string]bool)
	for _, provider := range EnabledProviders() {
		candidates[provider.Key] = true
	}
	for key := range bindings {
		candidates[key] = true
	}
	result := make([]BindingProvider, 0, len(candidates))
	for key := range candidates {
		binding, bound := bindings[key]
		name := strings.TrimSpace(names[key])
		if name == "" {
			name = key
		}
		item := BindingProvider{Key: key, DisplayName: name, Enabled: IsProviderEnabled(key), Bound: bound}
		if bound {
			item.BoundAt = &binding.CreatedAt
		}
		result = append(result, item)
	}
	slices.SortFunc(result, func(a, b BindingProvider) int {
		return strings.Compare(a.DisplayName, b.DisplayName)
	})
	return result
}

func enabledProviderKeys() []string {
	providers := EnabledProviders()
	keys := make([]string, 0, len(providers))
	for _, provider := range providers {
		keys = append(keys, provider.Key)
	}
	return keys
}

func ReloadProviders(config pageConfig.OAuthSettingsConfig) error {
	siteURL := hotdataserve.GetSiteSettingsConfigCache().SiteUrl
	providers, enabled, err := buildProviders(config, siteURL)
	if err != nil {
		return err
	}
	activateProviders(providers, enabled)
	return nil
}

func ReloadCurrentProviders() error {
	return ReloadProviders(loadSettings())
}

func activateProviders(providers []goth.Provider, enabled map[string]runtimeProvider) {
	if len(providers) > 0 {
		goth.UseProviders(providers...)
	}
	providerRuntime.Lock()
	providerRuntime.enabled = enabled
	providerRuntime.Unlock()
}

func buildProviders(config pageConfig.OAuthSettingsConfig, siteURL string) ([]goth.Provider, map[string]runtimeProvider, error) {
	hasEnabledProvider := config.GitHub.Enabled || config.Google.Enabled || config.Discord.Enabled
	for _, custom := range config.Custom {
		hasEnabledProvider = hasEnabledProvider || custom.Enabled
	}
	if hasEnabledProvider {
		if strings.TrimSpace(siteURL) == "" {
			return nil, nil, errors.New("site URL is required before enabling OAuth")
		}
		if parsed, err := url.ParseRequestURI(siteURL); err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return nil, nil, errors.New("site URL must be an absolute HTTP(S) URL")
		}
	}

	providers := make([]goth.Provider, 0, 3+len(config.Custom))
	enabled := make(map[string]runtimeProvider)
	addBuiltIn := func(key, displayName string, cfg pageConfig.OAuthProviderConfig, factory func(string) goth.Provider) error {
		if !cfg.Enabled {
			return nil
		}
		if strings.TrimSpace(cfg.ClientID) == "" || strings.TrimSpace(cfg.ClientSecret) == "" {
			return fmt.Errorf("%s requires both client ID and client secret", displayName)
		}
		providers = append(providers, factory(callbackURL(siteURL, key)))
		enabled[key] = runtimeProvider{PublicProvider: PublicProvider{Key: key, DisplayName: displayName}, Kind: key}
		return nil
	}

	if err := addBuiltIn(ProviderGitHub, "GitHub", config.GitHub, func(callback string) goth.Provider {
		return github.New(config.GitHub.ClientID, config.GitHub.ClientSecret, callback, "read:user", "user:email")
	}); err != nil {
		return nil, nil, err
	}
	if err := addBuiltIn(ProviderGoogle, "Google", config.Google, func(callback string) goth.Provider {
		return google.New(config.Google.ClientID, config.Google.ClientSecret, callback, "openid", "profile", "email")
	}); err != nil {
		return nil, nil, err
	}
	if err := addBuiltIn(ProviderDiscord, "Discord", config.Discord, func(callback string) goth.Provider {
		return discord.New(config.Discord.ClientID, config.Discord.ClientSecret, callback, discord.ScopeIdentify, discord.ScopeEmail)
	}); err != nil {
		return nil, nil, err
	}

	seen := map[string]bool{ProviderGitHub: true, ProviderGoogle: true, ProviderDiscord: true}
	for _, custom := range config.Custom {
		custom = normalizeOIDCConfig(custom)
		if err := validateOIDCConfig(custom, seen); err != nil {
			return nil, nil, err
		}
		seen[custom.Key] = true
		if !custom.Enabled {
			continue
		}
		provider, err := newOIDCProvider(custom, callbackURL(siteURL, custom.Key))
		if err != nil {
			return nil, nil, fmt.Errorf("initialize OIDC provider %s: %w", custom.Key, err)
		}
		providers = append(providers, provider)
		enabled[custom.Key] = runtimeProvider{PublicProvider: PublicProvider{Key: custom.Key, DisplayName: custom.DisplayName}, Kind: "oidc"}
	}
	return providers, enabled, nil
}

func normalizeOIDCConfig(config pageConfig.OIDCProviderConfig) pageConfig.OIDCProviderConfig {
	config.Key = strings.ToLower(strings.TrimSpace(config.Key))
	config.DisplayName = strings.TrimSpace(config.DisplayName)
	config.ClientID = strings.TrimSpace(config.ClientID)
	config.ClientSecret = strings.TrimSpace(config.ClientSecret)
	config.DiscoveryURL = strings.TrimSpace(config.DiscoveryURL)
	if len(config.Scopes) == 0 {
		config.Scopes = []string{"openid", "profile", "email"}
	} else {
		unique := make(map[string]struct{}, len(config.Scopes)+1)
		scopes := make([]string, 0, len(config.Scopes)+1)
		for _, scope := range config.Scopes {
			scope = strings.TrimSpace(scope)
			if scope == "" {
				continue
			}
			if _, ok := unique[scope]; !ok {
				unique[scope] = struct{}{}
				scopes = append(scopes, scope)
			}
		}
		if _, ok := unique["openid"]; !ok {
			scopes = append([]string{"openid"}, scopes...)
		}
		config.Scopes = scopes
	}
	return config
}

func validateOIDCConfig(config pageConfig.OIDCProviderConfig, seen map[string]bool) error {
	if !providerKeyPattern.MatchString(config.Key) {
		return fmt.Errorf("OIDC provider key %q must match %s", config.Key, providerKeyPattern.String())
	}
	if seen[config.Key] {
		return fmt.Errorf("duplicate OAuth provider key %q", config.Key)
	}
	if config.DisplayName == "" {
		return fmt.Errorf("OIDC provider %s requires a display name", config.Key)
	}
	if !config.Enabled {
		return nil
	}
	if config.ClientID == "" || config.ClientSecret == "" || config.DiscoveryURL == "" {
		return fmt.Errorf("OIDC provider %s requires client ID, client secret, and discovery URL", config.Key)
	}
	if !validSecureEndpoint(config.DiscoveryURL) {
		return fmt.Errorf("OIDC provider %s discovery URL must use HTTPS (HTTP is allowed only for loopback hosts)", config.Key)
	}
	for _, scope := range config.Scopes {
		if len(scope) > 128 || !scopeTokenPattern.MatchString(scope) {
			return fmt.Errorf("OIDC provider %s has invalid scope %q", config.Key, scope)
		}
	}
	return nil
}

func newOIDCProvider(config pageConfig.OIDCProviderConfig, callback string) (*openidConnect.Provider, error) {
	discovery, err := fetchOIDCDiscovery(config.DiscoveryURL)
	if err != nil {
		return nil, err
	}
	if err := validateDiscoveredOIDC(config.Key, discovery); err != nil {
		return nil, err
	}
	provider, err := openidConnect.NewCustomisedURL(
		config.ClientID,
		config.ClientSecret,
		callback,
		discovery.AuthEndpoint,
		discovery.TokenEndpoint,
		discovery.Issuer,
		discovery.UserInfoEndpoint,
		discovery.EndSessionEndpoint,
		config.Scopes...,
	)
	if err != nil {
		return nil, err
	}
	provider.SetName(config.Key)
	return provider, nil
}

func fetchOIDCDiscovery(discoveryURL string) (*openidConnect.OpenIDConfig, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, _ []*http.Request) error {
			if !validSecureEndpoint(req.URL.String()) {
				return errors.New("OIDC discovery redirect uses an insecure endpoint")
			}
			return nil
		},
	}
	response, err := client.Get(discoveryURL)
	if err != nil {
		return nil, fmt.Errorf("fetch discovery document: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("discovery endpoint returned HTTP %d", response.StatusCode)
	}
	reader := io.LimitReader(response.Body, oidcDiscoveryMaxBytes+1)
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read discovery document: %w", err)
	}
	if len(body) > oidcDiscoveryMaxBytes {
		return nil, errors.New("discovery document exceeds 1 MiB")
	}
	var discovery openidConnect.OpenIDConfig
	if err := json.Unmarshal(body, &discovery); err != nil {
		return nil, fmt.Errorf("decode discovery document: %w", err)
	}
	return &discovery, nil
}

func validateDiscoveredOIDC(key string, config *openidConnect.OpenIDConfig) error {
	if config == nil || strings.TrimSpace(config.Issuer) == "" || strings.TrimSpace(config.AuthEndpoint) == "" ||
		strings.TrimSpace(config.TokenEndpoint) == "" || strings.TrimSpace(config.UserInfoEndpoint) == "" {
		return fmt.Errorf("OIDC provider %s discovery document must include issuer, authorization_endpoint, token_endpoint, and userinfo_endpoint", key)
	}
	for _, endpoint := range []string{config.Issuer, config.AuthEndpoint, config.TokenEndpoint, config.UserInfoEndpoint} {
		if !validSecureEndpoint(endpoint) {
			return fmt.Errorf("OIDC provider %s returned an insecure endpoint %q", key, endpoint)
		}
	}
	return nil
}

func validSecureEndpoint(value string) bool {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(value))
	if err != nil || parsed.Host == "" {
		return false
	}
	if parsed.Scheme == "https" {
		return true
	}
	host := strings.ToLower(parsed.Hostname())
	return parsed.Scheme == "http" && (host == "localhost" || host == "127.0.0.1" || host == "::1")
}

func loadSettings() pageConfig.OAuthSettingsConfig {
	return hotdataserve.GetOAuthSettingsConfigCache()
}

func persistSettings(config pageConfig.OAuthSettingsConfig) error {
	if config.Custom == nil {
		config.Custom = []pageConfig.OIDCProviderConfig{}
	}
	if err := pageConfig.SaveConfig(pageConfig.OAuthSettings, jsonopt.Encode(config)); err != nil {
		return err
	}
	hotdataserve.ClearOAuthSettingsConfigCache()
	return nil
}

func AdminSettings() AdminSettingsView {
	config := loadSettings()
	siteURL := hotdataserve.GetSiteSettingsConfigCache().SiteUrl
	views := []AdminProviderView{
		adminBuiltInView(ProviderGitHub, "GitHub", config.GitHub, siteURL),
		adminBuiltInView(ProviderGoogle, "Google", config.Google, siteURL),
		adminBuiltInView(ProviderDiscord, "Discord", config.Discord, siteURL),
	}
	for _, custom := range config.Custom {
		custom = normalizeOIDCConfig(custom)
		views = append(views, AdminProviderView{
			Key: custom.Key, DisplayName: custom.DisplayName, Kind: "oidc", Enabled: custom.Enabled,
			ClientID: custom.ClientID, ClientSecretConfigured: custom.ClientSecret != "",
			CallbackURL: callbackURL(siteURL, custom.Key), DiscoveryURL: custom.DiscoveryURL, Scopes: custom.Scopes,
		})
	}
	return AdminSettingsView{Providers: views}
}

func adminBuiltInView(key, name string, config pageConfig.OAuthProviderConfig, siteURL string) AdminProviderView {
	return AdminProviderView{
		Key: key, DisplayName: name, Kind: key, Enabled: config.Enabled, ClientID: config.ClientID,
		ClientSecretConfigured: config.ClientSecret != "", CallbackURL: callbackURL(siteURL, key),
	}
}

func SaveSettings(update SettingsUpdate) error {
	current := loadSettings()
	customByKey := make(map[string]pageConfig.OIDCProviderConfig, len(current.Custom))
	for _, item := range current.Custom {
		customByKey[item.Key] = item
	}
	next := pageConfig.OAuthSettingsConfig{Custom: []pageConfig.OIDCProviderConfig{}}
	seen := make(map[string]bool)
	for _, item := range update.Providers {
		item.Key = strings.ToLower(strings.TrimSpace(item.Key))
		if seen[item.Key] {
			return fmt.Errorf("duplicate OAuth provider key %q", item.Key)
		}
		seen[item.Key] = true
		switch item.Key {
		case ProviderGitHub:
			next.GitHub = mergeBuiltIn(current.GitHub, item)
		case ProviderGoogle:
			next.Google = mergeBuiltIn(current.Google, item)
		case ProviderDiscord:
			next.Discord = mergeBuiltIn(current.Discord, item)
		default:
			old := customByKey[item.Key]
			custom := pageConfig.OIDCProviderConfig{
				Key: item.Key, DisplayName: item.DisplayName, Enabled: item.Enabled, ClientID: item.ClientID,
				ClientSecret: mergedSecret(old.ClientSecret, item), DiscoveryURL: item.DiscoveryURL, Scopes: item.Scopes,
			}
			next.Custom = append(next.Custom, normalizeOIDCConfig(custom))
		}
	}
	for _, required := range []string{ProviderGitHub, ProviderGoogle, ProviderDiscord} {
		if !seen[required] {
			return fmt.Errorf("missing built-in OAuth provider %q", required)
		}
	}
	providers, enabled, err := buildProviders(next, hotdataserve.GetSiteSettingsConfigCache().SiteUrl)
	if err != nil {
		return err
	}
	if err := persistSettings(next); err != nil {
		return err
	}
	activateProviders(providers, enabled)
	return nil
}

func mergeBuiltIn(current pageConfig.OAuthProviderConfig, update ProviderUpdate) pageConfig.OAuthProviderConfig {
	return pageConfig.OAuthProviderConfig{
		Enabled: update.Enabled, ClientID: strings.TrimSpace(update.ClientID),
		ClientSecret: mergedSecret(current.ClientSecret, update),
	}
}

func mergedSecret(current string, update ProviderUpdate) string {
	if update.ClearClientSecret {
		return ""
	}
	if secret := strings.TrimSpace(update.ClientSecret); secret != "" {
		return secret
	}
	return current
}
