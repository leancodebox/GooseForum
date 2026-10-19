package routes

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/connect/db4fileconnect"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/markdownext"
	"github.com/leancodebox/GooseForum/app/bundles/trustlevel"
	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroupMembers"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroups"
	"github.com/leancodebox/GooseForum/app/models/forum/authsessions"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/categoryGroupPermissions"
	"github.com/leancodebox/GooseForum/app/models/forum/dailyStats"
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
	"github.com/leancodebox/GooseForum/app/models/forum/fileUsage"
	"github.com/leancodebox/GooseForum/app/models/forum/moderators"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/pointsRecord"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/role"
	"github.com/leancodebox/GooseForum/app/models/forum/rolePermissionRs"
	"github.com/leancodebox/GooseForum/app/models/forum/topicCategoryIndex"
	"github.com/leancodebox/GooseForum/app/models/forum/topicUserAction"
	"github.com/leancodebox/GooseForum/app/models/forum/topicUserStat"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/trusthistory"
	"github.com/leancodebox/GooseForum/app/models/forum/userActivities"
	"github.com/leancodebox/GooseForum/app/models/forum/userBadges"
	"github.com/leancodebox/GooseForum/app/models/forum/userPoints"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
	"github.com/leancodebox/GooseForum/app/service/datamigration"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"github.com/leancodebox/GooseForum/app/service/trustlevelservice"
	"github.com/leancodebox/GooseForum/app/service/userservice"
)

// Exercise the registered HTTP routes, including binding, session auth and ACLs.
func TestTrustHTTPContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := dbconnect.Connect()
	fileDB := db4fileconnect.Connect()
	if err := db.AutoMigrate(&eventNotification.Entity{}); err != nil {
		t.Fatal(err)
	}
	if err := fileDB.AutoMigrate(&filedata.Entity{}); err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&users.EntityComplete{}, &authsessions.Token{}, &authsessions.Log{}, &role.Entity{}, &rolePermissionRs.Entity{}, &pageConfig.Entity{}, &trusthistory.History{}, &accessGroups.Entity{}, &accessGroupMembers.Entity{}, &categoryGroupPermissions.Entity{}, &category.Entity{}, &topics.Entity{}, &posts.Entity{}, &topicCategoryIndex.Entity{}, &moderators.Entity{}, &userStatistics.Entity{}, &dailyStats.Entity{}, &fileUsage.Entity{}, &topicUserStat.Entity{}, &topicUserAction.Entity{}, &userActivities.Entity{}, &userPoints.Entity{}, &pointsRecord.Entity{}, &userBadges.Entity{}); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{pageConfig.TrustSettings, pageConfig.SecuritySettings} {
		old := pageConfig.GetByPageType(key)
		t.Cleanup(func() {
			if old.Id != 0 {
				_ = pageConfig.SaveConfig(key, old.Config)
			} else {
				db.Where("page_type = ?", key).Delete(&pageConfig.Entity{})
			}
			hotdataserve.ClearTrustSettingsCache()
			hotdataserve.ClearSecuritySettingsConfigCache()
			trustlevelservice.ClearReviewCache()
		})
	}
	if err := pageConfig.SaveConfig(pageConfig.SecuritySettings, `{"enableEmailVerification":false}`); err != nil {
		t.Fatal(err)
	}
	hotdataserve.ClearSecuritySettingsConfigCache()
	cfg := pageConfig.DefaultTrustSettings()
	save := func() {
		t.Helper()
		raw, _ := json.Marshal(cfg)
		if err := pageConfig.SaveConfig(pageConfig.TrustSettings, string(raw)); err != nil {
			t.Fatal(err)
		}
		hotdataserve.ClearTrustSettingsCache()
		trustlevelservice.ClearReviewCache()
	}
	save()
	adminRole := role.Entity{RoleName: "trust-http-admin", Effective: 1}
	if err := db.Create(&adminRole).Error; err != nil {
		t.Fatal(err)
	}
	grant := rolePermissionRs.Entity{RoleId: adminRole.Id, PermissionId: permission.Admin.Id(), Effective: 1}
	if err := db.Create(&grant).Error; err != nil {
		t.Fatal(err)
	}
	permission.InvalidateRole(adminRole.Id)
	t.Cleanup(func() {
		db.Unscoped().Delete(&grant)
		db.Unscoped().Delete(&adminRole)
		permission.InvalidateRole(adminRole.Id)
	})
	makeUser := func(name string, level trustlevel.Level, roleID uint64, locked bool) (users.EntityComplete, string) {
		t.Helper()
		u := users.EntityComplete{Username: name, UsernameLower: name, IsActivated: 1, TrustLevel: level, RoleId: roleID, TrustLevelLocked: locked, CreatedAt: time.Now().Add(-48 * time.Hour)}
		if err := db.Create(&u).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Create(&userStatistics.Entity{UserId: u.Id}).Error; err != nil {
			t.Fatal(err)
		}
		raw := fmt.Sprintf("%043d", u.Id)
		hash := sha256.Sum256([]byte(raw))
		session := authsessions.Token{UserId: u.Id, TokenHash: hex.EncodeToString(hash[:]), CreatedAt: time.Now(), LastSeenAt: time.Now(), ExpiresAt: time.Now().Add(30 * 24 * time.Hour)}
		if err := db.Create(&session).Error; err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			db.Where("user_id = ?", u.Id).Delete(&eventNotification.Entity{})
			db.Unscoped().Where("user_id = ?", u.Id).Delete(&posts.Entity{})
			db.Unscoped().Where("user_id = ?", u.Id).Delete(&topics.Entity{})
			db.Where("user_id = ?", u.Id).Delete(&trusthistory.History{})
			db.Where("user_id = ?", u.Id).Delete(&authsessions.Token{})
			db.Delete(&userStatistics.Entity{}, u.Id)
			db.Unscoped().Delete(&u)
			userservice.InvalidateUserCaches(u.Id)
		})
		return u, raw
	}
	admin, adminToken := makeUser("trust-http-admin", trustlevel.Core, adminRole.Id, true)
	user, token := makeUser("trust-http-user", trustlevel.Newcomer, 0, true)
	_, memberToken := makeUser("trust-http-member", trustlevel.Member, 0, true)
	cat := category.Entity{Name: "trust-http", Slug: "trust-http"}
	if err := db.Create(&cat).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Delete(&cat) })
	if result := datamigration.BackfillAccessControlDefaultsWithDB(db); result.Failed != 0 {
		t.Fatal(result.LastFailed)
	}
	accesscontrol.InvalidateSystemGroups()
	var groups []accessGroups.Entity
	db.Find(&groups)
	for _, g := range groups {
		accesscontrol.InvalidateGroup(g.Id)
	}
	router := gin.New()
	apiRoute(router)
	fileServer(router)
	request := func(t *testing.T, method, path, bearer string, body any, status int, message string) map[string]any {
		t.Helper()
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(method, path, strings.NewReader(string(raw)))
		req.Header.Set("Content-Type", "application/json")
		if bearer != "" {
			req.Header.Set("Authorization", "Bearer "+bearer)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != status {
			t.Fatalf("%s %s: HTTP %d: %s", method, path, w.Code, w.Body.String())
		}
		var out map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
			t.Fatal(w.Body.String(), err)
		}
		if message != "" && out["messageCode"] != message {
			t.Fatalf("%s %s: want %s, got %s", method, path, message, w.Body.String())
		}
		if message == "" && out["code"] != float64(0) {
			t.Fatalf("%s %s: %s", method, path, w.Body.String())
		}
		return out
	}
	t.Run("authentication-and-admin-boundary", func(t *testing.T) {
		request(t, "GET", "/api/trust-level", "", nil, 401, "auth.required")
		for _, route := range []struct{ method, path string }{{"GET", "get-trust-settings"}, {"POST", "save-trust-settings"}, {"GET", "user-trust-level"}, {"POST", "set-user-trust-level"}, {"GET", "user-trust-level-history"}} {
			request(t, route.method, "/api/admin/"+route.path, token, map[string]any{}, 403, "permission.denied")
		}
	})
	t.Run("config-validation-and-removed-switch", func(t *testing.T) {
		raw, _ := json.Marshal(cfg)
		var value map[string]any
		_ = json.Unmarshal(raw, &value)
		value["autoPromotionEnabled"] = false
		request(t, "POST", "/api/admin/save-trust-settings", adminToken, map[string]any{"settings": value}, 200, "")
		out := request(t, "GET", "/api/admin/get-trust-settings", adminToken, nil, 200, "")["result"].(map[string]any)
		if _, ok := out["autoPromotionEnabled"]; ok {
			t.Fatal("retired switch leaked")
		}
		value["unknownPolicy"] = true
		request(t, "POST", "/api/admin/save-trust-settings", adminToken, map[string]any{"settings": value}, 200, "common.request.invalidParams")
		for _, invalid := range []func(*pageConfig.TrustConfig){
			func(c *pageConfig.TrustConfig) { c.LevelLimits[0].MaxDailyUploads = new(-1) },
			func(c *pageConfig.TrustConfig) { c.LevelLimits[0].MaxDailyTopics = new(-1) },
			func(c *pageConfig.TrustConfig) { c.LevelLimits[0].MaxDailyReplies = new(-1) },
			func(c *pageConfig.TrustConfig) { c.LevelLimits[0].MaxImagesPerPost = new(101) },
			func(c *pageConfig.TrustConfig) { c.RegistrationCooldown.MinPostAccountAgeMinutes = -1 },
			func(c *pageConfig.TrustConfig) { c.RegistrationCooldown.MinUploadAccountAgeMinutes = -1 },
			func(c *pageConfig.TrustConfig) { c.PromotionRules[0].MinTopicsVisited = 21 },
			func(c *pageConfig.TrustConfig) { c.PromotionRules[1] = c.PromotionRules[0] },
		} {
			bad := pageConfig.DefaultTrustSettings()
			invalid(&bad)
			request(t, "POST", "/api/admin/save-trust-settings", adminToken, map[string]any{"settings": bad}, 200, "common.request.invalidParams")
		}
		preserved := request(t, "GET", "/api/admin/get-trust-settings", adminToken, nil, 200, "")["result"].(map[string]any)
		if preserved["schemaVersion"] != float64(3) {
			t.Fatal("invalid save corrupted config")
		}
		for _, field := range []string{"maxDailyTopics", "maxDailyReplies", "maxDailyUploads", "maxImagesPerPost", "maxExternalLinksPerPost", "maxMentionsPerPost", "allowImageUpload"} {
			raw, _ := json.Marshal(pageConfig.DefaultTrustSettings())
			var bad map[string]any
			if err := json.Unmarshal(raw, &bad); err != nil {
				t.Fatal(err)
			}
			limits := bad["levelLimits"].([]any)
			if field == "allowImageUpload" {
				limits[1].(map[string]any)[field] = false
			} else {
				limits[0].(map[string]any)[field] = nil
				limits[1].(map[string]any)[field] = 0
			}
			out := request(t, "POST", "/api/admin/save-trust-settings", adminToken, map[string]any{"settings": bad}, 200, "common.request.invalidParams")
			params := out["params"].(map[string]any)
			if params["field"] != field || params["lowerLevel"] != float64(0) || params["higherLevel"] != float64(100) {
				t.Fatalf("missing limit conflict details: %v", out)
			}
		}
		after := request(t, "GET", "/api/admin/get-trust-settings", adminToken, nil, 200, "")["result"]
		beforeJSON, _ := json.Marshal(preserved)
		afterJSON, _ := json.Marshal(after)
		if string(beforeJSON) != string(afterJSON) {
			t.Fatal("rejected order changed stored policy")
		}
	})
	write := func(t *testing.T, bearer string, id uint64, status int, content string, message string) map[string]any {
		return request(t, "POST", "/api/forum/topics/write", bearer, map[string]any{"sourceVersion": 1, "topicId": id, "topicStatus": status, "title": "HTTP trust topic", "content": content, "categoryId": []uint64{cat.Id}}, 200, message)
	}
	t.Run("content-limits-and-level-benefits", func(t *testing.T) {
		for _, content := range []string{"![a](/a.png) ![b](/b.png)", "[a](https://a.example) [b](https://b.example) [c](https://c.example)", markdownext.Canonical(admin.Id, admin.Username) + " " + markdownext.Canonical(user.Id, user.Username)} {
			cfg.LevelLimits[0].MaxMentionsPerPost = new(1)
			save()
			write(t, token, 0, 1, content, "trust.contentLimit")
		}
		write(t, memberToken, 0, 1, "![a](/a.png) ![b](/b.png)", "")
	})
	t.Run("draft-publish-cooldown-both-endpoints", func(t *testing.T) {
		cfg.LevelLimits[0].MaxDailyTopics = nil
		save()
		id := uint64(write(t, token, 0, 0, "Draft with enough content", "")["result"].(float64))
		second := uint64(write(t, token, 0, 0, "Second draft content", "")["result"].(float64))
		cfg.RegistrationCooldown.MinPostAccountAgeMinutes = 4320
		save()
		write(t, token, id, 1, "Changed draft content", "topic.post.cooldown")
		request(t, "POST", "/api/forum/topics/status", token, map[string]any{"topicId": id, "topicStatus": 1}, 200, "topic.post.cooldown")
		if topics.Get(id).Status != 0 {
			t.Fatal("rejected publish changed draft")
		}
		write(t, token, 0, 1, "New topic blocked too", "topic.post.cooldown")
		cfg.RegistrationCooldown.MinPostAccountAgeMinutes = 0
		cfg.LevelLimits[0].MaxDailyTopics = new(0)
		save()
		request(t, "POST", "/api/forum/topics/status", token, map[string]any{"topicId": id, "topicStatus": 1}, 200, "")
		write(t, token, second, 1, "Publish draft through write endpoint", "")
		write(t, token, 0, 1, "New topic exceeds quota", "trust.dailyTopics")
		write(t, token, id, 1, "Published edit allowed", "")
		cfg.RegistrationCooldown.MinPostAccountAgeMinutes = 4320
		save()
		write(t, token, id, 1, "Published edit ignores age", "")
		cfg.RegistrationCooldown.MinPostAccountAgeMinutes = 0
		cfg.LevelLimits[0].MaxDailyReplies = new(0)
		save()
		request(t, "POST", "/api/forum/posts/create", token, map[string]any{"topicId": id, "content": "Reply denied by zero quota"}, 200, "trust.dailyReplies")
		cfg.LevelLimits[0].MaxDailyReplies = nil
		cfg.RegistrationCooldown.MinPostAccountAgeMinutes = 4320
		save()
		request(t, "POST", "/api/forum/posts/create", token, map[string]any{"topicId": id, "content": "Reply denied by account age"}, 200, "topic.post.cooldown")
		cfg.RegistrationCooldown.MinPostAccountAgeMinutes = 0
		save()
		request(t, "POST", "/api/forum/posts/create", token, map[string]any{"topicId": id, "content": "Reply allowed with no quota"}, 200, "")
	})
	t.Run("upload-policy-and-role-avatar-boundaries", func(t *testing.T) {
		cfg.LevelLimits[0].AllowImageUpload = false
		save()
		request(t, "POST", "/file/img-upload", token, nil, 403, "permission.denied")
		request(t, "POST", "/file/img-upload/complete", token, map[string]any{"name": "pending-before-change"}, 403, "permission.denied")
		cfg.LevelLimits[0].AllowImageUpload = true
		cfg.RegistrationCooldown.MinUploadAccountAgeMinutes = 4320
		save()
		request(t, "POST", "/file/img-upload", token, nil, 400, "upload.cooldown")
		request(t, "POST", "/api/upload-avatar", token, nil, 200, "upload.cooldown")
		cfg.RegistrationCooldown.MinUploadAccountAgeMinutes = 0
		cfg.LevelLimits[0].MaxDailyUploads = new(0)
		save()
		request(t, "POST", "/file/img-upload", token, nil, 400, "upload.dailyLimit")
		cfg.LevelLimits[0].MaxDailyUploads = new(1)
		save()
		pending := filedata.Entity{Name: "trust-http-pending", UserId: user.Id, StorageStatus: filedata.StorageStatusPending}
		if err := fileDB.Create(&pending).Error; err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { fileDB.Delete(&pending) })
		request(t, "POST", "/file/img-upload", token, nil, 400, "upload.dailyLimit")
		cfg.RegistrationCooldown.MinUploadAccountAgeMinutes = 4320
		cfg.LevelLimits[3].MaxDailyUploads = new(0)
		save()
		// An admin role bypasses upload age/quota, reaching file validation.
		request(t, "POST", "/file/img-upload", adminToken, nil, 400, "upload.file.missing")
		request(t, "POST", "/api/upload-avatar", adminToken, nil, 200, "upload.cooldown")
	})
	t.Run("exact-limits-for-every-level", func(t *testing.T) {
		for i, d := range trustlevel.Definitions() {
			_, bearer := makeUser(fmt.Sprintf("trust-http-level-%d", d.Level), d.Level, 0, true)
			cfg.RegistrationCooldown.MinPostAccountAgeMinutes = 4320
			save()
			write(t, bearer, 0, 1, "Shared cooldown applies before every level", "topic.post.cooldown")
			cfg.RegistrationCooldown.MinPostAccountAgeMinutes = 0
			cfg.LevelLimits[i].MaxDailyTopics = new(0)
			save()
			write(t, bearer, 0, 1, "Denied for this exact level", "trust.dailyTopics")
			cfg.LevelLimits[i].MaxDailyTopics = nil
			save()
			write(t, bearer, 0, 1, "Allowed without topic quota", "")
			request(t, "GET", "/api/admin/get-trust-settings", bearer, nil, 403, "permission.denied")
		}
	})
	t.Run("manual-level-lock-conflict-history", func(t *testing.T) {
		path := fmt.Sprintf("/api/admin/user-trust-level?userId=%d", user.Id)
		progress := request(t, "GET", path, adminToken, nil, 200, "")["result"].(map[string]any)
		change := map[string]any{"userId": user.Id, "level": 200, "locked": true, "expectedVersion": progress["version"], "reason": "HTTP tested promotion"}
		request(t, "POST", "/api/admin/set-user-trust-level", adminToken, change, 200, "")
		request(t, "POST", "/api/admin/set-user-trust-level", adminToken, change, 200, "trust.adjustmentFailed")
		current := request(t, "GET", "/api/trust-level", token, nil, 200, "")["result"].(map[string]any)
		if current["level"] != float64(200) || current["locked"] != true {
			t.Fatal(current)
		}
		change["level"] = 0
		change["expectedVersion"] = current["version"]
		request(t, "POST", "/api/admin/set-user-trust-level", adminToken, change, 200, "trust.adjustmentFailed")
		for _, level := range []int{-1, 101} {
			change["level"] = level
			request(t, "POST", "/api/admin/set-user-trust-level", adminToken, change, 200, "trust.adjustmentFailed")
		}
		change["level"] = 32768
		request(t, "POST", "/api/admin/set-user-trust-level", adminToken, change, 400, "common.request.parseFailed")
		history := request(t, "GET", fmt.Sprintf("/api/admin/user-trust-level-history?userId=%d", user.Id), adminToken, nil, 200, "")["result"].([]any)
		if len(history) != 1 {
			t.Fatalf("history=%v", history)
		}
		entry := history[0].(map[string]any)
		if entry["fromLevel"] != float64(0) || entry["toLevel"] != float64(200) {
			t.Fatalf("history levels=%v", entry)
		}
		notices := request(t, "GET", "/api/forum/notifications", token, nil, 200, "")["result"].(map[string]any)["items"].([]any)
		if len(notices) != 1 {
			t.Fatalf("promotion notices=%v", notices)
		}
		notice := notices[0].(map[string]any)
		body := notice["payload"].(map[string]any)
		if notice["eventType"] != "system" || body["templateKey"] != eventNotification.TemplateTrustPromotion || body["templateParams"].(map[string]any)["trustLevel"] != float64(200) {
			t.Fatalf("wrong promotion notice=%v", notice)
		}
		request(t, "POST", "/api/forum/notification/mark-read", token, map[string]any{"notificationId": notice["id"]}, 200, "")
		request(t, "POST", "/api/admin/set-user-trust-level", adminToken, map[string]any{"userId": user.Id, "level": 200, "locked": false, "expectedVersion": current["version"], "reason": "Unlock only"}, 200, "")
		var count int64
		if err := db.Model(&eventNotification.Entity{}).Where("user_id = ?", user.Id).Count(&count).Error; err != nil || count != 1 {
			t.Fatalf("lock change duplicated notice: count=%d err=%v", count, err)
		}
	})
	t.Run("restricted-account-cannot-write-or-promote", func(t *testing.T) {
		u, bearer := makeUser("trust-http-suspended", trustlevel.Newcomer, 0, false)
		if err := db.Model(&users.EntityComplete{}).Where("id = ?", u.Id).Updates(map[string]any{"restriction_status": users.RestrictionSuspended, "trust_visit_days": 2, "trust_topic_ids": "[1,2,3,4,5]"}).Error; err != nil {
			t.Fatal(err)
		}
		userservice.InvalidateUserCaches(u.Id)
		trustlevelservice.ClearReviewCache()
		out := request(t, "GET", "/api/trust-level", bearer, nil, 200, "")["result"].(map[string]any)
		if out["level"] != float64(0) {
			t.Fatal("suspended account promoted", out)
		}
		request(t, "POST", "/api/forum/topics/write", bearer, map[string]any{}, 403, "permission.userFrozen")
	})
	t.Run("level-groups-control-classification-not-admin-rights", func(t *testing.T) {
		key := "trust:200"
		group := accessGroups.Entity{Name: "HTTP regular group", SystemKey: &key, JoinMode: accessGroups.JoinModeSystem, Status: accessGroups.StatusEnabled}
		if err := db.Where("system_key = ?", key).FirstOrCreate(&group).Error; err != nil {
			t.Fatal(err)
		}
		private := category.Entity{Name: "HTTP regular category", Slug: "http-regular"}
		if err := db.Create(&private).Error; err != nil {
			t.Fatal(err)
		}
		grant := categoryGroupPermissions.Entity{CategoryId: private.Id, AccessGroupId: group.Id, PermissionLevel: categoryGroupPermissions.PermissionCreate, Status: categoryGroupPermissions.StatusEnabled}
		if err := db.Create(&grant).Error; err != nil {
			t.Fatal(err)
		}
		accesscontrol.InvalidateSystemGroups()
		accesscontrol.InvalidateGroup(group.Id)
		t.Cleanup(func() { db.Delete(&grant); db.Delete(&private); accesscontrol.InvalidateGroup(group.Id) })
		for _, level := range []trustlevel.Level{trustlevel.Newcomer, trustlevel.Member, trustlevel.Regular, trustlevel.Core} {
			_, bearer := makeUser(fmt.Sprintf("trust-http-acl-%d", level), level, 0, true)
			want := ""
			if level < trustlevel.Regular {
				want = "permission.denied"
			}
			request(t, "POST", "/api/forum/topics/write", bearer, map[string]any{"topicStatus": 1, "title": "Private category topic", "content": "Trust group authorization test", "categoryId": []uint64{private.Id}}, 200, want)
		}
	})
	t.Run("automatic-promotion-via-authenticated-request", func(t *testing.T) {
		u, bearer := makeUser("trust-http-auto", trustlevel.Newcomer, 0, false)
		if err := db.Model(&users.EntityComplete{}).Where("id = ?", u.Id).Updates(map[string]any{"trust_visit_days": 2, "trust_topic_ids": "[1,2,3,4,5]"}).Error; err != nil {
			t.Fatal(err)
		}
		trustlevelservice.ClearReviewCache()
		out := request(t, "GET", "/api/trust-level", bearer, nil, 200, "")["result"].(map[string]any)
		if out["level"] != float64(100) {
			t.Fatal(out)
		}
		if out["limits"].(map[string]any)["level"] != float64(100) || out["nextLimits"].(map[string]any)["level"] != float64(200) {
			t.Fatalf("missing capability comparison=%v", out)
		}
		for i := 0; i < 2; i++ {
			request(t, "GET", "/api/trust-level", bearer, nil, 200, "")
		}
		var count int64
		if err := db.Model(&eventNotification.Entity{}).Where("user_id = ?", u.Id).Count(&count).Error; err != nil || count != 1 {
			t.Fatalf("auto notice duplicate: count=%d err=%v", count, err)
		}
		if _, ok := out["autoPromotionEnabled"]; ok {
			t.Fatal("retired switch in progress")
		}
	})
}
