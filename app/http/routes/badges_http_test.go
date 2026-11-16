package routes

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/authsessions"
	"github.com/leancodebox/GooseForum/app/models/forum/badgeStats"
	"github.com/leancodebox/GooseForum/app/models/forum/badges"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/role"
	"github.com/leancodebox/GooseForum/app/models/forum/rolePermissionRs"
	"github.com/leancodebox/GooseForum/app/models/forum/userBadges"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/badgeservice"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"github.com/leancodebox/GooseForum/app/service/userservice"
)

func TestBadgeRecipientRoutesRequireSiteAndUserManagement(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&badgeStats.Entity{}, &users.EntityComplete{}, &authsessions.Token{}, &authsessions.Log{}, &role.Entity{}, &rolePermissionRs.Entity{}, &pageConfig.Entity{}, &badges.Entity{}, &userBadges.Entity{}); err != nil {
		t.Fatal(err)
	}
	for _, code := range badgeservice.SystemCodes() {
		if err := badgeStats.EnsureWithDB(db, code); err != nil {
			t.Fatal(err)
		}
	}
	router := gin.New()
	apiRoute(router)
	badgeservice.InvalidateDefinitions()
	t.Cleanup(badgeservice.InvalidateDefinitions)
	for _, test := range []struct {
		name        string
		permissions []permission.Enum
		wantStatus  int
	}{
		{"site only", []permission.Enum{permission.SiteManager}, http.StatusForbidden},
		{"users only", []permission.Enum{permission.UserManager}, http.StatusForbidden},
		{"site and users", []permission.Enum{permission.SiteManager, permission.UserManager}, http.StatusOK},
	} {
		t.Run(test.name, func(t *testing.T) {
			roleEntity := role.Entity{RoleName: fmt.Sprintf("badge-http-%d", time.Now().UnixNano()), Effective: 1}
			if err := db.Create(&roleEntity).Error; err != nil {
				t.Fatal(err)
			}
			for _, p := range test.permissions {
				if err := db.Create(&rolePermissionRs.Entity{RoleId: roleEntity.Id, PermissionId: p.Id(), Effective: 1}).Error; err != nil {
					t.Fatal(err)
				}
			}
			permission.InvalidateRole(roleEntity.Id)
			user := users.EntityComplete{Username: roleEntity.RoleName, RoleId: roleEntity.Id, IsActivated: 1}
			if err := db.Create(&user).Error; err != nil {
				t.Fatal(err)
			}
			token := fmt.Sprintf("%043d", user.Id)
			hash := sha256.Sum256([]byte(token))
			session := authsessions.Token{UserId: user.Id, TokenHash: hex.EncodeToString(hash[:]), CreatedAt: time.Now(), LastSeenAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)}
			if err := db.Create(&session).Error; err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				db.Where("role_id = ?", roleEntity.Id).Delete(&rolePermissionRs.Entity{})
				db.Delete(&roleEntity)
				db.Delete(&session)
				db.Unscoped().Delete(&user)
				permission.InvalidateRole(roleEntity.Id)
				userservice.InvalidateUserCaches(user.Id)
			})
			for _, path := range []string{"/api/admin/badge-recipients", "/api/admin/badge-recipient-update"} {
				// An invalid mutation body reaches validation only when both ACLs permit it.
				body := `{"code":"first_post","action":"invalid","reason":"test"}`
				req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+token)
				response := httptest.NewRecorder()
				router.ServeHTTP(response, req)
				if response.Code != test.wantStatus {
					t.Fatalf("%s status=%d body=%s", path, response.Code, response.Body.String())
				}
				if test.wantStatus == http.StatusForbidden {
					var result struct {
						MessageCode string `json:"messageCode"`
					}
					if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
						t.Fatal(err)
					}
					if result.MessageCode != "permission.denied" {
						t.Fatalf("unexpected denial: %s", response.Body.String())
					}
				}
			}
		})
	}
}
