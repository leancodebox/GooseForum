package api

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/connect/db4fileconnect"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/trustlevel"
	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
	"github.com/leancodebox/GooseForum/app/models/forum/attachments"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
)

type countedAttachmentBody struct {
	io.Reader
	read int
}

func (b *countedAttachmentBody) Read(p []byte) (int, error) {
	n, err := b.Reader.Read(p)
	b.read += n
	return n, err
}
func (b *countedAttachmentBody) Close() error { return nil }

func TestAttachmentHTTPDownloadHeadersAndRawRouteIsolation(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&attachments.Entity{}, &users.EntityComplete{}, &pageConfig.Entity{}); err != nil {
		t.Fatal(err)
	}
	if err := db4fileconnect.Connect().AutoMigrate(&filedata.Entity{}); err != nil {
		t.Fatal(err)
	}
	user := users.EntityComplete{Id: 98876543, Username: "attachment-http", UsernameLower: "attachment-http", TrustLevel: trustlevel.Member, CreatedAt: time.Now().Add(-72 * time.Hour)}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Unscoped().Delete(&user) })
	router := gin.New()
	router.Use(func(c *gin.Context) {
		if c.GetHeader("X-Test-Owner") == "yes" {
			c.Set("userId", user.Id)
		}
	})
	router.POST("/attachments", UploadAttachment)
	router.GET("/attachments/:id", DownloadAttachment)
	router.GET("/img/*filename", GetFileByFileName)
	t.Run("rejects permission and cooldown before reading body", func(t *testing.T) {
		for _, tc := range []struct {
			level   trustlevel.Level
			created time.Time
			code    string
		}{
			{trustlevel.Newcomer, user.CreatedAt, "permission.denied"},
			{trustlevel.Member, time.Now(), "upload.cooldown"},
		} {
			if err := db.Table("users").Where("id = ?", user.Id).Updates(map[string]any{"trust_level": tc.level, "created_at": tc.created}).Error; err != nil {
				t.Fatal(err)
			}
			body := &countedAttachmentBody{Reader: bytes.NewBufferString("must not be read")}
			req := httptest.NewRequest(http.MethodPost, "/attachments", body)
			req.Header.Set("X-Test-Owner", "yes")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code < 400 || !bytes.Contains(rec.Body.Bytes(), []byte(tc.code)) || body.read != 0 {
				t.Fatalf("preflight failed: status=%d read=%d body=%s", rec.Code, body.read, rec.Body.String())
			}
		}
		if err := db.Table("users").Where("id = ?", user.Id).Updates(map[string]any{"trust_level": trustlevel.Member, "created_at": time.Now().Add(-72 * time.Hour)}).Error; err != nil {
			t.Fatal(err)
		}
	})
	t.Run("bounds streamed file reads by configured size", func(t *testing.T) {
		maximum := hotdataserve.GetPostingSettingsConfigCache().Attachments.MaxFileSizeKb * 1024
		var form bytes.Buffer
		writer := multipart.NewWriter(&form)
		part, err := writer.CreateFormFile("file", "large.txt")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write(bytes.Repeat([]byte("x"), maximum+128*1024)); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		body := &countedAttachmentBody{Reader: &form}
		req := httptest.NewRequest(http.MethodPost, "/attachments", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("X-Test-Owner", "yes")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != 400 || !bytes.Contains(rec.Body.Bytes(), []byte("attachment.tooLarge")) || body.read > maximum+64*1024 {
			t.Fatalf("unbounded read: status=%d read=%d body=%s", rec.Code, body.read, rec.Body.String())
		}
		if req.MultipartForm != nil && (len(req.MultipartForm.File) != 0 || len(req.MultipartForm.Value) != 0) {
			t.Fatal("streamed upload populated a buffered multipart form")
		}
	})
	for _, filename := range []string{"../secret.txt", "folder/secret.txt", `folder\secret.txt`} {
		t.Run(filename, func(t *testing.T) {
			body := new(bytes.Buffer)
			writer := multipart.NewWriter(body)
			part, err := writer.CreateFormFile("file", filename)
			if err != nil {
				t.Fatal(err)
			}
			part.Write([]byte("hello"))
			writer.Close()
			req := httptest.NewRequest(http.MethodPost, "/attachments", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			req.Header.Set("X-Test-Owner", "yes")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != http.StatusBadRequest || !bytes.Contains(rec.Body.Bytes(), []byte("attachment.invalidFile")) {
				t.Fatalf("path filename accepted: %d %s", rec.Code, rec.Body.String())
			}
		})
	}
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "说明.txt")
	part.Write([]byte("hello attachment"))
	writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/attachments", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Test-Owner", "yes")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("upload %d: %s", rec.Code, rec.Body.String())
	}
	var result struct {
		Result attachments.Info `json:"result"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Result.ID == 0 {
		t.Fatal(rec.Body.String())
	}
	entity, err := attachments.Get(result.Result.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		path   string
		owner  bool
		status int
	}{
		{"/attachments/" + strconv.FormatUint(entity.ID, 10), false, 404},
		{"/attachments/" + strconv.FormatUint(entity.ID, 10), true, 200},
		{"/img/" + entity.ObjectName, false, 404}, {"/img/" + entity.ObjectName, true, 404},
	} {
		req := httptest.NewRequest("GET", tc.path, nil)
		if tc.owner {
			req.Header.Set("X-Test-Owner", "yes")
		}
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != tc.status {
			t.Fatalf("%s owner=%v: %d", tc.path, tc.owner, rec.Code)
		}
		if rec.Code == 200 {
			if rec.Header().Get("Content-Type") != "application/octet-stream" || rec.Header().Get("X-Content-Type-Options") != "nosniff" || rec.Header().Get("Cache-Control") != "private, no-store" || !bytes.Contains([]byte(rec.Header().Get("Content-Disposition")), []byte("attachment;")) {
				t.Fatal(rec.Header())
			}
			if rec.Body.String() != "hello attachment" {
				t.Fatal("wrong body")
			}
		}
	}
}
