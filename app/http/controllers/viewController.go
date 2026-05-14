package controllers

import (
	"context"
	"strings"

	"github.com/leancodebox/GooseForum/app/bundles/captchaOpt"
	"github.com/leancodebox/GooseForum/app/bundles/eventbus"
	jwt "github.com/leancodebox/GooseForum/app/bundles/jwtopt"
	"github.com/leancodebox/GooseForum/app/bundles/logincrypto"
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
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
		c.JSON(200, component.FailData("请求参数格式错误"))
		return
	}
	if err := validate.Valid(r); err != nil {
		c.JSON(200, component.FailData(validate.FormatError(err)))
		return
	}

	securityConfig := hotdataserve.GetSecuritySettingsConfigCache()

	if !securityConfig.EnableSignup {
		c.JSON(200, component.FailData("目前已关闭注册功能"))
		return
	}

	r.Username = strings.TrimSpace(r.Username)
	r.Email = strings.TrimSpace(strings.ToLower(r.Email))

	if err := component.ValidateEmailDomain(r.Email); err != nil {
		c.JSON(200, component.FailData(err.Error()))
		return
	}

	if !component.ValidateUsername(r.Username) {
		c.JSON(200, component.FailData("用户名仅允许字母、数字、下划线、连字符，长度6-32"))
		return
	}

	if err := component.ValidatePassword(r.Password, 6); err != nil {
		c.JSON(200, component.FailData(err.Error()))
		return
	}

	if !captchaOpt.VerifyCaptcha(r.CaptchaId, r.CaptchaCode) {
		c.JSON(200, component.FailData("验证码错误或已过期"))
		return
	}

	if users.ExistUsername(r.Username) {
		c.JSON(200, component.FailData("用户名已存在"))
		return
	}

	if users.ExistEmail(r.Email) {
		c.JSON(200, component.FailData("邮箱已被使用"))
		return
	}

	userEntity, err := userservice.CreateUser(r.Username, r.Password, r.Email, securityConfig.EnableEmailVerification)
	if userEntity == nil || err != nil {
		c.JSON(200, component.FailData("注册失败"))
		return
	}

	if err = component.SendAEmail4User(userEntity); err != nil {
		slog.Error("添加邮件任务到队列失败", "error", err)
	}

	eventbus.Publish(context.Background(), &eventhandlers.UserSignUpEvent{
		UserId:   userEntity.Id,
		Username: userEntity.Username,
	})

	if userEntity.Id == 1 {
		WriteArticles(component.BetterRequest[WriteArticleReq]{
			Params: WriteArticleReq{
				Id:         0,
				Content:    userservice.GetInitBlog(),
				Title:      "Hi With GooseForum",
				Type:       1,
				CategoryId: []uint64{1},
			},
			UserId: userEntity.Id,
		})
	}

	if securityConfig.EnableEmailVerification {
		c.JSON(http.StatusOK, component.SuccessData(
			"注册成功，请前往邮箱验证您的账号",
		))
		return
	}

	token, err := jwt.CreateNewTokenDefault(userEntity.Id)
	if err != nil {

		c.JSON(200, component.FailData("注册异常，尝试登陆"))
	}
	jwt.TokenSetting(c, token)

	c.JSON(http.StatusOK, component.SuccessData(
		"登录成功",
	))
}

type LoginReq struct {
	Username          string `json:"username" validate:"required"` // 可以是用户名或邮箱
	EncryptedPassword string `json:"encryptedPassword" validate:"required"`
	CaptchaId         string `json:"captchaId"`
	CaptchaCode       string `json:"captchaCode"`
}

func LoginPublicKey(c *gin.Context) {
	c.JSON(http.StatusOK, component.SuccessData(map[string]string{
		"publicKey": logincrypto.PublicKeyPEM(),
		"algorithm": "RSA-OAEP-256",
	}))
}

// Login 处理登录请求
func Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, component.FailData("请求参数格式错误"))
		return
	}

	if err := validate.Valid(req); err != nil {
		c.JSON(200, component.FailData("请求参数验证失败"))
		return
	}

	username := strings.TrimSpace(req.Username)
	captchaId := req.CaptchaId
	captchaCode := req.CaptchaCode

	if username == "" {
		c.JSON(200, component.FailData("用户名或邮箱不能为空"))
		return
	}

	password, err := logincrypto.DecryptPassword(req.EncryptedPassword)
	if err != nil {
		slog.Info("登录密码解密失败", "username", username, "error", err)
		c.JSON(200, component.FailData("登录请求无效，请刷新页面后重试"))
		return
	}

	if len(password) < 6 {
		c.JSON(200, component.FailData("密码格式错误"))
		return
	}

	if !captchaOpt.VerifyCaptcha(captchaId, captchaCode) {
		c.JSON(200, component.FailData("验证码错误或已过期"))
		return
	}

	userEntity, err := users.Verify(username, password)
	if err != nil {
		slog.Info("登录失败", "username", username, "error", err)
		c.JSON(200, component.FailData("用户名/邮箱或密码错误"))
		return
	}

	if userEntity.IsFrozen == users.StatusFrozen {
		c.JSON(200, component.FailData("账户已被冻结，请联系管理员"))
		return
	}

	securityConfig := hotdataserve.GetSecuritySettingsConfigCache()
	if securityConfig.EnableEmailVerification && userEntity.IsActivated == users.ActivationPending {
		c.JSON(200, component.FailData("账户邮箱未验证，请先验证您的邮箱"))
		return
	}

	token, err := jwt.CreateNewTokenDefault(userEntity.Id)
	if err != nil {
		slog.Error("生成 token 失败", "userId", userEntity.Id, "error", err)
		c.JSON(200, component.FailData("登录异常，请稍后重试"))
		return
	}

	jwt.TokenSetting(c, token)
	c.JSON(http.StatusOK, component.SuccessData(
		"登录成功",
	))
}
