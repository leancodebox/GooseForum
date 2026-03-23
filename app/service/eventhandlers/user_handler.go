package eventhandlers

import (
	"context"
	"time"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/leancodebox/GooseForum/app/service/userservice"
)

// UserLastActiveUpdatedEvent 用户最后活跃时间更新事件
type UserLastActiveUpdatedEvent struct {
	UserId     uint64
	ActiveTime time.Time
}

// UserLastActiveUpdatedHandler 用户最后活跃时间更新处理器
func NewUserLastActiveUpdatedHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler(
		"UserLastActiveUpdatedHandler",
		func(ctx context.Context, event *UserLastActiveUpdatedEvent) error {
			userservice.UpdateUserActivity(event.UserId)
			return nil
		},
	)
}

// UserSignUpEvent 用户注册事件
type UserSignUpEvent struct {
	UserId   uint64
	Username string
}
