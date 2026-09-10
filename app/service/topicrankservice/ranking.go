package topicrankservice

import (
	"context"
	"errors"
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/topicrank"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"gorm.io/gorm"
)

// Recalculate reads one topic and writes its score without a transaction.
// Concurrent changes are eventually reflected by another event or a rebuild.
func Recalculate(db *gorm.DB, id uint64, now time.Time) error {

	topic, err := topics.GetForRankingWithDB(db, id)
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
	return topicrank.Save(db, id, score, next)
}

// ProcessDue is bounded both by row count and by the caller's context. Failed
// rows stay due for retry; an error does not starve the rest of the batch.
func ProcessDue(ctx context.Context, db *gorm.DB, now time.Time, limit int) error {
	if limit <= 0 {
		return nil
	}
	db = db.WithContext(ctx)
	ids, err := topicrank.DueIDs(db, now, limit)
	if err != nil {
		return err
	}
	var errs []error
	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			return errors.Join(append(errs, err)...)
		}
		if err := Recalculate(db, id, now); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Backfill is a one-time versioned migration, never a periodic table scan.
func Backfill(db *gorm.DB) error {
	upper, err := topicrank.MaxID(db)
	if err != nil {
		return err
	}
	var cursor uint64
	now := time.Now()
	for {
		ids, err := topicrank.IDsAfter(db, cursor, upper, 200)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		for _, id := range ids {
			row, err := topics.GetForRankingWithDB(db, id)
			if err != nil {
				return err
			}
			if row.Status == 1 {
				if err := topicrank.SetPublishedAt(db, id, row.CreatedAt); err != nil {
					return err
				}
			}
			if err := Recalculate(db, id, now); err != nil {
				return err
			}
			cursor = id
		}
	}
}

// Rebuild explicitly visits a bounded snapshot of all topic IDs, including
// dormant topics. It preserves first publication timestamps and supports reruns.
func Rebuild(ctx context.Context, db *gorm.DB, progress func(int64)) (int64, error) {
	db = db.WithContext(ctx).Session(&gorm.Session{SkipDefaultTransaction: true})
	upper, err := topicrank.MaxID(db)
	if err != nil {
		return 0, err
	}
	var cursor uint64
	var count int64
	now := time.Now()
	for {
		ids, err := topicrank.IDsAfter(db, cursor, upper, 200)
		if err != nil {
			return count, err
		}
		if len(ids) == 0 {
			return count, nil
		}
		for _, id := range ids {
			if err := Recalculate(db, id, now); err != nil {
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
