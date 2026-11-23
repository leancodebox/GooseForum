package vo

// RegReq is the user registration request payload.
type RegReq struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"userName"  validate:"required"`
	Password string `json:"passWord"  validate:"required"`
	Locale   string `json:"locale,omitempty"`
	// InvitationCode is optional until the site enables an invitation-only policy.
	InvitationCode string `json:"invitationCode,omitempty"`
	CaptchaToken   string `json:"captchaToken"`
	CaptchaId      string `json:"captchaId"`
	CaptchaCode    string `json:"captchaCode"`
}
