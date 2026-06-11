package chatservice

import (
	"errors"
	"time"

	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/chat/imConversations"
	"github.com/leancodebox/GooseForum/app/models/chat/imUserChatConfigs"
	"github.com/leancodebox/GooseForum/app/models/chat/messages"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/unreadservice"
	"github.com/samber/lo"
)

const (
	defaultMessageLimit = 30
	maxMessageLimit     = 100
)

type MessageCursorResult struct {
	List          []*vo.MessageVo `json:"list"`
	HasMoreBefore bool            `json:"hasMoreBefore"`
	HasMoreAfter  bool            `json:"hasMoreAfter"`
	NextBeforeId  uint64          `json:"nextBeforeId"`
	LatestId      uint64          `json:"latestId"`
}

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
		MsgType:   msgType,
		IsRead:    0,
		CreatedAt: time.Now(),
	}
	messages.SaveOrCreateById(msg)
	imUserChatConfigs.InvalidateConversationAccess(senderId, convId)
	imUserChatConfigs.InvalidateConversationAccess(peerId, convId)
	unreadservice.Invalidate(peerId)

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

// GetMessages returns cursor-paginated messages for a conversation.
func GetMessages(userId, convId uint64, beforeId, afterId uint64, limit int) (*MessageCursorResult, error) {
	if !imUserChatConfigs.CanAccessConversation(userId, convId) {
		return nil, errors.New("conversation not found")
	}
	if beforeId > 0 && afterId > 0 {
		return nil, errors.New("beforeId and afterId cannot be used together")
	}
	if limit <= 0 {
		limit = defaultMessageLimit
	}
	if limit > maxMessageLimit {
		limit = maxMessageLimit
	}

	queryLimit := limit + 1
	var msgs []messages.Entity
	switch {
	case beforeId > 0:
		msgs = messages.GetBeforeId(convId, beforeId, queryLimit)
	case afterId > 0:
		msgs = messages.GetAfterId(convId, afterId, queryLimit)
	default:
		msgs = messages.GetLatestByConvId(convId, queryLimit)
	}

	hasMore := len(msgs) > limit
	if hasMore {
		msgs = msgs[:limit]
	}
	if afterId == 0 {
		reverseMessages(msgs)
	}

	list := lo.Map(msgs, func(m messages.Entity, _ int) *vo.MessageVo {
		return &vo.MessageVo{
			Id:        m.Id,
			SenderId:  m.SenderId,
			Content:   m.Content,
			MsgType:   m.MsgType,
			IsRead:    m.IsRead,
			CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
			IsSelf:    m.SenderId == userId,
		}
	})

	result := &MessageCursorResult{
		List: list,
	}
	if len(list) > 0 {
		result.NextBeforeId = list[0].Id
		result.LatestId = list[len(list)-1].Id
	}
	if afterId > 0 {
		result.HasMoreAfter = hasMore
	} else {
		result.HasMoreBefore = hasMore
	}
	return result, nil
}

func reverseMessages(msgs []messages.Entity) {
	for left, right := 0, len(msgs)-1; left < right; left, right = left+1, right-1 {
		msgs[left], msgs[right] = msgs[right], msgs[left]
	}
}

// MarkRead clears unread state for a conversation.
func MarkRead(userId, convId uint64) error {
	imUserChatConfigs.ClearUnread(convId, userId)
	messages.MarkMessagesRead(convId, userId)
	unreadservice.Invalidate(userId)
	return nil
}
