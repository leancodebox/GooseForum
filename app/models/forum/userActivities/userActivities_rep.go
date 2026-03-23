package userActivities

import (
	"time"
)

// ActionType 行为类型枚举
type ActionType int

const (
	ActionSignUp  ActionType = 1 // 注册帐号
	ActionPost    ActionType = 2 // 发帖/发布话题
	ActionLike    ActionType = 3 // 点赞
	ActionFollow  ActionType = 4 // 关注
	ActionComment ActionType = 5 // 回复/评论
)

// SubjectType 目标对象类型
const (
	SubjectTopic = "Topic"
	SubjectPost  = "Post"
	SubjectUser  = "User"
)

// Record 记录一条用户行为
func Record(userId uint64, action ActionType, subjectType string, subjectId uint64, preview string) error {
	entity := &Entity{
		UserId:         userId,
		Action:         int(action),
		SubjectType:    subjectType,
		SubjectId:      subjectId,
		ContentPreview: preview,
		CreatedAt:      time.Now(),
	}
	return builder().Create(entity).Error
}

// RecordWithTime 记录一条带有指定时间的用户行为（用于数据修复）
func RecordWithTime(userId uint64, action ActionType, subjectType string, subjectId uint64, preview string, t time.Time) error {
	entity := &Entity{
		UserId:         userId,
		Action:         int(action),
		SubjectType:    subjectType,
		SubjectId:      subjectId,
		ContentPreview: preview,
		CreatedAt:      t,
	}
	return builder().Create(entity).Error
}

// GetUserTimeline 获取用户的动态时间轴（基于主键的分页）
func GetUserTimeline(userId uint64, lastId uint64, limit int) (entities []*Entity, err error) {
	db := builder().Where("user_id = ?", userId)
	if lastId > 0 {
		db = db.Where("id < ?", lastId)
	}
	err = db.Order("id DESC").
		Limit(limit).
		Find(&entities).Error
	return
}
