package api

import (
	"strconv"

	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

type MentionUsersReq struct {
	Query string `form:"query" json:"query"`
}

func MentionUsers(req component.BetterRequest[MentionUsersReq]) component.Response {
	if req.UserId == 0 {
		return component.FailResponseCode(component.MessagePermissionDenied, nil)
	}
	query := req.Params.Query
	if len(query) > 32 {
		return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
	}
	for _, char := range query {
		if !(char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || char == '_' || char == '-') {
			return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
		}
	}
	users, err := users.MentionCandidates(query)
	if err != nil {
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}
	items := make([]map[string]string, 0, len(users))
	for _, user := range users {
		items = append(items, map[string]string{"id": strconv.FormatUint(user.Id, 10), "username": user.Username})
	}
	return component.SuccessResponse(items)
}
