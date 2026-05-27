package api

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/captchaOpt"
	"github.com/leancodebox/GooseForum/app/http/controllers/transform"
	"github.com/leancodebox/GooseForum/app/models/forum/userFollow"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/mailservice"
	"github.com/leancodebox/GooseForum/app/service/tokenservice"
	"github.com/leancodebox/GooseForum/app/service/urlconfig"
	"github.com/leancodebox/GooseForum/app/service/usercardservice"

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
	card, ok := usercardservice.GetCard(userId)
	if !ok {
		return component.FailResponse("User not found")
	}
	currentUserId := req.UserId
	card.IsSelf = currentUserId == userId
	card.IsFollowing = false
	if currentUserId > 0 && currentUserId != userId {
		card.IsFollowing = userFollow.IsFollowing(currentUserId, userId)
	}

	return component.SuccessResponse(card)
}

func GetUserHoverCard(req component.BetterRequest[GetUserCardReq]) component.Response {
	userId := req.Params.UserId
	card, ok := usercardservice.GetHoverCard(userId)
	if !ok {
		return component.FailResponse("User not found")
	}
	currentUserId := req.UserId
	card.IsFollowing = false
	if currentUserId > 0 && currentUserId != userId {
		card.IsFollowing = userFollow.IsFollowing(currentUserId, userId)
	}

	return component.SuccessResponse(card)
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

	if err = component.SendAEmail4User(&userEntity); err != nil {
		slog.Info("验证邮件发送失败", "error", err)
	}

	return component.SuccessResponse("更新成功")
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

// UploadAvatar stores a new avatar for the current user.
func UploadAvatar(c *gin.Context) {
	postingConfig := hotdataserve.GetPostingSettingsConfigCache()
	if !postingConfig.UploadControl.AllowAttachments {
		c.JSON(200, component.FailData("目前已关闭附件上传功能"))
		return
	}

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

	if err, code := component.CheckUserPermission(&userEntity, "上传附件"); err != nil {
		c.JSON(code, component.FailData(err.Error()))
		return
	}

	if postingConfig.UploadControl.NewUserUploadCooldownMinutes > 0 {
		cooldownTime := userEntity.CreatedAt.Add(time.Duration(postingConfig.UploadControl.NewUserUploadCooldownMinutes) * time.Minute)
		if time.Now().Before(cooldownTime) {
			c.JSON(200, component.FailData(fmt.Sprintf("新用户注册%d分钟后才能上传，请在 %s 后再试",
				postingConfig.UploadControl.NewUserUploadCooldownMinutes,
				cooldownTime.Format("2006-01-02 15:04:05"))))
			return
		}
	}

	files, err := avatarFormFiles(c)
	if err != nil {
		slog.Error(err.Error())
		c.JSON(200, component.FailData("获取上传文件失败"))
		return
	}

	fileCount := files.Count()
	if postingConfig.UploadControl.MaxDailyUploadsPerUser > 0 {
		count := filedata.CountDailyUploads(userId)
		if count+int64(fileCount) > int64(postingConfig.UploadControl.MaxDailyUploadsPerUser) {
			c.JSON(200, component.FailData(fmt.Sprintf("您今日已上传 %d 个文件，上传头像需要 %d 个名额，已超过每日限制", count, fileCount)))
			return
		}
	}

	maxSize := avatarUploadMaxSize()
	if configMaxSize := int64(postingConfig.UploadControl.MaxAttachmentSizeKb) * 1024; configMaxSize > 0 && configMaxSize < maxSize {
		maxSize = configMaxSize
	}
	allowedExts := postingConfig.UploadControl.AuthorizedExtensions

	mainData, err := readAvatarUploadFile(files.Main, maxSize, allowedExts)
	if err != nil {
		c.JSON(200, component.FailData(err.Error()))
		return
	}

	var fileEntities []*filedata.Entity
	if files.AvatarMedium == nil {
		fileEntity, err := filedata.SaveAvatar(userId, mainData, files.Main.Filename)
		if err != nil {
			c.JSON(200, component.FailData("保存文件失败: "+err.Error()))
			return
		}
		fileEntities = []*filedata.Entity{fileEntity}
	} else {
		uploads := []filedata.AvatarUpload{{
			Filename: files.Main.Filename,
			Data:     mainData,
		}}
		fileData, err := readAvatarUploadFile(files.AvatarMedium, maxSize, allowedExts)
		if err != nil {
			c.JSON(200, component.FailData(err.Error()))
			return
		}
		uploads = append(uploads, filedata.AvatarUpload{
			Filename: files.AvatarMedium.Filename,
			Data:     fileData,
		})

		fileEntities, err = filedata.SaveAvatarSet(userId, uploads)
		if err != nil {
			c.JSON(200, component.FailData("保存文件失败: "+err.Error()))
			return
		}
	}

	userEntity.AvatarUrl = fileEntities[0].Name
	if err := SaveUser(&userEntity); err != nil {
		c.JSON(200, component.FailData("更新用户信息失败"))
		return
	}

	response := map[string]string{
		"avatarUrl": urlconfig.FilePath(fileEntities[0].Name),
	}
	if len(fileEntities) > 1 {
		response["avatarMediumUrl"] = urlconfig.FilePath(fileEntities[1].Name)
	}
	c.JSON(200, component.SuccessData(response))
}

type avatarUploadFiles struct {
	Main         *multipart.FileHeader
	AvatarMedium *multipart.FileHeader
}

func (files avatarUploadFiles) Count() int {
	count := 0
	for _, file := range []*multipart.FileHeader{files.Main, files.AvatarMedium} {
		if file != nil {
			count++
		}
	}
	return count
}

func avatarFormFiles(c *gin.Context) (avatarUploadFiles, error) {
	main, err := c.FormFile("avatar")
	if err != nil {
		return avatarUploadFiles{}, err
	}

	files := avatarUploadFiles{Main: main}
	files.AvatarMedium, _ = c.FormFile("avatarMedium")
	return files, nil
}

func avatarUploadMaxSize() int64 {
	return int64(filedata.MaxFileSize)
}

func readAvatarUploadFile(file *multipart.FileHeader, maxSize int64, allowedExts []string) ([]byte, error) {
	if file.Filename == "" {
		return nil, fmt.Errorf("文件名不能为空")
	}
	if file.Size > maxSize {
		return nil, fmt.Errorf("文件大小超过限制，最大允许%dKB", maxSize/1024)
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if len(allowedExts) > 0 {
		if !isAllowedExtension(ext, allowedExts) {
			return nil, fmt.Errorf("不支持的文件格式，允许的格式为: %s", strings.Join(allowedExts, ", "))
		}
	} else if _, err := filedata.CheckImageType(file.Filename); err != nil {
		return nil, fmt.Errorf("不支持的图片格式，仅支持 JPG、PNG、GIF、WebP、BMP 格式")
	}

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("打开文件失败")
	}
	defer src.Close()

	header := make([]byte, 512)
	n, _ := io.ReadFull(src, header)
	if n > 0 && !isValidImageContent(header[:n]) {
		return nil, fmt.Errorf("文件内容不是有效的图片格式")
	}

	remainingData, err := io.ReadAll(io.LimitReader(src, maxSize-int64(n)+1))
	if err != nil {
		return nil, fmt.Errorf("读取文件失败")
	}
	fileData := append(bytes.Clone(header[:n]), remainingData...)
	if int64(len(fileData)) > maxSize {
		return nil, fmt.Errorf("文件大小超过限制，最大允许%dKB", maxSize/1024)
	}
	return fileData, nil
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
		usercardservice.Invalidate(userEntity.Id)
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
