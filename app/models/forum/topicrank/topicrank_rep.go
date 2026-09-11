package topicrank

import (
	"context"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// One atomic upsert coalesces triggers, retains queue age and advances version.
func MarkAt(ctx context.Context, topicID uint64, now time.Time) error {
	if topicID == 0 {
		return nil
	}
	result := builder().WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "topic_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"next_run_at": gorm.Expr("CASE WHEN next_run_at IS NULL OR next_run_at > ? THEN ? ELSE next_run_at END", now, now),
			"version":     gorm.Expr("version + 1"),
		}),
	}).Create(&Entity{TopicID: topicID, NextRunAt: &now, Version: 1})
	if result.Error == nil {
		select {
		case wakeups <- struct{}{}:
		default:
		}
	}
	return result.Error
}

func Get(ctx context.Context, topicID uint64) (Entity, error) {
	var entry Entity
	err := builder().WithContext(ctx).Where(queryopt.Eq("topic_id", topicID)).Take(&entry).Error
	return entry, err
}

// SetNext confirms only the generation that was processed. A new trigger wins.
// Even completion/retry advances version so stale acknowledgements cannot win.
func SetNext(ctx context.Context, entry ScheduledTopic, next *time.Time) error {
	return builder().WithContext(ctx).Where(queryopt.Eq("topic_id", entry.ID)).
		Where(queryopt.Eq("version", entry.Version)).UpdateColumns(map[string]any{
		"next_run_at": next, "version": gorm.Expr("version + 1"),
	}).Error
}

func Pending(ctx context.Context, limit int) ([]ScheduledTopic, error) {
	var entries []ScheduledTopic
	err := builder().WithContext(ctx).Select("topic_id", "next_run_at", "version").
		Where(queryopt.IsNotNull("next_run_at")).
		Order(queryopt.Asc("next_run_at")).Order(queryopt.Asc("topic_id")).
		Limit(limit).Find(&entries).Error
	return entries, err
}

func Defer(ctx context.Context, entry ScheduledTopic, until time.Time) error {
	return SetNext(ctx, entry, &until)
}

// Import preserves any schedule already written by a previous migration attempt
// or a running producer. It never resets an existing version or deadline.
func Import(ctx context.Context, entries []Entity) error {
	if len(entries) == 0 {
		return nil
	}
	return builder().WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&entries).Error
}
