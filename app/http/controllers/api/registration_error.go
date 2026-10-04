package api

import (
	"errors"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/registrationservice"
)

func registrationError(err error) error {
	code := component.MessageAuthRegisterFailed
	switch {
	case errors.Is(err, registrationservice.ErrClosed):
		code = component.MessageAuthSignupDisabled
	case errors.Is(err, registrationservice.ErrEmailRequired):
		code = component.MessageAuthEmailDomainInvalid
	case errors.Is(err, registrationservice.ErrDomain):
		code = component.MessageAuthEmailDomainNotAllowed
	case errors.Is(err, registrationservice.ErrRateLimited):
		code = component.MessageRegistrationRateLimited
	case errors.Is(err, registrationservice.ErrMailUnavailable):
		code = component.MessageAuthActivationResendFailed
	case errors.Is(err, users.ErrUsernameExists):
		code = component.MessageAuthUsernameExists
	case errors.Is(err, users.ErrEmailExists):
		code = component.MessageAuthEmailExists
	}
	return component.NewMessageError(code, err.Error(), nil)
}
