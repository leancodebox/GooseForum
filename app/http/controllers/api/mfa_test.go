package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/authsessions"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/usermfa"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/authsessionservice"
	"github.com/leancodebox/GooseForum/app/service/mfaservice"
	"github.com/pquerna/otp/totp"
)

func mfaAPIContext(body string, cookie *http.Cookie) (*gin.Context, *httptest.ResponseRecorder) {
	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/mfa/login", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		c.Request.AddCookie(cookie)
	}
	return c, r
}

func TestMFAAPIChallengeRequiresCookieAndDoesNotIssueSessionEarly(t *testing.T) {
	useMailSettings(t, pageConfig.MailSettingsConfig{EnableMail: true, SmtpHost: "smtp.example.com", SmtpPort: 587, FromEmail: "forum@example.com"})
	oldKey := preferences.Get("app.secretKey")
	preferences.Set("app.secretKey", base64.RawURLEncoding.EncodeToString(make([]byte, 32)))
	t.Cleanup(func() { preferences.Set("app.secretKey", oldKey) })
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&users.EntityComplete{}, &usermfa.Factor{}, &usermfa.RecoveryCode{}, &authsessions.Token{}, &authsessions.Log{}); err != nil {
		t.Fatal(err)
	}
	user := users.MakeUser(fmt.Sprintf("mfa-api-%d", time.Now().UnixNano()), "password123", "")
	user.IsActivated = users.ActivationSuccess
	if err := db.Create(user).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Where("user_id = ?", user.Id).Delete(&usermfa.Factor{})
		db.Where("user_id = ?", user.Id).Delete(&usermfa.RecoveryCode{})
		db.Where("user_id = ?", user.Id).Delete(&authsessions.Token{})
		db.Where("user_id = ?", user.Id).Delete(&authsessions.Log{})
		db.Unscoped().Delete(user)
	})
	c, _ := mfaAPIContext("", nil)
	setup, err := mfaservice.Begin(c, user.Id, "password123")
	if err != nil {
		t.Fatal(err)
	}
	code, _ := totp.GenerateCode(setup["secret"].(string), time.Now())
	codes, err := mfaservice.Change(c, user.Id, "password123", code, "enable")
	if err != nil {
		t.Fatal(err)
	}
	user.TokenVersion++
	securityContext, _ := mfaAPIContext("", nil)
	emailResponse := EditUserEmail(component.BetterRequest[EditUserEmailReq]{UserId: user.Id, GinContext: securityContext, Params: EditUserEmailReq{Email: fmt.Sprintf("mfa-%d@example.com", user.Id)}})
	if emailResponse.Data.MessageCode != component.MessageMFAInvalid {
		t.Fatalf("email changed without MFA: %+v", emailResponse)
	}
	unbindResponse := UnbindOAuth(component.BetterRequest[UnbindOAuthReq]{UserId: user.Id, GinContext: securityContext})
	if unbindResponse.Data.MessageCode != component.MessageMFAInvalid {
		t.Fatalf("OAuth unbound without MFA: %+v", unbindResponse)
	}
	securityContext.Set("userId", user.Id)
	securityContext.Set("sessionId", uint64(123))
	securityContext.Request.Method = http.MethodGet
	if _, code := authorizeOAuthBinding(securityContext); code != component.MessageMFAInvalid {
		t.Fatalf("GET OAuth binding bypassed MFA: %s", code)
	}
	bindContext, _ := mfaAPIContext(`{}`, nil)
	bindContext.Set("userId", user.Id)
	bindContext.Set("sessionId", uint64(123))
	if _, code := authorizeOAuthBinding(bindContext); code != component.MessageMFAInvalid {
		t.Fatalf("POST OAuth binding bypassed MFA: %s", code)
	}
	bindContext, _ = mfaAPIContext(fmt.Sprintf(`{"mfaCode":%q}`, codes[9]), nil)
	bindContext.Set("userId", user.Id)
	bindContext.Set("sessionId", uint64(123))
	if authority, code := authorizeOAuthBinding(bindContext); code != "" || authority.SessionID != 123 || authority.TokenVersion != user.TokenVersion {
		t.Fatalf("verified OAuth binding: %+v %s", authority, code)
	}
	c, r := mfaAPIContext("", nil)
	challenge, err := mfaservice.CompleteFirstFactor(c, user.Id, user.TokenVersion, authsessionservice.LoginDetails{Method: "oauth", Provider: "github"}, "/settings")
	if err != nil || !challenge || r.Header().Get("New-Token") != "" {
		t.Fatalf("first factor: %v %v", challenge, err)
	}
	var cookie *http.Cookie
	for _, item := range r.Result().Cookies() {
		if item.Name == "mfa_challenge" {
			cookie = item
		}
	}
	if cookie == nil {
		t.Fatal("missing challenge")
	}
	body := fmt.Sprintf(`{"code":%q}`, codes[0])
	noCookie, noCookieResponse := mfaAPIContext(body, nil)
	MFALogin(noCookie)
	var failure component.ResultStruct
	if err = json.Unmarshal(noCookieResponse.Body.Bytes(), &failure); err != nil {
		t.Fatal(err)
	}
	if failure.Code != component.FAIL || failure.MessageCode != component.MessageMFAInvalid || noCookieResponse.Header().Get("New-Token") != "" {
		t.Fatal("missing challenge accepted")
	}
	valid, validResponse := mfaAPIContext(body, cookie)
	MFALogin(valid)
	var success struct {
		Code   int `json:"code"`
		Result struct {
			Redirect string `json:"redirect"`
		} `json:"result"`
	}
	if err = json.Unmarshal(validResponse.Body.Bytes(), &success); err != nil {
		t.Fatal(err)
	}
	if success.Code != int(component.SUCCESS) || success.Result.Redirect != "/settings" || validResponse.Header().Get("New-Token") == "" {
		t.Fatalf("invalid success: %s", validResponse.Body.String())
	}
	replay, replayResponse := mfaAPIContext(body, cookie)
	MFALogin(replay)
	if replayResponse.Header().Get("New-Token") != "" {
		t.Fatal("replay issued session")
	}
}

func TestMFAAPIRejectsOversizedCode(t *testing.T) {
	c, r := mfaAPIContext(fmt.Sprintf(`{"code":%q}`, strings.Repeat("x", 129)), nil)
	MFALogin(c)
	var result component.ResultStruct
	if err := json.Unmarshal(r.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.MessageCode != component.MessageRequestInvalidParams {
		t.Fatalf("code: %s", result.MessageCode)
	}
}
