package api

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	core "github.com/leancodebox/GooseForum/app/bundles/oidcprovider"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
)

type fakeOIDCClientStore struct {
	clients map[string]core.Client
	err     error
}

func newFakeOIDCClientStore() *fakeOIDCClientStore {
	return &fakeOIDCClientStore{clients: make(map[string]core.Client)}
}

func (s *fakeOIDCClientStore) ListClients(context.Context) ([]core.Client, error) {
	if s.err != nil {
		return nil, s.err
	}
	result := make([]core.Client, 0, len(s.clients))
	for _, client := range s.clients {
		result = append(result, cloneOIDCClient(client))
	}
	return result, nil
}

func (s *fakeOIDCClientStore) GetClient(_ context.Context, id string) (*core.Client, error) {
	if s.err != nil {
		return nil, s.err
	}
	client, ok := s.clients[id]
	if !ok {
		return nil, nil
	}
	owned := cloneOIDCClient(client)
	return &owned, nil
}

func (s *fakeOIDCClientStore) SaveClient(_ context.Context, client *core.Client) error {
	if s.err != nil {
		return s.err
	}
	s.clients[client.ID] = cloneOIDCClient(*client)
	return nil
}

func (s *fakeOIDCClientStore) CreateClient(_ context.Context, client *core.Client) error {
	if s.err != nil {
		return s.err
	}
	if _, exists := s.clients[client.ID]; exists {
		return errors.New("client already exists")
	}
	s.clients[client.ID] = cloneOIDCClient(*client)
	return nil
}

func cloneOIDCClient(client core.Client) core.Client {
	client.RedirectURIs = append([]string(nil), client.RedirectURIs...)
	client.Scopes = append([]string(nil), client.Scopes...)
	client.GrantTypes = append([]string(nil), client.GrantTypes...)
	return client
}

func TestCreateOIDCConfidentialClientReturnsSecretOnce(t *testing.T) {
	store := newFakeOIDCClientStore()
	response := createOIDCClient(context.Background(), store, CreateOIDCClientReq{
		Name:                    " Wiki ",
		RedirectURIs:            []string{" https://wiki.example/callback "},
		Scopes:                  []string{"openid", "profile"},
		GrantTypes:              []string{"authorization_code"},
		TokenEndpointAuthMethod: string(core.ClientSecretBasic),
	}, func() (string, error) { return "gf_client", nil }, func() (string, error) {
		return strings.Repeat("s", 43), nil
	})

	credentials := successResult[OIDCClientCredentials](t, response)
	if credentials.ClientSecret != strings.Repeat("s", 43) {
		t.Fatalf("clientSecret = %q", credentials.ClientSecret)
	}
	stored := store.clients["gf_client"]
	if stored.Name != "Wiki" || stored.SecretHash != core.HashClientSecret(credentials.ClientSecret) {
		t.Fatalf("stored client = %#v", stored)
	}
	assertNoSecretHashJSON(t, response.Data.Result)
}

func TestCreateOIDCPublicClientForcesNoneAndPKCE(t *testing.T) {
	store := newFakeOIDCClientStore()
	response := createOIDCClient(context.Background(), store, CreateOIDCClientReq{
		Name:                    "Mobile",
		RedirectURIs:            []string{"com.example.app:/callback"},
		Scopes:                  []string{"openid"},
		GrantTypes:              []string{"authorization_code"},
		TokenEndpointAuthMethod: string(core.ClientSecretPost),
		Public:                  true,
		RequirePKCE:             false,
	}, func() (string, error) { return "gf_public", nil }, func() (string, error) {
		t.Fatal("public clients must not generate a secret")
		return "", nil
	})

	credentials := successResult[OIDCClientCredentials](t, response)
	if credentials.ClientSecret != "" {
		t.Fatalf("public client secret leaked: %q", credentials.ClientSecret)
	}
	if !credentials.Client.RequirePKCE || credentials.Client.TokenEndpointAuthMethod != core.ClientAuthNone {
		t.Fatalf("public policy not enforced: %#v", credentials.Client)
	}
	if store.clients["gf_public"].SecretHash != "" {
		t.Fatal("public client persisted a secret hash")
	}
}

func TestUpdateOIDCClientPreservesSecretAndType(t *testing.T) {
	store := newFakeOIDCClientStore()
	originalHash := core.HashClientSecret("old-secret")
	store.clients["gf_client"] = core.Client{
		ID: "gf_client", Name: "Old", SecretHash: originalHash,
		RedirectURIs: []string{"https://old.example/callback"}, Scopes: []string{"openid"},
		GrantTypes: []string{"authorization_code"}, TokenEndpointAuthMethod: core.ClientSecretBasic, Enabled: true,
	}
	response := updateOIDCClient(context.Background(), store, UpdateOIDCClientReq{
		ClientID: "gf_client", Name: "New", RedirectURIs: []string{"https://new.example/callback"},
		Scopes: []string{"openid", "email"}, GrantTypes: []string{"authorization_code"},
		TokenEndpointAuthMethod: string(core.ClientSecretPost),
	})

	view := successResult[OIDCClientView](t, response)
	if view.Name != "New" || view.TokenEndpointAuthMethod != core.ClientSecretPost {
		t.Fatalf("updated view = %#v", view)
	}
	if store.clients["gf_client"].SecretHash != originalHash {
		t.Fatal("update changed the client secret")
	}
	public := true
	response = updateOIDCClient(context.Background(), store, UpdateOIDCClientReq{ClientID: "gf_client", Public: &public})
	if response.Data.Code != component.FAIL {
		t.Fatal("client type change should be rejected")
	}
}

func TestRotateOIDCClientSecret(t *testing.T) {
	store := newFakeOIDCClientStore()
	store.clients["gf_client"] = core.Client{
		ID: "gf_client", Name: "Wiki", SecretHash: core.HashClientSecret("old-secret"),
		RedirectURIs: []string{"https://wiki.example/callback"}, Scopes: []string{"openid"},
		GrantTypes: []string{"authorization_code"}, TokenEndpointAuthMethod: core.ClientSecretBasic, Enabled: true,
	}
	response := rotateOIDCClientSecret(context.Background(), store, RotateOIDCClientSecretReq{ClientID: "gf_client"}, func() (string, error) {
		return strings.Repeat("n", 43), nil
	})
	credentials := successResult[OIDCClientCredentials](t, response)
	if credentials.ClientSecret != strings.Repeat("n", 43) {
		t.Fatalf("rotated secret = %q", credentials.ClientSecret)
	}
	if store.clients["gf_client"].SecretHash != core.HashClientSecret(credentials.ClientSecret) {
		t.Fatal("rotated secret hash was not persisted")
	}
	assertNoSecretHashJSON(t, response.Data.Result)
}

func TestOIDCClientManagementRejectsInvalidAndStorageFailures(t *testing.T) {
	store := newFakeOIDCClientStore()
	response := createOIDCClient(context.Background(), store, CreateOIDCClientReq{
		Name: "Invalid", RedirectURIs: []string{"https://app.example/callback"}, Scopes: []string{"profile"},
		GrantTypes: []string{"authorization_code"}, TokenEndpointAuthMethod: "not-supported",
	}, func() (string, error) { return "gf_invalid", nil }, func() (string, error) { return "secret", nil })
	if response.Data.Code != component.FAIL || response.Data.MessageCode != component.MessageRequestInvalidParams {
		t.Fatalf("invalid client response = %#v", response.Data)
	}
	response = createOIDCClient(context.Background(), store, CreateOIDCClientReq{
		Name: "Unsupported scope", RedirectURIs: []string{"https://app.example/callback"}, Scopes: []string{"openid", "groups"},
		GrantTypes: []string{"authorization_code"}, TokenEndpointAuthMethod: string(core.ClientSecretBasic),
	}, func() (string, error) { return "gf_scope", nil }, func() (string, error) { return "secret", nil })
	if response.Data.Code != component.FAIL || response.Data.MessageCode != component.MessageRequestInvalidParams {
		t.Fatalf("unsupported scope response = %#v", response.Data)
	}

	store.err = errors.New("database unavailable")
	response = listOIDCClients(context.Background(), store)
	if response.Data.Code != component.FAIL || response.Data.MessageCode != component.MessageOperationFailed {
		t.Fatalf("storage failure response = %#v", response.Data)
	}
}

func TestListOIDCClientsNeverExposesSecretHash(t *testing.T) {
	store := newFakeOIDCClientStore()
	store.clients["gf_client"] = core.Client{
		ID: "gf_client", Name: "Wiki", SecretHash: core.HashClientSecret("secret"),
		RedirectURIs: []string{"https://wiki.example/callback"}, Scopes: []string{"openid"},
		GrantTypes: []string{"authorization_code"}, TokenEndpointAuthMethod: core.ClientSecretBasic, Enabled: true,
	}
	response := listOIDCClients(context.Background(), store)
	views := successResult[[]OIDCClientView](t, response)
	if len(views) != 1 || views[0].ClientID != "gf_client" {
		t.Fatalf("views = %#v", views)
	}
	assertNoSecretHashJSON(t, response.Data.Result)
}

func successResult[T any](t *testing.T, response component.Response) T {
	t.Helper()
	if response.Data.Code != component.SUCCESS {
		t.Fatalf("response = %#v", response.Data)
	}
	result, ok := response.Data.Result.(T)
	if !ok {
		t.Fatalf("result type = %T", response.Data.Result)
	}
	return result
}

func assertNoSecretHashJSON(t *testing.T, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.ToLower(string(data)), "secrethash") {
		t.Fatalf("response exposes secret hash: %s", data)
	}
}
