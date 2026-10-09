package topics

import (
	"time"

	"gorm.io/gorm"
)

type AgentTopic struct {
	Id                                  uint64
	Title                               string
	UserId, MainCategoryId, FirstPostId uint64
	CreatedAt, UpdatedAt                time.Time
}

type AgentTopicQuery struct {
	ReadableCategoryIDs  []uint64
	GlobalManage         bool
	CategoryID, BeforeID uint64
}

func visibleAgentTopics(filter AgentTopicQuery) *gorm.DB {
	q := builder().Model(&Entity{}).Where("status = 1 AND process_status = 0").Where("EXISTS (SELECT 1 FROM posts WHERE posts.id = topics.first_post_id AND posts.deleted_at IS NULL AND posts.process_status = 0)")
	if !filter.GlobalManage {
		q = q.Where("main_category_id IN ?", filter.ReadableCategoryIDs)
	}
	return q
}

func GetAgentTopic(filter AgentTopicQuery, id uint64) (AgentTopic, bool, error) {
	var row AgentTopic
	result := visibleAgentTopics(filter).Where("id = ?", id).Limit(1).Find(&row)
	return row, result.RowsAffected > 0, result.Error
}

func ListAgentTopics(filter AgentTopicQuery, limit int) ([]AgentTopic, error) {
	q := visibleAgentTopics(filter)
	if filter.CategoryID > 0 {
		q = q.Where("main_category_id = ?", filter.CategoryID)
	}
	if filter.BeforeID > 0 {
		q = q.Where("id < ?", filter.BeforeID)
	}
	var rows []AgentTopic
	err := q.Order("id DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func GetAgentTopicsByIDs(filter AgentTopicQuery, ids []uint64) ([]AgentTopic, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var rows []AgentTopic
	err := visibleAgentTopics(filter).Where("id IN ?", ids).Find(&rows).Error
	return rows, err
}

func GetSubmissionTopic(id uint64) (Entity, error) {
	var row Entity
	err := builder().Unscoped().Where("id = ?", id).Take(&row).Error
	return row, err
}
