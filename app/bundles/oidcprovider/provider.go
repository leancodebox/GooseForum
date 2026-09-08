// Package oidcprovider contains the HTTP-agnostic OAuth 2.0/OIDC provider core.
//
// The package deliberately knows nothing about Gin, net/http, GORM, cookies, or
// login pages. Applications provide authenticated users and persistence through
// the interfaces in this package.
package oidcprovider

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"
)

var (
	ErrInvalidClient           = errors.New("invalid client")
	ErrInvalidRequest          = errors.New("invalid authorization request")
	ErrInvalidGrant            = errors.New("invalid grant")
	ErrInvalidScope            = errors.New("invalid scope")
	ErrLoginRequired           = errors.New("login required")
	ErrConsentRequired         = errors.New("consent required")
	ErrAccessDenied            = errors.New("access denied")
	ErrTokenRevoked            = errors.New("token revoked")
	ErrUnsupportedGrant        = errors.New("unsupported grant type")
	ErrUnsupportedResponse     = errors.New("unsupported response type")
	ErrUnsupportedResponseMode = errors.New("unsupported response mode")
	ErrServer                  = errors.New("provider internal error")
	ErrUserUnavailable         = errors.New("user is unavailable")
	ErrAuthorizationCodeReplay = errors.New("authorization code replay")
)

// OAuthError is the protocol-level error returned by the core. An HTTP
// adapter can map Code to the OAuth error field and choose the appropriate
// status code without the core importing net/http.
type OAuthError struct {
	Code            string
	Description     string
	URI             string
	Cause           error
	RedirectAllowed bool
	RedirectURI     string
	State           string
	ResponseMode    string
}

func (e *OAuthError) Unwrap() error { return e.Cause }

func (e *OAuthError) Error() string {
	if e.Description == "" {
		return e.Code
	}
	return e.Code + ": " + e.Description
}

func NewOAuthError(code, description string) *OAuthError {
	return &OAuthError{Code: code, Description: description}
}

func protocolError(code, description string, cause error) *OAuthError {
	return &OAuthError{Code: code, Description: description, Cause: cause}
}

func authorizationError(req AuthorizeRequest, code, description string, cause error) *OAuthError {
	return &OAuthError{Code: code, Description: description, Cause: cause, RedirectAllowed: true, RedirectURI: req.RedirectURI, State: req.State, ResponseMode: safeResponseMode(req.ResponseMode)}
}

// JWK is the transport-neutral representation of a public signing key.
// N and E are base64url encoded for RSA keys; other key types may use X, Y,
// or a future extension field in the adapter.
type JWK struct {
	KTY string `json:"kty"`
	Use string `json:"use,omitempty"`
	Alg string `json:"alg,omitempty"`
	Kid string `json:"kid"`
	N   string `json:"n,omitempty"`
	E   string `json:"e,omitempty"`
	X   string `json:"x,omitempty"`
	Y   string `json:"y,omitempty"`
}

type JWKSet struct {
	Keys []JWK `json:"keys"`
}

// DiscoveryMetadata is the OIDC discovery document before JSON encoding.
type DiscoveryMetadata struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	UserinfoEndpoint                  string   `json:"userinfo_endpoint"`
	JWKSEndpoint                      string   `json:"jwks_uri"`
	RevocationEndpoint                string   `json:"revocation_endpoint"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
	IDTokenSigningAlgValues           []string `json:"id_token_signing_alg_values_supported"`
	ScopesSupported                   []string `json:"scopes_supported"`
	ClaimsSupported                   []string `json:"claims_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
	TokenEndpointAuthMethods          []string `json:"token_endpoint_auth_methods_supported"`
	ResponseModesSupported            []string `json:"response_modes_supported"`
	AuthorizationResponseISSSupported bool     `json:"authorization_response_iss_parameter_supported"`
	RevocationEndpointAuthMethods     []string `json:"revocation_endpoint_auth_methods_supported"`
	RequestParameterSupported         bool     `json:"request_parameter_supported"`
	RequestURIParameterSupported      bool     `json:"request_uri_parameter_supported"`
	ClaimsParameterSupported          bool     `json:"claims_parameter_supported"`
}

type Client struct {
	ID                      string
	Name                    string
	SecretHash              string
	RedirectURIs            []string
	Scopes                  []string
	GrantTypes              []string
	TokenEndpointAuthMethod ClientAuthenticationMethod
	RequirePKCE             bool
	Public                  bool
	Enabled                 bool
}

// AuthorizationClient is the non-sensitive client view returned to login and
// consent UIs. It intentionally excludes secret hashes and redirect lists.
type AuthorizationClient struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Public bool   `json:"public"`
}

type ClientAuthenticationMethod string

const (
	ClientAuthNone    ClientAuthenticationMethod = "none"
	ClientSecretBasic ClientAuthenticationMethod = "client_secret_basic"
	ClientSecretPost  ClientAuthenticationMethod = "client_secret_post"
)

type User struct {
	Subject           string
	Name              string
	PreferredUsername string
	Email             string
	EmailVerified     bool
	Picture           string
}

type AuthorizeRequest struct {
	Request             string
	RequestURI          string
	IDTokenHint         string
	ClientID            string
	RedirectURI         string
	ResponseType        string
	ResponseMode        string
	Scope               []string
	State               string
	Nonce               string
	CodeChallenge       string
	CodeChallengeMethod string
	Prompt              []string
	MaxAge              *time.Duration
}

// Authentication is trusted state supplied by the host application, never
// parsed from authorization-request parameters.
type Authentication struct {
	UserID   string
	AuthTime time.Time
	// Fresh may only be set by a trusted host after it has verified a new
	// authentication ceremony for this authorization request. It satisfies
	// prompt=login and max_age even when AuthTime has second-level precision.
	Fresh bool
	// AuthID identifies a credential-verified authentication ceremony.
	// Federated sessions without upstream reauthentication proof leave it empty.
	AuthID string
}

type AuthorizationAction string

const (
	AuthorizationNeedLogin   AuthorizationAction = "login"
	AuthorizationNeedConsent AuthorizationAction = "consent"
	AuthorizationReady       AuthorizationAction = "ready"
)

type AuthorizationDecision struct {
	Action AuthorizationAction
	Client AuthorizationClient
	Scopes []string
}

type AuthorizeResult struct {
	Code         string
	RedirectURI  string
	State        string
	Issuer       string
	ResponseMode string
}

type TokenRequest struct {
	GrantType    string
	Code         string
	RefreshToken string
	ClientID     string
	ClientSecret string
	AuthMethod   ClientAuthenticationMethod
	RedirectURI  string
	CodeVerifier string
	Scope        []string
}

type TokenRevocationRequest struct {
	Token         string
	ClientID      string
	ClientSecret  string
	AuthMethod    ClientAuthenticationMethod
	TokenTypeHint string
}

type TokenResponse struct {
	AccessToken  string
	TokenType    string
	ExpiresIn    int64
	IDToken      string
	RefreshToken string
	Scope        []string
}

// UserInfoClaims is the typed result for an OIDC UserInfo response.
type UserInfoClaims struct {
	Subject           string `json:"sub"`
	Name              string `json:"name,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	Email             string `json:"email,omitempty"`
	EmailVerified     *bool  `json:"email_verified,omitempty"`
	Picture           string `json:"picture,omitempty"`
}

// IDTokenClaims contains the claims emitted by the built-in code flow. The
// concrete type prevents adapters and signers from silently omitting required
// OpenID Connect claims.
type IDTokenClaims struct {
	Issuer            string `json:"iss"`
	Subject           string `json:"sub"`
	Audience          string `json:"aud"`
	ExpiresAt         int64  `json:"exp"`
	IssuedAt          int64  `json:"iat"`
	AuthTime          int64  `json:"auth_time,omitempty"`
	Nonce             string `json:"nonce,omitempty"`
	Name              string `json:"name,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	Email             string `json:"email,omitempty"`
	EmailVerified     *bool  `json:"email_verified,omitempty"`
	Picture           string `json:"picture,omitempty"`
}

func (c IDTokenClaims) Validate() error {
	if c.Issuer == "" || c.Subject == "" || c.Audience == "" {
		return errors.New("issuer, subject, and audience are required")
	}
	if c.IssuedAt <= 0 || c.ExpiresAt <= c.IssuedAt {
		return errors.New("ID token timestamps are invalid")
	}
	if c.AuthTime < 0 || c.AuthTime > c.IssuedAt {
		return errors.New("ID token authentication time is invalid")
	}
	return nil
}

type Consent struct {
	UserID   string
	ClientID string
	Scopes   []string
}

type AuthorizationCode struct {
	Hash                string
	GrantID             string
	UserID              string
	ClientID            string
	RedirectURI         string
	Scopes              []string
	Nonce               string
	CodeChallenge       string
	CodeChallengeMethod string
	ExpiresAt           time.Time
	AuthTime            time.Time
	Used                bool
}

// AuthorizationCodeExchange describes the values that must still match when
// an authorization code is consumed and its tokens are persisted atomically.
type AuthorizationCodeExchange struct {
	CodeHash     string
	ClientID     string
	RedirectURI  string
	CodeVerifier string
	Now          time.Time
	AccessToken  *Token
	RefreshToken *Token
}

// RefreshTokenRotation describes an atomic refresh-token rotation. A Store
// must revoke the complete family if OldHash has already been consumed.
type RefreshTokenRotation struct {
	OldHash      string
	ClientID     string
	Now          time.Time
	AccessToken  *Token
	RefreshToken *Token
}

type PurgeRequest struct {
	Now           time.Time
	RevokedBefore time.Time
}

type PurgeResult struct {
	AuthorizationCodes int
	AccessTokens       int
	RefreshTokens      int
}

type Token struct {
	Hash            string
	GrantID         string
	FamilyID        string
	UserID          string
	ClientID        string
	Scopes          []string
	ExpiresAt       time.Time
	FamilyExpiresAt time.Time
	RevokedAt       *time.Time
	AuthTime        time.Time
}

// Store is the persistence boundary for the provider. ExchangeAuthorizationCode
// and RotateRefreshToken are transaction boundaries: either all changes are
// committed or none are.
type Store interface {
	GetClient(context.Context, string) (*Client, error)
	GetConsent(context.Context, string, string) (*Consent, error)
	SaveConsent(context.Context, *Consent) error
	SaveAuthorizationCode(context.Context, *AuthorizationCode) error
	GetAuthorizationCode(context.Context, string) (*AuthorizationCode, error)
	ExchangeAuthorizationCode(context.Context, AuthorizationCodeExchange) error
	GetAccessToken(context.Context, string) (*Token, error)
	GetRefreshToken(context.Context, string) (*Token, error)
	RotateRefreshToken(context.Context, RefreshTokenRotation) error
	RevokeToken(context.Context, string, string, time.Time) error
	RevokeTokenFamily(context.Context, string, time.Time) error
	RevokeGrant(context.Context, string, string, time.Time) error
	RevokeAuthorizationCodeGrant(context.Context, string, time.Time) error
	Purge(context.Context, PurgeRequest) (PurgeResult, error)
}

type UserResolver interface {
	ResolveUser(context.Context, string) (*User, error)
}

type IDTokenSigner interface {
	SignIDToken(context.Context, IDTokenClaims) (string, error)
}

// PublicKeyProvider is optionally implemented by an IDTokenSigner. It keeps
// JWKS generation independent from a concrete RSA/ECDSA implementation.
type PublicKeyProvider interface {
	PublicKeys(context.Context) ([]JWK, error)
}

type SigningKeyProvider interface {
	IDTokenSigner
	PublicKeyProvider
}

type Config struct {
	Issuer                  string
	AuthorizationCodeTTL    time.Duration
	AccessTokenTTL          time.Duration
	RefreshTokenTTL         time.Duration
	RefreshTokenMaxLifetime time.Duration
	IDTokenTTL              time.Duration
	Store                   Store
	Users                   UserResolver
	Signer                  SigningKeyProvider
	Now                     func() time.Time
	AllowInsecureIssuer     bool
	Random                  io.Reader
	SupportedScopes         []string
}

type Provider struct{ cfg Config }

func New(cfg Config) (*Provider, error) {
	cfg.Issuer = strings.TrimRight(cfg.Issuer, "/")
	issuerURL, issuerErr := url.Parse(cfg.Issuer)
	validIssuer := issuerErr == nil && issuerURL.IsAbs() && issuerURL.Host != "" && issuerURL.User == nil && issuerURL.RawQuery == "" && !issuerURL.ForceQuery && issuerURL.Fragment == "" && !strings.Contains(cfg.Issuer, "#") && (strings.EqualFold(issuerURL.Scheme, "https") || cfg.AllowInsecureIssuer && strings.EqualFold(issuerURL.Scheme, "http"))
	if !validIssuer || cfg.Store == nil || cfg.Users == nil || cfg.Signer == nil {
		return nil, errors.New("issuer, store, users, and signer are required")
	}
	if cfg.AccessTokenTTL <= 0 {
		cfg.AccessTokenTTL = time.Hour
	}
	if cfg.AuthorizationCodeTTL <= 0 {
		cfg.AuthorizationCodeTTL = 5 * time.Minute
	}
	if cfg.RefreshTokenTTL <= 0 {
		cfg.RefreshTokenTTL = 30 * 24 * time.Hour
	}
	if cfg.RefreshTokenMaxLifetime <= 0 {
		cfg.RefreshTokenMaxLifetime = 90 * 24 * time.Hour
	}
	if cfg.IDTokenTTL <= 0 {
		cfg.IDTokenTTL = 10 * time.Minute
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.Random == nil {
		cfg.Random = rand.Reader
	}
	if len(cfg.SupportedScopes) == 0 {
		cfg.SupportedScopes = []string{"openid", "profile", "email", "offline_access"}
	}
	if !validScopeSet(cfg.SupportedScopes) || !contains(cfg.SupportedScopes, "openid") {
		return nil, errors.New("supported scopes are invalid or omit openid")
	}
	cfg.SupportedScopes = cloneStrings(cfg.SupportedScopes)
	return &Provider{cfg: cfg}, nil
}

// Metadata returns the standard discovery document as typed data. The caller
// is responsible for JSON encoding it at the well-known HTTP endpoint.
func (p *Provider) Metadata() DiscoveryMetadata {
	base := p.cfg.Issuer
	return DiscoveryMetadata{
		Issuer: p.cfg.Issuer, AuthorizationEndpoint: base + "/authorize", TokenEndpoint: base + "/token",
		UserinfoEndpoint: base + "/userinfo", JWKSEndpoint: base + "/jwks.json", RevocationEndpoint: base + "/revoke",
		ResponseTypesSupported: []string{"code"}, SubjectTypesSupported: []string{"public"},
		IDTokenSigningAlgValues: []string{"RS256"}, ScopesSupported: cloneStrings(p.cfg.SupportedScopes),
		ClaimsSupported:     []string{"iss", "sub", "aud", "iat", "exp", "auth_time", "nonce", "name", "preferred_username", "picture", "email", "email_verified"},
		GrantTypesSupported: []string{"authorization_code", "refresh_token"}, CodeChallengeMethodsSupported: []string{"S256"},
		TokenEndpointAuthMethods: []string{"client_secret_basic", "client_secret_post", "none"},
		ResponseModesSupported:   []string{"query"}, AuthorizationResponseISSSupported: true,
		RevocationEndpointAuthMethods: []string{"client_secret_basic", "client_secret_post", "none"},
		RequestParameterSupported:     false, RequestURIParameterSupported: false, ClaimsParameterSupported: false,
	}
}

// PublicJWKS returns the current public keys for the JWKS endpoint.
func (p *Provider) PublicJWKS(ctx context.Context) (JWKSet, error) {
	keys, err := p.cfg.Signer.PublicKeys(ctx)
	if err != nil {
		return JWKSet{}, errors.Join(ErrServer, err)
	}
	return JWKSet{Keys: keys}, nil
}

// BeginAuthorization validates an authorization request and tells the host
// application which user interaction is required. It never creates a code.
func (p *Provider) BeginAuthorization(ctx context.Context, req AuthorizeRequest, auth Authentication) (*AuthorizationDecision, error) {
	client, scopes, err := p.validateAuthorizationRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := p.validateIDTokenHint(ctx, req, auth); err != nil {
		return nil, err
	}
	now := p.cfg.Now()
	if auth.AuthTime.After(now) {
		return nil, authorizationError(req, "server_error", "authentication time is in the future", ErrServer)
	}
	decision := &AuthorizationDecision{Action: AuthorizationNeedLogin, Client: AuthorizationClient{ID: client.ID, Name: client.Name, Public: client.Public}, Scopes: scopes}
	if auth.UserID == "" || auth.AuthTime.IsZero() || !auth.Fresh && (contains(req.Prompt, "login") || req.MaxAge != nil && (auth.AuthID == "" || *req.MaxAge == 0 || now.Sub(auth.AuthTime) > *req.MaxAge)) {
		if contains(req.Prompt, "none") {
			return nil, authorizationError(req, "login_required", "prompt=none cannot satisfy login", ErrLoginRequired)
		}
		return decision, nil
	}
	consent, err := p.cfg.Store.GetConsent(ctx, auth.UserID, client.ID)
	if err != nil {
		return nil, authorizationError(req, "server_error", "consent store failed", errors.Join(ErrServer, err))
	}
	decision.Action = AuthorizationNeedConsent
	if consent != nil && allAllowed(scopes, consent.Scopes) && !contains(req.Prompt, "consent") {
		decision.Action = AuthorizationReady
	}
	if decision.Action == AuthorizationNeedConsent && contains(req.Prompt, "none") {
		return nil, authorizationError(req, "consent_required", "prompt=none cannot satisfy consent", ErrConsentRequired)
	}
	return decision, nil
}

func (p *Provider) Authorize(ctx context.Context, req AuthorizeRequest, auth Authentication, approved bool) (*AuthorizeResult, error) {
	client, scopes, err := p.validateAuthorizationRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := p.validateIDTokenHint(ctx, req, auth); err != nil {
		return nil, err
	}
	now := p.cfg.Now()
	if auth.AuthTime.After(now) {
		return nil, authorizationError(req, "server_error", "authentication time is in the future", ErrServer)
	}
	if auth.UserID == "" || auth.AuthTime.IsZero() || !auth.Fresh && (contains(req.Prompt, "login") || req.MaxAge != nil && (auth.AuthID == "" || *req.MaxAge == 0 || now.Sub(auth.AuthTime) > *req.MaxAge)) {
		return nil, protocolError("invalid_request", "an authenticated user is required", ErrInvalidRequest)
	}
	if !approved {
		return nil, authorizationError(req, "access_denied", "resource owner denied the request", ErrAccessDenied)
	}
	consentScopes := cloneStrings(scopes)
	if existing, getErr := p.cfg.Store.GetConsent(ctx, auth.UserID, client.ID); getErr != nil {
		return nil, authorizationError(req, "server_error", "consent store failed", errors.Join(ErrServer, getErr))
	} else if existing != nil {
		consentScopes = unionStrings(existing.Scopes, scopes)
	}
	if err := p.cfg.Store.SaveConsent(ctx, &Consent{UserID: auth.UserID, ClientID: client.ID, Scopes: consentScopes}); err != nil {
		return nil, authorizationError(req, "server_error", "consent store failed", errors.Join(ErrServer, err))
	}
	raw, err := p.randomString(32)
	if err != nil {
		return nil, authorizationError(req, "server_error", "secure random source failed", errors.Join(ErrServer, err))
	}
	codeHash := hash(raw)
	code := &AuthorizationCode{Hash: codeHash, GrantID: codeHash, UserID: auth.UserID, ClientID: client.ID, RedirectURI: req.RedirectURI, Scopes: scopes, Nonce: req.Nonce, CodeChallenge: req.CodeChallenge, CodeChallengeMethod: req.CodeChallengeMethod, ExpiresAt: now.Add(p.cfg.AuthorizationCodeTTL), AuthTime: auth.AuthTime}
	if err := p.cfg.Store.SaveAuthorizationCode(ctx, code); err != nil {
		return nil, authorizationError(req, "server_error", "authorization code store failed", errors.Join(ErrServer, err))
	}
	return &AuthorizeResult{Code: raw, RedirectURI: req.RedirectURI, State: req.State, Issuer: p.cfg.Issuer, ResponseMode: normalizedResponseMode(req.ResponseMode)}, nil
}

func (p *Provider) validateAuthorizationRequest(ctx context.Context, req AuthorizeRequest) (*Client, []string, error) {
	client, err := p.cfg.Store.GetClient(ctx, req.ClientID)
	if err != nil {
		return nil, nil, protocolError("server_error", "client store failed", errors.Join(ErrServer, err))
	}
	if client == nil || !client.Enabled {
		return nil, nil, protocolError("unauthorized_client", "client is unknown or disabled", ErrInvalidClient)
	}
	if err := ValidateClient(*client); err != nil {
		return nil, nil, protocolError("unauthorized_client", "client registration is invalid", errors.Join(ErrInvalidClient, err))
	}
	if !validRedirectURI(req.RedirectURI) || !contains(client.RedirectURIs, req.RedirectURI) {
		return nil, nil, protocolError("invalid_request", "redirect_uri does not match the registered client", ErrInvalidRequest)
	}
	if req.Request != "" {
		return nil, nil, authorizationError(req, "request_not_supported", "request objects are not supported", ErrInvalidRequest)
	}
	if req.RequestURI != "" {
		return nil, nil, authorizationError(req, "request_uri_not_supported", "request URI objects are not supported", ErrInvalidRequest)
	}
	if !validUniqueStrings(req.Prompt, func(value string) bool { return value == "none" || value == "login" || value == "consent" }) && len(req.Prompt) > 0 {
		return nil, nil, authorizationError(req, "invalid_request", "prompt contains an unsupported or duplicate value", ErrInvalidRequest)
	}
	if contains(req.Prompt, "none") && len(req.Prompt) != 1 {
		return nil, nil, authorizationError(req, "invalid_request", "prompt none cannot be combined with other values", ErrInvalidRequest)
	}
	if req.MaxAge != nil && *req.MaxAge < 0 {
		return nil, nil, authorizationError(req, "invalid_request", "max_age must not be negative", ErrInvalidRequest)
	}
	if req.ResponseType != "code" {
		return nil, nil, authorizationError(req, "unsupported_response_type", "only the code response type is supported", ErrUnsupportedResponse)
	}
	if req.ResponseMode != "" && req.ResponseMode != "query" {
		return nil, nil, authorizationError(req, "unsupported_response_mode", "only query response mode is supported", ErrUnsupportedResponseMode)
	}
	if !contains(client.GrantTypes, "authorization_code") {
		return nil, nil, authorizationError(req, "unauthorized_client", "client may not use authorization_code", ErrUnsupportedGrant)
	}
	if !validScopeSet(req.Scope) || !contains(req.Scope, "openid") || !allAllowed(req.Scope, client.Scopes) || !allAllowed(req.Scope, p.cfg.SupportedScopes) {
		return nil, nil, authorizationError(req, "invalid_scope", "scope must contain openid and be allowed for the client", ErrInvalidScope)
	}
	if req.CodeChallenge == "" {
		if req.CodeChallengeMethod != "" || client.Public || client.RequirePKCE {
			return nil, nil, authorizationError(req, "invalid_request", "this client requires an S256 PKCE challenge", ErrInvalidRequest)
		}
	} else if !validPKCEChallenge(req.CodeChallenge) || req.CodeChallengeMethod != "S256" {
		return nil, nil, authorizationError(req, "invalid_request", "PKCE challenge must use S256", ErrInvalidRequest)
	}
	scopes := cloneStrings(req.Scope)
	if contains(scopes, "offline_access") && !contains(req.Prompt, "consent") {
		scopes = removeString(scopes, "offline_access")
	}
	return client, scopes, nil
}

func (p *Provider) ExchangeCode(ctx context.Context, req TokenRequest) (*TokenResponse, error) {
	if req.GrantType != "authorization_code" {
		return nil, protocolError("unsupported_grant_type", "expected authorization_code", ErrUnsupportedGrant)
	}
	client, err := p.authenticateClient(ctx, req.ClientID, req.ClientSecret, req.AuthMethod)
	if err != nil {
		return nil, err
	}
	if !contains(client.GrantTypes, "authorization_code") {
		return nil, protocolError("unauthorized_client", "client may not use authorization_code", ErrUnsupportedGrant)
	}
	codeHash := hash(req.Code)
	code, err := p.cfg.Store.GetAuthorizationCode(ctx, codeHash)
	if err != nil {
		return nil, protocolError("server_error", "authorization code store failed", errors.Join(ErrServer, err))
	}
	if code != nil && (code.Hash != codeHash || code.UserID == "" || code.ClientID == "" || !validScopeSet(code.Scopes) || !contains(code.Scopes, "openid") || code.CodeChallenge == "" && code.CodeChallengeMethod != "" || code.CodeChallenge != "" && (!validPKCEChallenge(code.CodeChallenge) || code.CodeChallengeMethod != "S256")) {
		return nil, protocolError("server_error", "authorization code record is corrupt", ErrServer)
	}
	if code != nil && code.Used {
		if revokeErr := p.cfg.Store.RevokeAuthorizationCodeGrant(ctx, codeHash, p.cfg.Now()); revokeErr != nil {
			return nil, protocolError("server_error", "authorization code replay revocation failed", errors.Join(ErrServer, revokeErr))
		}
		return nil, protocolError("invalid_grant", "authorization code was already used", ErrInvalidGrant)
	}
	if code == nil || !p.cfg.Now().Before(code.ExpiresAt) || code.ClientID != client.ID || code.RedirectURI != req.RedirectURI {
		return nil, protocolError("invalid_grant", "authorization code is invalid, expired, or already used", ErrInvalidGrant)
	}
	if code.CodeChallenge != "" && !verifyPKCE(req.CodeVerifier, code.CodeChallenge) {
		return nil, protocolError("invalid_grant", "PKCE verification failed", ErrInvalidGrant)
	}
	return p.issueAuthorizationCodeTokens(ctx, codeHash, code.UserID, client.ID, code.RedirectURI, code.Scopes, code.Nonce, code.AuthTime, req.CodeVerifier)
}

func (p *Provider) Refresh(ctx context.Context, req TokenRequest) (*TokenResponse, error) {
	if req.GrantType != "refresh_token" {
		return nil, protocolError("unsupported_grant_type", "expected refresh_token", ErrUnsupportedGrant)
	}
	client, err := p.authenticateClient(ctx, req.ClientID, req.ClientSecret, req.AuthMethod)
	if err != nil {
		return nil, err
	}
	if !contains(client.GrantTypes, "refresh_token") {
		return nil, protocolError("unauthorized_client", "client may not use refresh_token", ErrUnsupportedGrant)
	}
	refreshHash := hash(req.RefreshToken)
	old, err := p.cfg.Store.GetRefreshToken(ctx, refreshHash)
	if err != nil {
		return nil, protocolError("server_error", "refresh token store failed", errors.Join(ErrServer, err))
	}
	if old != nil && (old.Hash != refreshHash || old.FamilyID == "" || old.UserID == "" || old.ClientID == "" || old.FamilyExpiresAt.IsZero() || !validScopeSet(old.Scopes) || !contains(old.Scopes, "openid")) {
		return nil, protocolError("server_error", "refresh token record is corrupt", ErrServer)
	}
	now := p.cfg.Now()
	if old == nil || old.ClientID != client.ID || !now.Before(old.ExpiresAt) || !old.FamilyExpiresAt.IsZero() && !now.Before(old.FamilyExpiresAt) {
		return nil, protocolError("invalid_grant", "refresh token is invalid or expired", ErrInvalidGrant)
	}
	if old.RevokedAt != nil {
		if err := p.cfg.Store.RevokeTokenFamily(ctx, old.FamilyID, now); err != nil {
			return nil, protocolError("server_error", "refresh token family revocation failed", errors.Join(ErrServer, err))
		}
		return nil, protocolError("invalid_grant", "refresh token reuse detected", ErrInvalidGrant)
	}
	scopes := cloneStrings(old.Scopes)
	if len(req.Scope) > 0 {
		if !validScopeSet(req.Scope) || !contains(req.Scope, "openid") || !allAllowed(req.Scope, old.Scopes) {
			return nil, protocolError("invalid_scope", "refresh scope may not exceed the original grant", ErrInvalidScope)
		}
		scopes = cloneStrings(req.Scope)
	}
	accessRaw, access, err := p.newAccessToken(old.UserID, old.ClientID, old.GrantID, old.FamilyID, scopes, now)
	if err != nil {
		return nil, err
	}
	newRaw, err := p.randomString(32)
	if err != nil {
		return nil, protocolError("server_error", "secure random source failed", errors.Join(ErrServer, err))
	}
	newExpiresAt := now.Add(p.cfg.RefreshTokenTTL)
	if !old.FamilyExpiresAt.IsZero() && old.FamilyExpiresAt.Before(newExpiresAt) {
		newExpiresAt = old.FamilyExpiresAt
	}
	newToken := &Token{Hash: hash(newRaw), GrantID: old.GrantID, FamilyID: old.FamilyID, UserID: old.UserID, ClientID: old.ClientID, Scopes: cloneStrings(scopes), ExpiresAt: newExpiresAt, FamilyExpiresAt: old.FamilyExpiresAt, AuthTime: old.AuthTime}
	user, err := p.resolveUser(ctx, old.UserID)
	if err != nil {
		if errors.Is(err, ErrUserUnavailable) {
			if revokeErr := p.cfg.Store.RevokeTokenFamily(ctx, old.FamilyID, now); revokeErr != nil {
				return nil, protocolError("server_error", "unavailable user token revocation failed", errors.Join(ErrServer, revokeErr))
			}
			return nil, protocolError("invalid_grant", "refresh token subject is unavailable", ErrInvalidGrant)
		}
		return nil, protocolError("server_error", "user resolution failed", err)
	}
	id, err := p.signIDToken(ctx, user, old.ClientID, "", old.AuthTime, now)
	if err != nil {
		return nil, protocolError("server_error", "ID token signing failed", errors.Join(ErrServer, err))
	}
	if err := p.cfg.Store.RotateRefreshToken(ctx, RefreshTokenRotation{OldHash: old.Hash, ClientID: client.ID, Now: now, AccessToken: access, RefreshToken: newToken}); err != nil {
		if errors.Is(err, ErrInvalidGrant) {
			return nil, protocolError("invalid_grant", "refresh token rotation failed", ErrInvalidGrant)
		}
		return nil, protocolError("server_error", "refresh token store failed", errors.Join(ErrServer, err))
	}
	return &TokenResponse{AccessToken: accessRaw, TokenType: "Bearer", ExpiresIn: int64(p.cfg.AccessTokenTTL.Seconds()), IDToken: id, RefreshToken: newRaw, Scope: cloneStrings(scopes)}, nil
}

func (p *Provider) Revoke(ctx context.Context, req TokenRevocationRequest) error {
	client, err := p.authenticateClient(ctx, req.ClientID, req.ClientSecret, req.AuthMethod)
	if err != nil {
		return err
	}
	if req.Token == "" {
		return protocolError("invalid_request", "token is required", ErrInvalidRequest)
	}
	// RFC 7009 requires success for an unknown token, so Store implementations
	// deliberately do not disclose whether the digest existed.
	if err := p.cfg.Store.RevokeToken(ctx, hash(req.Token), client.ID, p.cfg.Now()); err != nil {
		return protocolError("server_error", "token revocation failed", errors.Join(ErrServer, err))
	}
	return nil
}

// ValidateAccessToken resolves an opaque access token and enforces its
// expiration and revocation state. HTTP adapters can use the returned token
// to authorize UserInfo or resource requests.
func (p *Provider) ValidateAccessToken(ctx context.Context, raw string) (*Token, error) {
	if raw == "" {
		return nil, protocolError("invalid_token", "access token is missing", ErrInvalidGrant)
	}
	token, err := p.cfg.Store.GetAccessToken(ctx, hash(raw))
	if err != nil {
		return nil, protocolError("server_error", "access token store failed", errors.Join(ErrServer, err))
	}
	if token != nil && (token.Hash != hash(raw) || token.UserID == "" || token.ClientID == "" || !validScopeSet(token.Scopes) || !contains(token.Scopes, "openid")) {
		return nil, protocolError("server_error", "access token record is corrupt", ErrServer)
	}
	if token == nil {
		return nil, protocolError("invalid_token", "access token is unknown", ErrInvalidGrant)
	}
	if token.RevokedAt != nil || !p.cfg.Now().Before(token.ExpiresAt) {
		return nil, protocolError("invalid_token", "access token is expired or revoked", ErrTokenRevoked)
	}
	return token, nil
}

// UserInfo validates an opaque access token and resolves claims according to
// the scopes granted to that token.
func (p *Provider) UserInfo(ctx context.Context, rawAccessToken string) (*UserInfoClaims, error) {
	token, err := p.ValidateAccessToken(ctx, rawAccessToken)
	if err != nil {
		return nil, err
	}
	user, err := p.resolveUser(ctx, token.UserID)
	if err != nil {
		if errors.Is(err, ErrUserUnavailable) {
			return nil, protocolError("invalid_token", "access token subject is unavailable", ErrUserUnavailable)
		}
		return nil, protocolError("server_error", "user resolution failed", err)
	}
	claims := &UserInfoClaims{Subject: user.Subject}
	if contains(token.Scopes, "profile") {
		claims.Name, claims.PreferredUsername, claims.Picture = user.Name, user.PreferredUsername, user.Picture
	}
	if contains(token.Scopes, "email") {
		verified := user.EmailVerified
		claims.Email, claims.EmailVerified = user.Email, &verified
	}
	return claims, nil
}

// RevokeRefreshTokenFamily revokes all refresh tokens produced by one
// rotation family, which is used when token reuse is detected.
func (p *Provider) RevokeRefreshTokenFamily(ctx context.Context, familyID string) error {
	if familyID == "" {
		return protocolError("invalid_request", "token family is required", ErrInvalidRequest)
	}
	if err := p.cfg.Store.RevokeTokenFamily(ctx, familyID, p.cfg.Now()); err != nil {
		return protocolError("server_error", "token family revocation failed", errors.Join(ErrServer, err))
	}
	return nil
}

// RevokeGrant removes a user's consent and atomically revokes every access and
// refresh token belonging to that user/client grant.
func (p *Provider) RevokeGrant(ctx context.Context, userID, clientID string) error {
	if userID == "" || clientID == "" {
		return protocolError("invalid_request", "user and client are required", ErrInvalidRequest)
	}
	if err := p.cfg.Store.RevokeGrant(ctx, userID, clientID, p.cfg.Now()); err != nil {
		return protocolError("server_error", "grant revocation failed", errors.Join(ErrServer, err))
	}
	return nil
}

// Purge removes terminal credentials without starting background goroutines.
// The host decides when to call it and how long revoked-token audit records are
// retained.
func (p *Provider) Purge(ctx context.Context, revokedRetention time.Duration) (PurgeResult, error) {
	if revokedRetention < 0 {
		return PurgeResult{}, protocolError("invalid_request", "revoked retention must not be negative", ErrInvalidRequest)
	}
	now := p.cfg.Now()
	result, err := p.cfg.Store.Purge(ctx, PurgeRequest{Now: now, RevokedBefore: now.Add(-revokedRetention)})
	if err != nil {
		return PurgeResult{}, protocolError("server_error", "credential purge failed", errors.Join(ErrServer, err))
	}
	return result, nil
}

func (p *Provider) issueAuthorizationCodeTokens(ctx context.Context, codeHash, userID, clientID, redirectURI string, scopes []string, nonce string, authTime time.Time, verifier string) (*TokenResponse, error) {
	now := p.cfg.Now()
	var refreshRaw string
	var refresh *Token
	var family string
	if contains(scopes, "offline_access") {
		var err error
		refreshRaw, err = p.randomString(32)
		if err != nil {
			return nil, protocolError("server_error", "secure random source failed", errors.Join(ErrServer, err))
		}
		family, err = p.randomString(16)
		if err != nil {
			return nil, protocolError("server_error", "secure random source failed", errors.Join(ErrServer, err))
		}
		familyExpiresAt := now.Add(p.cfg.RefreshTokenMaxLifetime)
		expiresAt := now.Add(p.cfg.RefreshTokenTTL)
		if familyExpiresAt.Before(expiresAt) {
			expiresAt = familyExpiresAt
		}
		refresh = &Token{Hash: hash(refreshRaw), GrantID: codeHash, FamilyID: family, UserID: userID, ClientID: clientID, Scopes: cloneStrings(scopes), ExpiresAt: expiresAt, FamilyExpiresAt: familyExpiresAt, AuthTime: authTime}
	}
	accessRaw, access, err := p.newAccessToken(userID, clientID, codeHash, family, scopes, now)
	if err != nil {
		return nil, err
	}
	user, err := p.resolveUser(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserUnavailable) {
			return nil, protocolError("invalid_grant", "authorization code subject is unavailable", ErrInvalidGrant)
		}
		return nil, protocolError("server_error", "user resolution failed", err)
	}
	id, err := p.signIDToken(ctx, user, clientID, nonce, authTime, now)
	if err != nil {
		return nil, protocolError("server_error", "ID token signing failed", errors.Join(ErrServer, err))
	}
	if err := p.cfg.Store.ExchangeAuthorizationCode(ctx, AuthorizationCodeExchange{CodeHash: codeHash, ClientID: clientID, RedirectURI: redirectURI, CodeVerifier: verifier, Now: now, AccessToken: access, RefreshToken: refresh}); err != nil {
		if errors.Is(err, ErrAuthorizationCodeReplay) {
			return nil, protocolError("invalid_grant", "authorization code replay detected", ErrInvalidGrant)
		}
		if errors.Is(err, ErrInvalidGrant) {
			return nil, protocolError("invalid_grant", "authorization code exchange lost an atomic race", ErrInvalidGrant)
		}
		return nil, protocolError("server_error", "authorization code store failed", errors.Join(ErrServer, err))
	}
	return &TokenResponse{AccessToken: accessRaw, TokenType: "Bearer", ExpiresIn: int64(p.cfg.AccessTokenTTL.Seconds()), IDToken: id, RefreshToken: refreshRaw, Scope: cloneStrings(scopes)}, nil
}

func (p *Provider) newAccessToken(userID, clientID, grantID, familyID string, scopes []string, now time.Time) (string, *Token, error) {
	raw, err := p.randomString(32)
	if err != nil {
		return "", nil, protocolError("server_error", "secure random source failed", errors.Join(ErrServer, err))
	}
	return raw, &Token{Hash: hash(raw), GrantID: grantID, FamilyID: familyID, UserID: userID, ClientID: clientID, Scopes: cloneStrings(scopes), ExpiresAt: now.Add(p.cfg.AccessTokenTTL)}, nil
}

func (p *Provider) signIDToken(ctx context.Context, user *User, clientID string, nonce string, authTime, now time.Time) (string, error) {
	claims := IDTokenClaims{Issuer: p.cfg.Issuer, Subject: user.Subject, Audience: clientID, IssuedAt: now.Unix(), ExpiresAt: now.Add(p.cfg.IDTokenTTL).Unix(), Nonce: nonce}
	if !authTime.IsZero() {
		claims.AuthTime = authTime.Unix()
	}
	// In code flow, profile and email scopes release claims through UserInfo.
	// Do not copy those claims into the authentication token implicitly.
	return p.cfg.Signer.SignIDToken(ctx, claims)
}

func (p *Provider) authenticateClient(ctx context.Context, id, secret string, method ClientAuthenticationMethod) (*Client, error) {
	client, err := p.cfg.Store.GetClient(ctx, id)
	if err != nil {
		return nil, protocolError("server_error", "client store failed", errors.Join(ErrServer, err))
	}
	if client == nil || !client.Enabled {
		return nil, protocolError("invalid_client", "client authentication failed", ErrInvalidClient)
	}
	if err := ValidateClient(*client); err != nil {
		return nil, protocolError("invalid_client", "client registration is invalid", errors.Join(ErrInvalidClient, err))
	}
	expectedMethod := client.TokenEndpointAuthMethod
	if client.Public {
		expectedMethod = ClientAuthNone
	}
	if method != expectedMethod {
		return nil, protocolError("invalid_client", "client authentication method is not allowed", ErrInvalidClient)
	}
	if client.Public {
		if secret != "" {
			return nil, protocolError("invalid_client", "public clients must not send a client secret", ErrInvalidClient)
		}
		return client, nil
	}
	if secret == "" || subtle.ConstantTimeCompare([]byte(hash(secret)), []byte(client.SecretHash)) != 1 {
		return nil, protocolError("invalid_client", "client authentication failed", ErrInvalidClient)
	}
	return client, nil
}

func (p *Provider) resolveUser(ctx context.Context, userID string) (*User, error) {
	user, err := p.cfg.Users.ResolveUser(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserUnavailable) {
			return nil, ErrUserUnavailable
		}
		return nil, errors.Join(ErrServer, err)
	}
	if user == nil || !validSubject(user.Subject) {
		return nil, errors.Join(ErrServer, errors.New("resolved user subject is invalid"))
	}
	return user, nil
}

func validSubject(subject string) bool {
	if len(subject) == 0 || len(subject) > 255 {
		return false
	}
	for _, c := range []byte(subject) {
		if c < 0x20 || c > 0x7e {
			return false
		}
	}
	return true
}

// ValidateClient checks the persistent registration independently of a
// request, allowing administrative code to reject invalid clients up front.
func ValidateClient(client Client) error {
	if client.ID == "" {
		return errors.New("client ID is required")
	}
	if len(client.RedirectURIs) == 0 || !validUniqueStrings(client.RedirectURIs, validRedirectURI) {
		return errors.New("client redirect URIs are invalid or duplicated")
	}
	if !validScopeSet(client.Scopes) || !contains(client.Scopes, "openid") {
		return errors.New("client scopes are invalid or omit openid")
	}
	if len(client.GrantTypes) == 0 || !validUniqueStrings(client.GrantTypes, func(v string) bool { return v == "authorization_code" || v == "refresh_token" }) || !contains(client.GrantTypes, "authorization_code") {
		return errors.New("client grant types are invalid")
	}
	if contains(client.Scopes, "offline_access") && !contains(client.GrantTypes, "refresh_token") {
		return errors.New("offline_access requires refresh_token grant")
	}
	if client.Public {
		if client.TokenEndpointAuthMethod != ClientAuthNone || client.SecretHash != "" {
			return errors.New("public client must use none and omit a secret")
		}
		return nil
	}
	if client.TokenEndpointAuthMethod != ClientSecretBasic && client.TokenEndpointAuthMethod != ClientSecretPost {
		return errors.New("confidential client auth method is invalid")
	}
	digest, err := base64.RawURLEncoding.DecodeString(client.SecretHash)
	if err != nil || len(digest) != sha256.Size || subtle.ConstantTimeCompare([]byte(client.SecretHash), []byte(hash(""))) == 1 {
		return errors.New("client secret hash is invalid")
	}
	return nil
}

func verifyPKCE(verifier, challenge string) bool {
	if !validPKCEValue(verifier) || !validPKCEChallenge(challenge) {
		return false
	}
	sum := sha256.Sum256([]byte(verifier))
	actual := base64.RawURLEncoding.EncodeToString(sum[:])
	return subtle.ConstantTimeCompare([]byte(actual), []byte(challenge)) == 1
}
func validPKCEValue(value string) bool {
	if len(value) < 43 || len(value) > 128 {
		return false
	}
	for _, c := range value {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '.' || c == '_' || c == '~' {
			continue
		}
		return false
	}
	return true
}
func validPKCEChallenge(value string) bool { return len(value) == 43 && validPKCEValue(value) }
func validRedirectURI(value string) bool {
	u, err := url.Parse(value)
	if err != nil || !u.IsAbs() || u.Fragment != "" || strings.Contains(value, "#") || u.User != nil {
		return false
	}
	isHTTP := strings.EqualFold(u.Scheme, "http") || strings.EqualFold(u.Scheme, "https")
	if isHTTP && u.Host == "" {
		return false
	}
	if !isHTTP && (u.Opaque != "" || (u.Host == "" && u.Path == "") || strings.EqualFold(u.Scheme, "javascript") || strings.EqualFold(u.Scheme, "data") || strings.EqualFold(u.Scheme, "file")) {
		return false
	}
	return true
}
func validScopeSet(scopes []string) bool {
	return len(scopes) > 0 && validUniqueStrings(scopes, func(scope string) bool {
		if scope == "" {
			return false
		}
		for _, c := range []byte(scope) {
			if c != 0x21 && (c < 0x23 || c > 0x5b) && (c < 0x5d || c > 0x7e) {
				return false
			}
		}
		return true
	})
}
func validUniqueStrings(values []string, valid func(string) bool) bool {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if !valid(value) {
			return false
		}
		if _, ok := seen[value]; ok {
			return false
		}
		seen[value] = struct{}{}
	}
	return true
}
func hash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// HashClientSecret returns the stable digest stored for a high-entropy client
// secret. Client secrets should be generated with GenerateClientSecret.
func HashClientSecret(secret string) string { return hash(secret) }

// GenerateClientSecret returns a 256-bit, URL-safe client secret.
func GenerateClientSecret() (string, error)            { return randomString(rand.Reader, 32) }
func (p *Provider) randomString(n int) (string, error) { return randomString(p.cfg.Random, n) }
func randomString(reader io.Reader, n int) (string, error) {
	b := make([]byte, n)
	if _, err := io.ReadFull(reader, b); err != nil {
		return "", fmt.Errorf("random token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func contains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}
func allAllowed(values, allowed []string) bool {
	for _, v := range values {
		if !contains(allowed, v) {
			return false
		}
	}
	return true
}
func cloneStrings(values []string) []string { return append([]string(nil), values...) }
func unionStrings(a, b []string) []string {
	result := cloneStrings(a)
	for _, value := range b {
		if !contains(result, value) {
			result = append(result, value)
		}
	}
	return result
}
func removeString(values []string, target string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value != target {
			result = append(result, value)
		}
	}
	return result
}
func normalizedResponseMode(value string) string {
	if value == "" {
		return "query"
	}
	return value
}
func safeResponseMode(value string) string {
	if value == "query" {
		return value
	}
	return "query"
}
