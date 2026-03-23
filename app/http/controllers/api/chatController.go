package api

import (
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/userFollow"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/chatservice"
	"github.com/samber/lo"
)

// SendMessageReq 发送私信请求
type SendMessageReq struct {
	PeerId  uint64 `json:"peerId" validate:"required"`
	Content string `json:"content" validate:"required"`
	MsgType int8   `json:"msgType" validate:"oneof=1 2 3"` // 1: Text, 2: Image, 3: Voice
}

// SendMessage 发送私信
func SendMessage(req component.BetterRequest[SendMessageReq]) component.Response {
	// Set default msg type to text if not provided or 0 (though validate should handle it if required, let's assume default 1)
	msgType := req.Params.MsgType
	if msgType == 0 {
		msgType = 1
	}

	convId, err := chatservice.SendMessage(req.UserId, req.Params.PeerId, req.Params.Content, msgType)
	if err != nil {
		return component.FailResponse(err.Error())
	}
	return component.SuccessResponse(component.DataMap{
		"convId": convId,
	})
}

// GetChatListReq 获取私信列表请求
type GetChatListReq struct {
}

// GetChatList 获取私信列表
func GetChatList(req component.BetterRequest[GetChatListReq]) component.Response {
	list, err := chatservice.GetChatList(req.UserId)
	if err != nil {
		return component.FailResponse("Failed to get chat list")
	}
	return component.SuccessResponse(component.DataMap{
		"list": list,
	})
}

// GetMessagesReq 获取消息记录请求
type GetMessagesReq struct {
	ConvId   uint64 `json:"convId" validate:"required"`
	Page     int    `json:"page" validate:"required,min=1"`
	PageSize int    `json:"pageSize" validate:"required,min=1,max=100"`
}

// GetMessages 获取消息记录
func GetMessages(req component.BetterRequest[GetMessagesReq]) component.Response {
	msgs, err := chatservice.GetMessages(req.UserId, req.Params.ConvId, req.Params.Page, req.Params.PageSize)
	if err != nil {
		return component.FailResponse("Failed to get messages")
	}
	return component.SuccessResponse(component.DataMap{
		"list": msgs,
	})
}

// MarkReadReq 标记已读请求
type MarkReadReq struct {
	ConvId uint64 `json:"convId" validate:"required"`
}

// MarkChatRead 标记已读
func MarkChatRead(req component.BetterRequest[MarkReadReq]) component.Response {
	err := chatservice.MarkRead(req.UserId, req.Params.ConvId)
	if err != nil {
		return component.FailResponse("Failed to mark read")
	}
	return component.SuccessResponse(nil)
}

// DeleteChatReq 删除对话请求
type DeleteChatReq struct {
	ConvId uint64 `json:"convId" validate:"required"`
}

// DeleteChat 删除对话
func DeleteChat(req component.BetterRequest[DeleteChatReq]) component.Response {
	err := chatservice.DeleteChat(req.UserId, req.Params.ConvId)
	if err != nil {
		return component.FailResponse("Failed to delete chat")
	}
	return component.SuccessResponse(nil)
}

// GetSuggestedUsers 获取推荐用户（关注的人）
func GetSuggestedUsers(req component.BetterRequest[component.Null]) component.Response {
	list := userFollow.GetFollowingList(req.UserId, 1, 20)
	users := lo.Map(list, func(u *users.EntityComplete, _ int) component.DataMap {
		return component.DataMap{
			"id":       u.Id,
			"name":     u.Username,
			"username": u.Username,
			"avatar":   u.GetWebAvatarUrl(),
		}
	})
	return component.SuccessResponse(component.DataMap{
		"list": users,
	})
}
