package component

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/transform"
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/mailservice"
	"github.com/leancodebox/GooseForum/app/service/tokenservice"
)

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{6,32}$`)
)

var permissionActionCodes = map[string]string{
	"上传附件": "uploadAttachment",
	"发帖":   "post",
	"评论":   "comment",
}

func ValidateUsername(username string) bool {
	return usernameRegex.MatchString(username)
}

// ValidatePassword 验证密码复杂度
func ValidatePassword(password string, minLength int) error {
	if minLength <= 0 {
		minLength = 8
	}
	if len(password) < minLength {
		return NewMessageError(
			MessageAuthPasswordTooShort,
			fmt.Sprintf("密码长度不能少于%d位", minLength),
			MessageParams{"minLength": minLength},
		)
	}
	if len(password) > 64 {
		return NewMessageError(MessageAuthPasswordTooLong, "密码长度不能超过64位", nil)
	}

	// 检查是否包含数字
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	// 检查是否包含字母
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)

	if !hasDigit || !hasLetter {
		return NewMessageError(MessageAuthPasswordNeedsLetterNumber, "密码必须包含字母和数字", nil)
	}

	return nil
}

func LoginUserId(c *gin.Context) uint64 {
	return c.GetUint64("userId")
}

func GetLoginUser(c *gin.Context) *vo.UserInfoShow {
	userId := LoginUserId(c)
	return GetUserShowByUserId(userId)
}

func GetUserShowByUserId(userId uint64) *vo.UserInfoShow {
	if userId == 0 {
		return &vo.UserInfoShow{}
	}
	return hotdataserve.GetOrLoad(fmt.Sprintf("user:%v", userId), func() (*vo.UserInfoShow, error) {
		user, _ := users.Get(userId)
		if user.Id == 0 {
			return &vo.UserInfoShow{}, errors.New("no found")
		}
		return transform.User2userShow(user), nil
	})
}

func SendAEmail4User(userEntity *users.EntityComplete) error {
	token, err := tokenservice.GenerateActivationTokenByUser(*userEntity)
	if err != nil {
		slog.Debug("生成激活邮件 Token 失败", "userId", userEntity.Id, "email", userEntity.Email, "err", err)
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
		slog.Debug("激活邮件任务入队失败", "userId", userEntity.Id, "email", userEntity.Email, "err", err)
		return err
	}
	slog.Debug("激活邮件任务入队成功", "userId", userEntity.Id, "email", userEntity.Email)
	return nil
}

// CheckUserPermission 统一检查用户操作权限（封禁状态、邮箱验证等）
func CheckUserPermission(userEntity *users.EntityComplete, action string) (error, int) {
	if userEntity == nil || userEntity.Id == 0 {
		return NewMessageError(MessageAuthRequired, "用户不存在或未登录", nil), 401
	}

	// 1. 检查用户是否被冻结
	if userEntity.IsFrozen == users.StatusFrozen {
		return NewMessageError(
			MessagePermissionUserFrozen,
			fmt.Sprintf("您的账号已被封禁，无法进行%s操作", action),
			permissionActionParams(action),
		), 403
	}

	// 2. 检查邮箱验证（如果系统开启了强制要求）
	securityConfig := hotdataserve.GetSecuritySettingsConfigCache()
	if securityConfig.EnableEmailVerification && userEntity.IsActivated == users.ActivationPending {
		return NewMessageError(
			MessagePermissionEmailRequired,
			fmt.Sprintf("请先完成邮箱验证后再进行%s操作", action),
			permissionActionParams(action),
		), 403
	}

	return nil, 200
}

func permissionActionParams(action string) MessageParams {
	params := MessageParams{"action": action}
	if actionCode, ok := permissionActionCodes[action]; ok {
		params["actionCode"] = actionCode
	}
	return params
}

// ValidateEmailDomain 验证邮箱域名是否符合白名单限制
func ValidateEmailDomain(email string) error {
	securityConfig := hotdataserve.GetSecuritySettingsConfigCache()
	if len(securityConfig.AllowedDomains) == 0 {
		return nil
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return NewMessageError(MessageAuthEmailDomainInvalid, "邮箱格式不正确", nil)
	}

	domain := parts[1]
	if slices.Contains(securityConfig.AllowedDomains, domain) {
		return nil
	}

	return NewMessageError(
		MessageAuthEmailDomainNotAllowed,
		"该邮箱域名不在允许的注册白名单中",
		MessageParams{"domain": domain},
	)
}
