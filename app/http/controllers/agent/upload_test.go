package agent

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/connect/db4fileconnect"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	core "github.com/leancodebox/GooseForum/app/bundles/oidcprovider"
	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
	"github.com/leancodebox/GooseForum/app/models/forum/agenttokens"
	"github.com/leancodebox/GooseForum/app/models/forum/fileUsage"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
)

func uploadRequest(t *testing.T, router http.Handler, token, filename string, data []byte) (int, map[string]any) {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	if filename != "" {
		part, err := w.CreateFormFile("file", filename)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "/api/agent/v1/images", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	out := httptest.NewRecorder()
	router.ServeHTTP(out, req)
	var result map[string]any
	if err := json.Unmarshal(out.Body.Bytes(), &result); err != nil {
		t.Fatalf("%d %s", out.Code, out.Body.String())
	}
	return out.Code, result
}

func enableAgentSettings(t *testing.T) {
	t.Helper()
	old := pageConfig.GetByPageType(pageConfig.AgentSettings)
	config := pageConfig.DefaultAgentSettings()
	config.Enabled, config.ManualTokens = true, true
	data, _ := json.Marshal(config)
	if err := pageConfig.SaveConfig(pageConfig.AgentSettings, string(data)); err != nil {
		t.Fatal(err)
	}
	hotdataserve.ClearAgentSettingsConfigCache()
	t.Cleanup(func() {
		if old.Id == 0 {
			dbconnect.Connect().Where("page_type = ?", pageConfig.AgentSettings).Delete(&pageConfig.Entity{})
		} else {
			_ = pageConfig.SaveConfig(pageConfig.AgentSettings, old.Config)
		}
		hotdataserve.ClearAgentSettingsConfigCache()
	})
}

func pngImage(t *testing.T) []byte {
	t.Helper()
	var data bytes.Buffer
	if err := png.Encode(&data, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	return data.Bytes()
}

func TestManualImageUploadPermissionsValidationAndQuota(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&users.EntityComplete{}, &agenttokens.Entity{}, &pageConfig.Entity{}, &fileUsage.Entity{}); err != nil {
		t.Fatal(err)
	}
	fileDB := db4fileconnect.Connect()
	enableAgentSettings(t)
	if err := fileDB.AutoMigrate(&filedata.Entity{}); err != nil {
		t.Fatal(err)
	}
	user := users.EntityComplete{Id: 99011, Username: "agent-upload-test", UsernameLower: "agent-upload-test", IsActivated: 1, CreatedAt: time.Now().Add(-365 * 24 * time.Hour)}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	old := pageConfig.GetByPageType(pageConfig.PostingSettings)
	t.Cleanup(func() {
		db.Unscoped().Where("user_id = ?", user.Id).Delete(&agenttokens.Entity{})
		db.Where("user_id = ?", user.Id).Delete(&fileUsage.Entity{})
		fileDB.Where("user_id = ?", user.Id).Delete(&filedata.Entity{})
		db.Unscoped().Delete(&user)
		if old.Id == 0 {
			db.Where("page_type = ?", pageConfig.PostingSettings).Delete(&pageConfig.Entity{})
		} else {
			_ = pageConfig.SaveConfig(pageConfig.PostingSettings, old.Config)
		}
		hotdataserve.ClearPostingSettingsConfigCache()
	})
	config := hotdataserve.GetPostingSettingsConfigCache()
	config.UploadControl.AllowAttachments = true
	config.UploadControl.AuthorizedExtensions = []string{".png"}
	config.UploadControl.MaxAttachmentSizeKb = 1
	config.UploadControl.MaxDailyUploadsPerUser = 1
	config.UploadControl.NewUserUploadCooldownMinutes = 60
	saveConfig := func() {
		data, _ := json.Marshal(config)
		if err := pageConfig.SaveConfig(pageConfig.PostingSettings, string(data)); err != nil {
			t.Fatal(err)
		}
		hotdataserve.ClearPostingSettingsConfigCache()
	}
	saveConfig()
	_, token, err := agenttokens.CreateToken(user.Id, user.TokenVersion, "Upload", []string{core.ScopeForumRead, core.ScopeImagesUpload}, 30)
	if err != nil {
		t.Fatal(err)
	}
	_, readOnly, err := agenttokens.CreateToken(user.Id, user.TokenVersion, "Read", []string{core.ScopeForumRead}, 30)
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	h := New()
	h.Resolve = accesscontrol.NewResolver(testAccessStore{}, nil, nil).Resolve
	h.Register(router)
	data := pngImage(t)
	for _, tc := range []struct {
		name, token, filename string
		data                  []byte
		status                int
	}{
		{"anonymous", "", "photo.png", data, 401},
		{"scope", readOnly, "photo.png", data, 403},
		{"missing", token, "", nil, 400},
		{"extension", token, "photo.jpg", data, 422},
		{"invalid", token, "photo.png", []byte("not an image"), 422},
		{"too_large", token, "photo.png", make([]byte, 1025), 413},
		{"request_too_large", token, "photo.png", make([]byte, 70*1024), 413},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if status, result := uploadRequest(t, router, tc.token, tc.filename, tc.data); status != tc.status {
				t.Fatalf("%d %#v", status, result)
			}
		})
	}
	if err := db.Table(user.TableName()).Where("id = ?", user.Id).Update("created_at", time.Now()).Error; err != nil {
		t.Fatal(err)
	}
	if status, result := uploadRequest(t, router, token, "photo.png", data); status != 429 {
		t.Fatalf("cooldown: %d %#v", status, result)
	}
	if err := db.Table(user.TableName()).Where("id = ?", user.Id).Update("created_at", time.Now().Add(-365*24*time.Hour)).Error; err != nil {
		t.Fatal(err)
	}
	config.UploadControl.AllowAttachments = false
	saveConfig()
	if status, result := uploadRequest(t, router, token, "photo.png", data); status != 403 {
		t.Fatalf("disabled: %d %#v", status, result)
	}
	config.UploadControl.AllowAttachments = true
	saveConfig()
	status, result := uploadRequest(t, router, token, "photo.png", data)
	if status != 201 {
		t.Fatalf("upload: %d %#v", status, result)
	}
	response := result["data"].(map[string]any)
	var file filedata.Entity
	if err := fileDB.Where("user_id = ?", user.Id).Take(&file).Error; err != nil {
		t.Fatal(err)
	}
	if file.GetAccessPath() != response["url"] || response["size"] != float64(len(data)) || !bytes.Equal(file.Data, data) {
		t.Fatalf("stored image mismatch: %#v", response)
	}
	var count int64
	if err := db.Model(&fileUsage.Entity{}).Where("file_name = ? AND user_id = ? AND usage_type = ?", file.Name, user.Id, fileUsage.UsageUploadOwner).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("owner usage: %d %v", count, err)
	}
	if status, result := uploadRequest(t, router, token, "photo.png", data); status != 429 {
		t.Fatalf("quota: %d %#v", status, result)
	}
	if err := db.Model(&user).Update("token_version", 1).Error; err != nil {
		t.Fatal(err)
	}
	if status, result := uploadRequest(t, router, token, "photo.png", data); status != 401 {
		t.Fatalf("stale token: %d %#v", status, result)
	}
}
