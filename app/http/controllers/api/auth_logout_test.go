package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/authsessions"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/authsessionservice"
	"gorm.io/gorm"
)

func TestLogoutReportsRevocationFailure(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&users.EntityComplete{}, &authsessions.Token{}, &authsessions.Log{}); err != nil {
		t.Fatal(err)
	}
	userID := uint64(800000000 + time.Now().UnixNano()%100000000)
	user := users.EntityComplete{Id: userID, Username: "logout-fixture", TokenVersion: 1}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Where("user_id = ?", userID).Delete(&authsessions.Token{})
		db.Where("user_id = ?", userID).Delete(&authsessions.Log{})
		db.Unscoped().Delete(&user)
	})
	issued := httptest.NewRecorder()
	issueContext, _ := gin.CreateTestContext(issued)
	issueContext.Request = httptest.NewRequest(http.MethodPost, "/api/login", nil)
	if err := authsessionservice.Issue(issueContext, userID, 1, authsessionservice.LoginDetails{Method: "password"}); err != nil {
		t.Fatal(err)
	}
	var cookie *http.Cookie
	for _, item := range issued.Result().Cookies() {
		if item.Name == "access_token" {
			cookie = item
		}
	}
	if cookie == nil {
		t.Fatal("no login cookie")
	}

	const callback = "test:logout_revoke_failure"
	if err := db.Callback().Update().Before("gorm:update").Register(callback, func(tx *gorm.DB) { tx.AddError(errors.New("write failed")) }); err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/logout", nil)
	c.Request.AddCookie(cookie)
	Logout(c)
	if err := db.Callback().Update().Remove(callback); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("logout status = %d", response.Code)
	}
	if strings.Contains(response.Header().Get("Set-Cookie"), "access_token=") {
		t.Fatal("failed logout cleared cookie")
	}
}
