package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/filestorage"
)

func TestCompleteDirectUploadRechecksTrustPolicyBeforeStorage(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&users.EntityComplete{}, &pageConfig.Entity{}); err != nil {
		t.Fatal(err)
	}
	u := users.EntityComplete{Username: "direct-policy-trust", UsernameLower: "direct-policy-trust", IsActivated: 1}
	if err := db.Create(&u).Error; err != nil {
		t.Fatal(err)
	}
	previous := pageConfig.GetByPageType(pageConfig.TrustSettings)
	t.Cleanup(func() {
		if previous.Id > 0 {
			_ = pageConfig.SaveConfig(pageConfig.TrustSettings, previous.Config)
		} else {
			db.Where("page_type = ?", pageConfig.TrustSettings).Delete(&pageConfig.Entity{})
		}
		hotdataserve.ClearTrustSettingsCache()
	})
	c := pageConfig.DefaultTrustSettings()
	c.RestrictionsEnabled = true
	c.LevelLimits[0].AllowImageUpload = false
	raw, _ := json.Marshal(c)
	if err := pageConfig.SaveConfig(pageConfig.TrustSettings, string(raw)); err != nil {
		t.Fatal(err)
	}
	hotdataserve.ClearTrustSettingsCache()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/file/img-upload/complete", bytes.NewBufferString(`{"name":"pending-before-policy-change"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set("userId", u.Id)
	CompleteDirectImageUpload(ctx)
	if w.Code != http.StatusForbidden || !bytes.Contains(w.Body.Bytes(), []byte("permission.denied")) {
		t.Fatalf("completion bypassed policy: %d %s", w.Code, w.Body.String())
	}
}

func TestInitDirectImageUploadReturnsProxyModeForDatabaseStorage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	if err := filestorage.Configure(filestorage.NewDatabaseStore()); err != nil {
		t.Fatalf("configure storage: %v", err)
	}
	t.Cleanup(func() { _ = filestorage.Configure(filestorage.NewDatabaseStore()) })

	request := httptest.NewRequest(http.MethodPost, "/file/img-upload/init", bytes.NewBufferString(`{"filename":"image.webp","contentType":"image/webp","size":10}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = request
	ctx.Set("userId", uint64(1))

	InitDirectImageUpload(ctx)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	var payload struct {
		Code   component.Status `json:"code"`
		Result struct {
			Mode string `json:"mode"`
		} `json:"result"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload.Code != component.SUCCESS || payload.Result.Mode != "proxy" {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestInitDirectImageUploadRejectsMalformedJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	request := httptest.NewRequest(http.MethodPost, "/file/img-upload/init", bytes.NewBufferString(`{"filename":`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = request

	InitDirectImageUpload(ctx)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestWriteDirectUploadErrorDoesNotExposeOwnership(t *testing.T) {
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	writeDirectUploadError(ctx, filestorage.ErrDirectUploadOwnerMismatch)
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d", response.Code)
	}
	if bytes.Contains(response.Body.Bytes(), []byte("owner")) {
		t.Fatalf("response leaked ownership: %s", response.Body.String())
	}
}
