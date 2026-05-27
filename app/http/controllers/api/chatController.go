package api

import (
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/service/chatservice"
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
	return successDataMap("convId", convId)
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
	return successDataMap("list", msgs)
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
