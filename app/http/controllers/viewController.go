package controllers

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"regexp"
	"strings"

	"github.com/leancodebox/GooseForum/app/bundles/captchaOpt"
	jwt "github.com/leancodebox/GooseForum/app/bundles/jwtopt"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/service/userservice"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/validate"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/pointservice"

	"log/slog"
	"net/http"
)

func Logout(c *gin.Context) {
	jwt.TokenClean(c)
	c.JSON(http.StatusOK, component.SuccessData(
		"👋",
	))
}

// ValidatePassword 验证密码复杂度
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("密码长度不能少于8位")
	}
	if len(password) > 128 {
		return fmt.Errorf("密码长度不能超过128位")
	}

	// 检查是否包含数字
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	// 检查是否包含字母
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)

	if !hasDigit || !hasLetter {
		return fmt.Errorf("密码必须包含字母和数字")
	}

	return nil
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

	// 清理输入数据
	r.Username = strings.TrimSpace(r.Username)
	r.Email = strings.TrimSpace(strings.ToLower(r.Email))

	if !component.ValidateUsername(r.Username) {
		c.JSON(200, component.FailData("用户名仅允许字母、数字、下划线、连字符，长度6-32"))
		return
	}

	// 验证密码复杂度
	if err := ValidatePassword(r.Password); err != nil {
		c.JSON(200, component.FailData(err.Error()))
		return
	}

	// 首先验证验证码
	if !captchaOpt.VerifyCaptcha(r.CaptchaId, r.CaptchaCode) {
		c.JSON(200, component.FailData("验证码错误或已过期"))
		return
	}

	// 检查用户名是否已存在
	if users.ExistUsername(r.Username) {
		c.JSON(200, component.FailData("用户名已存在"))
		return
	}

	// 检查邮箱是否已存在
	if users.ExistEmail(r.Email) {
		c.JSON(200, component.FailData("邮箱已被使用"))
		return
	}

	userEntity := users.MakeUser(r.Username, r.Password, r.Email)
	userEntity.Nickname = component.GenerateGooseNickname()
	err := users.Create(userEntity)
	if err != nil {
		c.JSON(200, component.FailData("注册失败"))
	}
	userSt := userStatistics.Entity{UserId: userEntity.Id}
	userStatistics.SaveOrCreateById(&userSt)

	if err = component.SendAEmail4User(userEntity); err != nil {
		slog.Error("添加邮件任务到队列失败", "error", err)
	}

	// 初始化用户积分
	pointservice.InitUserPoints(userEntity.Id, 100)

	if userEntity.Id == 1 {
		// For the first user registered, elevate it to admin group.
		userservice.FirstUserInit(userEntity)
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

	// 生成 token
	token, err := jwt.CreateNewTokenDefault(userEntity.Id)
	if err != nil {

		c.JSON(200, component.FailData("注册异常，尝试登陆"))
	}
	// 设置Cookie
	jwt.TokenSetting(c, token)

	c.JSON(http.StatusOK, component.SuccessData(
		"登录成功",
	))
}

type LoginReq struct {
	Username    string `json:"username" validate:"required"` // 可以是用户名或邮箱
	Password    string `json:"password" validate:"required"`
	CaptchaId   string `json:"captchaId"`
	CaptchaCode string `json:"captchaCode"`
}

// Login 处理登录请求
func Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, component.FailData("请求参数格式错误"))
		return
	}

	// 验证输入参数
	if err := validate.Valid(req); err != nil {
		c.JSON(200, component.FailData("请求参数验证失败"))
		return
	}

	username := strings.TrimSpace(req.Username)
	password := req.Password
	captchaId := req.CaptchaId
	captchaCode := req.CaptchaCode

	// 验证用户名/邮箱格式
	if username == "" {
		c.JSON(200, component.FailData("用户名或邮箱不能为空"))
		return
	}

	// 验证密码长度（登录时只检查最小长度，避免暴露密码策略）
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

	// 检查用户状态
	if userEntity.Status != 0 {
		c.JSON(200, component.FailData("账户已被冻结，请联系管理员"))
		return
	}

	token, err := jwt.CreateNewTokenDefault(userEntity.Id)
	if err != nil {
		slog.Error("生成token失败", "userId", userEntity.Id, "error", err)
		c.JSON(200, component.FailData("登录异常，请稍后重试"))
		return
	}

	jwt.TokenSetting(c, token)
	c.JSON(http.StatusOK, component.SuccessData(
		"登录成功",
	))
}
