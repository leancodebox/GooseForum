package eventnotice

import (
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
	"github.com/spf13/cast"
)

// Send4User 发送用户通知
func Send4User(userId uint64, msg string, eventType string) {
	entity := eventNotification.Entity{UserId: cast.ToString(userId), ReceivedNotification: msg, EventType: eventType}
	eventNotification.Create(&entity)
}

// Send4All 发送系统通知
func Send4All(userId uint64, msg string, eventType string) {
	entity := eventNotification.Entity{UserId: cast.ToString(0), ReceivedNotification: msg, EventType: eventType}
	eventNotification.Create(&entity)
}
