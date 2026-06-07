package jwtopt

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestCreateNewToken(t *testing.T) {
	const userID uint64 = 123456
	token, err := CreateNewToken(userID, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	userId, newToken, err := VerifyTokenWithFresh(token)
	if err != nil {
		t.Fatal(err)
	}
	if userId != userID {
		t.Fatalf("VerifyTokenWithFresh userId = %d, want %d", userId, userID)
	}
	if newToken == "" {
		t.Fatal("expected refreshed token")
	}

	userId, err = VerifyToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if userId != userID {
		t.Fatalf("VerifyToken userId = %d, want %d", userId, userID)
	}

	time.Sleep(1100 * time.Millisecond)
	if _, err = VerifyToken(token); err == nil {
		t.Fatal("expected expired token error")
	}
}

func TestCreateNewTokenDefault(t *testing.T) {
	token, err := CreateNewTokenDefault(7)
	if err != nil {
		t.Fatal(err)
	}
	userID, err := VerifyToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if userID != 7 {
		t.Fatalf("userID = %d, want 7", userID)
	}
}

func TestCreateNewTokenWithVersion(t *testing.T) {
	const userID uint64 = 7
	const tokenVersion uint64 = 3

	token, err := CreateNewTokenWithVersion(userID, tokenVersion, 15*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	claims, newToken, err := VerifyTokenWithFreshClaims(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserId != userID || claims.TokenVersion != tokenVersion {
		t.Fatalf("claims = (%d, %d), want (%d, %d)", claims.UserId, claims.TokenVersion, userID, tokenVersion)
	}
	if newToken == token {
		t.Fatal("expected refreshed token")
	}

	refreshedClaims, _, err := VerifyTokenWithFreshClaims(newToken)
	if err != nil {
		t.Fatal(err)
	}
	if refreshedClaims.TokenVersion != tokenVersion {
		t.Fatalf("refreshed tokenVersion = %d, want %d", refreshedClaims.TokenVersion, tokenVersion)
	}
}

func TestGetGinAccessToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	headerRecorder := httptest.NewRecorder()
	headerContext, _ := gin.CreateTestContext(headerRecorder)
	headerContext.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	headerContext.Request.Header.Set("Authorization", "Bearer header-token")
	if got := GetGinAccessToken(headerContext); got != "header-token" {
		t.Fatalf("header token = %q, want header-token", got)
	}

	cookieRecorder := httptest.NewRecorder()
	cookieContext, _ := gin.CreateTestContext(cookieRecorder)
	cookieContext.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	cookieContext.Request.AddCookie(&http.Cookie{Name: "access_token", Value: "cookie-token"})
	if got := GetGinAccessToken(cookieContext); got != "cookie-token" {
		t.Fatalf("cookie token = %q, want cookie-token", got)
	}
}

func TestTokenSettingAndClean(t *testing.T) {
	gin.SetMode(gin.TestMode)

	setRecorder := httptest.NewRecorder()
	setContext, _ := gin.CreateTestContext(setRecorder)
	TokenSetting(setContext, "fresh-token")
	if got := setRecorder.Header().Get("New-Token"); got != "fresh-token" {
		t.Fatalf("New-Token = %q, want fresh-token", got)
	}
	if cookies := setRecorder.Result().Cookies(); len(cookies) != 1 || cookies[0].Name != "access_token" || cookies[0].Value != "fresh-token" {
		t.Fatalf("set cookies = %#v", cookies)
	}

	cleanRecorder := httptest.NewRecorder()
	cleanContext, _ := gin.CreateTestContext(cleanRecorder)
	TokenClean(cleanContext)
	setCookie := cleanRecorder.Header().Get("Set-Cookie")
	if !strings.Contains(setCookie, "access_token=") || !strings.Contains(setCookie, "Max-Age=0") {
		t.Fatalf("clear cookie header = %q", setCookie)
	}
}

func TestVerifyTokenWithFresh(t *testing.T) {
	const userID uint64 = 123456
	token, err := CreateNewToken(123456, 15*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	userId, newToken, err := VerifyTokenWithFresh(token)
	if err != nil {
		t.Fatal(err)
	}
	if userId != userID {
		t.Fatalf("VerifyTokenWithFresh userId = %d, want %d", userId, userID)
	}
	if newToken != token {
		token = newToken
	}

	time.Sleep(time.Second * 2)
	userId, err = VerifyToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if userId != userID {
		t.Fatalf("VerifyToken userId = %d, want %d", userId, userID)
	}
}
