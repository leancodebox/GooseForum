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

// SendMessage creates or updates a direct conversation and stores a message.
func SendMessage(senderId, peerId uint64, content string, msgType int8) (uint64, error) {
	if senderId == peerId {
		return 0, errors.New("cannot send message to yourself")
	}

	senderConfig := imUserChatConfigs.GetConfig(senderId, peerId)
	var convId uint64

	if senderConfig == nil {
		conv := &imConversations.Entity{
			Type:           1,
			LastMsgContent: content,
			LastMsgTime:    time.Now(),
		}
		imConversations.SaveOrCreateById(conv)
		convId = conv.Id

		imUserChatConfigs.SaveOrCreateById(&imUserChatConfigs.Entity{
			UserId:    senderId,
			PeerId:    peerId,
			ConvId:    convId,
			UpdatedAt: time.Now(),
		})

		imUserChatConfigs.SaveOrCreateById(&imUserChatConfigs.Entity{
			UserId:      peerId,
			PeerId:      senderId,
			ConvId:      convId,
			UnreadCount: 1,
			UpdatedAt:   time.Now(),
		})
	} else {
		convId = senderConfig.ConvId
		imConversations.UpdateLastMsg(convId, content)
		imUserChatConfigs.Touch(convId, senderId)

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

// GetChatList returns the current user's conversations.
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

// GetMessages returns paginated messages for a conversation.
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

// MarkRead clears unread state for a conversation.
func MarkRead(userId, convId uint64) error {
	imUserChatConfigs.ClearUnread(convId, userId)
	messages.MarkMessagesRead(convId, userId)
	return nil
}

// DeleteChat hides a conversation for the user.
func DeleteChat(userId, convId uint64) error {
	imUserChatConfigs.DeleteConfig(convId, userId)
	return nil
}
