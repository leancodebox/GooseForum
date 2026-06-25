package eventhandlers

import (
	"context"
	"time"

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
	return nil
}

// UserSignUpEvent 用户注册事件
type UserSignUpEvent struct {
	UserId   uint64
	Username string
}
