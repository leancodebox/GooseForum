package api

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/emailactivationservice"
	"github.com/leancodebox/GooseForum/app/service/mailservice"
	"github.com/leancodebox/GooseForum/app/service/registrationservice"
)

// First-factor authentication is required before calling this recovery path.
func resendPendingLogin(c *gin.Context, user users.EntityComplete) component.ResultStruct {
	if err := registrationservice.AllowMail(c.ClientIP(), user.Email); err != nil {
		return component.FailDataCode(component.MessageRegistrationRateLimited, nil)
	}
	if err := mailservice.CheckConfigured(); err != nil {
		return component.FailDataCode(component.MessageAuthActivationResendFailed, nil)
	}
	result, err := emailactivationservice.Resend(user)
	if errors.Is(err, emailactivationservice.ErrCooldown) {
		return component.FailDataCode(component.MessageAuthActivationResendCooldown, component.MessageParams{"retryAfterSeconds": result.RetryAfterSeconds})
	}
	if errors.Is(err, emailactivationservice.ErrDailyLimit) {
		return component.FailDataCode(component.MessageAuthActivationResendDaily, component.MessageParams{"limit": result.DailyLimit})
	}
	if err != nil {
		return component.FailDataCode(component.MessageAuthActivationResendFailed, nil)
	}
	// A queued verification email is not a successful login.
	return component.FailDataCode(component.MessageAuthActivationResendSuccess, component.MessageParams{"remainingToday": result.RemainingToday})
}
