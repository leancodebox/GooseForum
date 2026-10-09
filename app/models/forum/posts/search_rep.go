package posts

import (
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
)

type SearchSource struct {
	Id                uint64
	TopicId           uint64
	PostNo            uint64
	Content           string
	UpdatedAt         time.Time
	ModerationVersion uint64
}

type SearchReference struct {
	Id                uint64
	TopicId           uint64
	PostNo            uint64
	UpdatedAt         time.Time
	ModerationVersion uint64
}

func GetSearchPostReferences(ids []uint64) ([]SearchReference, error) {
	var rows []SearchReference
	if len(ids) == 0 {
		return rows, nil
	}
	err := builder().Model(&Entity{}).Where("id IN ? AND process_status = 0", ids).Find(&rows).Error
	return rows, err
}

func GetSearchReplies(topicID uint64) ([]SearchSource, error) {
	var replies []SearchSource
	err := builder().Model(&Entity{}).
		Where("topic_id = ? AND post_no > 1 AND process_status = 0", topicID).
		Order(queryopt.Asc("post_no")).Find(&replies).Error
	return replies, err
}
