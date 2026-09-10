// Package topicrank marks topics for asynchronous score updates.
package topicrank

import (
	"context"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/eventbus"

	"gorm.io/gorm"
)

func Mark(tx *gorm.DB, topicID uint64) error {
	return MarkAt(tx, topicID, time.Now())
}

func MarkAt(tx *gorm.DB, topicID uint64, now time.Time) error {
	if topicID == 0 {
		return nil
	}
	return query(tx).Where("id = ?", topicID).UpdateColumns(map[string]any{
		"next_rank_at": now,
		"published_at": gorm.Expr("CASE WHEN published_at IS NULL AND status = 1 THEN ? ELSE published_at END", now),
	}).Error
}

// TopicRankRequested carries values only, never a request context or transaction handle.
type TopicRankRequested struct {
	TopicID    uint64
	OccurredAt time.Time
}

// Notify uses the existing asynchronous event bus; it performs no database I/O.
func Notify(topicID uint64) {
	if topicID == 0 {
		return
	}
	eventbus.Publish(context.Background(), &TopicRankRequested{TopicID: topicID, OccurredAt: time.Now()})
}
