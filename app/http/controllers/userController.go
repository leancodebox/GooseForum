package controllers

import (
	"io"
	"log/slog"
	"strconv"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/algorithm"
	jwt "github.com/leancodebox/GooseForum/app/bundles/goose/jwtopt"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/userPoints"
	"github.com/leancodebox/GooseForum/app/models/forum/userRoleRs"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/mailservice"
	"github.com/leancodebox/GooseForum/app/service/pointservice"
	"github.com/leancodebox/GooseForum/app/service/tokenservice"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
	"github.com/spf13/cast"
)

const (
	expireTime = time.Second * 86400 * 7
)

type RegReq struct {
	Email          string `json:"email" validate:"required"`
	Username       string `json:"userName"  validate:"required"`
	Password       string `json:"passWord"  validate:"required"`
	NickName       string `json:"nickName" gorm:"default:'QMPlusUser'"`
	InvitationCode string `json:"invitationCode"`
	CaptchaId      string `json:"captchaId" validate:"required"`
	CaptchaCode    string `json:"captchaCode" validate:"required"`
}

// Register 注册
func Register(r RegReq) component.Response {
	// 首先验证验证码
	if !VerifyCaptcha(r.CaptchaId, r.CaptchaCode) {
		return component.FailResponse("验证码错误或已过期")
	}

	// 检查用户名是否已存在
	if users.ExistUsername(r.Username) {
		return component.FailResponse("用户名已存在")
	}

	// 检查邮箱是否已存在
	if users.ExistEmail(r.Email) {
		return component.FailResponse("邮箱已被使用")
	}

	userEntity := users.MakeUser(r.Username, r.Password, r.Email)
	err := users.Create(userEntity)
	if err != nil {
		return component.FailResponse(cast.ToString(err))
	}

	// 生成激活令牌
	token, err := tokenservice.GenerateActivationToken(userEntity.Id, userEntity.Email)
	if err != nil {
		return component.FailResponse("生成激活令牌失败")
	}

	// 将邮件任务加入队列
	err = mailservice.AddToQueue(mailservice.EmailTask{
		To:       userEntity.Email,
		Username: userEntity.Username,
		Token:    token,
		Type:     "activation",
	})
	if err != nil {
		slog.Error("添加邮件任务到队列失败", "error", err)
		// 不要因为发送邮件失败而阻止注册
	}

	// 初始化用户积分
	pointservice.InitUserPoints(userEntity.Id, 100)

	// 生成 token
	token, err = jwt.CreateNewToken(userEntity.Id, expireTime)
	if err != nil {
		return component.FailResponse(cast.ToString(err))
	}

	return component.SuccessResponse(component.DataMap{
		"username":  userEntity.Username,
		"userId":    userEntity.Id,
		"token":     token,
		"avatarUrl": "",
	})
}

type LoginReq struct {
	Username    string `json:"userName"   validate:"required"`
	Password    string `json:"password"   validate:"required"`
	CaptchaId   string `json:"captchaId"`
	CaptchaCode string `json:"captchaCode"`
}

// Login 登录只返回 token
func Login(r LoginReq) component.Response {
	if !VerifyCaptcha(r.CaptchaId, r.CaptchaCode) {
		return component.FailResponse("验证失败")
	}
	userEntity, err := users.Verify(r.Username, r.Password)
	if err != nil {
		slog.Info(cast.ToString(err))
		return component.FailResponse("验证失败")
	}
	token, err := jwt.CreateNewToken(userEntity.Id, expireTime)
	if err != nil {
		slog.Info(cast.ToString(err))
		return component.FailResponse("验证失败")
	}

	return component.SuccessResponse(component.DataMap{
		"token": token,
	})
}

func GetCaptcha() component.Response {
	captchaId, captchaImg := GenerateCaptcha()
	return component.SuccessResponse(map[string]any{
		"captchaId":  captchaId,
		"captchaImg": captchaImg,
	})
}

type null struct {
}

// UserInfo 获取登录用户信息
func UserInfo(req component.BetterRequest[null]) component.Response {
	userEntity, err := req.GetUser()
	if err != nil {
		return component.FailResponse("账号异常" + err.Error())
	}

	// 处理头像URL
	avatarUrl := ""
	if userEntity.AvatarUrl != "" {
		avatarUrl = component.FilePath(userEntity.AvatarUrl)
	}

	// 检查用户是否有任意角色（是否是管理员）
	isAdmin := len(userRoleRs.GetRoleIdsByUserId(userEntity.Id)) > 0

	return component.SuccessResponse(component.DataMap{
		"username":  userEntity.Username,
		"userId":    userEntity.Id,
		"avatarUrl": avatarUrl,
		"email":     userEntity.Email,
		"nickname":  userEntity.Nickname,
		"isAdmin":   isAdmin,
	})
}

type EditUserInfoReq struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
}

// EditUserInfo 编辑用户
func EditUserInfo(req component.BetterRequest[EditUserInfoReq]) component.Response {
	userEntity, err := req.GetUser()
	if err != nil {
		return component.FailResponse("获取用户信息失败")
	}

	// 如果要修改邮箱,需要检查邮箱是否已被使用
	if req.Params.Email != "" && req.Params.Email != userEntity.Email {
		if users.ExistEmail(req.Params.Email) {
			return component.FailResponse("邮箱已被使用")
		}
		userEntity.Email = req.Params.Email
	}

	// 更新昵称
	if req.Params.Nickname != "" {
		userEntity.Nickname = req.Params.Nickname
	}

	err = users.Save(&userEntity)
	if err != nil {
		return component.FailResponse("更新用户信息失败")
	}

	return component.SuccessResponse("更新成功")
}

func Invitation(req component.BetterRequest[null]) component.Response {
	base36 := strconv.FormatInt(int64(req.UserId), 36)
	return component.SuccessResponse(map[string]any{
		"invitation": base36,
	})
}

type GetUserInfoReq struct {
	UserId uint64 `json:"userId"`
}

// GetUserInfo 游客访问某些用户时
func GetUserInfo(req GetUserInfoReq) component.Response {
	user, _ := users.Get(req.UserId)
	if user.Id == 0 {
		return component.FailResponse("用户不存在")
	}
	userPoint := userPoints.Get(user.Id)

	// 如果有头像，添加域名前缀
	avatarUrl := ""
	if user.AvatarUrl != "" {
		avatarUrl = component.FilePath(user.AvatarUrl)
	}

	return component.SuccessResponse(UserInfoShow{
		Username:  user.Username,
		Prestige:  user.Prestige,
		AvatarUrl: avatarUrl,
		UserPoint: userPoint.CurrentPoints,
	})
}

type UserInfoShow struct {
	Username  string `json:"username"`
	Prestige  int64  `json:"prestige"`
	AvatarUrl string `json:"avatarUrl"`
	UserPoint int64  `json:"userPoint"`
}

// UploadAvatar 头像上传处理函数
func UploadAvatar(c *gin.Context) {
	// 从 context 中获取用户 ID
	userIdData, exists := c.Get("userId")
	if !exists {
		c.JSON(200, component.FailData("未登录"))
		return
	}
	userId := cast.ToUint64(userIdData)

	// 获取用户信息
	userEntity, err := users.Get(userId)
	if err != nil {
		c.JSON(200, component.FailData("获取用户信息失败"))
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(200, component.FailData("获取上传文件失败"))
		return
	}

	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		c.JSON(200, component.FailData("打开文件失败"))
		return
	}
	defer src.Close()

	// 读取文件内容
	fileData, err := io.ReadAll(src)
	if err != nil {
		c.JSON(200, component.FailData("读取文件失败"))
		return
	}

	// 保存头像
	fileEntity, err := filedata.SaveAvatar(userId, fileData, file.Filename)
	if err != nil {
		c.JSON(200, component.FailData("保存文件失败: "+err.Error()))
		return
	}

	// 更新用户头像信息
	userEntity.AvatarUrl = fileEntity.Name
	if err := users.Save(&userEntity); err != nil {
		c.JSON(200, component.FailData("更新用户信息失败"))
		return
	}

	c.JSON(200, component.SuccessData(map[string]string{
		"avatarUrl": component.FilePath(fileEntity.Name),
	}))
}

// 添加新的请求结构体
type ChangePasswordReq struct {
	OldPassword string `json:"oldPassword" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required"`
}

// 添加修改密码的处理函数
func ChangePassword(req component.BetterRequest[ChangePasswordReq]) component.Response {
	userEntity, err := req.GetUser()
	if err != nil {
		return component.FailResponse("获取用户信息失败")
	}

	// 验证旧密码
	err = algorithm.VerifyEncryptPassword(userEntity.Password, req.Params.OldPassword)
	if err != nil {
		return component.FailResponse("原密码错误")
	}

	// 更新密码
	userEntity.SetPassword(req.Params.NewPassword)
	err = users.Save(&userEntity)
	if err != nil {
		return component.FailResponse("更新密码失败")
	}

	return component.SuccessResponse("密码修改成功")
}

// 添加激活处理函数
func ActivateAccount(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		renderActivationPage(c, false, "无效的激活链接")
		return
	}

	// 解析激活令牌
	claims, err := tokenservice.ParseActivationToken(token)
	if err != nil {
		renderActivationPage(c, false, "激活链接已过期或无效")
		return
	}

	// 获取用户信息
	user, err := users.Get(claims.UserId)
	if err != nil {
		renderActivationPage(c, false, "用户不存在")
		return
	}

	// 检查邮箱是否匹配
	if user.Email != claims.Email {
		renderActivationPage(c, false, "激活链接无效")
		return
	}

	// 激活账号
	err = user.Activate()
	if err != nil {
		renderActivationPage(c, false, "激活失败")
		return
	}

	renderActivationPage(c, true, "账号激活成功")
}

// 添加新的辅助函数
func renderActivationPage(c *gin.Context, success bool, message string) {
	status := "失败"
	if success {
		status = "成功"
	}

	html := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>账号激活` + status + ` - Goose Forum</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            background-color: #f0f2f5;
            color: #333;
        }
        .container {
            text-align: center;
            padding: 2.5rem;
            background-color: white;
            border-radius: 12px;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
            max-width: 90%;
            width: 420px;
        }
        .icon {
            font-size: 48px;
            margin-bottom: 1rem;
        }
        .title {
            font-size: 24px;
            font-weight: 600;
            margin: 1rem 0;
            color: #1a1a1a;
        }
        .message {
            margin: 1rem 0;
            line-height: 1.6;
            color: ` + func() string {
		if success {
			return "#4caf50"
		}
		return "#f44336"
	}() + `;
            font-size: 16px;
        }
        .description {
            color: #666;
            margin: 1rem 0;
            font-size: 14px;
            line-height: 1.6;
        }
        .button-group {
            margin-top: 1.5rem;
            display: flex;
            gap: 1rem;
            justify-content: center;
        }
        .home-link {
            display: inline-block;
            padding: 0.75rem 1.5rem;
            background-color: #1890ff;
            color: white;
            text-decoration: none;
            border-radius: 6px;
            font-weight: 500;
            transition: all 0.3s ease;
        }
        .home-link:hover {
            background-color: #096dd9;
            transform: translateY(-1px);
        }
        .login-link {
            display: inline-block;
            padding: 0.75rem 1.5rem;
            background-color: #f0f0f0;
            color: #333;
            text-decoration: none;
            border-radius: 6px;
            font-weight: 500;
            transition: all 0.3s ease;
        }
        .login-link:hover {
            background-color: #d9d9d9;
            transform: translateY(-1px);
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">` + func() string {
		if success {
			return "✅"
		}
		return "❌"
	}() + `</div>
        <h2 class="title">账号激活` + status + `</h2>
        <p class="message">` + message + `</p>
        <p class="description">` + func() string {
		if success {
			return "您的账号已成功激活！现在您可以使用完整的论坛功能，包括发帖、回复、个人中心等服务。"
		}
		return "激活失败可能是因为链接已过期或无效。您可以尝试重新注册或联系管理员获取帮助。"
	}() + `</p>
        <div class="button-group">
            <a href="/" class="home-link">返回首页</a>
            ` + func() string {
		if success {
			return `<a href="/login" class="login-link">立即登录</a>`
		}
		return `<a href="/register" class="login-link">重新注册</a>`
	}() + `
        </div>
    </div>
</body>
</html>`

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, html)
}
