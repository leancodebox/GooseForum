package topicrankservice

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/leancodebox/GooseForum/app/models/forum/topicrank"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
)

// Recalculate is the explicit rebuild entry point. Normal worker processing
// already has the task version and goes straight to recalculateScheduled.
func Recalculate(ctx context.Context, id uint64, now time.Time) error {
	entry, err := topicrank.Get(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := topicrank.MarkAt(ctx, id, now); err != nil {
			return err
		}
		entry, err = topicrank.Get(ctx, id)
	}
	if err != nil {
		return err
	}
	return recalculateScheduled(ctx, topicrank.ScheduledTopic{ID: id, Version: entry.Version}, now)
}

func recalculateScheduled(ctx context.Context, entry topicrank.ScheduledTopic, now time.Time) error {
	topic, err := topics.GetForRanking(ctx, entry.ID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return topicrank.SetNext(ctx, entry, nil)
	}
	if err != nil {
		return err
	}
	var score int64
	var next *time.Time
	if topic.Status == 1 && topic.ProcessStatus == 0 && !topic.DeletedAt.Valid {
		published := topic.CreatedAt
		if topic.PublishedAt != nil {
			published = *topic.PublishedAt
		}
		score, next = Score(Signals{Likes: topic.LikeCount, Replies: topic.ReplyCount, LastPostedAt: topic.LastPostedAt}, published, now)
	}
	if err := topics.SaveRankScore(ctx, entry.ID, score); err != nil {
		return err
	}
	return topicrank.SetNext(ctx, entry, next)
}

// Backfill is a one-time versioned migration, never a periodic table scan.
func Backfill(ctx context.Context) error {
	upper, err := topics.MaxRankingID(ctx)
	if err != nil {
		return err
	}
	var cursor uint64
	now := time.Now()
	for {
		ids, err := topics.RankingIDsAfter(ctx, cursor, upper, 200)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		for _, id := range ids {
			row, err := topics.GetForRanking(ctx, id)
			if err != nil {
				return err
			}
			if row.Status == 1 {
				if err := topics.SetPublishedAt(ctx, id, row.CreatedAt); err != nil {
					return err
				}
			}
			if err := Recalculate(ctx, id, now); err != nil {
				return err
			}
			cursor = id
		}
	}
}

// Rebuild explicitly visits a bounded snapshot of all topic IDs, including
// dormant topics. It preserves first publication timestamps and supports reruns.
func Rebuild(ctx context.Context, progress func(int64)) (int64, error) {
	upper, err := topics.MaxRankingID(ctx)
	if err != nil {
		return 0, err
	}
	var cursor uint64
	var count int64
	now := time.Now()
	for {
		ids, err := topics.RankingIDsAfter(ctx, cursor, upper, 200)
		if err != nil {
			return count, err
		}
		if len(ids) == 0 {
			return count, nil
		}
		for _, id := range ids {
			if err := Recalculate(ctx, id, now); err != nil {
				return count, err
			}
			cursor = id
			count++
		}
		if progress != nil {
			progress(count)
		}
	}
}
