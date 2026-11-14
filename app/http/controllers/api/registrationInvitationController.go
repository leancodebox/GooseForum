package api

import (
	"errors"
	"time"

	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/invitations"
	"github.com/leancodebox/GooseForum/app/service/optlogger"
	"github.com/leancodebox/GooseForum/app/service/registrationinvitationservice"
)

type ListRegistrationInvitationsReq struct {
	Page     int `form:"page"`
	PageSize int `form:"pageSize"`
}

type CreateRegistrationInvitationReq struct {
	MaxUses   int        `json:"maxUses"`
	ExpiresAt *time.Time `json:"expiresAt"`
	Email     string     `json:"email"`
}

type RevokeRegistrationInvitationReq struct {
	ID uint64 `json:"id"`
}

func ListRegistrationInvitations(req component.BetterRequest[ListRegistrationInvitationsReq]) component.Response {
	rows, page, pageSize, total, err := invitations.Page(invitations.PageQuery{Page: req.Params.Page, PageSize: req.Params.PageSize})
	if err != nil {
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}
	return component.SuccessPage(rows, page, pageSize, total)
}

func CreateRegistrationInvitation(req component.BetterRequest[CreateRegistrationInvitationReq]) component.Response {
	entity, raw, emailQueued, err := registrationinvitationservice.Create(req.UserId, req.Params.MaxUses, req.Params.ExpiresAt, req.Params.Email)
	if err != nil {
		if errors.Is(err, invitations.ErrInvalidParameters) || errors.Is(err, registrationinvitationservice.ErrInvalidEmail) || errors.Is(err, registrationinvitationservice.ErrMailUnavailable) {
			return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
		}
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}
	optlogger.UserOpt(req.UserId, optlogger.ManageRegistrationInvitation, entity.Id, "Create registration invitation")
	return component.SuccessResponse(map[string]any{"code": raw, "entry": entity, "emailQueued": emailQueued})
}

func RevokeRegistrationInvitation(req component.BetterRequest[RevokeRegistrationInvitationReq]) component.Response {
	if req.Params.ID == 0 {
		return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
	}
	if err := invitations.Revoke(req.Params.ID, time.Now()); err != nil {
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}
	optlogger.UserOpt(req.UserId, optlogger.ManageRegistrationInvitation, req.Params.ID, "Revoke registration invitation")
	return component.SuccessResponse(true)
}

func RetryRegistrationInvitationEmail(req component.BetterRequest[RevokeRegistrationInvitationReq]) component.Response {
	if req.Params.ID == 0 {
		return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
	}
	if err := registrationinvitationservice.RetryEmail(req.Params.ID); err != nil {
		if errors.Is(err, registrationinvitationservice.ErrInvitationNotFound) || errors.Is(err, registrationinvitationservice.ErrInvitationEmailUnavailable) {
			return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
		}
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}
	optlogger.UserOpt(req.UserId, optlogger.ManageRegistrationInvitation, req.Params.ID, "Retry registration invitation email")
	return component.SuccessResponse(true)
}
