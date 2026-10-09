package agent

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/leancodebox/GooseForum/app/bundles/connect/db4fileconnect"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	core "github.com/leancodebox/GooseForum/app/bundles/oidcprovider"
	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroups"
	"github.com/leancodebox/GooseForum/app/models/forum/agenttokens"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/categoryGroupPermissions"
	"github.com/leancodebox/GooseForum/app/models/forum/fileUsage"
	"github.com/leancodebox/GooseForum/app/models/forum/oidcProviderStore"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topicCategoryIndex"
	"github.com/leancodebox/GooseForum/app/models/forum/topicUserStat"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
	"github.com/leancodebox/GooseForum/app/service/oidcproviderservice"
)

type testAccessStore struct{}

func (testAccessStore) SystemGroupIDs() (map[string]uint64, error) {
	return map[string]uint64{"everyone": 1, "registered": 2}, nil
}
func (testAccessStore) ActiveCustomGroupIDs(uint64) ([]uint64, error) { return nil, nil }
func (testAccessStore) EnabledCategoryGrants(groupID uint64) ([]accesscontrol.CategoryGrant, error) {
	if groupID == 1 {
		return []accesscontrol.CategoryGrant{{CategoryID: 99001, Capability: accesscontrol.CapabilityRead}}, nil
	}
	return []accesscontrol.CategoryGrant{{CategoryID: 99001, Capability: accesscontrol.CapabilityCreate}, {CategoryID: 99002, Capability: accesscontrol.CapabilityCreate}}, nil
}

func TestAuthorizationForumWritesAndIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := dbconnect.Connect()
	fileDB := db4fileconnect.Connect()
	if err := fileDB.AutoMigrate(&filedata.Entity{}); err != nil {
		t.Fatal(err)
	}
	models := append(oidcProviderStore.Models(), &users.EntityComplete{}, &agenttokens.Entity{}, &pageConfig.Entity{}, &posts.Entity{}, &topics.Entity{}, &category.Entity{}, &accessGroups.Entity{}, &categoryGroupPermissions.Entity{}, &topicCategoryIndex.Entity{}, &topicUserStat.Entity{}, &filedata.Entity{}, &fileUsage.Entity{})
	if err := db.AutoMigrate(models...); err != nil {
		t.Fatal(err)
	}
	enableAgentSettings(t)
	user := users.EntityComplete{Id: 99001, Username: "agent-test", UsernameLower: "agent-test", IsActivated: 1, CreatedAt: time.Now().Add(-365 * 24 * time.Hour)}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&[]category.Entity{{Id: 99001, Name: "Public"}, {Id: 99002, Name: "Private"}}).Error; err != nil {
		t.Fatal(err)
	}
	everyone := "everyone"
	if err := db.Create(&accessGroups.Entity{Id: 99001, Name: "Everyone", SystemKey: &everyone, Status: accessGroups.StatusEnabled}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&categoryGroupPermissions.Entity{CategoryId: 99001, AccessGroupId: 99001, PermissionLevel: categoryGroupPermissions.PermissionRead, Status: categoryGroupPermissions.StatusEnabled}).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Where("user_id = ?", user.Id).Delete(&fileUsage.Entity{})
		fileDB.Where("user_id = ?", user.Id).Delete(&filedata.Entity{})
		db.Unscoped().Where("user_id = ?", user.Id).Delete(&posts.Entity{})
		db.Unscoped().Where("user_id = ?", user.Id).Delete(&topics.Entity{})
		db.Where("user_id = ?", user.Id).Delete(&agenttokens.Entity{})
		db.Unscoped().Delete(&user)
		db.Delete(&category.Entity{}, []uint64{99001, 99002})
		db.Where("access_group_id = ?", 99001).Delete(&categoryGroupPermissions.Entity{})
		db.Delete(&accessGroups.Entity{}, 99001)
		db.Where("client_id = ?", "agent-integration").Delete(&oidcProviderStore.TokenEntity{})
		db.Where("client_id = ?", "agent-integration").Delete(&oidcProviderStore.AuthorizationCodeEntity{})
		db.Where("client_id = ?", "agent-integration").Delete(&oidcProviderStore.ConsentEntity{})
		db.Where("client_id = ?", "agent-integration").Delete(&oidcProviderStore.ClientEntity{})
	})
	previous := accesscontrol.Default
	resolver := accesscontrol.NewResolver(testAccessStore{}, nil, nil)
	accesscontrol.Default = resolver
	t.Cleanup(func() { accesscontrol.Default = previous })
	service, err := oidcproviderservice.New(context.Background(), oidcproviderservice.Options{DB: db, SiteURL: "https://forum.example", KeyEncryptionSecret: []byte(strings.Repeat("k", 32)), SupportedScopes: []string{"openid", "offline_access", core.ScopeForumRead, core.ScopeTopicsCreate, core.ScopePostsCreate, core.ScopeImagesUpload}})
	if err != nil {
		t.Fatal(err)
	}
	provider := service.Provider()
	client := &core.Client{ID: "agent-integration", Name: "Agent", RedirectURIs: []string{"http://127.0.0.1:9876/callback"}, Scopes: []string{"openid", "offline_access", core.ScopeForumRead, core.ScopeTopicsCreate, core.ScopePostsCreate, core.ScopeImagesUpload}, GrantTypes: []string{"authorization_code", "refresh_token"}, TokenEndpointAuthMethod: core.ClientAuthNone, Public: true, RequirePKCE: true, Enabled: true}
	if err := service.InteractionStore().CreateClient(context.Background(), client); err != nil {
		t.Fatal(err)
	}
	verifier := strings.Repeat("v", 43)
	digest := sha256.Sum256([]byte(verifier))
	authReq := core.AuthorizeRequest{ClientID: client.ID, RedirectURI: client.RedirectURIs[0], ResponseType: "code", Scope: client.Scopes, State: "random-state", Nonce: "random-nonce", CodeChallenge: base64.RawURLEncoding.EncodeToString(digest[:]), CodeChallengeMethod: "S256", Prompt: []string{"consent"}}
	code, err := provider.Authorize(context.Background(), authReq, core.Authentication{UserID: fmt.Sprint(user.Id), AuthTime: time.Now()}, true)
	if err != nil {
		t.Fatal(err)
	}
	tokens, err := provider.ExchangeCode(context.Background(), core.TokenRequest{GrantType: "authorization_code", ClientID: client.ID, AuthMethod: core.ClientAuthNone, RedirectURI: client.RedirectURIs[0], Code: code.Code, CodeVerifier: verifier})
	if err != nil {
		t.Fatal(err)
	}
	h := New()
	h.Provider = func() (*core.Provider, error) { return provider, nil }
	h.Resolve = resolver.Resolve
	router := gin.New()
	h.Register(router)
	call := func(method, path, token string, body any) (int, map[string]any) {
		t.Helper()
		data, _ := json.Marshal(body)
		req := httptest.NewRequest(method, path, bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		out := httptest.NewRecorder()
		router.ServeHTTP(out, req)
		var result map[string]any
		if err := json.Unmarshal(out.Body.Bytes(), &result); err != nil {
			t.Fatalf("%s %s: %d %s", method, path, out.Code, out.Body.String())
		}
		return out.Code, result
	}
	if status, _ := call("GET", "/api/agent/v1/me", tokens.IDToken, nil); status != 401 {
		t.Fatalf("ID token accepted: %d", status)
	}
	if status, _ := call("GET", "/api/agent/v1/categories", "invalid", nil); status != 401 {
		t.Fatalf("invalid credential fell back to anonymous: %d", status)
	}
	if status, _ := call("GET", "/api/agent/v1/me", tokens.AccessToken, nil); status != 200 {
		t.Fatalf("me status %d", status)
	}
	uploadStatus, uploaded := uploadRequest(t, router, tokens.AccessToken, "agent.png", pngImage(t))
	if uploadStatus != 201 {
		t.Fatalf("OAuth image upload: %d %#v", uploadStatus, uploaded)
	}
	imageURL := uploaded["data"].(map[string]any)["url"].(string)
	key := uuid.NewString()
	body := map[string]any{"title": "Agent integration topic", "content": "A sufficiently long integration test message.\n\n![Agent image](" + imageURL + ")", "categoryIds": []string{"99001"}, "clientRequestId": key}
	status, result := call("POST", "/api/agent/v1/topics", tokens.AccessToken, body)
	if status != 201 {
		t.Fatalf("create: %d %#v", status, result)
	}
	created := result["data"].(map[string]any)
	topicID := created["topicId"].(string)
	if status, _ := call("GET", "/api/agent/v1/topics/"+topicID, "", nil); status != 200 {
		t.Fatalf("public topic: %d", status)
	}
	refreshed, err := provider.Refresh(context.Background(), core.TokenRequest{GrantType: "refresh_token", RefreshToken: tokens.RefreshToken, ClientID: client.ID, AuthMethod: core.ClientAuthNone})
	if err != nil {
		t.Fatal(err)
	}
	status, result = call("POST", "/api/agent/v1/topics", refreshed.AccessToken, body)
	if status != 200 || result["data"].(map[string]any)["reused"] != true {
		t.Fatalf("refresh dedup: %d %#v", status, result)
	}
	body["title"] = "Changed title"
	if status, _ := call("POST", "/api/agent/v1/topics", refreshed.AccessToken, body); status != 409 {
		t.Fatalf("key conflict status %d", status)
	}
	body["title"] = "Agent integration topic"
	status, result = call("GET", "/api/agent/v1/me/submissions?clientRequestId="+key, refreshed.AccessToken, nil)
	if status != 200 || result["data"].(map[string]any)["topicId"] != topicID {
		t.Fatalf("query submission: %d %#v", status, result)
	}
	reply := map[string]any{"content": "A sufficiently long reply to the topic.", "clientRequestId": uuid.NewString()}
	status, result = call("POST", "/api/agent/v1/topics/"+topicID+"/posts", refreshed.AccessToken, reply)
	if status != 201 {
		t.Fatalf("reply: %d %#v", status, result)
	}
	if status, _ := call("POST", "/api/agent/v1/topics/"+topicID+"/posts", refreshed.AccessToken, reply); status != 200 {
		t.Fatalf("duplicate reply: %d", status)
	}
	manual, raw, err := agenttokens.Create(db, user.Id, user.TokenVersion, "Read only", []string{core.ScopeForumRead}, 30)
	if err != nil {
		t.Fatal(err)
	}
	if status, _ := call("POST", "/api/agent/v1/topics", raw, body); status != 403 {
		t.Fatalf("scope escalated: %d", status)
	}
	if status, _ := call("GET", "/api/agent/v1/me/submissions?clientRequestId="+key, raw, nil); status != 404 {
		t.Fatalf("cross credential submission leak: %d", status)
	}
	if err := db.Model(&agenttokens.Entity{}).Where("id = ?", manual.ID).Update("revoked_at", time.Now()).Error; err != nil {
		t.Fatal(err)
	}
	if status, _ := call("GET", "/api/agent/v1/me", raw, nil); status != 401 {
		t.Fatalf("manual revocation: %d", status)
	}
	// Concurrent requests still have one database-unique business result.
	concurrentKey := uuid.NewString()
	concurrentBody := map[string]any{"content": "A concurrent reply with a stable request key.", "clientRequestId": concurrentKey}
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			call("POST", "/api/agent/v1/topics/"+topicID+"/posts", refreshed.AccessToken, concurrentBody)
		}()
	}
	wg.Wait()
	var count int64
	if err := db.Model(&posts.Entity{}).Where("client_request_id = ?", concurrentKey).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("concurrent count %d error %v", count, err)
	}
	h.limits = limiter{}
	if status, _ := call("POST", "/api/agent/v1/topics/"+topicID+"/posts", refreshed.AccessToken, map[string]any{"content": "Unknown fields are rejected.", "clientRequestId": uuid.NewString(), "userId": "another-user"}); status != 400 {
		t.Fatalf("unknown field status %d", status)
	}
	if status, _ := call("GET", "/api/agent/v1/topics?limit=51", refreshed.AccessToken, nil); status != 400 {
		t.Fatalf("invalid page size status %d", status)
	}
	if status, _ := call("GET", "/api/agent/v1/topics?cursor="+cursorToken("topics", "different-filter", 1), refreshed.AccessToken, nil); status != 400 {
		t.Fatalf("mismatched cursor status %d", status)
	}
	if status, _ := call("GET", "/api/agent/v1/topics/"+topicID+"/posts?limit=1", refreshed.AccessToken, nil); status != 200 {
		t.Fatalf("post window status %d", status)
	}
	// The author may query a pending submission without exposing it publicly.
	if err := db.Model(&topics.Entity{}).Where("id = ?", topicID).Updates(map[string]any{"status": 0, "moderation_status": "pending"}).Error; err != nil {
		t.Fatal(err)
	}
	status, result = call("GET", "/api/agent/v1/me/submissions?clientRequestId="+key, refreshed.AccessToken, nil)
	if status != 200 || result["data"].(map[string]any)["visible"] != false || result["data"].(map[string]any)["moderationStatus"] != "pending" {
		t.Fatalf("pending query: %d %#v", status, result)
	}
	if status, _ := call("GET", "/api/agent/v1/topics/"+topicID, "", nil); status != 404 {
		t.Fatalf("pending topic leak: %d", status)
	}
	if err := db.Model(&topics.Entity{}).Where("id = ?", topicID).Updates(map[string]any{"status": 1, "moderation_status": "approved"}).Error; err != nil {
		t.Fatal(err)
	}
	// Reading a child must not disclose the ID of a hidden parent.
	parent := posts.Entity{TopicId: mustID(t, topicID), PostNo: 100, UserId: user.Id, Content: "Hidden parent", ProcessStatus: 1}
	if err := db.Create(&parent).Error; err != nil {
		t.Fatal(err)
	}
	child := posts.Entity{TopicId: parent.TopicId, PostNo: 101, UserId: user.Id, Content: "Visible child", ReplyToPostId: parent.Id}
	if err := db.Create(&child).Error; err != nil {
		t.Fatal(err)
	}
	status, result = call("GET", "/api/agent/v1/posts/"+fmt.Sprint(child.Id), refreshed.AccessToken, nil)
	if status != 200 || result["data"].(map[string]any)["replyToPostId"] != nil {
		t.Fatalf("hidden parent exposed: %d %#v", status, result)
	}
	if status, _ := call("POST", "/api/agent/v1/topics/"+topicID+"/posts", refreshed.AccessToken, map[string]any{"content": "Cannot reply to a hidden parent.", "replyToPostId": fmt.Sprint(parent.Id), "clientRequestId": uuid.NewString()}); status != 404 {
		t.Fatalf("hidden parent reply status %d", status)
	}
	// Moving the topic into a private category removes anonymous access.
	if err := db.Model(&topics.Entity{}).Where("id = ?", topicID).Update("main_category_id", 99002).Error; err != nil {
		t.Fatal(err)
	}
	if status, _ := call("GET", "/api/agent/v1/topics/"+topicID, "", nil); status != 404 {
		t.Fatalf("private topic leak: %d", status)
	}
	if err := db.Delete(&topics.Entity{}, topicID).Error; err != nil {
		t.Fatal(err)
	}
	if status, _ := call("GET", "/api/agent/v1/me/submissions/"+created["submissionId"].(string), refreshed.AccessToken, nil); status != 410 {
		t.Fatalf("deleted submission status %d", status)
	}
	if err := db.Model(&users.EntityComplete{}).Where("id = ?", user.Id).Update("token_version", 1).Error; err != nil {
		t.Fatal(err)
	}
	if status, _ := call("GET", "/api/agent/v1/me", refreshed.AccessToken, nil); status != 401 {
		t.Fatalf("account version change: %d", status)
	}
	if _, err := provider.Refresh(context.Background(), core.TokenRequest{GrantType: "refresh_token", RefreshToken: refreshed.RefreshToken, ClientID: client.ID, AuthMethod: core.ClientAuthNone}); err == nil {
		t.Fatal("refresh survived account credential reset")
	}
	// Fresh authorization works after reset; revoking the app then invalidates it.
	newCode, err := provider.Authorize(context.Background(), authReq, core.Authentication{UserID: fmt.Sprint(user.Id), AuthTime: time.Now()}, true)
	if err != nil {
		t.Fatal(err)
	}
	newTokens, err := provider.ExchangeCode(context.Background(), core.TokenRequest{GrantType: "authorization_code", ClientID: client.ID, AuthMethod: core.ClientAuthNone, RedirectURI: client.RedirectURIs[0], Code: newCode.Code, CodeVerifier: verifier})
	if err != nil {
		t.Fatal(err)
	}
	if err := provider.RevokeGrant(context.Background(), fmt.Sprint(user.Id), client.ID); err != nil {
		t.Fatal(err)
	}
	if status, _ := call("GET", "/api/agent/v1/me", newTokens.AccessToken, nil); status != 401 {
		t.Fatalf("OAuth grant revocation: %d", status)
	}
}

func mustID(t *testing.T, value string) uint64 {
	t.Helper()
	n, ok := id(value)
	if !ok {
		t.Fatal("invalid ID")
	}
	return n
}

func TestOpenAPIAndSkillRoutes(t *testing.T) {
	doc := openAPIDocument("https://forum.example/prefix")
	paths := doc["paths"].(gin.H)
	router := gin.New()
	New().Register(router)
	for _, route := range router.Routes() {
		if !strings.HasPrefix(route.Path, "/api/agent/v1/") || strings.HasSuffix(route.Path, "openapi.json") || strings.HasSuffix(route.Path, "API.md") {
			continue
		}
		path := strings.TrimPrefix(route.Path, "/api/agent/v1")
		for _, name := range []string{"topicId", "postId", "submissionId"} {
			path = strings.ReplaceAll(path, ":"+name, "{"+name+"}")
		}
		methods, ok := paths[path].(gin.H)
		if !ok || methods[strings.ToLower(route.Method)] == nil {
			t.Errorf("missing operation %s %s", route.Method, path)
		}
	}
	if len(paths) != 11 {
		t.Fatalf("unexpected path count %d", len(paths))
	}
	if !bytes.Contains(skillDocument, []byte("code_verifier")) || !bytes.Contains(skillDocument, []byte("clientRequestId")) {
		t.Fatal("Skill is missing authorization or retry workflow")
	}
	if !bytes.Contains(apiDocument, []byte("[mention user=")) {
		t.Fatal("Skill does not describe canonical mentions")
	}
	if !bytes.Contains(skillDocument, []byte("/api/agent/v1/API.md")) {
		t.Fatal("Skill must prefer the compact reference")
	}
	encoded, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if len(encoded) >= 30000 {
		t.Fatalf("OpenAPI grew to %d bytes", len(encoded))
	}
	t.Logf("OpenAPI %d bytes; Markdown %d bytes", len(encoded), len(apiDocument))
	responses := doc["components"].(gin.H)["responses"].(gin.H)
	for path, methods := range paths {
		for method, value := range methods.(gin.H) {
			op := value.(gin.H)
			if op["description"] != nil {
				t.Errorf("repeated operation prose: %s %s", method, path)
			}
			for status, value := range op["responses"].(gin.H) {
				if status == "200" || status == "201" {
					continue
				}
				ref := value.(gin.H)["$ref"].(string)
				if responses[strings.TrimPrefix(ref, "#/components/responses/")] == nil {
					t.Errorf("missing response %s", ref)
				}
				if method == "get" && (status == "409" || status == "413" || status == "422") {
					t.Errorf("write-only error on read: %s %s", path, status)
				}
			}
		}
	}
	for _, path := range []string{"/site", "/me", "/categories", "/topics", "/topics/{topicId}", "/topics/{topicId}/posts", "/posts/{postId}", "/search", "/me/submissions", "/me/submissions/{submissionId}"} {
		if !bytes.Contains(apiDocument, []byte("`"+path+"`")) {
			t.Errorf("Markdown is missing %s", path)
		}
	}
}

func TestMarkdownReferencesArePublicAndCacheable(t *testing.T) {
	router := gin.New()
	New().Register(router)
	for path, document := range map[string][]byte{"/api/agent/SKILL.md": skillDocument, "/api/agent/v1/API.md": apiDocument} {
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("Authorization", "Bearer invalid")
		out := httptest.NewRecorder()
		router.ServeHTTP(out, req)
		if out.Code != 200 || !bytes.Equal(out.Body.Bytes(), document) || !strings.HasPrefix(out.Header().Get("Content-Type"), "text/markdown") {
			t.Fatalf("bad reference response: %s %d", path, out.Code)
		}
		etag := out.Header().Get("ETag")
		if etag == "" {
			t.Fatal("missing ETag")
		}
		req = httptest.NewRequest("GET", path, nil)
		req.Header.Set("If-None-Match", etag)
		out = httptest.NewRecorder()
		router.ServeHTTP(out, req)
		if out.Code != 304 || out.Body.Len() != 0 {
			t.Fatalf("cache validation failed: %s", path)
		}
	}
}
