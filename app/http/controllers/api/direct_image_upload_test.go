package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/service/filestorage"
)

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
