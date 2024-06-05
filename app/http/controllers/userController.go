package controllers

import (
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/userPoints"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/pointservice"
	jwt "github.com/leancodebox/goose/jwtopt"
	"log/slog"
	"strconv"
	"time"

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
}

// Register 注册
func Register(r RegReq) component.Response {
	userEntity := users.MakeUser(r.Username, r.Password, r.Email)
	err := users.Create(userEntity)
	if err != nil {
		return component.FailResponse(cast.ToString(err))
	}
	pointservice.InitUserPoints(userEntity.Id, 100)
	token, err := jwt.CreateNewToken(userEntity.Id, expireTime)
	if err != nil {
		return component.FailResponse(cast.ToString(err))
	}
	return component.SuccessResponse(component.DataMap{
		"username": userEntity.Username,
		"token":    token,
	})
}

type LoginReq struct {
	Username    string `json:"userName"   validate:"required"`
	Password    string `json:"password"   validate:"required"`
	CaptchaId   string `json:"captchaId"`
	CaptchaCode string `json:"captchaCode"`
}

func Login(r LoginReq) component.Response {
	if !VerifyCaptcha(r.CaptchaId, r.CaptchaCode) {
		return component.SuccessResponse("ok")
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
		"username": userEntity.Username,
		"token":    token,
	})
}

func GetCaptcha() component.Response {
	captchaId, captchaImg := GenerateCaptcha()
	return component.SuccessResponse(map[string]any{
		"captchaId":  captchaId,
		"captchaImg": captchaImg,
	})
}

func GenerateCaptcha() (string, string) {
	return "", ""
}

func VerifyCaptcha(captchaId, captchaCode string) bool {
	return true
}

type null struct {
}

func UserInfo(req component.BetterRequest[null]) component.Response {
	userEntity, err := req.GetUser()
	if err != nil {
		return component.FailResponse("账号异常" + err.Error())
	}
	return component.SuccessResponse(userEntity)
}

type EditUserInfoReq struct {
	Nickname string `json:"nickname"`
}

// EditUserInfo 编辑用户
func EditUserInfo(req component.BetterRequest[EditUserInfoReq]) component.Response {
	return component.SuccessResponse("success")
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
	return component.SuccessResponse(UserInfoShow{
		Username:  user.Username,
		AvatarUrl: "",
		UserPoint: userPoint.CurrentPoints,
	})
}

type UserInfoShow struct {
	Username  string `json:"username"`
	AvatarUrl string `json:"avatarUrl"`
	UserPoint int64  `json:"userPoint"`
}
