package controllers

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/setting"

	jwt "github.com/leancodebox/GooseForum/app/bundles/goose/jwtopt"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/userPoints"
	"github.com/leancodebox/GooseForum/app/models/forum/userRoleRs"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/pointservice"

	"github.com/gin-gonic/gin"
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

	// 初始化用户积分
	pointservice.InitUserPoints(userEntity.Id, 100)

	// 生成 token
	token, err := jwt.CreateNewToken(userEntity.Id, expireTime)
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
		avatarUrl = "/api" + userEntity.AvatarUrl
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
		avatarUrl = "/api" + user.AvatarUrl
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

	// 验证文件类型
	if !isValidImageType(file.Header.Get("Content-Type")) {
		c.JSON(200, component.FailData("不支持的文件类型"))
		return
	}

	// 验证文件大小（2MB）
	if file.Size > 2*1024*1024 {
		c.JSON(200, component.FailData("文件大小不能超过2MB"))
		return
	}

	// 生成文件名
	filename := fmt.Sprintf("avatar_%d_%d%s",
		userId,
		time.Now().Unix(),
		path.Ext(file.Filename))

	// 保存路径
	avatarPath := path.Join(setting.GetStorage(), "avatars", filename)

	// 确保目录存在
	if err := os.MkdirAll(path.Dir(avatarPath), 0755); err != nil {
		c.JSON(200, component.FailData("创建目录失败"))
		return
	}

	// 保存文件
	if err := c.SaveUploadedFile(file, avatarPath); err != nil {
		c.JSON(200, component.FailData("保存文件失败"))
		return
	}

	// 更新用户头像信息
	avatarUrl := "/avatars/" + filename
	userEntity.AvatarUrl = avatarUrl
	if err := users.Save(&userEntity); err != nil {
		c.JSON(200, component.FailData("更新用户信息失败"))
		return
	}

	c.JSON(200, component.SuccessData(map[string]string{
		"avatarUrl": "/api" + avatarUrl,
	}))
}

func isValidImageType(contentType string) bool {
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
	}
	return validTypes[contentType]
}
