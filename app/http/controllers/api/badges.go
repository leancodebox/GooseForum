package api

import (
	"errors"
	"strings"

	"github.com/leancodebox/GooseForum/app/bundles/pageutil"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/badgeStats"
	"github.com/leancodebox/GooseForum/app/models/forum/userBadges"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/badgeservice"
	"github.com/leancodebox/GooseForum/app/service/userservice"
)

type BadgeSaveReq = badgeservice.Badge

func BadgeList(req component.BetterRequest[component.Null]) component.Response {
	items, err := badgeservice.ListForAdmin()
	if err != nil {
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}
	return component.SuccessResponse(items)
}

func badgeFailure(err error, fallback component.MessageCode) component.Response {
	var field *badgeservice.InvalidFieldError
	if errors.As(err, &field) {
		return component.FailResponseCode(component.MessageRequestInvalidParams, component.MessageParams{"field": field.Field})
	}
	switch {
	case errors.Is(err, badgeservice.ErrBadgeInUse):
		return component.FailResponseCode(component.MessageAdminBadgeInUse, nil)
	case errors.Is(err, badgeservice.ErrSystemBadge):
		return component.FailResponseCode(component.MessageAdminBadgeSystemDeleteBlock, nil)
	case errors.Is(err, badgeservice.ErrRecipientNotFound):
		return component.FailResponseCode(component.MessageUserNotFound, nil)
	case errors.Is(err, badgeservice.ErrBadgeNotFound):
		return component.FailResponseCode(component.MessageRequestInvalidParams, component.MessageParams{"field": "code"})
	default:
		return component.FailResponseCode(fallback, nil)
	}
}

func SaveBadge(req component.BetterRequest[BadgeSaveReq]) component.Response {
	if err := badgeservice.SaveDefinition(req.Params); err != nil {
		return badgeFailure(err, component.MessageAdminBadgeSaveFailed)
	}
	userservice.ClearUserPublicProfileCache()
	return component.SuccessResponseCode("success", component.MessageOperationSuccess, nil)
}

type BadgeDeleteReq struct {
	Code string `json:"code"`
}

func DeleteBadge(req component.BetterRequest[BadgeDeleteReq]) component.Response {
	code := strings.TrimSpace(req.Params.Code)
	if code == "" {
		return component.FailResponseCode(component.MessageAdminBadgeCodeRequired, nil)
	}
	if err := badgeservice.DeleteDefinition(code); err != nil {
		return badgeFailure(err, component.MessageAdminBadgeDeleteFailed)
	}
	userservice.ClearUserPublicProfileCache()
	return component.SuccessResponseCode("success", component.MessageOperationSuccess, nil)
}

type UserBadgeOptionsReq struct {
	UserId uint64 `json:"userId"`
}
type UserBadgeOptionsResp struct {
	Options []badgeservice.Badge     `json:"options"`
	Active  []badgeservice.UserBadge `json:"active"`
}

func UserBadgeOptions(req component.BetterRequest[UserBadgeOptionsReq]) component.Response {
	return component.SuccessResponse(UserBadgeOptionsResp{Options: badgeservice.ManualGrantBadgesForAdmin(), Active: badgeservice.GetUserBadges(req.Params.UserId)})
}

type SaveUserBadgesReq struct {
	UserId     uint64   `json:"userId"`
	BadgeCodes []string `json:"badgeCodes"`
}

func SaveUserBadges(req component.BetterRequest[SaveUserBadgesReq]) component.Response {
	// A multi-selection can partially succeed; always evict affected user views.
	defer userservice.InvalidateUserPublicProfileCache(req.Params.UserId)
	if err := badgeservice.SaveManualSelection(req.Params.UserId, req.UserId, req.Params.BadgeCodes); err != nil {
		return badgeFailure(err, component.MessageUserUpdateFailed)
	}
	return component.SuccessResponseCode("success", component.MessageOperationSuccess, nil)
}

type BadgeRecipientsReq struct {
	Code     string `json:"code"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Search   string `json:"search"`
}

type BadgeRecipientsPage struct {
	List    []userBadges.Recipient `json:"list"`
	Page    int                    `json:"page"`
	Size    int                    `json:"size"`
	HasNext bool                   `json:"hasNext"`
	Total   *int64                 `json:"total,omitempty"`
}

func BadgeRecipients(req component.BetterRequest[BadgeRecipientsReq]) component.Response {
	code := strings.TrimSpace(req.Params.Code)
	if badgeservice.ResolveOne(code).Code == "" {
		return badgeFailure(badgeservice.ErrBadgeNotFound, component.MessageOperationFailed)
	}
	page := max(1, req.Params.Page)
	size := pageutil.NormalizePageSize(req.Params.PageSize, 20, 100)
	search := strings.TrimSpace(req.Params.Search)
	list, hasNext, err := userBadges.ListRecipients(code, page, size, search)
	if err != nil {
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}
	for i := range list {
		user := users.EntityComplete{AvatarUrl: list[i].AvatarUrl, RestrictionStatus: list[i].RestrictionStatus, RestrictionUntil: list[i].RestrictionUntil}
		list[i].AvatarUrl = user.GetWebAvatarUrl()
	}
	result := BadgeRecipientsPage{List: list, Page: page, Size: size, HasNext: hasNext}
	if search == "" {
		stat, err := badgeStats.Get(code)
		if err != nil {
			return component.FailResponseCode(component.MessageOperationFailed, nil)
		}
		result.Total = &stat.GrantedCount
	}
	return component.SuccessResponse(result)
}

type UpdateBadgeRecipientReq struct {
	Code   string `json:"code"`
	UserId uint64 `json:"userId"`
	Action string `json:"action"`
	Reason string `json:"reason"`
}

func UpdateBadgeRecipient(req component.BetterRequest[UpdateBadgeRecipientReq]) component.Response {
	if err := badgeservice.UpdateRecipient(strings.TrimSpace(req.Params.Code), req.Params.UserId, req.Params.Action, req.Params.Reason, req.UserId); err != nil {
		return badgeFailure(err, component.MessageUserUpdateFailed)
	}
	userservice.InvalidateUserPublicProfileCache(req.Params.UserId)
	return component.SuccessResponseCode("success", component.MessageOperationSuccess, nil)
}
