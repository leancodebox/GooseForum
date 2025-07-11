package controllers

import (
	"io"
	"log/slog"
	"strconv"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/captchaOpt"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/service/urlconfig"

	"github.com/leancodebox/GooseForum/app/bundles/algorithm"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/mailservice"
	"github.com/leancodebox/GooseForum/app/service/tokenservice"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
)

type RegReq struct {
	Email          string `json:"email" validate:"required,email"`
	Username       string `json:"userName"  validate:"required"`
	Password       string `json:"passWord"  validate:"required"`
	InvitationCode string `json:"invitationCode,omitempty"`
	CaptchaId      string `json:"captchaId" validate:"required"`
	CaptchaCode    string `json:"captchaCode" validate:"required"`
}

func SendAEmail4User(userEntity *users.Entity) error {
	token, err := tokenservice.GenerateActivationTokenByUser(*userEntity)
	if err != nil {
		return err
	}

	// 将邮件任务加入队列
	err = mailservice.AddToQueue(mailservice.EmailTask{
		To:       userEntity.Email,
		Username: userEntity.Username,
		Token:    token,
		Type:     "activation",
	})
	if err != nil {
		return nil
	}
	return nil
}

func GetCaptcha() component.Response {
	captchaId, captchaImg := captchaOpt.GenerateCaptcha()
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
	avatarUrl := userEntity.GetWebAvatarUrl()
	authorInfoStatistics := userStatistics.Get(userEntity.Id)

	return component.SuccessResponse(component.DataMap{
		"username":             userEntity.Username,
		"userId":               userEntity.Id,
		"avatarUrl":            avatarUrl,
		"email":                userEntity.Email,
		"nickname":             userEntity.Nickname,
		"isAdmin":              userEntity.RoleId != 0,
		"bio":                  userEntity.Bio,
		"signature":            userEntity.Signature,
		"website":              userEntity.Website,
		"websiteName":          userEntity.WebsiteName,
		"externalInformation":  userEntity.GetExternalInformation(),
		"authorInfoStatistics": authorInfoStatistics,
	})
}

type EditUserEmailReq struct {
	Email string `json:"email" validate:"required,email"`
}

// EditUserEmail 编辑用户
func EditUserEmail(req component.BetterRequest[EditUserEmailReq]) component.Response {
	userEntity, err := req.GetUser()
	if err != nil {
		return component.FailResponse("获取用户信息失败")
	}

	newEmail := req.GetParams().Email
	// 如果要修改邮箱,需要检查邮箱是否已被使用
	// 检查邮箱是否已存在
	if users.ExistEmail(newEmail) {
		return component.FailResponse("邮箱已被使用")
	}
	userEntity.Email = newEmail

	err = users.Save(&userEntity)
	if err != nil {
		return component.FailResponse("更新用户信息失败")
	}

	if err = SendAEmail4User(&userEntity); err == nil {
		slog.Info("验证邮件发送失败", "error", err)
	}

	return component.SuccessResponse("更新成功")
}

type EditUsernameReq struct {
	Username string `json:"username" validate:"required"`
}

// EditUsername 编辑用户
func EditUsername(req component.BetterRequest[EditUsernameReq]) component.Response {
	userEntity, err := req.GetUser()
	if err != nil {
		return component.FailResponse("获取用户信息失败")
	}
	newUsername := req.GetParams().Username
	if !ValidateUsername(newUsername) {
		return component.FailResponse("用户名仅允许字母、数字、下划线、连字符，长度6-32")
	}
	// 检查用户名是否已存在
	if users.ExistUsername(newUsername) {
		return component.FailResponse("用户名已存在")
	}
	userEntity.Username = newUsername
	err = users.Save(&userEntity)
	if err != nil {
		return component.FailResponse("更新用户信息失败")
	}

	return component.SuccessResponse("更新成功")
}

type EditUserInfoReq struct {
	Nickname            string                    `json:"nickname"`
	Bio                 string                    `json:"bio"`
	Signature           string                    `json:"signature"`
	Website             string                    `json:"website"`
	WebsiteName         string                    `json:"websiteName"`
	ExternalInformation users.ExternalInformation `json:"externalInformation"`
}

// EditUserInfo 编辑用户
func EditUserInfo(req component.BetterRequest[EditUserInfoReq]) component.Response {
	userEntity, err := req.GetUser()
	if err != nil {
		return component.FailResponse("获取用户信息失败")
	}

	userEntity.Nickname = req.Params.Nickname
	if req.Params.Bio != "" {
		userEntity.Bio = req.Params.Bio
	}
	if req.Params.Signature != "" {
		userEntity.Signature = req.Params.Signature
	}
	userEntity.Website = req.Params.Website
	userEntity.WebsiteName = req.Params.WebsiteName
	userEntity.SetExternalInformation(req.Params.ExternalInformation)

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

type UserInfoShow struct {
	UserId     uint64    `json:"userId,omitempty"`
	Username   string    `json:"username"`
	Bio        string    `json:"bio"`
	Signature  string    `json:"Signature"`
	Prestige   int64     `json:"prestige"`
	AvatarUrl  string    `json:"avatarUrl"`
	UserPoint  int64     `json:"userPoint"`
	CreateTime time.Time `json:"createTime"`
	IsAdmin    bool      `json:"isAdmin"`
}

// UploadAvatar 头像上传处理函数
func UploadAvatar(c *gin.Context) {
	// 从 context 中获取用户 ID
	userId := c.GetUint64("userId")

	if userId == 0 {
		c.JSON(200, component.FailData("未登录"))
		return
	}

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
		"avatarUrl": urlconfig.FilePath(fileEntity.Name),
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
