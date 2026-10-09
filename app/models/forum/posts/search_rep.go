package posts

import (
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
)

type SearchReply struct {
	Content   string
	UpdatedAt time.Time
}

func GetSearchReplies(topicID uint64) ([]SearchReply, error) {
	var replies []SearchReply
	err := builder().Model(&Entity{}).
		Where("topic_id = ? AND post_no > 1 AND process_status = 0", topicID).
		Order(queryopt.Asc("post_no")).Find(&replies).Error
	return replies, err
}
