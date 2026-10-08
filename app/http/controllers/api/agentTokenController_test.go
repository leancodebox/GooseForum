package api

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/algorithm"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/agenttokens"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/usermfa"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

func TestManualTokenManagementVerifiesPasswordAndOwnership(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&users.EntityComplete{}, &agenttokens.Entity{}, &pageConfig.Entity{}, &usermfa.Factor{}, &usermfa.RecoveryCode{}); err != nil {
		t.Fatal(err)
	}
	hash, err := algorithm.MakePassword("Password123")
	if err != nil {
		t.Fatal(err)
	}
	user := users.EntityComplete{Id: 98756121, Username: "agent-token-owner", UsernameLower: "agent-token-owner", Password: hash, IsActivated: 1}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Where("user_id = ?", user.Id).Delete(&agenttokens.Entity{}); db.Unscoped().Delete(&user) })
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	req := component.BetterRequest[CreateAgentTokenReq]{UserId: user.Id, GinContext: c, Params: CreateAgentTokenReq{Name: "Integration", Scopes: []string{"forum:read"}, Days: 30, Password: "WrongPassword"}}
	if response := CreateAgentToken(req); response.Data.Code != component.FAIL {
		t.Fatal("wrong password created a credential")
	}
	req.Params.Password = "Password123"
	response := CreateAgentToken(req)
	if response.Data.Code != component.SUCCESS {
		t.Fatalf("create failed: %#v", response.Data)
	}
	result := response.Data.Result.(map[string]any)
	raw := result["token"].(string)
	entry := result["entry"].(agenttokens.Entity)
	if !strings.HasPrefix(raw, "gf_agent_") || entry.Hash == raw || entry.Hash != agenttokens.Digest(raw) {
		t.Fatal("token not securely hashed")
	}
	listed := ListAgentTokens(component.BetterRequest[component.Null]{UserId: user.Id})
	data, err := json.Marshal(listed.Data)
	if err != nil || strings.Contains(string(data), raw) || strings.Contains(string(data), entry.Hash) {
		t.Fatal("token list leaked credential material")
	}
	RevokeAgentToken(component.BetterRequest[RevokeAgentTokenReq]{UserId: user.Id + 1, Params: RevokeAgentTokenReq{ID: entry.ID}})
	var stored agenttokens.Entity
	if err := db.First(&stored, "id = ?", entry.ID).Error; err != nil {
		t.Fatal(err)
	}
	if stored.RevokedAt != nil {
		t.Fatal("another user revoked this token")
	}
	if response := RevokeAllAgentTokens(component.BetterRequest[component.Null]{UserId: user.Id}); response.Data.Code != component.SUCCESS {
		t.Fatal("revoke all failed")
	}
	if err := db.First(&stored, "id = ?", entry.ID).Error; err != nil || stored.RevokedAt == nil {
		t.Fatal("owner revoke did not persist")
	}
	DeleteAgentToken(component.BetterRequest[RevokeAgentTokenReq]{UserId: user.Id + 1, Params: RevokeAgentTokenReq{ID: entry.ID}})
	if err := db.First(&stored, "id = ?", entry.ID).Error; err != nil {
		t.Fatal("another user deleted this token")
	}
	DeleteAgentToken(component.BetterRequest[RevokeAgentTokenReq]{UserId: user.Id, Params: RevokeAgentTokenReq{ID: entry.ID}})
	if err := db.First(&agenttokens.Entity{}, "id = ?", entry.ID).Error; err == nil {
		t.Fatal("deleted token remains authenticatable")
	}
	var retained agenttokens.Entity
	if err := db.Unscoped().First(&retained, "id = ?", entry.ID).Error; err != nil || !retained.DeletedAt.Valid {
		t.Fatal("source identity was not retained")
	}
	if entries := ListAgentTokens(component.BetterRequest[component.Null]{UserId: user.Id}).Data.Result.([]agenttokens.Entity); len(entries) != 0 {
		t.Fatal("deleted token remains listed")
	}
}
