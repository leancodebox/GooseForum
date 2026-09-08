package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	core "github.com/leancodebox/GooseForum/app/bundles/oidcprovider"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/oidcProviderStore"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/service/oidcproviderservice"
)

const oidcClientIDPrefix = "gf_"

var managedOIDCScopes = map[string]struct{}{
	"openid": {}, "profile": {}, "email": {}, "offline_access": {},
}

type oidcClientStore interface {
	ListClients(context.Context) ([]core.Client, error)
	GetClient(context.Context, string) (*core.Client, error)
	CreateClient(context.Context, *core.Client) error
	SaveClient(context.Context, *core.Client) error
}

type OIDCClientView struct {
	ClientID                string                          `json:"clientId"`
	Name                    string                          `json:"name"`
	RedirectURIs            []string                        `json:"redirectUris"`
	Scopes                  []string                        `json:"scopes"`
	GrantTypes              []string                        `json:"grantTypes"`
	TokenEndpointAuthMethod core.ClientAuthenticationMethod `json:"tokenEndpointAuthMethod"`
	RequirePKCE             bool                            `json:"requirePkce"`
	Public                  bool                            `json:"public"`
	Enabled                 bool                            `json:"enabled"`
}

type OIDCClientCredentials struct {
	Client       OIDCClientView `json:"client"`
	ClientSecret string         `json:"clientSecret,omitempty"`
}

type CreateOIDCClientReq struct {
	Name                    string   `json:"name"`
	RedirectURIs            []string `json:"redirectUris"`
	Scopes                  []string `json:"scopes"`
	GrantTypes              []string `json:"grantTypes"`
	TokenEndpointAuthMethod string   `json:"tokenEndpointAuthMethod"`
	RequirePKCE             bool     `json:"requirePkce"`
	Public                  bool     `json:"public"`
	Enabled                 *bool    `json:"enabled"`
}

type UpdateOIDCClientReq struct {
	ClientID                string   `json:"clientId"`
	Name                    string   `json:"name"`
	RedirectURIs            []string `json:"redirectUris"`
	Scopes                  []string `json:"scopes"`
	GrantTypes              []string `json:"grantTypes"`
	TokenEndpointAuthMethod string   `json:"tokenEndpointAuthMethod"`
	RequirePKCE             bool     `json:"requirePkce"`
	Public                  *bool    `json:"public"`
	Enabled                 *bool    `json:"enabled"`
}

type RotateOIDCClientSecretReq struct {
	ClientID string `json:"clientId"`
}

type SaveOIDCProviderSettingsReq struct {
	Enabled bool `json:"enabled"`
}

type OIDCGrantView struct {
	ClientID  string    `json:"clientId"`
	Name      string    `json:"name"`
	Scopes    []string  `json:"scopes"`
	GrantedAt time.Time `json:"grantedAt"`
	Enabled   bool      `json:"enabled"`
}

type RevokeOIDCGrantReq struct {
	ClientID string `json:"clientId"`
}

func ListMyOIDCGrants(req component.BetterRequest[component.Null]) component.Response {
	store, err := newOIDCClientStore()
	if err != nil {
		return oidcClientFailure("initialize store", err)
	}
	grants, err := store.ListUserGrants(requestContext(req.GinContext), fmt.Sprint(req.UserId))
	if err != nil {
		return oidcClientFailure("list user grants", err)
	}
	result := make([]OIDCGrantView, 0, len(grants))
	for _, grant := range grants {
		result = append(result, OIDCGrantView{
			ClientID: grant.ClientID, Name: grant.Name, Scopes: grant.Scopes,
			GrantedAt: grant.GrantedAt, Enabled: grant.Enabled,
		})
	}
	return component.SuccessResponse(result)
}

func RevokeMyOIDCGrant(req component.BetterRequest[RevokeOIDCGrantReq]) component.Response {
	clientID := strings.TrimSpace(req.Params.ClientID)
	if clientID == "" {
		return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
	}
	store, err := newOIDCClientStore()
	if err != nil {
		return oidcClientFailure("initialize store", err)
	}
	if err := store.RevokeGrant(requestContext(req.GinContext), fmt.Sprint(req.UserId), clientID, time.Now()); err != nil {
		return oidcClientFailure("revoke user grant", err)
	}
	return component.SuccessResponse(true)
}

func GetOIDCProviderStatus(component.BetterRequest[component.Null]) component.Response {
	return component.SuccessResponse(oidcproviderservice.Status())
}

func SaveOIDCProviderSettings(req component.BetterRequest[SaveOIDCProviderSettingsReq]) component.Response {
	savePageConfig(pageConfig.OIDCProvider, pageConfig.OIDCProviderSettingsConfig{Enabled: req.Params.Enabled}, func() {})
	if err := oidcproviderservice.ReloadDefault(); err != nil {
		slog.Error("reload OIDC provider after settings update", "err", err)
	}
	return component.SuccessResponse(oidcproviderservice.Status())
}

func RotateOIDCSigningKey(req component.BetterRequest[component.Null]) component.Response {
	if err := oidcproviderservice.RotateDefaultSigningKey(requestContext(req.GinContext)); err != nil {
		return oidcClientFailure("rotate signing key", err)
	}
	return component.SuccessResponse(oidcproviderservice.Status())
}

func ListOIDCClients(req component.BetterRequest[component.Null]) component.Response {
	store, err := newOIDCClientStore()
	if err != nil {
		return oidcClientFailure("initialize store", err)
	}
	return listOIDCClients(requestContext(req.GinContext), store)
}

func CreateOIDCClient(req component.BetterRequest[CreateOIDCClientReq]) component.Response {
	store, err := newOIDCClientStore()
	if err != nil {
		return oidcClientFailure("initialize store", err)
	}
	return createOIDCClient(requestContext(req.GinContext), store, req.Params, generateOIDCClientID, core.GenerateClientSecret)
}

func UpdateOIDCClient(req component.BetterRequest[UpdateOIDCClientReq]) component.Response {
	store, err := newOIDCClientStore()
	if err != nil {
		return oidcClientFailure("initialize store", err)
	}
	return updateOIDCClient(requestContext(req.GinContext), store, req.Params)
}

func RotateOIDCClientSecret(req component.BetterRequest[RotateOIDCClientSecretReq]) component.Response {
	store, err := newOIDCClientStore()
	if err != nil {
		return oidcClientFailure("initialize store", err)
	}
	return rotateOIDCClientSecret(requestContext(req.GinContext), store, req.Params, core.GenerateClientSecret)
}

func newOIDCClientStore() (*oidcProviderStore.Store, error) {
	return oidcProviderStore.New(dbconnect.Connect())
}

func listOIDCClients(ctx context.Context, store oidcClientStore) component.Response {
	clients, err := store.ListClients(ctx)
	if err != nil {
		return oidcClientFailure("list clients", err)
	}
	views := make([]OIDCClientView, 0, len(clients))
	for i := range clients {
		views = append(views, oidcClientView(clients[i]))
	}
	return component.SuccessResponse(views)
}

func createOIDCClient(ctx context.Context, store oidcClientStore, req CreateOIDCClientReq, generateID, generateSecret func() (string, error)) component.Response {
	clientID, err := generateID()
	if err != nil {
		return oidcClientFailure("generate client ID", err)
	}
	secret := ""
	client := core.Client{
		ID:                      clientID,
		Name:                    strings.TrimSpace(req.Name),
		RedirectURIs:            normalizeStrings(req.RedirectURIs),
		Scopes:                  normalizeStrings(req.Scopes),
		GrantTypes:              normalizeStrings(req.GrantTypes),
		TokenEndpointAuthMethod: core.ClientAuthenticationMethod(strings.TrimSpace(req.TokenEndpointAuthMethod)),
		RequirePKCE:             req.RequirePKCE,
		Public:                  req.Public,
		Enabled:                 true,
	}
	if req.Enabled != nil {
		client.Enabled = *req.Enabled
	}
	if client.Public {
		client.TokenEndpointAuthMethod = core.ClientAuthNone
		client.RequirePKCE = true
	} else {
		secret, err = generateSecret()
		if err != nil {
			return oidcClientFailure("generate client secret", err)
		}
		client.SecretHash = core.HashClientSecret(secret)
	}
	if err := validateManagedOIDCClient(client); err != nil {
		return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
	}
	if existing, err := store.GetClient(ctx, client.ID); err != nil {
		return oidcClientFailure("check generated client ID", err)
	} else if existing != nil {
		return oidcClientFailure("generate unique client ID", errors.New("generated OIDC client ID collision"))
	}
	if err := store.CreateClient(ctx, &client); err != nil {
		return oidcClientFailure("create client", err)
	}
	return component.SuccessResponse(OIDCClientCredentials{Client: oidcClientView(client), ClientSecret: secret})
}

func updateOIDCClient(ctx context.Context, store oidcClientStore, req UpdateOIDCClientReq) component.Response {
	clientID := strings.TrimSpace(req.ClientID)
	if clientID == "" {
		return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
	}
	client, err := store.GetClient(ctx, clientID)
	if err != nil {
		return oidcClientFailure("get client", err)
	}
	if client == nil {
		return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
	}
	if req.Public != nil && *req.Public != client.Public {
		// Changing client type would either silently create a credential or leave
		// one behind. Create a new registration instead.
		return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
	}
	client.Name = strings.TrimSpace(req.Name)
	client.RedirectURIs = normalizeStrings(req.RedirectURIs)
	client.Scopes = normalizeStrings(req.Scopes)
	client.GrantTypes = normalizeStrings(req.GrantTypes)
	client.RequirePKCE = req.RequirePKCE
	if req.Enabled != nil {
		client.Enabled = *req.Enabled
	}
	if client.Public {
		client.TokenEndpointAuthMethod = core.ClientAuthNone
		client.RequirePKCE = true
		client.SecretHash = ""
	} else {
		client.TokenEndpointAuthMethod = core.ClientAuthenticationMethod(strings.TrimSpace(req.TokenEndpointAuthMethod))
	}
	if err := validateManagedOIDCClient(*client); err != nil {
		return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
	}
	if err := store.SaveClient(ctx, client); err != nil {
		return oidcClientFailure("update client", err)
	}
	return component.SuccessResponse(oidcClientView(*client))
}

func rotateOIDCClientSecret(ctx context.Context, store oidcClientStore, req RotateOIDCClientSecretReq, generate func() (string, error)) component.Response {
	clientID := strings.TrimSpace(req.ClientID)
	client, err := store.GetClient(ctx, clientID)
	if err != nil {
		return oidcClientFailure("get client", err)
	}
	if client == nil || client.Public {
		return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
	}
	secret, err := generate()
	if err != nil {
		return oidcClientFailure("generate secret", err)
	}
	client.SecretHash = core.HashClientSecret(secret)
	if err := validateManagedOIDCClient(*client); err != nil {
		return oidcClientFailure("validate stored client", err)
	}
	if err := store.SaveClient(ctx, client); err != nil {
		return oidcClientFailure("rotate client secret", err)
	}
	return component.SuccessResponse(OIDCClientCredentials{Client: oidcClientView(*client), ClientSecret: secret})
}

func validateManagedOIDCClient(client core.Client) error {
	if client.Name == "" || len(client.Name) > 255 || len(client.RedirectURIs) > 20 || len(client.Scopes) > 32 || len(client.GrantTypes) > 2 {
		return errors.New("OIDC client registration exceeds management limits")
	}
	for _, redirectURI := range client.RedirectURIs {
		if len(redirectURI) > 2048 {
			return errors.New("OIDC redirect URI is too long")
		}
	}
	for _, scope := range client.Scopes {
		if len(scope) > 255 {
			return errors.New("OIDC scope is too long")
		}
		if _, supported := managedOIDCScopes[scope]; !supported {
			return errors.New("OIDC scope is not supported by GooseForum")
		}
	}
	return core.ValidateClient(client)
}

func generateOIDCClientID() (string, error) {
	randomID, err := core.GenerateClientSecret()
	if err != nil {
		return "", err
	}
	return oidcClientIDPrefix + randomID, nil
}

func oidcClientView(client core.Client) OIDCClientView {
	return OIDCClientView{
		ClientID:                client.ID,
		Name:                    client.Name,
		RedirectURIs:            append([]string(nil), client.RedirectURIs...),
		Scopes:                  append([]string(nil), client.Scopes...),
		GrantTypes:              append([]string(nil), client.GrantTypes...),
		TokenEndpointAuthMethod: client.TokenEndpointAuthMethod,
		RequirePKCE:             client.RequirePKCE,
		Public:                  client.Public,
		Enabled:                 client.Enabled,
	}
}

func normalizeStrings(values []string) []string {
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = strings.TrimSpace(value)
	}
	return result
}

func requestContext(ginContext *gin.Context) context.Context {
	if ginContext == nil || ginContext.Request == nil {
		return context.Background()
	}
	return ginContext.Request.Context()
}

func oidcClientFailure(operation string, err error) component.Response {
	slog.Error("OIDC client management failed", "operation", operation, "err", err)
	return component.FailResponseCode(component.MessageOperationFailed, nil)
}
