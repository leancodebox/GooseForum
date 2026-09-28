package api

import (
	"time"

	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/service/authsessionservice"
)

type sessionView struct {
	Id            uint64    `json:"id"`
	AuthMethod    string    `json:"authMethod"`
	OAuthProvider string    `json:"oauthProvider,omitempty"`
	ClientIP      string    `json:"clientIp"`
	UserAgent     string    `json:"userAgent"`
	CreatedAt     time.Time `json:"createdAt"`
	LastSeenAt    time.Time `json:"lastSeenAt"`
	Current       bool      `json:"current"`
}

func ListAuthSessions(req component.BetterRequest[component.Null]) component.Response {
	user, err := req.GetUser()
	if err != nil || user.Id == 0 {
		return component.FailResponseCode(component.MessageAuthRequired, nil)
	}
	sessions, err := authsessionservice.List(user.Id, user.TokenVersion)
	if err != nil {
		return component.FailResponseCode(component.MessageUserFetchFailed, nil)
	}
	views := make([]sessionView, 0, len(sessions))
	for _, session := range sessions {
		views = append(views, sessionView{
			Id: session.Id, AuthMethod: session.AuthMethod, OAuthProvider: session.OAuthProvider,
			ClientIP: session.ClientIP, UserAgent: session.UserAgent,
			CreatedAt: session.CreatedAt, LastSeenAt: session.LastSeenAt,
			Current: session.Id == req.GinContext.GetUint64("sessionId"),
		})
	}
	return component.SuccessResponse(views)
}

type revokeSessionReq struct {
	Id uint64 `json:"id" validate:"required"`
}

func RevokeAuthSession(req component.BetterRequest[revokeSessionReq]) component.Response {
	if err := authsessionservice.Revoke(req.GinContext, req.UserId, req.Params.Id, req.GinContext.GetUint64("sessionId")); err != nil {
		return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
	}
	return component.SuccessResponse(true)
}

func RevokeOtherAuthSessions(req component.BetterRequest[component.Null]) component.Response {
	if err := authsessionservice.RevokeOthers(req.GinContext, req.UserId, req.GinContext.GetUint64("sessionId")); err != nil {
		return component.FailResponseCode(component.MessageUserFetchFailed, nil)
	}
	return component.SuccessResponse(true)
}
