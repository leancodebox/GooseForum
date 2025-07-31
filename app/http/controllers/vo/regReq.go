package vo

type RegReq struct {
	Email          string `json:"email" validate:"required,email"`
	Username       string `json:"userName"  validate:"required"`
	Password       string `json:"passWord"  validate:"required"`
	InvitationCode string `json:"invitationCode,omitempty"`
	CaptchaId      string `json:"captchaId" validate:"required"`
	CaptchaCode    string `json:"captchaCode" validate:"required"`
}
