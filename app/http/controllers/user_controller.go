package controllers

import (
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/Users"
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

// Register
// @todo user表增加验证字段
// 创建后验证码存入redis，发认证送邮件。
// 邮件 附有 url?code=xxx
// 验证后更新验证字段
// 清除验证码
func Register(r RegReq) component.Response {
	userEntity := Users.MakeUser(r.Username, r.Password, r.Email)
	err := Users.Create(userEntity)

	if err != nil {
		return component.FailResponse(cast.ToString(err))
	}

	token, err := jwt.CreateNewToken(userEntity.Id, expireTime)
	if err != nil {
		return component.FailResponse(cast.ToString(err))
	}
	return component.SuccessResponse(component.DataMap{
		"message": "ok",
		"token":   token,
	})
}

type LoginReq struct {
	Username    string `json:"userName"   validate:"required"`
	Password    string `json:"password"   validate:"required"`
	CaptchaId   string `json:"captchaId"`
	CaptchaCode string `json:"captchaCode"`
}

func Login(r LoginReq) component.Response {
	if VerifyCaptcha(r.CaptchaId, r.CaptchaCode) {
		return component.SuccessResponse("ok")
	}
	userEntity, err := Users.Verify(r.Username, r.Password)
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
		"message": "ok",
		"token":   token,
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

func EditUserInfo(req component.BetterRequest[EditUserInfoReq]) component.Response {
	return component.SuccessResponse("success")
}

func Invitation(req component.BetterRequest[null]) component.Response {
	base36 := strconv.FormatInt(int64(req.UserId), 36)
	return component.SuccessResponse(map[string]any{
		"invitation": base36,
	})
}
