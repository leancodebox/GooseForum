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
	topicID := uint64(0)
	if subjectType == SubjectTopic {
		topicID = subjectId
	}
	return RecordForTopic(userId, action, subjectType, subjectId, topicID, preview)
}

func RecordForTopic(userId uint64, action ActionType, subjectType string, subjectId uint64, topicID uint64, preview string) error {
	entity := &Entity{
		UserId:         userId,
		TopicId:        topicID,
		Action:         int(action),
		SubjectType:    subjectType,
		SubjectId:      subjectId,
		ContentPreview: preview,
		CreatedAt:      time.Now(),
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

func GetUserTimelineForAudience(userID uint64, lastID uint64, limit int, readableCategoryIDs []uint64, filterAudience bool) (entities []*Entity, err error) {
	if userID == 0 || limit <= 0 {
		return []*Entity{}, nil
	}
	query := builder().
		Select("user_activities.*").
		Joins("LEFT JOIN topics ON topics.id = user_activities.topic_id").
		Where("user_activities.user_id = ?", userID)
	if lastID > 0 {
		query = query.Where("user_activities.id < ?", lastID)
	}
	topicVisible := `(
		topics.id IS NOT NULL
		AND topics.status = 1
		AND topics.process_status = 0
	)`
	if filterAudience {
		if len(readableCategoryIDs) == 0 {
			query = query.Where("user_activities.topic_id = 0")
		} else {
			topicVisible = `(
				topics.id IS NOT NULL
				AND topics.status = 1
				AND topics.process_status = 0
				AND topics.main_category_id IN ?
			)`
			query = query.Where("user_activities.topic_id = 0 OR "+topicVisible, readableCategoryIDs)
		}
	} else {
		query = query.Where("user_activities.topic_id = 0 OR " + topicVisible)
	}
	err = query.Order("user_activities.id DESC").Limit(limit).Find(&entities).Error
	return entities, err
}
