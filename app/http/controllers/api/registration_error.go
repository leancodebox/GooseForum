package api

import (
	"errors"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/invitations"
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
	case errors.Is(err, invitations.ErrRequired):
		code = component.MessageAuthInvitationRequired
	case errors.Is(err, invitations.ErrNotFound), errors.Is(err, invitations.ErrExpired), errors.Is(err, invitations.ErrRevoked), errors.Is(err, invitations.ErrExhausted), errors.Is(err, invitations.ErrReserved), errors.Is(err, invitations.ErrReservationExpired):
		code = component.MessageAuthInvitationInvalid
	}
	return component.NewMessageError(code, err.Error(), nil)
}
