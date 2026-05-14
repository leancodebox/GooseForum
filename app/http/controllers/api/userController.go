package api

import (
	"fmt"
	"io"
	"log/slog"
	"strconv"

	"github.com/leancodebox/GooseForum/app/bundles/captchaOpt"
	"github.com/leancodebox/GooseForum/app/http/controllers/transform"
	"github.com/leancodebox/GooseForum/app/models/forum/userActivities"
	"github.com/leancodebox/GooseForum/app/models/forum/userFollow"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/mailservice"
	"github.com/leancodebox/GooseForum/app/service/tokenservice"
	"github.com/leancodebox/GooseForum/app/service/urlconfig"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/algorithm"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

func GetCaptcha(req component.BetterRequest[component.Null]) component.Response {
	captchaId, captchaImg := captchaOpt.GenerateCaptcha()
	return component.SuccessResponse(map[string]any{
		"captchaId":  captchaId,
		"captchaImg": captchaImg,
	})
}

type GetUserCardReq struct {
	UserId uint64 `form:"userId" json:"userId" binding:"required"`
}

func GetUserCard(req component.BetterRequest[GetUserCardReq]) component.Response {
	userId := req.Params.UserId
	userEntity, err := users.Get(userId)
	if err != nil || userEntity.Id == 0 {
		return component.FailResponse("User not found")
	}

	userStats := userStatistics.Get(userId)
	currentUserId := req.UserId
	var isFollowingAuthor bool

	if currentUserId > 0 && currentUserId != userId {
		isFollowingAuthor = userFollow.IsFollowing(currentUserId, userId)
	}

	card := transform.User2UserCard(userEntity, userStats, isFollowingAuthor, currentUserId)

	return component.SuccessResponse(card)
}

// UserInfo returns the current user's profile and statistics.
func UserInfo(req component.BetterRequest[component.Null]) component.Response {
	userEntity, err := req.GetUser()
	if err != nil {
		return component.FailResponse("账号异常" + err.Error())
	}

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
		"externalInformation":  userEntity.ExternalInformation,
		"authorInfoStatistics": authorInfoStatistics,
	})
}

type EditUserEmailReq struct {
	Email string `json:"email" validate:"required,email"`
}

// EditUserEmail updates the current user's email and resets activation state.
func EditUserEmail(req component.BetterRequest[EditUserEmailReq]) component.Response {
	userEntity, err := req.GetUser()
	if err != nil {
		return component.FailResponse("获取用户信息失败")
	}

	newEmail := req.GetParams().Email

	if err := component.ValidateEmailDomain(newEmail); err != nil {
		return component.FailResponse(err.Error())
	}

	if users.ExistEmail(newEmail) {
		return component.FailResponse("邮箱已被使用")
	}
	userEntity.Email = newEmail
	userEntity.IsActivated = users.ActivationPending
	userEntity.ActivatedAt = nil

	err = SaveUser(&userEntity)
	if err != nil {
		return component.FailResponse("更新用户信息失败")
	}

	if err = component.SendAEmail4User(&userEntity); err == nil {
		slog.Info("验证邮件发送失败", "error", err)
	}

	return component.SuccessResponse("更新成功")
}

type GetUserActivitiesReq struct {
	UserId uint64 `form:"userId" validate:"required"`
	LastId uint64 `form:"lastId"`
	Limit  int    `form:"limit" validate:"max=50"`
}

// GetUserActivities returns a user's activity timeline.
func GetUserActivities(req component.BetterRequest[GetUserActivitiesReq]) component.Response {
	limit := req.Params.Limit
	if limit <= 0 {
		limit = 20
	}

	activities, err := userActivities.GetUserTimeline(req.Params.UserId, req.Params.LastId, limit)
	if err != nil {
		return component.FailResponse("获取活动记录失败")
	}

	return component.SuccessResponse(activities)
}

type EditUsernameReq struct {
	Username string `json:"username" validate:"required"`
}

// EditUsername updates the current user's username.
func EditUsername(req component.BetterRequest[EditUsernameReq]) component.Response {
	userEntity, err := req.GetUser()
	if err != nil {
		return component.FailResponse("获取用户信息失败")
	}
	newUsername := req.GetParams().Username
	if !component.ValidateUsername(newUsername) {
		return component.FailResponse("用户名仅允许字母、数字、下划线、连字符，长度6-32")
	}
	if users.ExistUsername(newUsername) {
		return component.FailResponse("用户名已存在")
	}
	userEntity.Username = newUsername
	err = SaveUser(&userEntity)
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

// EditUserInfo updates the current user's profile fields.
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
	userEntity.ExternalInformation = req.Params.ExternalInformation

	err = SaveUser(&userEntity)
	if err != nil {
		return component.FailResponse("更新用户信息失败")
	}
	return component.SuccessResponse("更新成功")
}

func Invitation(req component.BetterRequest[component.Null]) component.Response {
	base36 := strconv.FormatInt(int64(req.UserId), 36)
	return component.SuccessResponse(map[string]any{
		"invitation": base36,
	})
}

// UploadAvatar stores a new avatar for the current user.
func UploadAvatar(c *gin.Context) {
	userId := c.GetUint64("userId")

	if userId == 0 {
		c.JSON(200, component.FailData("未登录"))
		return
	}

	userEntity, err := users.Get(userId)
	if err != nil {
		c.JSON(200, component.FailData("获取用户信息失败"))
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		slog.Error(err.Error())
		c.JSON(200, component.FailData("获取上传文件失败"))
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(200, component.FailData("打开文件失败"))
		return
	}
	defer src.Close()

	fileData, err := io.ReadAll(src)
	if err != nil {
		c.JSON(200, component.FailData("读取文件失败"))
		return
	}

	fileEntity, err := filedata.SaveAvatar(userId, fileData, file.Filename)
	if err != nil {
		c.JSON(200, component.FailData("保存文件失败: "+err.Error()))
		return
	}

	userEntity.AvatarUrl = fileEntity.Name
	if err := SaveUser(&userEntity); err != nil {
		c.JSON(200, component.FailData("更新用户信息失败"))
		return
	}

	c.JSON(200, component.SuccessData(map[string]string{
		"avatarUrl": urlconfig.FilePath(fileEntity.Name),
	}))
}

// ChangePasswordReq is the password change request.
type ChangePasswordReq struct {
	OldPassword string `json:"oldPassword" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required"`
}

// ChangePassword updates the current user's password.
func ChangePassword(req component.BetterRequest[ChangePasswordReq]) component.Response {
	userEntity, err := req.GetUser()
	if err != nil {
		return component.FailResponse("获取用户信息失败")
	}
	if err = component.ValidatePassword(req.Params.NewPassword, 6); err != nil {
		return component.FailResponse(err.Error())
	}
	err = algorithm.VerifyEncryptPassword(userEntity.Password, req.Params.OldPassword)
	if err != nil {
		return component.FailResponse("原密码错误")
	}

	userEntity.SetPassword(req.Params.NewPassword)
	err = SaveUser(&userEntity)
	if err != nil {
		return component.FailResponse("更新密码失败")
	}

	return component.SuccessResponse("密码修改成功")
}

func SaveUser(userEntity *users.EntityComplete) error {
	err := users.Save(userEntity)
	if err == nil {
		if cacheErr := hotdataserve.Reload(fmt.Sprintf("user:%v", userEntity.Id), transform.User2userShow(*userEntity)); cacheErr != nil {
			slog.Error(cacheErr.Error())
		}
	}
	return err
}

// ForgotPasswordReq is the password reset email request.
type ForgotPasswordReq struct {
	Email       string `json:"email" validate:"required,email"`
	CaptchaId   string `json:"captchaId" validate:"required"`
	CaptchaCode string `json:"captchaCode" validate:"required"`
}

// ForgotPassword 忘记密码 - 发送重置邮件
func ForgotPassword(req component.BetterRequest[ForgotPasswordReq]) component.Response {
	if !captchaOpt.VerifyCaptcha(req.Params.CaptchaId, req.Params.CaptchaCode) {
		return component.FailResponse("验证码错误或已过期")
	}

	userEntity, err := users.GetByEmail(req.Params.Email)
	if err != nil {
		// 为了安全考虑，即使邮箱不存在也返回成功消息
		return component.SuccessResponse("操作成功：如果该邮箱已注册，您将收到密码重置邮件")
	}

	token, err := tokenservice.GeneratePasswordResetToken(userEntity.Id, userEntity.Email)
	if err != nil {
		return component.FailResponse("生成重置令牌失败")
	}

	err = mailservice.AddToQueue(mailservice.EmailTask{
		To:       userEntity.Email,
		Username: userEntity.Username,
		Token:    token,
		Type:     "reset_password",
	})
	if err != nil {
		slog.Error("添加密码重置邮件任务到队列失败", "error", err)
		return component.FailResponse("发送重置邮件失败")
	}

	return component.SuccessResponse("操作成功：如果该邮箱已注册，您将收到密码重置邮件")
}

// ResetPasswordReq is the password reset confirmation request.
type ResetPasswordReq struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required"`
}

// ResetPassword 重置密码
func ResetPassword(req component.BetterRequest[ResetPasswordReq]) component.Response {
	claims, err := tokenservice.ParsePasswordResetToken(req.Params.Token)
	if err != nil {
		return component.FailResponse("重置链接已过期或无效")
	}

	userEntity, err := users.Get(claims.UserId)
	if err != nil {
		return component.FailResponse("用户不存在")
	}

	if userEntity.Email != claims.Email {
		return component.FailResponse("重置链接无效")
	}

	if err = component.ValidatePassword(req.Params.NewPassword, 6); err != nil {
		return component.FailResponse(err.Error())
	}

	userEntity.SetPassword(req.Params.NewPassword)
	err = SaveUser(&userEntity)
	if err != nil {
		return component.FailResponse("重置密码失败")
	}

	return component.SuccessResponse("密码重置成功")
}
