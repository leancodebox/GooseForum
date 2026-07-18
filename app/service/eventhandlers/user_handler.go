package eventhandlers

import (
	"context"
	"log/slog"
	"time"

	"github.com/leancodebox/GooseForum/app/service/topicunseenservice"
	"github.com/leancodebox/GooseForum/app/service/userservice"
)

// UserLastActiveUpdatedEvent 用户最后活跃时间更新事件
type UserLastActiveUpdatedEvent struct {
	UserId     uint64
	ActiveTime time.Time
}

// handleUserLastActiveUpdated 更新用户最后活跃时间
func handleUserLastActiveUpdated(ctx context.Context, event *UserLastActiveUpdatedEvent) error {
	userservice.UpdateUserActivityAt(event.UserId, event.ActiveTime)
	if err := topicunseenservice.TouchUser(event.UserId, event.ActiveTime); err != nil {
		slog.Warn("touch topic unseen activity failed", "userId", event.UserId, "error", err)
	}
	return nil
}

// UserSignUpEvent 用户注册事件
type UserSignUpEvent struct {
	UserId   uint64
	Username string
}
