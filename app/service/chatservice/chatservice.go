package chatservice

import (
	"errors"
	"time"

	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/chat/imConversations"
	"github.com/leancodebox/GooseForum/app/models/chat/imUserChatConfigs"
	"github.com/leancodebox/GooseForum/app/models/chat/messages"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/samber/lo"
)

// SendMessage 发送私信
func SendMessage(senderId, peerId uint64, content string, msgType int8) (uint64, error) {
	if senderId == peerId {
		return 0, errors.New("cannot send message to yourself")
	}

	// 1. Check if conversation exists for sender
	senderConfig := imUserChatConfigs.GetConfig(senderId, peerId)
	var convId uint64

	if senderConfig == nil {
		// Create new conversation
		conv := &imConversations.Entity{
			Type:           1, // C2C
			LastMsgContent: content,
			LastMsgTime:    time.Now(),
		}
		imConversations.SaveOrCreateById(conv)
		convId = conv.Id

		// Create config for sender
		imUserChatConfigs.SaveOrCreateById(&imUserChatConfigs.Entity{
			UserId:    senderId,
			PeerId:    peerId,
			ConvId:    convId,
			UpdatedAt: time.Now(),
		})

		// Create config for receiver
		imUserChatConfigs.SaveOrCreateById(&imUserChatConfigs.Entity{
			UserId:      peerId,
			PeerId:      senderId,
			ConvId:      convId,
			UnreadCount: 1,
			UpdatedAt:   time.Now(),
		})
	} else {
		convId = senderConfig.ConvId
		// Update conversation last msg
		imConversations.UpdateLastMsg(convId, content)

		// Update sender config (touch)
		imUserChatConfigs.Touch(convId, senderId)

		// Update receiver config (incr unread)
		// Check if receiver config exists (it should, but safety first)
		peerConfig := imUserChatConfigs.GetConfig(peerId, senderId)
		if peerConfig == nil {
			imUserChatConfigs.SaveOrCreateById(&imUserChatConfigs.Entity{
				UserId:      peerId,
				PeerId:      senderId,
				ConvId:      convId,
				UnreadCount: 1,
				UpdatedAt:   time.Now(),
			})
		} else {
			imUserChatConfigs.IncrUnread(convId, peerId)
		}
	}

	// Save message
	msg := &messages.Entity{
		ConvId:    convId,
		SenderId:  senderId,
		Content:   content,
		MsgType:   int8(msgType),
		IsRead:    0,
		CreatedAt: time.Now(),
	}
	messages.SaveOrCreateById(msg)

	return convId, nil
}

// GetChatList 获取私信列表
func GetChatList(userId uint64) ([]*vo.ChatItemVo, error) {
	configs := imUserChatConfigs.GetUserConfigs(userId)
	if len(configs) == 0 {
		return []*vo.ChatItemVo{}, nil
	}

	peerIds := lo.Map(configs, func(cfg imUserChatConfigs.Entity, _ int) uint64 {
		return cfg.PeerId
	})
	convIds := lo.Map(configs, func(cfg imUserChatConfigs.Entity, _ int) uint64 {
		return cfg.ConvId
	})

	peers := users.GetMapByIds(peerIds)
	convs := imConversations.GetMapByIds(convIds)

	list := lo.Map(configs, func(cfg imUserChatConfigs.Entity, _ int) *vo.ChatItemVo {
		peer := peers[cfg.PeerId]
		conv := convs[cfg.ConvId]

		chatItem := &vo.ChatItemVo{
			Id:          cfg.Id,
			PeerId:      cfg.PeerId,
			UnreadCount: cfg.UnreadCount,
			ConvId:      cfg.ConvId,
		}

		if peer != nil {
			chatItem.PeerUsername = peer.Username
			chatItem.PeerAvatar = peer.GetWebAvatarUrl()
		} else {
			chatItem.PeerUsername = "Unknown User"
		}

		if conv != nil {
			chatItem.LastMsg = conv.LastMsgContent
			chatItem.LastMsgTime = conv.LastMsgTime.Format("2006-01-02 15:04:05")
		}

		return chatItem
	})

	return list, nil
}

// GetMessages 获取消息记录
func GetMessages(userId, convId uint64, page, pageSize int) ([]*vo.MessageVo, error) {
	offset := (page - 1) * pageSize
	msgs := messages.GetByConvId(convId, offset, pageSize)

	return lo.Map(msgs, func(m messages.Entity, _ int) *vo.MessageVo {
		return &vo.MessageVo{
			Id:        m.Id,
			SenderId:  m.SenderId,
			Content:   m.Content,
			MsgType:   int8(m.MsgType),
			IsRead:    m.IsRead,
			CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
			IsSelf:    m.SenderId == userId,
		}
	}), nil
}

// MarkRead 标记已读
func MarkRead(userId, convId uint64) error {
	imUserChatConfigs.ClearUnread(convId, userId)
	messages.MarkMessagesRead(convId, userId)
	return nil
}

// DeleteChat 删除对话（逻辑删除）
func DeleteChat(userId, convId uint64) error {
	imUserChatConfigs.DeleteConfig(convId, userId)
	return nil
}
