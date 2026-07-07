package controllers

import (
	"context"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/captchaOpt"
	"github.com/leancodebox/GooseForum/app/bundles/eventbus"
	jwt "github.com/leancodebox/GooseForum/app/bundles/jwtopt"
	"github.com/leancodebox/GooseForum/app/bundles/logincrypto"
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/service/emailactivationservice"
	"github.com/leancodebox/GooseForum/app/service/eventhandlers"
	"github.com/leancodebox/GooseForum/app/service/userservice"

	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/validate"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
)

func Logout(c *gin.Context) {
	jwt.TokenClean(c)
	c.JSON(http.StatusOK, component.SuccessData(
		"👋",
	))
}

// Register 注册
func Register(c *gin.Context) {
	var r vo.RegReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(200, component.FailDataCode(component.MessageRequestInvalidFormat, nil))
		return
	}
	if err := validate.Valid(r); err != nil {
		c.JSON(200, component.FailDataCode(component.MessageRequestInvalidParams, nil))
		return
	}

	securityConfig := hotdataserve.GetSecuritySettingsConfigCache()

	if !securityConfig.EnableSignup {
		c.JSON(200, component.FailDataCode(component.MessageAuthSignupDisabled, nil))
		return
	}

	r.Username = strings.TrimSpace(r.Username)
	r.Email = strings.TrimSpace(strings.ToLower(r.Email))

	if err := component.ValidateEmailDomain(r.Email); err != nil {
		c.JSON(200, component.FailDataError(err))
		return
	}

	if !component.ValidateUsername(r.Username) {
		c.JSON(200, component.FailDataCode(component.MessageAuthUsernameInvalid, nil))
		return
	}

	if err := component.ValidatePassword(r.Password, 6); err != nil {
		c.JSON(200, component.FailDataError(err))
		return
	}

	if !captchaOpt.VerifyCaptcha(r.CaptchaId, r.CaptchaCode) {
		c.JSON(200, component.FailDataCode(component.MessageAuthCaptchaInvalid, nil))
		return
	}

	if users.ExistUsername(r.Username) {
		c.JSON(200, component.FailDataCode(component.MessageAuthUsernameExists, nil))
		return
	}

	if users.ExistEmail(r.Email) {
		c.JSON(200, component.FailDataCode(component.MessageAuthEmailExists, nil))
		return
	}

	userEntity, err := userservice.CreateUser(r.Username, r.Password, r.Email, true, r.Locale)
	if userEntity == nil || err != nil {
		slog.Error("注册创建用户失败", "username", r.Username, "email", r.Email, "error", err)
		c.JSON(200, component.FailDataCode(component.MessageAuthRegisterFailed, nil))
		return
	}

	slog.Debug("注册用户创建成功", "userId", userEntity.Id, "username", userEntity.Username, "email", userEntity.Email, "enableEmailVerification", securityConfig.EnableEmailVerification)
	if err = emailactivationservice.SendActivationEmail(userEntity); err != nil {
		slog.Error("添加邮件任务到队列失败", "userId", userEntity.Id, "email", userEntity.Email, "error", err)
	} else {
		slog.Debug("注册激活邮件任务已提交", "userId", userEntity.Id, "email", userEntity.Email, "enableEmailVerification", securityConfig.EnableEmailVerification)
	}

	eventbus.Publish(context.Background(), &eventhandlers.UserSignUpEvent{
		UserId:   userEntity.Id,
		Username: userEntity.Username,
	})

	if userEntity.Id == 1 {
		WriteTopic(component.BetterRequest[WriteTopicReq]{
			Params: WriteTopicReq{
				Content:       userservice.GetWelcomeArticleContent(),
				Title:         "Hi With GooseForum",
				Type:          1,
				CategoryId:    []uint64{1},
				ArticleStatus: 1,
			},
			UserId: userEntity.Id,
		})
	}

	token, err := jwt.CreateNewTokenDefaultWithVersion(userEntity.Id, userEntity.TokenVersion)
	if err != nil {
		c.JSON(200, component.FailDataCode(component.MessageAuthRegisterRetryLogin, nil))
		return
	}
	jwt.TokenSetting(c, token)

	if securityConfig.EnableEmailVerification {
		c.JSON(http.StatusOK, component.SuccessDataCode(
			"注册成功，请前往邮箱验证您的账号",
			component.MessageAuthRegisterEmailVerify,

			nil))
		return
	}

	c.JSON(http.StatusOK, component.SuccessDataCode("登录成功", component.MessageAuthLoginSuccess, nil))
}

type LoginReq struct {
	Username          string `json:"username" validate:"required"` // 可以是用户名或邮箱
	EncryptedPassword string `json:"encryptedPassword" validate:"required"`
	CaptchaId         string `json:"captchaId"`
	CaptchaCode       string `json:"captchaCode"`
}

func LoginPublicKey(c *gin.Context) {
	c.JSON(http.StatusOK, component.SuccessData(map[string]any{
		"publicKey": logincrypto.PublicKeyPEM(),
		"serverTs":  time.Now().UnixMilli(),
		"algorithm": "RSA-OAEP-256",
	}))
}

// Login 处理登录请求
func Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, component.FailDataCode(component.MessageRequestInvalidFormat, nil))
		return
	}

	if err := validate.Valid(req); err != nil {
		c.JSON(200, component.FailDataCode(component.MessageRequestInvalidParams, nil))
		return
	}

	username := strings.TrimSpace(req.Username)
	captchaId := req.CaptchaId
	captchaCode := req.CaptchaCode

	if username == "" {
		c.JSON(200, component.FailDataCode(component.MessageRequestInvalidParams, nil))
		return
	}

	password, err := logincrypto.DecryptPassword(req.EncryptedPassword)
	if err != nil {
		slog.Info("登录密码解密失败", "username", username, "error", err)
		c.JSON(200, component.FailDataCode(component.MessageAuthLoginInvalidRequest, nil))
		return
	}

	if len(password) < 6 {
		c.JSON(200, component.FailDataCode(component.MessageAuthPasswordInvalidFormat, nil))
		return
	}

	if !captchaOpt.VerifyCaptcha(captchaId, captchaCode) {
		c.JSON(200, component.FailDataCode(component.MessageAuthCaptchaInvalid, nil))
		return
	}

	userEntity, err := users.Verify(username, password)
	if err != nil {
		slog.Info("登录失败", "username", username, "error", err)
		c.JSON(200, component.FailDataCode(component.MessageAuthInvalidCredentials, nil))
		return
	}

	securityConfig := hotdataserve.GetSecuritySettingsConfigCache()
	if securityConfig.EnableEmailVerification && userEntity.IsActivated == users.ActivationPending {
		c.JSON(200, component.FailDataCode(component.MessageAuthEmailUnverified, nil))
		return
	}

	token, err := jwt.CreateNewTokenDefaultWithVersion(userEntity.Id, userEntity.TokenVersion)
	if err != nil {
		slog.Error("生成 token 失败", "userId", userEntity.Id, "error", err)
		c.JSON(200, component.FailDataCode(component.MessageAuthLoginFailed, nil))
		return
	}

	jwt.TokenSetting(c, token)
	c.JSON(http.StatusOK, component.SuccessDataCode("登录成功", component.MessageAuthLoginSuccess, nil))
}
