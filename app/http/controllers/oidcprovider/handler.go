// Package oidcprovider adapts the transport-neutral OIDC provider core to Gin.
package oidcprovider

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	core "github.com/leancodebox/GooseForum/app/bundles/oidcprovider"
	"github.com/leancodebox/GooseForum/app/models/forum/oidcProviderStore"
)

const (
	maxFormBytes   = 64 << 10
	maxJSONBytes   = 16 << 10
	interactionTTL = 10 * time.Minute
)

// AuthenticationResolver supplies trusted GooseForum login state to the
// authorization endpoint. AuthTime must be the actual authentication time;
// resolvers that cannot prove it must return a zero Authentication value.
type AuthenticationResolver func(*gin.Context) core.Authentication

type ProviderGetter func() (*core.Provider, error)

type Handler struct {
	provider     ProviderGetter
	authenticate AuthenticationResolver
	loginPath    string
	interactions interactionStore
	now          func() time.Time
	random       io.Reader
}

type interactionStore interface {
	CreateInteraction(context.Context, *oidcProviderStore.InteractionEntity) error
	GetInteraction(context.Context, string, string, time.Time) (*oidcProviderStore.InteractionEntity, error)
	ConsumeInteraction(context.Context, string, string, time.Time) (*oidcProviderStore.InteractionEntity, error)
	ConsumeLoginInteraction(context.Context, string, string, time.Time, time.Time) (*oidcProviderStore.InteractionEntity, error)
}

func New(provider ProviderGetter, authenticate AuthenticationResolver) (*Handler, error) {
	if provider == nil {
		return nil, errors.New("OIDC provider getter is required")
	}
	return &Handler{provider: provider, authenticate: authenticate, loginPath: "/login", now: time.Now, random: rand.Reader}, nil
}

func (h *Handler) WithInteractions(store interactionStore, now func() time.Time, random io.Reader) *Handler {
	h.interactions = store
	if now != nil {
		h.now = now
	}
	if random != nil {
		h.random = random
	}
	return h
}

// WithLoginPath changes the local page used to resume an unauthenticated
// authorization request. The path must remain local to GooseForum.
func (h *Handler) WithLoginPath(path string) *Handler {
	if strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "//") {
		h.loginPath = path
	}
	return h
}

func (h *Handler) Discovery(c *gin.Context) {
	setPublicJSON(c)
	provider, ok := h.getProvider(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, provider.Metadata())
}

func (h *Handler) JWKS(c *gin.Context) {
	setPublicJSON(c)
	provider, ok := h.getProvider(c)
	if !ok {
		return
	}
	keys, err := provider.PublicJWKS(c.Request.Context())
	if err != nil {
		writeOAuthError(c, err, false)
		return
	}
	c.JSON(http.StatusOK, keys)
}

func (h *Handler) Token(c *gin.Context) {
	setNoStore(c)
	provider, ok := h.getProvider(c)
	if !ok {
		return
	}
	form, ok := parseForm(c)
	if !ok {
		return
	}
	grantType, ok := requiredSingle(c, form, "grant_type")
	if !ok {
		return
	}
	auth, ok := parseClientAuthentication(c, form)
	if !ok {
		return
	}

	req := core.TokenRequest{GrantType: grantType, ClientID: auth.id, ClientSecret: auth.secret, AuthMethod: auth.method}
	var result *core.TokenResponse
	var err error
	switch grantType {
	case "authorization_code":
		if !rejectDuplicates(c, form, "code", "redirect_uri", "code_verifier", "scope", "refresh_token") {
			return
		}
		req.Code = form.Get("code")
		req.RedirectURI = form.Get("redirect_uri")
		req.CodeVerifier = form.Get("code_verifier")
		result, err = provider.ExchangeCode(c.Request.Context(), req)
	case "refresh_token":
		if !rejectDuplicates(c, form, "refresh_token", "scope", "code", "redirect_uri", "code_verifier") {
			return
		}
		req.RefreshToken = form.Get("refresh_token")
		req.Scope = strings.Fields(form.Get("scope"))
		result, err = provider.Refresh(c.Request.Context(), req)
	default:
		err = core.NewOAuthError("unsupported_grant_type", "grant_type is not supported")
	}
	if err != nil {
		writeOAuthError(c, err, true)
		return
	}
	c.JSON(http.StatusOK, tokenJSON{
		AccessToken: result.AccessToken, TokenType: result.TokenType, ExpiresIn: result.ExpiresIn,
		RefreshToken: result.RefreshToken, IDToken: result.IDToken, Scope: strings.Join(result.Scope, " "),
	})
}

func (h *Handler) UserInfo(c *gin.Context) {
	setNoStore(c)
	provider, ok := h.getProvider(c)
	if !ok {
		return
	}
	raw, ok := bearerToken(c)
	if !ok {
		return
	}
	claims, err := provider.UserInfo(c.Request.Context(), raw)
	if err != nil {
		writeUserInfoError(c, err)
		return
	}
	c.JSON(http.StatusOK, claims)
}

func (h *Handler) Revoke(c *gin.Context) {
	setNoStore(c)
	provider, ok := h.getProvider(c)
	if !ok {
		return
	}
	form, ok := parseForm(c)
	if !ok {
		return
	}
	if !rejectDuplicates(c, form, "token", "token_type_hint") {
		return
	}
	auth, ok := parseClientAuthentication(c, form)
	if !ok {
		return
	}
	err := provider.Revoke(c.Request.Context(), core.TokenRevocationRequest{
		Token: form.Get("token"), TokenTypeHint: form.Get("token_type_hint"),
		ClientID: auth.id, ClientSecret: auth.secret, AuthMethod: auth.method,
	})
	if err != nil {
		writeOAuthError(c, err, true)
		return
	}
	c.Status(http.StatusOK)
}

func (h *Handler) Authorize(c *gin.Context) {
	setNoStore(c)
	provider, ok := h.getProvider(c)
	if !ok {
		return
	}
	req, ok := parseAuthorizeRequest(c)
	if !ok {
		return
	}
	auth := core.Authentication{}
	if h.authenticate != nil {
		auth = h.authenticate(c)
	}
	decision, err := provider.BeginAuthorization(c.Request.Context(), req, auth)
	if err != nil {
		writeAuthorizationError(c, err, provider.Metadata().Issuer)
		return
	}
	switch decision.Action {
	case core.AuthorizationNeedLogin:
		if auth.UserID != "" || len(req.Prompt) > 0 || req.MaxAge != nil {
			interaction, createErr := h.createInteraction(c, req, auth, oidcProviderStore.InteractionPurposeLogin)
			if createErr != nil {
				writeOAuthError(c, core.NewOAuthError("server_error", "reauthentication interaction could not be created"), false)
				return
			}
			resume := "/oauth2/authorize/resume?interaction=" + url.QueryEscape(interaction)
			c.Redirect(http.StatusFound, h.loginPath+"?force=true&redirect="+url.QueryEscape(resume))
			return
		}
		resume := authorizeResumeURI(req)
		c.Redirect(http.StatusFound, h.loginPath+"?redirect="+url.QueryEscape(resume))
	case core.AuthorizationNeedConsent:
		interaction, createErr := h.createInteraction(c, req, auth, oidcProviderStore.InteractionPurposeConsent)
		if createErr != nil {
			writeOAuthError(c, core.NewOAuthError("server_error", "authorization interaction could not be created"), false)
			return
		}
		c.Redirect(http.StatusFound, "/oauth2/consent?interaction="+url.QueryEscape(interaction))
	case core.AuthorizationReady:
		result, authorizeErr := provider.Authorize(c.Request.Context(), req, auth, true)
		if authorizeErr != nil {
			writeAuthorizationError(c, authorizeErr, provider.Metadata().Issuer)
			return
		}
		location, buildErr := authorizationRedirect(result)
		if buildErr != nil {
			writeOAuthError(c, core.NewOAuthError("server_error", "authorization response could not be created"), false)
			return
		}
		c.Redirect(http.StatusFound, location)
	default:
		writeOAuthError(c, core.NewOAuthError("server_error", "invalid authorization decision"), false)
	}
}

// ResumeAuthorization completes a trusted reauthentication interaction. The
// browser supplies only an opaque handle; the authorization request is loaded
// from the database and the handle is consumed exactly once.
func (h *Handler) ResumeAuthorization(c *gin.Context) {
	setNoStore(c)
	c.Header("Referrer-Policy", "no-referrer")
	auth, ok := h.currentAuthentication(c)
	if !ok {
		return
	}
	if auth.AuthID == "" {
		c.JSON(http.StatusUnauthorized, oauthErrorJSON{Error: "login_required", ErrorDescription: "a new authenticated session is required"})
		return
	}
	hash, valid := interactionHash(c.Query("interaction"))
	if !valid || h.interactions == nil {
		writeConsentInteractionError(c)
		return
	}
	entity, err := h.interactions.ConsumeLoginInteraction(c.Request.Context(), hash, auth.AuthID, auth.AuthTime, h.now())
	if err != nil {
		writeConsentStoreError(c, err)
		return
	}
	req, err := decodeInteractionRequest(entity.RequestJSON)
	if err != nil {
		writeOAuthError(c, core.NewOAuthError("server_error", "reauthentication interaction is corrupt"), false)
		return
	}
	auth.Fresh = true
	provider, ok := h.getProvider(c)
	if !ok {
		return
	}
	decision, err := provider.BeginAuthorization(c.Request.Context(), req, auth)
	if err != nil {
		writeAuthorizationError(c, err, provider.Metadata().Issuer)
		return
	}
	switch decision.Action {
	case core.AuthorizationNeedConsent:
		interaction, createErr := h.createInteraction(c, req, auth, oidcProviderStore.InteractionPurposeConsent)
		if createErr != nil {
			writeOAuthError(c, core.NewOAuthError("server_error", "authorization interaction could not be created"), false)
			return
		}
		c.Redirect(http.StatusFound, "/oauth2/consent?interaction="+url.QueryEscape(interaction))
	case core.AuthorizationReady:
		result, authorizeErr := provider.Authorize(c.Request.Context(), req, auth, true)
		if authorizeErr != nil {
			writeAuthorizationError(c, authorizeErr, provider.Metadata().Issuer)
			return
		}
		location, buildErr := authorizationRedirect(result)
		if buildErr != nil {
			writeOAuthError(c, core.NewOAuthError("server_error", "authorization response could not be created"), false)
			return
		}
		c.Redirect(http.StatusFound, location)
	default:
		writeOAuthError(c, core.NewOAuthError("invalid_request", "fresh authentication is required"), false)
	}
}

type consentDetailsJSON struct {
	Client    core.AuthorizationClient `json:"client"`
	Scopes    []string                 `json:"scopes"`
	ExpiresAt time.Time                `json:"expires_at"`
}

type consentDecisionJSON struct {
	Interaction string `json:"interaction"`
	Decision    string `json:"decision"`
}

type consentResultJSON struct {
	RedirectURL string `json:"redirect_url"`
}

// ConsentDetails returns display-only data derived from the persisted trusted
// authorization request. It never returns redirect_uri, state, nonce, or PKCE.
func (h *Handler) ConsentDetails(c *gin.Context) {
	setNoStore(c)
	c.Header("Referrer-Policy", "no-referrer")
	auth, ok := h.currentAuthentication(c)
	if !ok {
		return
	}
	hash, ok := interactionHash(c.Query("interaction"))
	if !ok || h.interactions == nil {
		writeConsentInteractionError(c)
		return
	}
	entity, err := h.interactions.GetInteraction(c.Request.Context(), hash, auth.UserID, h.now())
	if err != nil {
		writeConsentStoreError(c, err)
		return
	}
	req, err := decodeInteractionRequest(entity.RequestJSON)
	if err != nil {
		writeOAuthError(c, core.NewOAuthError("server_error", "authorization interaction is corrupt"), false)
		return
	}
	provider, ok := h.getProvider(c)
	if !ok {
		return
	}
	decision, err := provider.BeginAuthorization(c.Request.Context(), req, storedAuthentication(entity))
	if err != nil {
		writeAuthorizationError(c, err, provider.Metadata().Issuer)
		return
	}
	c.JSON(http.StatusOK, consentDetailsJSON{Client: decision.Client, Scopes: decision.Scopes, ExpiresAt: entity.ExpiresAt})
}

// ConsentDecision atomically consumes an interaction before acting on its
// persisted request. The browser cannot replace any authorization parameter.
func (h *Handler) ConsentDecision(c *gin.Context) {
	setNoStore(c)
	c.Header("Referrer-Policy", "no-referrer")
	if !sameOrigin(c.Request) {
		writeOAuthError(c, core.NewOAuthError("invalid_request", "request origin is not allowed"), false)
		return
	}
	auth, ok := h.currentAuthentication(c)
	if !ok {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxJSONBytes)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	var input consentDecisionJSON
	decodeErr := decoder.Decode(&input)
	var trailing any
	trailingErr := decoder.Decode(&trailing)
	if c.ContentType() != "application/json" || decodeErr != nil || !errors.Is(trailingErr, io.EOF) || (input.Decision != "approve" && input.Decision != "deny") {
		writeOAuthError(c, core.NewOAuthError("invalid_request", "a JSON interaction and approve or deny decision are required"), false)
		return
	}
	hash, valid := interactionHash(input.Interaction)
	if !valid || h.interactions == nil {
		writeConsentInteractionError(c)
		return
	}
	entity, err := h.interactions.ConsumeInteraction(c.Request.Context(), hash, auth.UserID, h.now())
	if err != nil {
		writeConsentStoreError(c, err)
		return
	}
	req, err := decodeInteractionRequest(entity.RequestJSON)
	if err != nil {
		writeOAuthError(c, core.NewOAuthError("server_error", "authorization interaction is corrupt"), false)
		return
	}
	provider, ok := h.getProvider(c)
	if !ok {
		return
	}
	result, err := provider.Authorize(c.Request.Context(), req, storedAuthentication(entity), input.Decision == "approve")
	if err != nil {
		if input.Decision == "deny" {
			if location, redirectErr := authorizationErrorRedirect(err, provider.Metadata().Issuer); redirectErr == nil {
				c.JSON(http.StatusOK, consentResultJSON{RedirectURL: location})
				return
			}
		}
		writeAuthorizationError(c, err, provider.Metadata().Issuer)
		return
	}
	location, err := authorizationRedirect(result)
	if err != nil {
		writeOAuthError(c, core.NewOAuthError("server_error", "authorization response could not be created"), false)
		return
	}
	c.JSON(http.StatusOK, consentResultJSON{RedirectURL: location})
}

func sameOrigin(request *http.Request) bool {
	origin := request.Header.Get("Origin")
	if origin == "" {
		return false
	}
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil || parsed.Path != "" {
		return false
	}
	scheme := "http"
	if request.TLS != nil {
		scheme = "https"
	}
	if forwarded := request.Header.Get("X-Forwarded-Proto"); forwarded == "https" || forwarded == "http" {
		scheme = forwarded
	}
	return strings.EqualFold(parsed.Scheme, scheme) && strings.EqualFold(parsed.Host, request.Host)
}

func (h *Handler) createInteraction(c *gin.Context, req core.AuthorizeRequest, auth core.Authentication, purpose string) (string, error) {
	if h.interactions == nil {
		return "", errors.New("interaction store is unavailable")
	}
	rawBytes := make([]byte, 32)
	if _, err := io.ReadFull(h.random, rawBytes); err != nil {
		return "", err
	}
	raw := base64.RawURLEncoding.EncodeToString(rawBytes)
	clear(rawBytes)
	payload, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	now := h.now()
	return raw, h.interactions.CreateInteraction(c.Request.Context(), &oidcProviderStore.InteractionEntity{
		Hash: hashInteraction(raw), RequestJSON: payload, Purpose: purpose, UserID: auth.UserID,
		AuthTime: auth.AuthTime, AuthFresh: auth.Fresh, PreviousAuthID: auth.AuthID,
		StartedAt: now.Truncate(time.Second), ExpiresAt: now.Add(interactionTTL),
	})
}

func (h *Handler) currentAuthentication(c *gin.Context) (core.Authentication, bool) {
	auth := core.Authentication{}
	if h.authenticate != nil {
		auth = h.authenticate(c)
	}
	if auth.UserID == "" || auth.AuthTime.IsZero() {
		c.JSON(http.StatusUnauthorized, oauthErrorJSON{Error: "login_required", ErrorDescription: "an authenticated user is required"})
		return core.Authentication{}, false
	}
	return auth, true
}

func storedAuthentication(entity *oidcProviderStore.InteractionEntity) core.Authentication {
	return core.Authentication{UserID: entity.UserID, AuthTime: entity.AuthTime, Fresh: entity.AuthFresh, AuthID: entity.PreviousAuthID}
}

func decodeInteractionRequest(payload []byte) (core.AuthorizeRequest, error) {
	var req core.AuthorizeRequest
	err := json.Unmarshal(payload, &req)
	return req, err
}

func interactionHash(raw string) (string, bool) {
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil || len(decoded) != 32 {
		return "", false
	}
	clear(decoded)
	return hashInteraction(raw), true
}

func hashInteraction(raw string) string {
	digest := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(digest[:])
}

func writeConsentStoreError(c *gin.Context, err error) {
	if errors.Is(err, oidcProviderStore.ErrInteractionInvalid) {
		writeConsentInteractionError(c)
		return
	}
	writeOAuthError(c, core.NewOAuthError("server_error", "authorization interaction store failed"), false)
}

func writeConsentInteractionError(c *gin.Context) {
	c.JSON(http.StatusBadRequest, oauthErrorJSON{Error: "invalid_request", ErrorDescription: "authorization interaction is invalid or expired"})
}

func (h *Handler) getProvider(c *gin.Context) (*core.Provider, bool) {
	provider, err := h.provider()
	if err != nil || provider == nil {
		setNoStore(c)
		c.JSON(http.StatusServiceUnavailable, oauthErrorJSON{
			Error:            "temporarily_unavailable",
			ErrorDescription: "OIDC provider is unavailable",
		})
		return nil, false
	}
	return provider, true
}

type clientAuthentication struct {
	id, secret string
	method     core.ClientAuthenticationMethod
}

type tokenJSON struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token"`
	Scope        string `json:"scope,omitempty"`
}

type oauthErrorJSON struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorURI         string `json:"error_uri,omitempty"`
}

func parseForm(c *gin.Context) (url.Values, bool) {
	mediaType, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
	if err != nil || mediaType != "application/x-www-form-urlencoded" {
		writeOAuthError(c, core.NewOAuthError("invalid_request", "Content-Type must be application/x-www-form-urlencoded"), false)
		return nil, false
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxFormBytes)
	if err := c.Request.ParseForm(); err != nil {
		writeOAuthError(c, core.NewOAuthError("invalid_request", "form body is invalid or too large"), false)
		return nil, false
	}
	return c.Request.PostForm, true
}

func requiredSingle(c *gin.Context, form url.Values, name string) (string, bool) {
	values := form[name]
	if len(values) != 1 || values[0] == "" {
		writeOAuthError(c, core.NewOAuthError("invalid_request", name+" must occur exactly once"), false)
		return "", false
	}
	return values[0], true
}

func rejectDuplicates(c *gin.Context, form url.Values, names ...string) bool {
	for _, name := range names {
		if len(form[name]) > 1 {
			writeOAuthError(c, core.NewOAuthError("invalid_request", name+" must not be repeated"), false)
			return false
		}
	}
	return true
}

func parseClientAuthentication(c *gin.Context, form url.Values) (clientAuthentication, bool) {
	if !rejectDuplicates(c, form, "client_id", "client_secret") {
		return clientAuthentication{}, false
	}
	authorization := c.GetHeader("Authorization")
	if authorization != "" {
		if form.Get("client_id") != "" || form.Get("client_secret") != "" {
			writeOAuthError(c, core.NewOAuthError("invalid_request", "multiple client authentication methods are not allowed"), false)
			return clientAuthentication{}, false
		}
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			writeInvalidClient(c, "HTTP Basic client credentials are invalid")
			return clientAuthentication{}, false
		}
		id, idErr := url.QueryUnescape(username)
		secret, secretErr := url.QueryUnescape(password)
		if idErr != nil || secretErr != nil || id == "" {
			writeInvalidClient(c, "HTTP Basic client credentials are invalid")
			return clientAuthentication{}, false
		}
		return clientAuthentication{id: id, secret: secret, method: core.ClientSecretBasic}, true
	}
	id := form.Get("client_id")
	if id == "" {
		writeInvalidClient(c, "client_id is required")
		return clientAuthentication{}, false
	}
	if secret := form.Get("client_secret"); secret != "" {
		return clientAuthentication{id: id, secret: secret, method: core.ClientSecretPost}, true
	}
	return clientAuthentication{id: id, method: core.ClientAuthNone}, true
}

func bearerToken(c *gin.Context) (string, bool) {
	header := c.GetHeader("Authorization")
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		c.Header("WWW-Authenticate", `Bearer error="invalid_token"`)
		c.JSON(http.StatusUnauthorized, oauthErrorJSON{Error: "invalid_token", ErrorDescription: "a single Bearer access token is required"})
		return "", false
	}
	return parts[1], true
}

func parseAuthorizeRequest(c *gin.Context) (core.AuthorizeRequest, bool) {
	query := c.Request.URL.Query()
	if c.Request.Method == http.MethodPost {
		form, ok := parseForm(c)
		if !ok {
			return core.AuthorizeRequest{}, false
		}
		for name := range form {
			if len(query[name]) != 0 {
				writeOAuthError(c, core.NewOAuthError("invalid_request", name+" must not occur in both query and form"), false)
				return core.AuthorizeRequest{}, false
			}
		}
		query = form
	}
	for _, name := range []string{"client_id", "redirect_uri", "response_type", "response_mode", "scope", "state", "nonce", "code_challenge", "code_challenge_method", "prompt", "max_age", "request", "request_uri", "id_token_hint"} {
		if len(query[name]) > 1 {
			writeOAuthError(c, core.NewOAuthError("invalid_request", name+" must not be repeated"), false)
			return core.AuthorizeRequest{}, false
		}
	}
	req := core.AuthorizeRequest{
		Request: query.Get("request"), RequestURI: query.Get("request_uri"), IDTokenHint: query.Get("id_token_hint"),
		ClientID: query.Get("client_id"), RedirectURI: query.Get("redirect_uri"), ResponseType: query.Get("response_type"),
		ResponseMode: query.Get("response_mode"), Scope: strings.Fields(query.Get("scope")), State: query.Get("state"),
		Nonce: query.Get("nonce"), CodeChallenge: query.Get("code_challenge"), CodeChallengeMethod: query.Get("code_challenge_method"),
		Prompt: strings.Fields(query.Get("prompt")),
	}
	if raw := query.Get("max_age"); raw != "" {
		seconds, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || seconds < 0 {
			writeOAuthError(c, core.NewOAuthError("invalid_request", "max_age must be a non-negative integer"), false)
			return core.AuthorizeRequest{}, false
		}
		if seconds > int64((time.Duration(1<<63-1))/time.Second) {
			writeOAuthError(c, core.NewOAuthError("invalid_request", "max_age is too large"), false)
			return core.AuthorizeRequest{}, false
		}
		duration := time.Duration(seconds) * time.Second
		req.MaxAge = &duration
	}
	return req, true
}

func authorizeResumeURI(req core.AuthorizeRequest) string {
	values := url.Values{"client_id": {req.ClientID}, "redirect_uri": {req.RedirectURI}, "response_type": {req.ResponseType}, "scope": {strings.Join(req.Scope, " ")}}
	optional := map[string]string{"id_token_hint": req.IDTokenHint, "response_mode": req.ResponseMode, "state": req.State, "nonce": req.Nonce, "code_challenge": req.CodeChallenge, "code_challenge_method": req.CodeChallengeMethod, "prompt": strings.Join(req.Prompt, " ")}
	for name, value := range optional {
		if value != "" {
			values.Set(name, value)
		}
	}
	if req.MaxAge != nil {
		values.Set("max_age", strconv.FormatInt(int64(*req.MaxAge/time.Second), 10))
	}
	return "/oauth2/authorize?" + values.Encode()
}

func authorizationRedirect(result *core.AuthorizeResult) (string, error) {
	redirect, err := url.Parse(result.RedirectURI)
	if err != nil {
		return "", err
	}
	query := redirect.Query()
	query.Set("code", result.Code)
	query.Set("iss", result.Issuer)
	if result.State != "" {
		query.Set("state", result.State)
	}
	redirect.RawQuery = query.Encode()
	return redirect.String(), nil
}

func writeAuthorizationError(c *gin.Context, err error, issuer string) {
	if location, redirectErr := authorizationErrorRedirect(err, issuer); redirectErr == nil {
		c.Redirect(http.StatusFound, location)
		return
	}
	writeOAuthError(c, err, false)
}

func authorizationErrorRedirect(err error, issuer string) (string, error) {
	var oauthErr *core.OAuthError
	if !errors.As(err, &oauthErr) || !oauthErr.RedirectAllowed || oauthErr.RedirectURI == "" {
		return "", errors.New("authorization error is not redirectable")
	}
	redirect, err := url.Parse(oauthErr.RedirectURI)
	if err != nil {
		return "", err
	}
	query := redirect.Query()
	query.Set("error", oauthErr.Code)
	if issuer != "" {
		query.Set("iss", issuer)
	}
	if oauthErr.Description != "" {
		query.Set("error_description", oauthErr.Description)
	}
	if oauthErr.URI != "" {
		query.Set("error_uri", oauthErr.URI)
	}
	if oauthErr.State != "" {
		query.Set("state", oauthErr.State)
	}
	redirect.RawQuery = query.Encode()
	return redirect.String(), nil
}

func writeUserInfoError(c *gin.Context, err error) {
	var oauthErr *core.OAuthError
	if errors.As(err, &oauthErr) && oauthErr.Code == "invalid_token" {
		c.Header("WWW-Authenticate", fmt.Sprintf(`Bearer error="invalid_token", error_description=%q`, oauthErr.Description))
		c.JSON(http.StatusUnauthorized, oauthErrorJSON{Error: oauthErr.Code, ErrorDescription: oauthErr.Description, ErrorURI: oauthErr.URI})
		return
	}
	writeOAuthError(c, err, false)
}

func writeInvalidClient(c *gin.Context, description string) {
	c.Header("WWW-Authenticate", `Basic realm="oauth2/token"`)
	c.JSON(http.StatusUnauthorized, oauthErrorJSON{Error: "invalid_client", ErrorDescription: description})
}

func writeOAuthError(c *gin.Context, err error, tokenEndpoint bool) {
	setNoStore(c)
	oauthErr := &core.OAuthError{Code: "server_error", Description: "the authorization server encountered an unexpected condition"}
	if !errors.As(err, &oauthErr) {
		oauthErr = core.NewOAuthError("server_error", "the authorization server encountered an unexpected condition")
	}
	status := http.StatusBadRequest
	if oauthErr.Code == "server_error" {
		status = http.StatusInternalServerError
	} else if tokenEndpoint && oauthErr.Code == "invalid_client" {
		status = http.StatusUnauthorized
		c.Header("WWW-Authenticate", `Basic realm="oauth2/token"`)
	}
	c.JSON(status, oauthErrorJSON{Error: oauthErr.Code, ErrorDescription: oauthErr.Description, ErrorURI: oauthErr.URI})
}

func setNoStore(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
}

func setPublicJSON(c *gin.Context) {
	c.Header("Cache-Control", "public, max-age=300")
}
